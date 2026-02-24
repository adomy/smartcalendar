package service

import (
	"fmt"
	"time"

	"smartcalendar/model"
)

// GenerateReminderNotifications 生成 15 分钟内将开始的日程提醒通知。
func GenerateReminderNotifications(now time.Time) error {
	end := now.Add(15 * time.Minute)
	var events []model.Event
	if err := model.DB.Where("start_time >= ? AND start_time <= ?", now, end).
		Preload("Creator").
		Preload("Participants.User").
		Find(&events).Error; err != nil {
		return err
	}

	for _, event := range events {
		userIDs := []uint{event.UserID}
		for _, participant := range event.Participants {
			userIDs = append(userIDs, participant.UserID)
		}
		for _, userID := range uniqueUintList(userIDs) {
			exists, err := reminderExists(event.ID, userID)
			if err != nil {
				return err
			}
			if exists {
				continue
			}
			content := fmt.Sprintf("您的日程《%s》将在 15 分钟后开始", event.Title)
			notification := model.Notification{
				UserID:  userID,
				Type:    "reminder",
				Content: content,
				EventID: event.ID,
				IsRead:  false,
			}
			if err := model.DB.Create(&notification).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// reminderExists 判断提醒通知是否已存在。
func reminderExists(eventID uint, userID uint) (bool, error) {
	var count int64
	if err := model.DB.Model(&model.Notification{}).
		Where("type = ? AND event_id = ? AND user_id = ?", "reminder", eventID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// uniqueUintList 对 uint 列表去重。
func uniqueUintList(list []uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0, len(list))
	for _, item := range list {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
