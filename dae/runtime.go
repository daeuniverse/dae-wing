/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dae

import (
	"errors"
	"time"

	"github.com/daeuniverse/dae/control"
)

type RuntimeTrafficSample struct {
	Timestamp    time.Time
	UploadRate   uint64
	DownloadRate uint64
}

type RuntimeOverview struct {
	UpdatedAt         time.Time
	UploadRate        uint64
	DownloadRate      uint64
	UploadTotal       uint64
	DownloadTotal     uint64
	ActiveConnections int
	UDPSessions       int
	Samples           []RuntimeTrafficSample
}

func GetRuntimeOverview(windowSec int, maxPoints int) (*RuntimeOverview, error) {
	activeTCPConnections := 0
	ctl, err := ControlPlane()
	if err != nil {
		if !errors.Is(err, ErrControlPlaneNotInit) {
			return nil, err
		}
	} else {
		activeTCPConnections = ctl.ActiveTCPConnections()
	}

	snapshot := control.SnapshotRuntimeStats(activeTCPConnections, control.DefaultUdpEndpointPool.Count(), windowSec, maxPoints)

	samples := make([]RuntimeTrafficSample, 0, len(snapshot.Samples))
	for _, sample := range snapshot.Samples {
		samples = append(samples, RuntimeTrafficSample{
			Timestamp:    sample.Timestamp,
			UploadRate:   sample.UploadRate,
			DownloadRate: sample.DownloadRate,
		})
	}

	return &RuntimeOverview{
		UpdatedAt:         snapshot.UpdatedAt,
		UploadRate:        snapshot.UploadRate,
		DownloadRate:      snapshot.DownloadRate,
		UploadTotal:       snapshot.UploadTotal,
		DownloadTotal:     snapshot.DownloadTotal,
		ActiveConnections: snapshot.ActiveConnections,
		UDPSessions:       snapshot.UDPSessions,
		Samples:           samples,
	}, nil
}
