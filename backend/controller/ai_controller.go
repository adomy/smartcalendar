package controller

import (
	"errors"
	"time"

	"smartcalendar/ai"
	"smartcalendar/model"
	"smartcalendar/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AIController 负责 AI 对话入口与确认执行。
type AIController struct {
	Service *ai.AIService
}

// AIChatRequest 表示 AI 对话输入与确认参数。
type AIChatRequest struct {
	Message   string  `json:"message" binding:"required,min=1,max=1000"`
	ConfirmID string  `json:"confirm_id"`
	Confirm   bool    `json:"confirm"`
	EventID   *uint64 `json:"event_id"`
}

// Chat 处理 AI 对话、候选返回与确认执行。
func (a AIController) Chat(c *gin.Context) {
	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败："+err.Error())
		return
	}
	user := c.MustGet("user").(model.User)

	if req.Confirm && req.ConfirmID != "" {
		proposal, ok := a.Service.ConsumeProposal(req.ConfirmID)
		if !ok {
			Error(c, 40001, "确认已过期，请重新输入")
			return
		}
		if req.EventID != nil && proposal.EventID == nil {
			proposal.EventID = req.EventID
		}
		switch proposal.Action {
		case "create":
			event, err := createEventFromProposal(user, proposal)
			if err != nil {
				Error(c, 50000, "服务器内部错误")
				return
			}
			Success(c, gin.H{
				"status": "success",
				"intent": "create",
				"result": "已为你创建日程：" + event.Title + " " + event.StartTime.Format("2006-01-02 15:04") + "-" + event.EndTime.Format("15:04"),
				"event":  buildEventResponse(event, user.ID),
			})
			return
		case "update":
			event, err := updateEventFromProposal(user, proposal)
			if err != nil {
				if errors.Is(err, errNeedEventID) {
					Success(c, gin.H{
						"status": "need_confirm",
						"intent": "update",
						"result": "匹配到多个日程，请指定日程ID后确认",
					})
					return
				}
				Error(c, 50000, "服务器内部错误")
				return
			}
			Success(c, gin.H{
				"status": "success",
				"intent": "update",
				"result": "已为你更新日程：" + event.Title,
				"event":  buildEventResponse(event, user.ID),
			})
			return
		case "delete":
			event, err := deleteEventFromProposal(user, proposal)
			if err != nil {
				if errors.Is(err, errNeedEventID) {
					Success(c, gin.H{
						"status": "need_confirm",
						"intent": "delete",
						"result": "匹配到多个日程，请指定日程ID后确认",
					})
					return
				}
				Error(c, 50000, "服务器内部错误")
				return
			}
			Success(c, gin.H{
				"status": "success",
				"intent": "delete",
				"result": "已为你删除日程：" + event.Title,
			})
			return
		default:
			Error(c, 40001, "意图不支持")
			return
		}
	}

	result, err := a.Service.ParseMessage(req.Message)
	if err != nil {
		Error(c, 50000, "服务器内部错误："+err.Error())
		return
	}
	if !result.NeedConfirm {
		Success(c, gin.H{
			"status": "success",
			"intent": result.Intent,
			"result": result.Result,
		})
		return
	}

	var candidates []model.Event
	if result.Intent == "update" || result.Intent == "delete" {
		candidates, err = findCandidateEvents(user.ID, result.Proposal)
		if err != nil {
			Error(c, 50000, "服务器内部错误")
			return
		}
		if len(candidates) == 0 && result.Proposal.EventID == nil {
			Success(c, gin.H{
				"status": "success",
				"intent": result.Intent,
				"result": "未找到匹配的日程，请补充时间、关键词或日程ID",
			})
			return
		}
		if len(candidates) == 1 && result.Proposal.EventID == nil {
			candidateEventID := uint64(candidates[0].ID)
			result.Proposal.EventID = &candidateEventID
		}
	}

	confirmID := a.Service.StoreProposal(result.Proposal)
	Success(c, gin.H{
		"status":     "need_confirm",
		"intent":     result.Intent,
		"result":     result.Result,
		"confirm_id": confirmID,
		"proposal": gin.H{
			"action":               result.Proposal.Action,
			"title":                result.Proposal.Title,
			"type":                 result.Proposal.Type,
			"start_time":           result.Proposal.StartTime,
			"end_time":             result.Proposal.EndTime,
			"location":             result.Proposal.Location,
			"participant_keywords": result.Proposal.ParticipantKeywords,
			"description":          result.Proposal.Description,
			"event_id":             result.Proposal.EventID,
			"target_time":          result.Proposal.TargetTime,
			"target_keywords":      result.Proposal.TargetKeywords,
		},
		"candidates": buildCandidateResponse(candidates),
	})
}

