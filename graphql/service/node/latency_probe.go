/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package node

import (
	"context"
	"time"

	dialer "github.com/daeuniverse/dae/component/outbound/dialer"
	"github.com/daeuniverse/outbound/netproxy"
)

const latencyProbeTimeout = 1500 * time.Millisecond

type latencyProbeResult struct {
	Alive     bool
	Latency   time.Duration
	Message   string
	CheckedAt time.Time
}

// probeLatency runs one bounded HTTP check per IP family over the node's TCP
// path and keeps the best round trip. It is the ad-hoc probe dae exposed as
// Dialer.ProbeLatencyFast before v2.1.1; the dialer's own health state is
// left untouched because the node is not part of a running group here.
func probeLatency(ctx context.Context, d *dialer.Dialer) *latencyProbeResult {
	result := &latencyProbeResult{CheckedAt: time.Now()}
	opt, err := d.TcpCheckOptionRaw.Option()
	if err != nil {
		result.Message = err.Error()
		return result
	}
	var soMark uint32
	var mptcp bool
	if network, err := netproxy.ParseMagicNetwork(d.TcpCheckOptionRaw.ResolverNetwork); err == nil {
		soMark = network.Mark
		mptcp = network.Mptcp
	}

	var lastErr error
	for _, family := range []struct {
		idx int
		ip  interface{ IsValid() bool }
	}{{dialer.IdxTcp4, opt.Ip4}, {dialer.IdxTcp6, opt.Ip6}} {
		if !family.ip.IsValid() {
			continue
		}
		ip := opt.Ip4
		if family.idx == dialer.IdxTcp6 {
			ip = opt.Ip6
		}
		probeCtx, cancel := context.WithTimeout(ctx, latencyProbeTimeout)
		start := time.Now()
		ok, err := d.HttpCheck(probeCtx, family.idx, opt.Url, ip, opt.Method, soMark, mptcp)
		latency := time.Since(start)
		cancel()
		if err != nil {
			lastErr = err
		}
		if !ok || err != nil {
			continue
		}
		if !result.Alive || latency < result.Latency {
			result.Alive = true
			result.Latency = latency
		}
	}
	if result.Alive {
		return result
	}
	if lastErr != nil {
		result.Message = lastErr.Error()
	} else {
		result.Message = "no latency result"
	}
	return result
}
