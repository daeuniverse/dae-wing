/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package general

import (
	"context"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type Resolver struct {
}

func (r *Resolver) Dae() *DaeResolver {
	return &DaeResolver{Ctx: context.TODO()}
}

func (r *Resolver) Interfaces(args *struct {
	Up *bool
}) (rs []*InterfaceResolver, err error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}
	for _, link := range linkList {
		if args.Up != nil {
			if (link.Attrs().Flags&unix.RTF_UP == unix.RTF_UP) != *args.Up {
				continue
			}
		}

		rs = append(rs, &InterfaceResolver{Link: link})
	}
	return rs, nil
}
