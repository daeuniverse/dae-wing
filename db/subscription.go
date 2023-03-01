/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package db

import (
	"fmt"
	"gorm.io/gorm"
)

var (
	InvalidRemarkError = fmt.Errorf("invalid remarks; only support numbers and letters")
	BadLinkFormatError = fmt.Errorf("not a valid link")
)

type Subscription struct {
	gorm.Model
	Remarks string `gorm:"not null;unique"`
	Link    string `gorm:"not null"`
	Status  string `gorm:"not null"` // Latency, error info, etc.
	Info    string `gorm:"not null"` // Maybe include some info from provider

	Nodes []Node
}

//type subscription struct{}
//
//var Subscription subscription
//
//func (subscription) Create(ctx context.Context, model *Subscription) error {
//	if !ValidateRemarks(model.Remarks) {
//		return InvalidRemarkError
//	}
//	return db.DB(ctx).Create(model).Error
//}
//
//func (subscription) List(ctx context.Context, afterId uint, count int) (models []Subscription, err error) {
//	if err := db.DB(ctx).
//		Model(&Subscription{}).
//		Preload("Nodes").
//		Where("id > ?", afterId).
//		Limit(count).
//		Find(&models).Error; err != nil {
//		return nil, err
//	}
//	return models, nil
//}
//
//func (subscription) UpdateRemarks(ctx context.Context, model *Subscription) error {
//	if model.Remarks != "" && !ValidateRemarks(model.Remarks) {
//		return InvalidRemarkError
//	}
//	return db.DB(ctx).Model(&Subscription{}).Update("remarks", model).Error
//}
//
//func (subscription) UpdateStatus(ctx context.Context, model *Subscription) error {
//	return db.DB(ctx).Model(&Subscription{}).Update("status", model).Error
//}
//
//func (subscription) Delete(ctx context.Context, model *Subscription) error {
//	return db.DB(ctx).Select(model.Nodes).Delete(model).Error
//}
//
//func (subscription) UpdateNodes(ctx context.Context, model *Subscription, nodes []*Node) (err error) {
//	tx := db.DB(ctx).Begin(&sql.TxOptions{
//		Isolation: sql.LevelSerializable,
//		ReadOnly:  false,
//	})
//	defer func() {
//		if err != nil {
//			tx.Callback()
//		} else {
//			tx.Commit()
//		}
//	}()
//	if err = tx.Model(Node{}).Delete("subscription_id = ?", model.ID).Error; err != nil {
//		return err
//	}
//	if err = tx.Model(Node{}).Create("subscription_id = ?", model.ID).Error; err != nil {
//		return err
//	}
//}
