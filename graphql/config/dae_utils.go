/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package config

import (
	"github.com/v2rayA/dae/common"
	"github.com/v2rayA/dae/config"
)

func NecessaryOutbounds(routing *config.Routing) (outbounds []string) {
	f := config.FunctionOrStringToFunction(routing.Fallback)
	outbounds = append(outbounds, f.Name)
	for _, r := range routing.Rules {
		outbounds = append(outbounds, r.Outbound.Name)
	}
	return common.Deduplicate(outbounds)
}
