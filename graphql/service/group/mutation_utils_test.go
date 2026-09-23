package group

import (
	"context"
	"testing"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/graph-gophers/graphql-go"
)

func TestGroupMutationErrors(t *testing.T) {
	if err := db.InitDatabase(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	d := db.DB(ctx)
	groups := []db.Group{{Name: "one"}, {Name: "two"}}
	if err := d.Create(&groups).Error; err != nil {
		t.Fatal(err)
	}
	id := common.EncodeCursor(groups[0].ID)
	if _, err := Rename(ctx, id, "two"); err == nil {
		t.Fatal("duplicate name was accepted")
	}
	var persisted db.Group
	if err := d.First(&persisted, groups[0].ID).Error; err != nil || persisted.Name != "one" {
		t.Fatalf("failed rename changed group: %+v err=%v", persisted, err)
	}
	sqlDB, err := d.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := d.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatal(err)
	}
	if err := d.Exec("PRAGMA defer_foreign_keys = ON").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := AddSubscriptions(ctx, id, []graphql.ID{common.EncodeCursor(999)}, nil); err == nil {
		t.Fatal("deferred foreign-key commit failure was not returned")
	}
}
