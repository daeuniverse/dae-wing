/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dae

import (
	"fmt"
	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
	daeConfig "github.com/v2rayA/dae/config"
	"github.com/v2rayA/dae/control"
	"github.com/v2rayA/dae/pkg/config_parser"
	"github.com/v2rayA/dae/pkg/logger"
	"runtime"
)

type ReloadMessage struct {
	Config   *daeConfig.Config
	Callback chan<- bool
}

var ChReloadConfigs = make(chan *ReloadMessage, 16)
var EmptyConfig *daeConfig.Config

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

func Run(log *logrus.Logger, conf *daeConfig.Config, disableTimestamp bool, dry bool) (err error) {

	// Not really run dae.
	if dry {
	dryLoop:
		for newConf := range ChReloadConfigs {
			switch newConf {
			case nil:
				break dryLoop
			default:
			}
		}
		return nil
	}

	// New ControlPlane.
	c, err := newControlPlane(log, nil, conf)
	if err != nil {
		return err
	}

	// Serve tproxy TCP/UDP server util signals.
	var listener *control.Listener
	go func() {
		readyChan := make(chan bool, 1)
		go func() {
			<-readyChan
			log.Infoln("Ready")
		}()
		if listener, err = c.ListenAndServe(readyChan, conf.Global.TproxyPort); err != nil {
			log.Errorln("ListenAndServe:", err)
		}
		// Exit
		ChReloadConfigs <- nil
	}()
	reloading := false
	isRollback := false
	var chCallback chan<- bool
loop:
	for newReloadMsg := range ChReloadConfigs {
		switch newReloadMsg {
		case nil:
			// We will receive nil after control plane being Closed.
			// We'll judge if we are in a reloading.
			if reloading {
				// Serve.
				reloading = false
				log.Warnln("[Reload] Serve")
				readyChan := make(chan bool, 1)
				go func() {
					if err := c.Serve(readyChan, listener); err != nil {
						log.Errorln("ListenAndServe:", err)
					}
					// Exit
					ChReloadConfigs <- nil
				}()
				<-readyChan
				log.Warnln("[Reload] Finished")
				if !isRollback {
					chCallback <- true
				}
			} else {
				// Listening error.
				break loop
			}
		default:
			// Reload signal.
			log.Warnln("[Reload] Received reload signal; prepare to reload")

			// New logger.
			log = logger.NewLogger(newReloadMsg.Config.Global.LogLevel, disableTimestamp)
			logrus.SetLevel(log.Level)

			// New control plane.
			obj := c.EjectBpf()
			log.Warnln("[Reload] Load new control plane")
			newC, err := newControlPlane(log, obj, newReloadMsg.Config)
			if err != nil {
				log.WithFields(logrus.Fields{
					"err": err,
				}).Errorln("[Reload] Failed to reload; try to roll back configuration")
				// Load last config back.
				newC, err = newControlPlane(log, obj, conf)
				if err != nil {
					obj.Close()
					c.Close()
					log.WithFields(logrus.Fields{
						"err": err,
					}).Fatalln("[Reload] Failed to roll back configuration")
				}
				log.Warnln("[Reload] Last reload failed; rolled back configuration")
				newReloadMsg.Callback <- false
				isRollback = true
			} else {
				log.Warnln("[Reload] Stopped old control plane")
				isRollback = false
			}

			// Inject bpf objects into the new control plane life-cycle.
			newC.InjectBpf(obj)

			// Prepare new context.
			conf = newReloadMsg.Config
			reloading = true
			chCallback = newReloadMsg.Callback
			oldC := c
			c = newC

			// Ready to close.
			oldC.Close()
		}
	}
	if e := c.Close(); e != nil {
		return fmt.Errorf("close control plane: %w", e)
	}
	return nil
}

func newControlPlane(log *logrus.Logger, bpf interface{}, conf *daeConfig.Config) (c *control.ControlPlane, err error) {

	// Print configuration.
	if log.IsLevelEnabled(logrus.DebugLevel) {
		bConf, _ := conf.Marshal(2)
		log.Debugln(string(bConf))
	}

	// Deep copy to prevent modification.
	conf = deepcopy.Copy(conf).(*daeConfig.Config)

	/// Get subscription -> nodeList mapping.
	subscriptionToNodeList := map[string][]string{}
	if len(conf.Node) > 0 {
		for _, node := range conf.Node {
			subscriptionToNodeList[""] = append(subscriptionToNodeList[""], string(node))
		}
	}
	if len(conf.Subscription) > 0 {
		return nil, fmt.Errorf("daeConfig.subscription is not supported in dae-wing")
	}

	c, err = control.NewControlPlane(
		log,
		bpf,
		subscriptionToNodeList,
		conf.Group,
		&conf.Routing,
		&conf.Global,
		&conf.Dns,
	)
	if err != nil {
		return nil, err
	}
	// Call GC to release memory.
	runtime.GC()

	return c, nil
}
