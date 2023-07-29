/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dae

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/daeuniverse/dae/cmd"
	daeCommon "github.com/daeuniverse/dae/common"
	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/daeuniverse/dae/pkg/config_parser"
	"github.com/mzz2017/softwind/netproxy"
	"github.com/mzz2017/softwind/protocol/direct"
	"github.com/sirupsen/logrus"
)

var (
	EmptyGroupSection        = `group {}`
	EmptySubscriptionSection = `subscription {}`
	EmptyNodeSection         = `node {}`
	EmptyRoutingSection      = `routing {}`
	EmptyDnsSection          = `dns {}`
	EmptyGlobalSection       = `global {}`
)

func NecessaryOutbounds(routing *daeConfig.Routing) (outbounds []string) {
	f := daeConfig.FunctionOrStringToFunction(routing.Fallback)
	outbounds = append(outbounds, f.Name)
	for _, r := range routing.Rules {
		outbound := r.Outbound.Name
		if outbound != "must_rules" {
			outbound = strings.TrimPrefix(outbound, "must_")
		}
		outbounds = append(outbounds, outbound)
	}
	return daeCommon.Deduplicate(outbounds)
}

func ParseConfig(globalSection *string, dnsSection *string, routingSection *string) (*daeConfig.Config, error) {
	if globalSection == nil {
		globalSection = &EmptyGlobalSection
	}
	if dnsSection == nil {
		dnsSection = &EmptyDnsSection
	}
	if routingSection == nil {
		routingSection = &EmptyRoutingSection
	}
	strConfig := strings.Join([]string{
		*globalSection,
		*dnsSection,
		*routingSection,
		EmptyGroupSection,
		EmptySubscriptionSection,
		EmptyNodeSection,
	}, "\n")
	// Parse it to sections.
	sections, err := config_parser.Parse(strConfig)
	if err != nil {
		return nil, err
	}
	// New dae.Config from sections.
	c, err := daeConfig.New(sections)
	if err != nil {
		return nil, err
	}
	return c, err
}

func preprocessWanInterfaceAuto(params *daeConfig.Config) error {
	// preprocess "auto".
	ifs := make([]string, 0, len(params.Global.WanInterface)+2)
	for _, ifname := range params.Global.WanInterface {
		if ifname == "auto" {
			defaultIfs, err := daeCommon.GetDefaultIfnames()
			if err != nil {
				return fmt.Errorf("failed to convert 'auto': %w", err)
			}
			ifs = append(ifs, defaultIfs...)
		} else {
			ifs = append(ifs, ifname)
		}
	}
	params.Global.WanInterface = daeCommon.Deduplicate(ifs)
	return nil
}

func WaitForNetwork(log *logrus.Logger) {
	epo := 5 * time.Second
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (c net.Conn, err error) {
				cd := netproxy.ContextDialer{Dialer: direct.SymmetricDirect}
				conn, err := cd.DialContext(ctx, "tcp", addr)
				if err != nil {
					return nil, err
				}
				return &netproxy.FakeNetConn{
					Conn:  conn,
					LAddr: nil,
					RAddr: nil,
				}, nil
			},
		},
		Timeout: epo,
	}
	log.Infoln("Waiting for network...")
	for i := 0; ; i++ {
		resp, err := client.Get(cmd.CheckNetworkLinks[i%len(cmd.CheckNetworkLinks)])
		if err != nil {
			log.Debugln("CheckNetwork:", err)
			var neterr net.Error
			if errors.As(err, &neterr) && neterr.Timeout() {
				// Do not sleep.
				continue
			}
			time.Sleep(epo)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 500 {
			break
		}
		log.Infof("Bad status: %v (%v)", resp.Status, resp.StatusCode)
		time.Sleep(epo)
	}
	log.Infoln("Network online.")
}
