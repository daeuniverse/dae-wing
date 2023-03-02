/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package main

import (
	"github.com/json-iterator/go/extra"
	"github.com/v2rayA/dae-wing/cmd"
	"os"
)

import (
	_ "github.com/mzz2017/softwind/protocol/shadowsocks"
	_ "github.com/mzz2017/softwind/protocol/trojanc"
	_ "github.com/mzz2017/softwind/protocol/vless"
	_ "github.com/mzz2017/softwind/protocol/vmess"
	_ "github.com/v2rayA/dae/component/outbound/dialer/http"
	_ "github.com/v2rayA/dae/component/outbound/dialer/shadowsocks"
	_ "github.com/v2rayA/dae/component/outbound/dialer/shadowsocksr"
	_ "github.com/v2rayA/dae/component/outbound/dialer/socks"
	_ "github.com/v2rayA/dae/component/outbound/dialer/trojan"
	_ "github.com/v2rayA/dae/component/outbound/dialer/v2ray"
)

func main() {
	extra.RegisterFuzzyDecoders()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
