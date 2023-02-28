/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package common

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/graph-gophers/graphql-go"
	"strconv"
	"strings"
)

func PasswordToHash32(entropy []byte, password string) string {
	h := sha256.New()
	h.Write([]byte("1b74413f-f3b8-409f-ak47-e8c062e3472a"))
	h.Write(entropy)
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))[:32]
}

func EncodeCursor(id uint) (cursor graphql.ID) {
	cursor.UnmarshalGraphQL(base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("cursor%v", id)),
	))
	return cursor
}

func EncodeNullableCursor(nullableId sql.NullInt64) (cursor *graphql.ID) {
	if !nullableId.Valid {
		return nil
	}
	id := EncodeCursor(uint(nullableId.Int64))
	return &id
}

func DecodeCursor(cursor graphql.ID) (id uint, err error) {
	_id, err := base64.StdEncoding.DecodeString(string(cursor))
	if err != nil {
		return 0, fmt.Errorf("failed to parse cursor")
	}
	intId, err := strconv.Atoi(strings.TrimPrefix(string(_id), "cursor"))
	if err != nil {
		return 0, fmt.Errorf("failed to parse cursor")
	}
	return uint(intId), nil
}
