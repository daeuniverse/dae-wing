/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dae

import (
	"context"
	"fmt"
	"net/netip"
	"runtime"
	"sync"
	"sync/atomic"

	daeCommon "github.com/daeuniverse/dae/common"
	"github.com/daeuniverse/dae/common/netutils"
	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/daeuniverse/dae/control"
	"github.com/daeuniverse/dae/pkg/config_parser"
	"github.com/daeuniverse/dae/pkg/logger"
	"github.com/daeuniverse/outbound/protocol/direct"
	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
)

var ErrControlPlaneNotInit = fmt.Errorf("control plane doesn't init yet")

type ReloadMessage struct {
	Config   *daeConfig.Config
	Callback chan<- error
}

var ChReloadConfigs = make(chan *ReloadMessage)
var GracefullyExit = make(chan struct{})
var EmptyConfig *daeConfig.Config

// active is the generation currently serving traffic. It is nil before the
// first generation is ready and while a reload has retired the old one.
var active atomic.Pointer[control.ControlPlane]
var onceWaitingNetwork sync.Once

func init() {
	sections, err := config_parser.Parse(`global{} routing{}`)
	if err != nil {
		panic(err)
	}
	EmptyConfig, err = daeConfig.New(sections)
	if err != nil {
		panic(err)
	}
}

func ControlPlane() (*control.ControlPlane, error) {
	c := active.Load()
	if c == nil {
		return nil, ErrControlPlaneNotInit
	}
	return c, nil
}

func Run(log *logrus.Logger, conf *daeConfig.Config, externGeoDataDirs []string, disableTimestamp bool, dry bool) (err error) {
	defer close(GracefullyExit)
	// Not really run dae.
	if dry {
		log.Infoln("Dry run in api-only mode")
	dryLoop:
		for newConf := range ChReloadConfigs {
			switch newConf {
			case nil:
				break dryLoop
			default:
				newConf.Callback <- nil
			}
		}
		return nil
	}

	// New c.
	c, err := newControlPlane(log, nil, conf, externGeoDataDirs)
	if err != nil {
		return err
	}
	listener, err := serve(log, c, conf.Global.TproxyPort)
	if err != nil {
		c.Close()
		return err
	}
	active.Store(c)
	log.Infoln("Ready")

	// A nil message is the exit request: cmd sends it on shutdown and serve
	// sends it when the datapath dies underneath us.
loop:
	for newReloadMsg := range ChReloadConfigs {
		if newReloadMsg == nil {
			break loop
		}
		log.Warnln("[Reload] Received reload signal; prepare to reload")
		newConf := newReloadMsg.Config

		// New logger.
		oldLogOutput := log.Out
		log = logrus.New()
		logger.SetLogger(log, newConf.Global.LogLevel, disableTimestamp, nil)
		logger.SetLogger(logrus.StandardLogger(), newConf.Global.LogLevel, disableTimestamp, nil)
		log.SetOutput(oldLogOutput) // FIXME: THIS IS A HACK.
		logrus.SetOutput(oldLogOutput)

		// dae >= 2.1.1 replaces TC programs in place through a staged handoff
		// that borrows the previous generation's hook set; re-attaching the
		// shared programs from a second generation fails with EEXIST on TCX.
		// dae-wing does not carry that state machine, so a reload is a cold
		// restart of the control plane: the datapath is detached while the
		// new generation loads (about a second), and pinned maps keep the
		// connection state across the gap exactly as a dae process restart
		// does.
		var dnsCache map[string]*control.DnsCache
		if conf.Dns.IpVersionPrefer == newConf.Dns.IpVersionPrefer {
			// Only keep dns cache when ip version preference not change.
			dnsCache = c.CloneDnsCache()
		}
		log.Warnln("[Reload] Stop old control plane")
		active.Store(nil)
		stopControlPlane(log, c, listener)
		listener = nil

		log.Warnln("[Reload] Load new control plane")
		var errReload error
		newC, err := newControlPlane(log, dnsCache, newConf, externGeoDataDirs)
		if err == nil {
			// Health snapshots survive Close; carrying them over keeps the
			// groups routable instead of falling back until the first check.
			newC.InheritDialerHealthFrom(c)
			listener, err = serve(log, newC, newConf.Global.TproxyPort)
			if err != nil {
				stopControlPlane(log, newC, nil)
			}
		}
		if err != nil {
			errReload = err
			log.WithFields(logrus.Fields{
				"err": err,
			}).Errorln("[Reload] Failed to reload; try to roll back configuration")
			// Load last config back.
			newC, err = newControlPlane(log, dnsCache, conf, externGeoDataDirs)
			if err == nil {
				newC.InheritDialerHealthFrom(c)
				listener, err = serve(log, newC, conf.Global.TproxyPort)
			}
			if err != nil {
				log.WithFields(logrus.Fields{
					"err": err,
				}).Fatalln("[Reload] Failed to roll back configuration")
			}
			newConf = conf
			log.Errorln("[Reload] Last reload failed; rolled back configuration")
		}
		c = newC
		conf = newConf
		active.Store(c)
		log.Warnln("[Reload] Finished")
		newReloadMsg.Callback <- errReload
	}

	active.Store(nil)
	if listener != nil {
		if e := listener.Close(); e != nil {
			log.Warnf("close listener: %v", e)
		}
	}
	if e := c.DetachBpfHooks(); e != nil {
		log.Warnf("detach BPF hooks: %v", e)
	}
	if e := control.GetDaeNetns().Close(); e != nil {
		log.Warnf("close dae netns: %v", e)
	}
	if e := c.AbortConnections(); e != nil {
		log.Warnf("abort connections: %v", e)
	}
	closeErr := c.Close()
	control.ResetGlobalUdpState()
	if closeErr != nil {
		return fmt.Errorf("close control plane: %w", closeErr)
	}
	return nil
}

