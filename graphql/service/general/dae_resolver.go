/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package general

import (
	"context"
	"github.com/daeuniverse/dae-wing/db"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

type DaeResolver struct {
	Ctx context.Context
}

func (r *DaeResolver) Running() (bool, error) {
	var m db.System
	q := db.DB(r.Ctx).Select("running").Model(&db.System{}).FirstOrCreate(&m)
	if q.Error != nil {
		return false, q.Error
	}
	return m.Running, nil
}

func (r *DaeResolver) Modified() (bool, error) {
	var m db.System
	tx := db.BeginReadOnlyTx(r.Ctx)
	defer tx.Commit()
	q := tx.Model(&m).
		Preload(clause.Associations).
		FirstOrCreate(&m)
	if q.Error != nil {
		return false, q.Error
	}
	if !m.Running {
		return false, nil
	}
	if m.RunningConfig == nil || m.RunningConfig.Version != m.RunningConfigVersion ||
		m.RunningDns == nil || m.RunningDns.Version != m.RunningDnsVersion ||
		m.RunningRouting == nil || m.RunningRouting.Version != m.RunningRoutingVersion ||
		len(m.RunningGroups) == 0 {
		return true, nil
	}
	groupVersions := strings.Split(m.RunningGroupVersions, ",")
	for i, g := range m.RunningGroups {
		if strconv.FormatUint(uint64(g.Version), 10) != groupVersions[i] {
			return true, nil
		}
	}
	return false, nil
}
