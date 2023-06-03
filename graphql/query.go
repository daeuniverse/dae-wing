/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package graphql

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql/service/config"
	"github.com/daeuniverse/dae-wing/graphql/service/dns"
	"github.com/daeuniverse/dae-wing/graphql/service/general"
	"github.com/daeuniverse/dae-wing/graphql/service/group"
	"github.com/daeuniverse/dae-wing/graphql/service/node"
	"github.com/daeuniverse/dae-wing/graphql/service/routing"
	"github.com/daeuniverse/dae-wing/graphql/service/subscription"
	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/daeuniverse/dae/pkg/config_parser"
	"github.com/golang-jwt/jwt/v5"
	"github.com/graph-gophers/graphql-go"
	"github.com/tidwall/gjson"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"
	"io"
	"time"
)

type queryResolver struct{}

func (r *queryResolver) HealthCheck() int32 {
	return 1
}
func hashPassword(salt []byte, password string) (string, error) {
	h := sha3.NewShake256()
	_, err := h.Write(salt)
	if err != nil {
		return "", err
	}
	_, err = h.Write([]byte(password))
	if err != nil {
		return "", err
	}
	var hash [32]byte
	_, err = io.ReadFull(h, hash[:])
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash[:]), nil
}
func getToken(
	d *gorm.DB,
	username string,
	password string) (string, error) {
	var m db.User
	// Check username.
	q := d.Model(&db.User{}).Where("username = ?", username).First(&m)
	if q.Error != nil || q.RowsAffected == 0 {
		return "", fmt.Errorf("incorrect username or password")
	}
	// Check password.
	hashedPassword, err := hashPassword([]byte(m.JwtSecret), password)
	if err != nil {
		return "", err
	}
	if hashedPassword != m.PasswordHash {
		return "", fmt.Errorf("incorrect username or password")
	}

	// File a token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "admin",
		"sub":  m.Username,
		"exp":  time.Now().Add(30 * time.Hour * 24).UTC().Unix(),
	})
	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(m.JwtSecret))
}

func (r *queryResolver) Token(args *struct {
	Username string
	Password string
}) (string, error) {
	return getToken(db.DB(context.TODO()), args.Username, args.Password)
}
func numberUsers(d *gorm.DB) (int32, error) {
	var cnt int64
	if err := d.Model(&db.User{}).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return int32(cnt), nil

}
func (r *queryResolver) NumberUsers() (int32, error) {
	return numberUsers(db.DB(context.TODO()))
}

func userFromContext(ctx context.Context) (user *db.User, err error) {
	_user := ctx.Value("user")
	if _user == nil {
		return nil, fmt.Errorf("failed to get user")
	}
	user, ok := _user.(*db.User)
	if !ok {
		return nil, fmt.Errorf("failed to get user")
	}
	return user, nil
}
func (r *queryResolver) JsonStorage(ctx context.Context, args *struct {
	Paths *[]string
}) ([]string, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if args.Paths == nil {
		return []string{user.JsonStorage}, nil
	}
	results := gjson.GetMany(user.JsonStorage, *args.Paths...)
	var ret []string
	for _, r := range results {
		ret = append(ret, r.String())
	}
	return ret, nil
}

