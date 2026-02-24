package service

import (
	"encoding/json"

	"smartcalendar/model"

	"gorm.io/gorm"
)

// CreateOperationLog 写入操作记录，支持传入事务。
func CreateOperationLog(tx *gorm.DB, userID uint, action string, targetTitle string, detail interface{}) error {
	db := tx
	if db == nil {
		db = model.DB
	}
	var detailString string
	if detail != nil {
		if bytes, err := json.Marshal(detail); err == nil {
			detailString = string(bytes)
		}
	}
	log := model.OperationLog{
		UserID:      userID,
		Action:      action,
		TargetTitle: targetTitle,
		Detail:      detailString,
	}
	return db.Create(&log).Error
}
