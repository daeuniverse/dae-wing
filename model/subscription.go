/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package model

import (
	"fmt"
	"gorm.io/gorm"
)

var (
	InvalidRemarkError = fmt.Errorf("invalid remarks; only support numbers and letters")
	BadLinkFormatError = fmt.Errorf("not a valid link")
)

type SubscriptionModel struct {
	gorm.Model
	Remarks string `gorm:"not null;unique"`
	Link    string `gorm:"not null"`
	Status  string `gorm:"not null"` // Latency, error info, etc.
	Info    string `gorm:"not null"` // Maybe include some info from provider

	Nodes []NodeModel
}
//type subscription struct{}
//
//var Subscription subscription
//
//func (subscription) Create(ctx context.Context, model *SubscriptionModel) error {
//	if !ValidateRemarks(model.Remarks) {
//		return InvalidRemarkError
//	}
//	return db.DB(ctx).Create(model).Error
//}
//
//func (subscription) List(ctx context.Context, afterId uint, count int) (models []SubscriptionModel, err error) {
//	if err := db.DB(ctx).
//		Model(&SubscriptionModel{}).
//		Preload("Nodes").
//		Where("id > ?", afterId).
//		Limit(count).
//		Find(&models).Error; err != nil {
//		return nil, err
//	}
//	return models, nil
//}
//
//func (subscription) UpdateRemarks(ctx context.Context, model *SubscriptionModel) error {
//	if model.Remarks != "" && !ValidateRemarks(model.Remarks) {
//		return InvalidRemarkError
//	}
//	return db.DB(ctx).Model(&SubscriptionModel{}).Update("remarks", model).Error
//}
//
//func (subscription) UpdateStatus(ctx context.Context, model *SubscriptionModel) error {
//	return db.DB(ctx).Model(&SubscriptionModel{}).Update("status", model).Error
//}
//
//func (subscription) Delete(ctx context.Context, model *SubscriptionModel) error {
//	return db.DB(ctx).Select(model.Nodes).Delete(model).Error
//}
//
//func (subscription) UpdateNodes(ctx context.Context, model *SubscriptionModel, nodes []*NodeModel) (err error) {
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
//	if err = tx.Model(NodeModel{}).Delete("subscription_id = ?", model.ID).Error; err != nil {
//		return err
//	}
//	if err = tx.Model(NodeModel{}).Create("subscription_id = ?", model.ID).Error; err != nil {
//		return err
//	}
//}
