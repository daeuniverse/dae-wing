/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package internal

import (
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/config"
	"gorm.io/gorm"
)

func ReferenceGroups(d *gorm.DB) (groups []string, err error) {
	var sys db.System
	if err = d.Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
		return nil, err
	}

	if !sys.Running {
		return nil, nil
	}

	var conf db.Config
	if err = d.Model(&db.Config{}).Where("selected = ?", true).First(&conf).Error; err != nil {
		return nil, err
	}
	daeConf, err := conf.ToDaeConfig()
	if err != nil {
		return nil, err
	}
	groups = config.NecessaryOutbounds(&daeConf.Routing)
	return groups, nil
}

func SetModified(d *gorm.DB) (err error) {
	var sys db.System
	// Get ID.
	if err = d.Model(&sys).Select("id").FirstOrCreate(&sys).Error; err != nil {
		return err
	}
	return d.Model(&sys).Update("modified", true).Error
}
