package config

import (
	"testing"

	"github.com/daeuniverse/dae-wing/dae"
	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/sirupsen/logrus"
)

func TestAPIOnlyRejectsRealReload(t *testing.T) {
	oldReload, oldExit := dae.ChReloadConfigs, dae.GracefullyExit
	dae.ChReloadConfigs = make(chan *dae.ReloadMessage)
	dae.GracefullyExit = make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- dae.Run(logrus.New(), dae.EmptyConfig, nil, true, true) }()
	t.Cleanup(func() {
		dae.ChReloadConfigs <- nil
		if err := <-done; err != nil {
			t.Error(err)
		}
		dae.ChReloadConfigs, dae.GracefullyExit = oldReload, oldExit
	})
	callback := make(chan error, 1)
	dae.ChReloadConfigs <- &dae.ReloadMessage{Config: &daeConfig.Config{}, Callback: callback}
	if err := <-callback; err == nil {
		t.Fatal("api-only mode accepted a real run")
	}
	dae.ChReloadConfigs <- &dae.ReloadMessage{Config: dae.EmptyConfig, Callback: callback}
	if err := <-callback; err != nil {
		t.Fatalf("dry reload: %v", err)
	}
}
