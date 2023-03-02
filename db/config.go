/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package db

import (
	daeConfig "github.com/v2rayA/dae/config"
	"github.com/v2rayA/dae/pkg/config_parser"
	"strings"
)

type Config struct {
	ID      uint   `gorm:"primaryKey;autoIncrement"`
	Global  string `gorm:"not null"`
	Dns     string `gorm:"not null"`
	Routing string `gorm:"not null"`

	Selected bool `gorm:"not null"` // Redundancy for convenient.
}

func (m *Config) ToDaeConfig() (*daeConfig.Config, error) {
	strConfig := strings.Join([]string{
		m.Global,
		m.Dns,
		m.Routing,
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
