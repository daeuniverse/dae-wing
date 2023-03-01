/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package model

import (
	"database/sql"
	"github.com/v2rayA/dae/component/outbound/dialer"
	"gorm.io/gorm"
	"strings"
)

func NewNodeModel(link string, remarks string, subscriptionId sql.NullInt64) (*NodeModel, error) {
	if !strings.Contains(link, "://") {
		return nil, BadLinkFormatError
	}
	if remarks != "" && !ValidateRemarks(remarks) {
		return nil, InvalidRemarkError
	}
	d, err := dialer.NewFromLink(&dialer.GlobalOption{}, dialer.InstanceOption{CheckEnabled: false}, link)
	if err != nil {
		return nil, err
	}
	property := d.Property()
	return &NodeModel{
		Model:          gorm.Model{},
		Link:           link,
		Name:           property.Name,
		Address:        property.Address,
		Protocol:       property.Protocol,
		Remarks:        remarks,
		Status:         "",
		SubscriptionID: subscriptionId,
		Subscription:   SubscriptionModel{},
	}, nil
}
