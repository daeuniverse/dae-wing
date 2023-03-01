/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package db

import (
	"github.com/v2rayA/dae-wing/common"
	"github.com/v2rayA/dae/component/outbound/dialer"
	"gorm.io/gorm"
	"strings"
)

func NewNodeModel(link string, remarks *string, subscriptionId *uint) (*Node, error) {
	if !strings.Contains(link, "://") {
		return nil, BadLinkFormatError
	}
	if remarks != nil && !common.ValidateRemarks(*remarks) {
		return nil, InvalidRemarkError
	}
	d, err := dialer.NewFromLink(&dialer.GlobalOption{}, dialer.InstanceOption{CheckEnabled: false}, link)
	if err != nil {
		return nil, err
	}
	property := d.Property()
	return &Node{
		Model:          gorm.Model{},
		Link:           link,
		Name:           property.Name,
		Address:        property.Address,
		Protocol:       property.Protocol,
		Remarks:        remarks,
		SubscriptionID: subscriptionId,
		Subscription:   nil,
	}, nil
}
