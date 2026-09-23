package general

import (
	"context"
	"testing"

	"github.com/daeuniverse/dae-wing/db"
)

func TestModifiedWithMissingRunningIDs(t *testing.T) {
	if err := db.InitDatabase(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	d := db.DB(ctx)
	config := db.Config{Selected: true}
	dns := db.Dns{Selected: true}
	routing := db.Routing{Selected: true}
	for _, model := range []interface{}{&config, &dns, &routing} {
		if err := d.Create(model).Error; err != nil {
			t.Fatal(err)
		}
	}
	sys := db.System{Running: true, RunningConfigID: &config.ID, RunningDnsID: &dns.ID, RunningRoutingID: &routing.ID}
	if err := d.Create(&sys).Error; err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"running_config_id", "running_dns_id", "running_routing_id"} {
		if err := d.Model(&sys).Updates(map[string]interface{}{
			"running_config_id": config.ID, "running_dns_id": dns.ID, "running_routing_id": routing.ID,
		}).Error; err != nil {
			t.Fatal(err)
		}
		if err := d.Model(&sys).Update(field, nil).Error; err != nil {
			t.Fatal(err)
		}
		modified, err := (&DaeResolver{Ctx: ctx}).Modified()
		if err != nil || !modified {
			t.Fatalf("%s: modified=%v err=%v", field, modified, err)
		}
	}
}
