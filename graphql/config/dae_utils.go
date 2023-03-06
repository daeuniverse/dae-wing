/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package config

import (
	daeCommon "github.com/v2rayA/dae/common"
	daeConfig "github.com/v2rayA/dae/config"
)

func NecessaryOutbounds(routing *daeConfig.Routing) (outbounds []string) {
	f := daeConfig.FunctionOrStringToFunction(routing.Fallback)
	outbounds = append(outbounds, f.Name)
	for _, r := range routing.Rules {
		outbounds = append(outbounds, r.Outbound.Name)
	}
	return daeCommon.Deduplicate(outbounds)
}