// serve listens in the dae netns and serves c in the background. It returns
// once c reports ready, or with the listen/serve error. A generation that
// dies after it was ready asks the run loop to exit; one that never became
// ready is the caller's error to handle, so a failed reload candidate does
// not take the process down with it.
func serve(log *logrus.Logger, c *control.ControlPlane, port uint16) (listener *control.Listener, err error) {
	err = control.GetDaeNetns().With(func() error {
		var listenErr error
		listener, listenErr = c.Listen(port)
		return listenErr
	})
	if err != nil {
		return nil, fmt.Errorf("listen in dae netns: %w", err)
	}
	readyChan := make(chan bool, 1)
	served := make(chan error, 1)
	go func() {
		served <- c.Serve(readyChan, listener)
	}()
	if ready := <-readyChan; !ready {
		err = <-served
		_ = listener.Close()
		if err == nil {
			err = fmt.Errorf("control plane did not become ready")
		}
		return nil, fmt.Errorf("serve: %w", err)
	}
	go func() {
		if err := <-served; err != nil && active.Load() == c {
			log.Errorln("Serve:", err)
			// The datapath died underneath us; ask the run loop to exit.
			ChReloadConfigs <- nil
		}
	}()
	return listener, nil
}

// stopControlPlane retires one generation in the order dae's own shutdown
// uses: stop accepting, detach the datapath, then release the control plane.
// The dae netns is process-wide and is left in place for the next generation.
func stopControlPlane(log *logrus.Logger, c *control.ControlPlane, listener *control.Listener) {
	if listener != nil {
		if e := listener.Close(); e != nil {
			log.Warnf("close listener: %v", e)
		}
	}
	if e := c.DetachBpfHooks(); e != nil {
		log.Warnf("detach BPF hooks: %v", e)
	}
	if e := c.Close(); e != nil {
		log.Warnf("close control plane: %v", e)
	}
}

func newControlPlane(log *logrus.Logger, dnsCache map[string]*control.DnsCache, conf *daeConfig.Config, externGeoDataDirs []string) (c *control.ControlPlane, err error) {

	// Print configuration.
	if log.IsLevelEnabled(logrus.DebugLevel) {
		bConf, _ := conf.Marshal(2)
		log.Debugln(string(bConf))
	}

	// Deep copy to prevent modification.
	conf = deepcopy.Copy(conf).(*daeConfig.Config)

	// Mirror dae cmd: resolve the socket mark before the control plane sees the
	// config so dae's own egress is never captured by its own datapath.
	if conf.Global.SoMarkFromDae == 0 {
		var autoSelected bool
		conf.Global.SoMarkFromDae, autoSelected = daeCommon.ResolveSoMarkFromDae(conf.Global.SoMarkFromDae, conf.Global.SoMarkFromDaeSet)
		if autoSelected {
			log.Warnf("so_mark_from_dae is unset; using internal socket mark %#x to prevent dae UDP self-capture", conf.Global.SoMarkFromDae)
		}
	}

	// Purge classic TC filters left by an older process. Our own previous
	// generation is already detached by the time a reload gets here.
	control.PurgeStaleTCFilters(log)

	// Generation-scoped direct dialers and system resolver.
	directDialers := direct.NewDirectDialers(conf.Global.FallbackResolver)
	systemDNSResolver := netutils.NewSystemDNSResolver(netip.MustParseAddrPort(conf.Global.FallbackResolver))

	if !conf.Global.DisableWaitingNetwork && len(conf.Global.WanInterface) > 0 {
		// Wait for network for WAN ready.
		onceWaitingNetwork.Do(func() {
			WaitForNetwork(log, directDialers.Symmetric)
		})
	}

	/// Get subscription -> nodeList mapping.
	subscriptionToNodeList := map[string][]string{}
	if len(conf.Node) > 0 {
		for _, node := range conf.Node {
			subscriptionToNodeList[""] = append(subscriptionToNodeList[""], string(node))
		}
	}
	if len(conf.Subscription) > 0 {
		return nil, fmt.Errorf("daeConfig.subscription is not supported")
	}

	if err = preprocessWanInterfaceAuto(conf); err != nil {
		return nil, err
	}

	// New dae control plane.
	// Every generation loads its own BPF objects; pinned maps are reused.
	c, err = control.NewControlPlaneWithContextOptions(
		context.Background(),
		log,
		nil,
		dnsCache,
		subscriptionToNodeList,
		conf.Group,
		&conf.Routing,
		&conf.Global,
		&conf.Dns,
		externGeoDataDirs,
		control.ControlPlaneBuildOptions{
			DirectDialer:         directDialers.Symmetric,
			FullconeDirectDialer: directDialers.Fullcone,
			SystemDNSResolver:    systemDNSResolver,
		},
	)
	if err != nil {
		return nil, err
	}
	// Call GC to release memory.
	runtime.GC()

	return c, nil
}
