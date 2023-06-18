/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package common

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	daeCommon "github.com/daeuniverse/dae/common"
	"github.com/graph-gophers/graphql-go"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func PasswordToHash32(entropy []byte, password string) string {
	h := sha256.New()
	h.Write([]byte("1b74413f-f3b8-409f-ak47-e8c062e3472a"))
	h.Write(entropy)
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))[:32]
}

func EncodeCursor(id uint) (cursor graphql.ID) {
	cursor.UnmarshalGraphQL(base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(
		[]byte(fmt.Sprintf("cursor%v", id)),
	))
	return cursor
}

func EncodeNullableCursor(nullableId *uint) (cursor *graphql.ID) {
	if nullableId == nil {
		return nil
	}
	id := EncodeCursor(*nullableId)
	return &id
}

func DecodeCursor(cursor graphql.ID) (id uint, err error) {
	_id, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(string(cursor))
	if err != nil {
		return 0, fmt.Errorf("failed to parse cursor")
	}
	intId, err := strconv.Atoi(strings.TrimPrefix(string(_id), "cursor"))
	if err != nil {
		return 0, fmt.Errorf("failed to parse cursor")
	}
	return uint(intId), nil
}

func DecodeCursorBatch(_ids []graphql.ID) (ids []uint, err error) {
	for _, _id := range _ids {
		id, err := DecodeCursor(_id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func GetIfAddrs() (globalIfAddrs []string, err error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}
	for _, family := range []int{unix.AF_INET, unix.AF_INET6} {
		for _, link := range linkList {
			if link.Attrs().Flags&unix.RTF_UP != unix.RTF_UP {
				// Interface is down.
				continue
			}
			addrs, err := netlink.AddrList(link, family)
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if addr.IP == nil ||
					addr.IP.IsUnspecified() ||
					addr.IP.IsInterfaceLocalMulticast() ||
					addr.IP.IsMulticast() ||
					addr.IP.IsLinkLocalMulticast() {
					continue
				}
				globalIfAddrs = append(globalIfAddrs, addr.IP.String())
			}
		}
	}
	return daeCommon.Deduplicate(globalIfAddrs), nil
}
