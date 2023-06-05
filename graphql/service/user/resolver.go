/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package user

import (
	"github.com/daeuniverse/dae-wing/db"
)

type Resolver struct {
	User *db.User
}

func (r *Resolver) Username() string {
	return r.User.Username
}

func (r *Resolver) Name() *string {
	return r.User.Name
}

func (r *Resolver) Avatar() *string {
	return r.User.Avatar
}