var errNeedEventID = errors.New("need event id")

// findCandidateEvents 按 ID/时间/关键词匹配候选日程。
func findCandidateEvents(userID uint, proposal ai.Proposal) ([]model.Event, error) {
	query := model.DB.Model(&model.Event{}).Where("user_id = ?", userID)
	if proposal.EventID != nil {
		var event model.Event
		if err := query.Where("id = ?", *proposal.EventID).First(&event).Error; err != nil {
			return nil, err
		}
		return []model.Event{event}, nil
	}
	if proposal.TargetTime != nil {
		start := time.Date(proposal.TargetTime.Year(), proposal.TargetTime.Month(), proposal.TargetTime.Day(), 0, 0, 0, 0, proposal.TargetTime.Location())
		end := start.Add(24 * time.Hour)
		query = query.Where("start_time < ? AND end_time >= ?", end, start)
	}
	if len(proposal.TargetKeywords) > 0 {
		for _, keyword := range proposal.TargetKeywords {
			value := "%" + keyword + "%"
			query = query.Where("(title LIKE ? OR description LIKE ?)", value, value)
		}
	}
	var events []model.Event
	if err := query.Order("start_time desc").Limit(5).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// buildCandidateResponse 输出前端用于选择的候选列表。
func buildCandidateResponse(list []model.Event) []gin.H {
	if len(list) == 0 {
		return nil
	}
	result := make([]gin.H, 0, len(list))
	for _, event := range list {
		result = append(result, gin.H{
			"id":         event.ID,
			"title":      event.Title,
			"start_time": event.StartTime,
			"end_time":   event.EndTime,
			"location":   event.Location,
		})
	}
	return result
}

// selectSingleEvent 根据规则选中唯一日程。
func selectSingleEvent(user model.User, proposal ai.Proposal) (model.Event, error) {
	candidates, err := findCandidateEvents(user.ID, proposal)
	if err != nil {
		return model.Event{}, err
	}
	if len(candidates) == 0 {
		return model.Event{}, gorm.ErrRecordNotFound
	}
	if len(candidates) > 1 && proposal.EventID == nil {
		return model.Event{}, errNeedEventID
	}
	return candidates[0], nil
}

// createEventFromProposal 创建日程并写入操作日志与通知。
func createEventFromProposal(user model.User, proposal ai.Proposal) (model.Event, error) {
	if proposal.StartTime == nil || proposal.EndTime == nil {
		return model.Event{}, errors.New("invalid time")
	}
	event := model.Event{
		UserID:      user.ID,
		Title:       proposal.Title,
		Type:        proposal.Type,
		StartTime:   *proposal.StartTime,
		EndTime:     *proposal.EndTime,
		Location:    proposal.Location,
		Description: proposal.Description,
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		if len(proposal.ParticipantIDs) > 0 {
			participants := make([]model.EventParticipant, 0, len(proposal.ParticipantIDs))
			for _, userID := range proposal.ParticipantIDs {
				if userID == user.ID {
					continue
				}
				participants = append(participants, model.EventParticipant{
					EventID: event.ID,
					UserID:  userID,
				})
			}
			if len(participants) > 0 {
				if err := tx.Create(&participants).Error; err != nil {
					return err
				}
			}
		}
		if err := service.CreateOperationLog(tx, user.ID, "create", event.Title, map[string]interface{}{
			"title": event.Title,
			"type":  event.Type,
			"time":  time.Now(),
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.Event{}, err
	}

	if err := model.DB.Preload("Creator").Preload("Participants.User").First(&event, event.ID).Error; err != nil {
		return model.Event{}, err
	}
	participantIDs := collectParticipantIDs(event.Participants)
	_ = service.CreateInvitationNotifications(event, participantIDs)
	return event, nil
}

// updateEventFromProposal 更新日程并写入操作日志与通知。
func updateEventFromProposal(user model.User, proposal ai.Proposal) (model.Event, error) {
	event, err := selectSingleEvent(user, proposal)
	if err != nil {
		return model.Event{}, err
	}
	if event.UserID != user.ID {
		return model.Event{}, errors.New("no permission")
	}

	if err := model.DB.Preload("Participants").First(&event, event.ID).Error; err != nil {
		return model.Event{}, err
	}

	before := eventSnapshot(event)
	updates := map[string]interface{}{}

	if proposal.Title != "" {
		updates["title"] = proposal.Title
	}
	if proposal.Type != "" {
		if !isValidEventType(proposal.Type) {
			return model.Event{}, errors.New("invalid type")
		}
		updates["type"] = proposal.Type
	}
	if proposal.Location != "" {
		updates["location"] = proposal.Location
	}
	if proposal.Description != "" {
		updates["description"] = proposal.Description
	}

	startTime := event.StartTime
	endTime := event.EndTime
	if proposal.StartTime != nil {
		startTime = *proposal.StartTime
		updates["start_time"] = *proposal.StartTime
	}
	if proposal.EndTime != nil {
		endTime = *proposal.EndTime
		updates["end_time"] = *proposal.EndTime
	}
	if !endTime.After(startTime) {
		return model.Event{}, errors.New("invalid time range")
	}

	if len(updates) == 0 && proposal.ParticipantKeywords == nil {
		return event, nil
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&model.Event{}).Where("id = ?", event.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		if proposal.ParticipantKeywords != nil {
			if err := tx.Where("event_id = ?", event.ID).Delete(&model.EventParticipant{}).Error; err != nil {
				return err
			}
			participants := make([]model.EventParticipant, 0, len(proposal.ParticipantIDs))
			for _, userID := range proposal.ParticipantIDs {
				if userID == user.ID {
					continue
				}
				participants = append(participants, model.EventParticipant{
					EventID: event.ID,
					UserID:  userID,
				})
			}
			if len(participants) > 0 {
				if err := tx.Create(&participants).Error; err != nil {
					return err
				}
			}
		}
		if err := service.CreateOperationLog(tx, user.ID, "update", event.Title, map[string]interface{}{
			"before": before,
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.Event{}, err
	}

	if err := model.DB.Preload("Creator").Preload("Participants.User").First(&event, event.ID).Error; err != nil {
		return model.Event{}, err
	}
	after := eventSnapshot(event)
	_ = service.CreateOperationLog(nil, user.ID, "update", event.Title, map[string]interface{}{
		"before": before,
		"after":  after,
	})
	participantIDs := collectParticipantIDs(event.Participants)
	_ = service.CreateChangeNotifications(event, participantIDs)
	return event, nil
}

// deleteEventFromProposal 删除日程并写入操作日志与通知。
func deleteEventFromProposal(user model.User, proposal ai.Proposal) (model.Event, error) {
	event, err := selectSingleEvent(user, proposal)
	if err != nil {
		return model.Event{}, err
	}
	if event.UserID != user.ID {
		return model.Event{}, errors.New("no permission")
	}

	if err := model.DB.Preload("Participants").First(&event, event.ID).Error; err != nil {
		return model.Event{}, err
	}

	participantIDs := collectParticipantIDs(event.Participants)
	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", event.ID).Delete(&model.EventParticipant{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.Event{}, event.ID).Error; err != nil {
			return err
		}
		if err := service.CreateOperationLog(tx, user.ID, "delete", event.Title, map[string]interface{}{
			"title": event.Title,
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.Event{}, err
	}

	_ = service.CreateChangeNotifications(event, participantIDs)
	return event, nil
}