func (r *queryResolver) General() *general.Resolver {
	return &general.Resolver{}
}
func (r *queryResolver) Configs(args *struct {
	ID       *graphql.ID
	Selected *bool
}) (rs []*config.Resolver, err error) {
	// Check if query specific ID.
	var id uint
	q := db.DB(context.TODO()).Model(&db.Config{})
	if args.ID != nil {
		id, err = common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	if args.Selected != nil {
		q = q.Where("selected = ?", *args.Selected)
	}
	// Get configs from DB.
	var models []db.Config
	if err = q.Find(&models).Error; err != nil {
		return nil, err
	}
	for i := range models {
		m := &models[i]
		c, err := dae.ParseConfig(&m.Global, nil, nil)
		if err != nil {
			return nil, err
		}
		rs = append(rs, &config.Resolver{
			DaeGlobal: &c.Global,
			Model:     m,
		})
	}
	return rs, nil
}

func (r *queryResolver) Dnss(args *struct {
	ID       *graphql.ID
	Selected *bool
}) (rs []*dns.Resolver, err error) {
	// Check if query specific ID.
	var id uint
	q := db.DB(context.TODO()).Model(&db.Dns{})
	if args.ID != nil {
		id, err = common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	if args.Selected != nil {
		q = q.Where("selected = ?", *args.Selected)
	}
	// Get dns from DB.
	var models []db.Dns
	if err = q.Find(&models).Error; err != nil {
		return nil, err
	}
	for i := range models {
		m := &models[i]
		c, err := dae.ParseConfig(nil, &m.Dns, nil)
		if err != nil {
			return nil, err
		}
		rs = append(rs, &dns.Resolver{
			DaeDns: &c.Dns,
			Model:  m,
		})
	}
	return rs, nil
}

func (r *queryResolver) Routings(args *struct {
	ID       *graphql.ID
	Selected *bool
}) (rs []*routing.Resolver, err error) {
	// Check if query specific ID.
	var id uint
	q := db.DB(context.TODO()).Model(&db.Routing{})
	if args.ID != nil {
		id, err = common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	if args.Selected != nil {
		q = q.Where("selected = ?", *args.Selected)
	}
	// Get routing from DB.
	var models []db.Routing
	if err = q.Find(&models).Error; err != nil {
		return nil, err
	}
	for i := range models {
		m := &models[i]
		c, err := dae.ParseConfig(nil, nil, &m.Routing)
		if err != nil {
			return nil, err
		}
		rs = append(rs, &routing.Resolver{
			DaeRouting: &c.Routing,
			Model:      m,
		})
	}
	return rs, nil
}

func (r *queryResolver) ConfigFlatDesc() []*dae.FlatDesc {
	return dae.ExportFlatDesc()
}
func (r *queryResolver) ParsedRouting(args *struct{ Raw string }) (rr *routing.DaeResolver, err error) {
	sections, err := config_parser.Parse("global{} routing {" + args.Raw + "}")
	if err != nil {
		return nil, err
	}
	conf, err := daeConfig.New(sections)
	if err != nil {
		return nil, err
	}
	return &routing.DaeResolver{
		Routing: &conf.Routing,
	}, nil
}
func (r *queryResolver) ParsedDns(args *struct{ Raw string }) (dr *dns.DnsResolver, err error) {
	sections, err := config_parser.Parse("global{} dns {" + args.Raw + "} routing{}")
	if err != nil {
		return nil, err
	}
	conf, err := daeConfig.New(sections)
	if err != nil {
		return nil, err
	}
	return &dns.DnsResolver{
		Dns: &conf.Dns,
	}, nil
}
func (r *queryResolver) Subscriptions(args *struct{ ID *graphql.ID }) (rs []*subscription.Resolver, err error) {
	q := db.DB(context.TODO()).
		Model(&db.Subscription{})
	if args.ID != nil {
		id, err := common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	var models []db.Subscription
	if err = q.Find(&models).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	for _, _m := range models {
		m := _m
		rs = append(rs, &subscription.Resolver{
			Subscription: &m,
		})
	}
	return rs, nil
}

func (r *queryResolver) Group(args *struct{ Name string }) (rs *group.Resolver, err error) {
	var m db.Group
	if err = db.DB(context.TODO()).
		Model(&db.Group{}).
		Where("name = ?", args.Name).First(&m).Error; err != nil {
		return nil, err
	}
	return &group.Resolver{Group: &m}, nil
}
func (r *queryResolver) Groups(args *struct{ ID *graphql.ID }) (rs []*group.Resolver, err error) {
	q := db.DB(context.TODO()).
		Model(&db.Group{})
	if args.ID != nil {
		id, err := common.DecodeCursor(*args.ID)
		if err != nil {
			return nil, err
		}
		q = q.Where("id = ?", id)
	}
	var models []db.Group
	if err = q.Find(&models).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	for _, _m := range models {
		m := _m
		rs = append(rs, &group.Resolver{
			Group: &m,
		})
	}
	return rs, nil
}
func (r *queryResolver) Nodes(args *struct {
	ID             *graphql.ID
	SubscriptionID *graphql.ID
	First          *int32
	After          *graphql.ID
}) (rs *node.ConnectionResolver, err error) {
	return node.NewConnectionResolver(args.ID, args.SubscriptionID, args.First, args.After)
}
