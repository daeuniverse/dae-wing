/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package db

type System struct {
	ID      uint `gorm:"primaryKey;autoIncrement"`
	Running bool `gorm:"not null;default:false"`
}
