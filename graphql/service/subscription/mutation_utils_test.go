package subscription

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/graph-gophers/graphql-go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

func TestFetchLinksBodyLimitAndClose(t *testing.T) {
	const limit = 16 << 20
	for _, tc := range []struct {
		name         string
		status, size int
		wantError    bool
	}{
		{"status", http.StatusForbidden, 0, true},
		{"exact limit", http.StatusOK, limit, false},
		{"over limit", http.StatusOK, limit + 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := ""
			if tc.size > 0 {
				payload = "socks5://127.0.0.1:1080#" + strings.Repeat("x", tc.size-len("socks5://127.0.0.1:1080#"))
			}
			body := &trackedBody{Reader: strings.NewReader(payload)}
			transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})
			links, err := _fetchLinks("https://example.com/sub", transport, time.Second)
			if (err != nil) != tc.wantError || !body.closed {
				t.Fatalf("err=%v closed=%v", err, body.closed)
			}
			if !tc.wantError && (len(links) != 1 || links[0] != payload) {
				t.Fatal("exact-limit link was not preserved")
			}
		})
	}
}

func TestSubscriptionRefreshAndRemoval(t *testing.T) {
	if err := db.InitDatabase(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	d := db.DB(ctx)
	link := "socks5://127.0.0.1:1080#preserved"
	sub := db.Subscription{Link: "https://example.com/sub"}
	if err := d.Create(&sub).Error; err != nil {
		t.Fatal(err)
	}
	n := db.Node{Link: link, SubscriptionID: &sub.ID}
	if err := d.Create(&n).Error; err != nil {
		t.Fatal(err)
	}
	g := db.Group{Name: "bound", Node: []db.Node{n}}
	if err := d.Create(&g).Error; err != nil {
		t.Fatal(err)
	}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(link)), Header: make(http.Header)}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = oldTransport })
	if _, err := UpdateById(ctx, sub.ID); err != nil {
		t.Fatal(err)
	}
	var preserved db.Node
	if err := d.First(&preserved, n.ID).Error; err != nil || preserved.Link != link {
		t.Fatalf("preserved node changed: %+v err=%v", preserved, err)
	}
	if _, err := UpdateCron(ctx, common.EncodeCursor(sub.ID), "", true); err == nil {
		t.Fatal("empty enabled cron was accepted")
	}
	if _, err := UpdateCron(ctx, common.EncodeCursor(sub.ID), "0 0 1 1 *", true); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { RemoveUpdateScheduler(sub.ID) })
	schedulerMu.RLock()
	scheduled := schedulerCache[sub.ID] != nil
	schedulerMu.RUnlock()
	if !scheduled {
		t.Fatal("enabled cron was not scheduled after commit")
	}
	if _, err := Remove(ctx, []graphql.ID{common.EncodeCursor(sub.ID)}); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := d.Table("group_nodes").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("orphan bindings: count=%d err=%v", count, err)
	}
}
