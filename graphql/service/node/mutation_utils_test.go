package node

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/internal"
	"github.com/graph-gophers/graphql-go"
)

func TestStandaloneNodeMutations(t *testing.T) {
	if err := db.InitDatabase(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	d := db.DB(ctx)
	tag := "standalone"
	link := "socks5://127.0.0.1:1080#old"
	standalone, err := importNode(d, nil, &internal.ImportArgument{Link: link, Tag: &tag})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = importNode(d, nil, &internal.ImportArgument{Link: link}); !errors.Is(err, DuplicatedError) {
		t.Fatalf("duplicate import: %v", err)
	}
	sub := db.Subscription{Link: "https://example.com/sub"}
	if err = d.Create(&sub).Error; err != nil {
		t.Fatal(err)
	}
	owned, err := importNode(d, &sub.ID, &internal.ImportArgument{Link: link})
	if err != nil {
		t.Fatal(err)
	}
	newLink := "socks5://127.0.0.1:1081#new"
	updated, err := Update(d, common.EncodeCursor(standalone.ID), newLink)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Node.ID != standalone.ID || updated.Node.Tag == nil || *updated.Node.Tag != tag || updated.Node.Link != newLink {
		t.Fatalf("wrong persisted row: %+v", updated.Node)
	}
	for _, id := range []uint{owned.ID, owned.ID + 100} {
		cursor := common.EncodeCursor(id)
		if _, err = Update(d, cursor, newLink); err == nil || !strings.Contains(err.Error(), string(cursor)) {
			t.Fatalf("update %s: %v", cursor, err)
		}
		if _, err = Remove(ctx, []graphql.ID{common.EncodeCursor(standalone.ID), cursor}); err == nil || !strings.Contains(err.Error(), string(cursor)) {
			t.Fatalf("remove %s: %v", cursor, err)
		}
		var count int64
		if err = d.Model(&db.Node{}).Count(&count).Error; err != nil || count != 2 {
			t.Fatalf("partial deletion: count=%d err=%v", count, err)
		}
	}
	if n, err := Remove(ctx, []graphql.ID{common.EncodeCursor(standalone.ID)}); err != nil || n != 1 {
		t.Fatalf("remove standalone: count=%d err=%v", n, err)
	}
}
