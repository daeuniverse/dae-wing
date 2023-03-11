/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package db

type Dns struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null;default:''"`
	Dns  string `gorm:"not null"`

	Selected bool `gorm:"not null"` // Redundancy for convenient.
	Version  uint `gorm:"not null;default:0"`

	versionUpdated uint32
}
