/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package general

import (
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"net"
)

type InterfaceResolver struct {
	netlink.Link
}

func (r *InterfaceResolver) Name() string {
	return r.Link.Attrs().Name
}

func (r *InterfaceResolver) Ifindex() int32 {
	return int32(r.Attrs().Index)
}

func (r *InterfaceResolver) Flag() *flagResolver {
	return &flagResolver{
		flags: r.Link.Attrs().Flags,
		link:  r.Link,
	}
}

func (r *InterfaceResolver) Ip(args *struct {
	OnlyGlobalScope *bool
}) (list []string, err error) {
	for _, family := range []int{unix.AF_INET, unix.AF_INET6} {
		addrs, err := netlink.AddrList(r.Link, family)
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			if addr.Scope != unix.RT_SCOPE_UNIVERSE && args.OnlyGlobalScope != nil && *args.OnlyGlobalScope {
				continue
			}
			list = append(list, addr.IPNet.String())
		}
	}
	return list, nil
}

type flagResolver struct {
	flags net.Flags
	link  netlink.Link
}

func (r *flagResolver) Up() bool {
	return r.flags&unix.RTF_UP == unix.RTF_UP
}

type DefaultRoute struct {
	IpVersion string
	Gateway   *string
	Source    *string
}

func (r *flagResolver) Default() (dr *[]*DefaultRoute, err error) {
	dr = new([]*DefaultRoute)
	for _, family := range []int{unix.AF_INET, unix.AF_INET6} {
		rs, err := netlink.RouteList(r.link, family)
		if err != nil {
			return nil, err
		}
		for _, route := range rs {
			if route.Dst != nil {
				continue
			}
			r := &DefaultRoute{}
			if family == unix.AF_INET {
				r.IpVersion = "4"
			} else {
				r.IpVersion = "6"
			}
			if route.Gw != nil {
				strGw := route.Gw.String()
				r.Gateway = &strGw
			}
			if route.Src != nil {
				strSrc := route.Src.String()
				r.Source = &strSrc
			}
			*dr = append(*dr, r)
		}
	}
	if len(*dr) == 0 {
		dr = nil
	}
	return dr, nil
}
