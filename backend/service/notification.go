package service

import (
	"fmt"

	"smartcalendar/model"
)

// CreateInvitationNotifications 向参与人发送邀请通知。
func CreateInvitationNotifications(event model.Event, participantIDs []uint) error {
	for _, userID := range participantIDs {
		if userID == event.UserID {
			continue
		}
		content := fmt.Sprintf("%s 邀请你参加日程《%s》", event.Creator.Nickname, event.Title)
		notification := model.Notification{
			UserID:  userID,
			Type:    "invitation",
			Content: content,
			EventID: event.ID,
			IsRead:  false,
		}
		if err := model.DB.Create(&notification).Error; err != nil {
			return err
		}
	}
	return nil
}

// CreateChangeNotifications 向参与人发送变更通知。
func CreateChangeNotifications(event model.Event, participantIDs []uint) error {
	for _, userID := range participantIDs {
		if userID == event.UserID {
			continue
		}
		content := fmt.Sprintf("您参与的日程《%s》时间已更新", event.Title)
		notification := model.Notification{
			UserID:  userID,
			Type:    "change",
			Content: content,
			EventID: event.ID,
			IsRead:  false,
		}
		if err := model.DB.Create(&notification).Error; err != nil {
			return err
		}
	}
	return nil
}
