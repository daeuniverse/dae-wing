/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package db

import (
	"strings"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae/component/outbound/dialer"
)

func NewNodeModel(link string, tag *string, subscriptionId *uint) (*Node, error) {
	if !strings.Contains(link, "://") {
		return nil, BadLinkFormatError
	}
	var _tag string
	if tag != nil {
		if err := common.ValidateTag(*tag); err != nil {
			return nil, err
		}
		_tag = *tag
	}
	d, err := dialer.NewFromLink(&dialer.GlobalOption{}, dialer.InstanceOption{DisableCheck: false}, link, _tag)
	if err != nil {
		return nil, err
	}
	property := d.Property()
	return &Node{
		ID:             0,
		Link:           link,
		Name:           property.Name,
		Address:        property.Address,
		Protocol:       property.Protocol,
		Tag:            tag,
		SubscriptionID: subscriptionId,
		Subscription:   nil,
	}, nil
}
