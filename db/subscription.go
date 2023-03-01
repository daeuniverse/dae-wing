/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package db

import (
	"fmt"
	"gorm.io/gorm"
)

var (
	InvalidRemarkError = fmt.Errorf("invalid remarks; only support numbers and letters")
	BadLinkFormatError = fmt.Errorf("not a valid link")
)

type Subscription struct {
	gorm.Model
	Link   string `gorm:"not null"`
	Status string `gorm:"not null"` // Latency, error info, etc.
	Info   string `gorm:"not null"` // Maybe include some info from provider

	Remarks *string `gorm:"unique"`

	Nodes []Node
}
