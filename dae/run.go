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
	"golang.org/x/sys/unix"
)

var ErrControlPlaneNotInit = fmt.Errorf("control plane doesn't init yet")

type ReloadMessage struct {
	Config   *daeConfig.Config
	Callback chan<- error
	// failed is set by the serve monitor when a generation's datapath died.
	// The run loop acts on it only while that generation is still the active
	// one; a report from a generation that has since been replaced is stale.
	failed *generation
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

// generation is one control plane with its listener. retired is set before
// the plane is torn down so a Serve error from a plane that is being replaced
// is not mistaken for the active datapath dying.
type generation struct {
	c        *control.ControlPlane
	listener *control.Listener
	retired  atomic.Bool
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

	gen, err := startGeneration(log, nil, nil, conf, externGeoDataDirs)
	if err != nil {
		shutdownProcess(log, nil)
		return err
	}
	active.Store(gen.c)
	log.Infoln("Ready")

	// A nil message is the exit request: cmd sends it on shutdown and the
	// serve monitor sends it when the active datapath dies underneath us.
loop:
	for newReloadMsg := range ChReloadConfigs {
		if newReloadMsg == nil {
			break loop
		}
		if newReloadMsg.failed != nil {
			if newReloadMsg.failed == gen {
				log.Errorln("Serve exited with an error; shutting down")
				break loop
			}
			log.Warnln("[Reload] Ignoring serve failure of a replaced generation")
			continue
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
		// restart of the control plane: the old generation is aborted and
		// closed, the new one is built and served. Established flows through
		// the old generation are dropped and the datapath is detached while
		// the new one loads (about a second). Only the DNS cache and dialer
		// health are carried over.
		var dnsCache map[string]*control.DnsCache
		if conf.Dns.IpVersionPrefer == newConf.Dns.IpVersionPrefer {
			// Only keep dns cache when ip version preference not change.
			dnsCache = gen.c.CloneDnsCache()
		}
		log.Warnln("[Reload] Stop old control plane")
		active.Store(nil)
		stopGeneration(log, gen)

		log.Warnln("[Reload] Load new control plane")
		var errReload error
		newGen, err := startGeneration(log, gen.c, dnsCache, newConf, externGeoDataDirs)
		if err != nil {
			errReload = err
			log.WithFields(logrus.Fields{
				"err": err,
			}).Errorln("[Reload] Failed to reload; try to roll back configuration")
			// Load last config back.
			newGen, err = startGeneration(log, gen.c, dnsCache, conf, externGeoDataDirs)
			if err != nil {
				shutdownProcess(log, nil)
				log.WithFields(logrus.Fields{
					"err": err,
				}).Fatalln("[Reload] Failed to roll back configuration")
			}
			newConf = conf
			log.Errorln("[Reload] Last reload failed; rolled back configuration")
		}
		gen = newGen
		conf = newConf
		active.Store(gen.c)
		log.Warnln("[Reload] Finished")
		newReloadMsg.Callback <- errReload
	}

	active.Store(nil)
	return shutdownProcess(log, gen)
}

// startGeneration builds a control plane for conf, inherits dialer health
// from previous when given, and serves it. On failure nothing of the new
// generation is left behind.
func startGeneration(log *logrus.Logger, previous *control.ControlPlane, dnsCache map[string]*control.DnsCache, conf *daeConfig.Config, externGeoDataDirs []string) (*generation, error) {
	c, err := newControlPlane(log, dnsCache, conf, externGeoDataDirs)
	if err != nil {
		return nil, err
	}
	gen := &generation{c: c}
	if previous != nil {
		// Health snapshots survive Close; carrying them over keeps the
		// groups routable instead of falling back until the first check.
		c.InheritDialerHealthFrom(previous)
	}
	if gen.listener, err = serve(log, gen, conf.Global.TproxyPort); err != nil {
		stopGeneration(log, gen)
		return nil, err
	}
	return gen, nil
}

// serve listens in the dae netns and serves the generation in the background.
// It returns once the plane reports ready, or with the listen/serve error. A
// generation that dies after it was ready asks the run loop to exit unless
// it has been retired; one that never became ready is the caller's error to
// handle, so a failed reload candidate does not take the process down.
func serve(log *logrus.Logger, gen *generation, port uint16) (listener *control.Listener, err error) {
	err = control.GetDaeNetns().With(func() error {
		var listenErr error
		listener, listenErr = gen.c.Listen(port)
		return listenErr
	})
	if err != nil {
		// The netns wrapper can fail after Listen succeeded.
		if listener != nil {
			_ = listener.Close()
		}
		return nil, fmt.Errorf("listen in dae netns: %w", err)
	}
	readyChan := make(chan bool, 1)
	served := make(chan error, 1)
	go func() {
		served <- gen.c.Serve(readyChan, listener)
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
		if err := <-served; err != nil && !gen.retired.Load() {
			log.Errorln("Serve:", err)
			// Report with the generation's identity: by the time the run loop
			// receives this it may already have replaced the generation.
			ChReloadConfigs <- &ReloadMessage{failed: gen}
		}
	}()
	return listener, nil
}

// stopGeneration retires one generation in the order dae's own shutdown
// uses: stop accepting, detach the datapath, abort its flows, release the
// plane. The dae netns is process-wide and is left in place for the next
// generation.
func stopGeneration(log *logrus.Logger, gen *generation) {
	gen.retired.Store(true)
	if gen.listener != nil {
		if e := gen.listener.Close(); e != nil {
			log.Warnf("close listener: %v", e)
		}
		gen.listener = nil
	}
	if e := gen.c.DetachBpfHooks(); e != nil {
		log.Warnf("detach BPF hooks: %v", e)
	}
	// Close alone leaves generation-owned UDP endpoints and their receive
	// goroutines alive; they cannot outlive the datapath they belong to.
	if e := gen.c.AbortConnections(); e != nil {
		log.Warnf("abort connections: %v", e)
	}
	if e := gen.c.Close(); e != nil {
		log.Warnf("close control plane: %v", e)
	}
}

// configureTransparentHugePages applies disable_thp to the process, as dae's
// cmd does before every control plane build; prctl(PR_SET_THP_DISABLE) is
// per-mm and idempotent, so a rollback simply applies the old value again.
func configureTransparentHugePages(log *logrus.Logger, disable bool) {
	value := uintptr(0)
	if disable {
		value = 1
	}
	if err := unix.Prctl(unix.PR_SET_THP_DISABLE, value, 0, 0, 0); err != nil {
		log.WithError(err).Warnf("Failed to configure transparent huge pages (disable=%v)", disable)
	}
}

// shutdownProcess releases the process-wide state after the last generation
// (nil when none is serving) and returns the plane's close error.
func shutdownProcess(log *logrus.Logger, gen *generation) error {
	var closeErr error
	if gen != nil {
		gen.retired.Store(true)
		if gen.listener != nil {
			if e := gen.listener.Close(); e != nil {
				log.Warnf("close listener: %v", e)
			}
		}
		if e := gen.c.DetachBpfHooks(); e != nil {
			log.Warnf("detach BPF hooks: %v", e)
		}
	}
	if e := control.GetDaeNetns().Close(); e != nil {
		log.Warnf("close dae netns: %v", e)
	}
	if gen != nil {
		if e := gen.c.AbortConnections(); e != nil {
			log.Warnf("abort connections: %v", e)
		}
		closeErr = gen.c.Close()
	}
	control.ResetGlobalUdpState()
	if closeErr != nil {
		return fmt.Errorf("close control plane: %w", closeErr)
	}
	return nil
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

	// Generation-scoped direct dialers and system resolver. The config was
	// parsed, not validated: a bad fallback_resolver is an error, not a panic.
	fallbackResolver, err := netip.ParseAddrPort(conf.Global.FallbackResolver)
	if err != nil {
		return nil, fmt.Errorf("fallback_resolver %q: %w", conf.Global.FallbackResolver, err)
	}
	configureTransparentHugePages(log, conf.Global.DisableTHP)
	directDialers := direct.NewDirectDialers(conf.Global.FallbackResolver)
	systemDNSResolver := netutils.NewSystemDNSResolver(fallbackResolver)

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
