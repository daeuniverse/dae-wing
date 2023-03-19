/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package db

type User struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	Username     string `gorm:"unique;not null;index"`
	PasswordHash string `gorm:"not null"`
	JwtSecret    string `gorm:"not null"`
	JsonStorage  string `gorm:"not null;default:'{}'"`
}
