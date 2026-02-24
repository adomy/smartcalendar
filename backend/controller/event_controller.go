package controller

import (
	"time"

	"smartcalendar/model"
	"smartcalendar/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// EventController 负责日程相关接口。
type EventController struct{}

// EventCreateRequest 表示创建日程的请求体。
type EventCreateRequest struct {
	Title          string `json:"title" binding:"required,min=1,max=100"`
	Type           string `json:"type" binding:"required"`
	StartTime      string `json:"start_time" binding:"required"`
	EndTime        string `json:"end_time" binding:"required"`
	ParticipantIDs []uint `json:"participant_ids"`
	Location       string `json:"location" binding:"omitempty,max=200"`
	Description    string `json:"description" binding:"omitempty,max=500"`
}

// EventUpdateRequest 表示更新日程的请求体。
type EventUpdateRequest struct {
	Title          *string `json:"title" binding:"omitempty,min=1,max=100"`
	Type           *string `json:"type" binding:"omitempty"`
	StartTime      *string `json:"start_time"`
	EndTime        *string `json:"end_time"`
	ParticipantIDs *[]uint `json:"participant_ids"`
	Location       *string `json:"location" binding:"omitempty,max=200"`
	Description    *string `json:"description" binding:"omitempty,max=500"`
}

// CreateEvent 创建日程并写入参与人、操作日志与通知。
func (e EventController) CreateEvent(c *gin.Context) {
	var req EventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败："+err.Error())
		return
	}
	if !isValidEventType(req.Type) {
		Error(c, 40001, "参数校验失败：type 无效")
		return
	}
	startTime, err := parseRFC3339(req.StartTime)
	if err != nil {
		Error(c, 40001, "参数校验失败：start_time 无效")
		return
	}
	endTime, err := parseRFC3339(req.EndTime)
	if err != nil {
		Error(c, 40001, "参数校验失败：end_time 无效")
		return
	}
	if !endTime.After(startTime) {
		Error(c, 40001, "参数校验失败：end_time 必须晚于 start_time")
		return
	}

	user := c.MustGet("user").(model.User)
	event := model.Event{
		UserID:      user.ID,
		Title:       req.Title,
		Type:        req.Type,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    req.Location,
		Description: req.Description,
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		if len(req.ParticipantIDs) > 0 {
			participants := make([]model.EventParticipant, 0, len(req.ParticipantIDs))
			for _, userID := range uniqueUintList(req.ParticipantIDs) {
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
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	if err := model.DB.Preload("Creator").Preload("Participants.User").First(&event, event.ID).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	participantIDs := collectParticipantIDs(event.Participants)
	_ = service.CreateInvitationNotifications(event, participantIDs)

	Success(c, buildEventResponse(event, user.ID))
}

// ListEvents 列出用户创建或参与的日程。
func (e EventController) ListEvents(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	eventType := c.Query("type")
	startQuery := c.Query("start")
	endQuery := c.Query("end")

	query := model.DB.Table("events").
		Select("events.id").
		Joins("LEFT JOIN event_participants ON event_participants.event_id = events.id").
		Where("events.user_id = ? OR event_participants.user_id = ?", user.ID, user.ID).
		Distinct()

	if eventType != "" {
		query = query.Where("events.type = ?", eventType)
	}
	if startQuery != "" {
		if startTime, err := parseRFC3339(startQuery); err == nil {
			query = query.Where("events.end_time >= ?", startTime)
		}
	}
	if endQuery != "" {
		if endTime, err := parseRFC3339(endQuery); err == nil {
			query = query.Where("events.start_time <= ?", endTime)
		}
	}

	var eventIDs []uint
	if err := query.Scan(&eventIDs).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	if len(eventIDs) == 0 {
		Success(c, gin.H{"list": []interface{}{}})
		return
	}
	var events []model.Event
	if err := model.DB.Where("id IN ?", eventIDs).
		Preload("Creator").
		Preload("Participants.User").
		Order("start_time asc").
		Find(&events).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	list := make([]interface{}, 0, len(events))
	for _, event := range events {
		list = append(list, buildEventResponse(event, user.ID))
	}
	Success(c, gin.H{"list": list})
}

// GetEventDetail 返回日程详情并校验访问权限。
func (e EventController) GetEventDetail(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	var event model.Event
	if err := model.DB.Preload("Creator").Preload("Participants.User").First(&event, c.Param("id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, 40401, "资源不存在")
			return
		}
		Error(c, 50000, "服务器内部错误")
		return
	}
	if event.UserID != user.ID && !hasParticipant(event.Participants, user.ID) {
		Error(c, 40301, "无权限")
		return
	}
	Success(c, buildEventResponse(event, user.ID))
}

// UpdateEvent 更新日程并同步参与人、操作日志与通知。
func (e EventController) UpdateEvent(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	var event model.Event
	if err := model.DB.Preload("Participants").First(&event, c.Param("id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, 40401, "资源不存在")
			return
		}
		Error(c, 50000, "服务器内部错误")
		return
	}
	if event.UserID != user.ID {
		Error(c, 40301, "无权限")
		return
	}

	var req EventUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败："+err.Error())
		return
	}

	beforeSnapshot := eventSnapshot(event)
	updatedFields := map[string]interface{}{}

	if req.Title != nil {
		updatedFields["title"] = *req.Title
	}
	if req.Type != nil {
		if !isValidEventType(*req.Type) {
			Error(c, 40001, "参数校验失败：type 无效")
			return
		}
		updatedFields["type"] = *req.Type
	}
	if req.Location != nil {
		updatedFields["location"] = *req.Location
	}
	if req.Description != nil {
		updatedFields["description"] = *req.Description
	}

	startTime := event.StartTime
	endTime := event.EndTime
	if req.StartTime != nil {
		parsed, err := parseRFC3339(*req.StartTime)
		if err != nil {
			Error(c, 40001, "参数校验失败：start_time 无效")
			return
		}
		startTime = parsed
		updatedFields["start_time"] = parsed
	}
	if req.EndTime != nil {
		parsed, err := parseRFC3339(*req.EndTime)
		if err != nil {
			Error(c, 40001, "参数校验失败：end_time 无效")
			return
		}
		endTime = parsed
		updatedFields["end_time"] = parsed
	}
	if !endTime.After(startTime) {
		Error(c, 40001, "参数校验失败：end_time 必须晚于 start_time")
		return
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if len(updatedFields) > 0 {
			if err := tx.Model(&model.Event{}).Where("id = ?", event.ID).Updates(updatedFields).Error; err != nil {
				return err
			}
		}
		if req.ParticipantIDs != nil {
			if err := tx.Where("event_id = ?", event.ID).Delete(&model.EventParticipant{}).Error; err != nil {
				return err
			}
			participants := make([]model.EventParticipant, 0, len(*req.ParticipantIDs))
			for _, userID := range uniqueUintList(*req.ParticipantIDs) {
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
		return nil
	}); err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}

	if err := model.DB.Preload("Creator").Preload("Participants.User").First(&event, event.ID).Error; err != nil {
		Error(c, 50000, "服务器内部错误")
		return
	}
	afterSnapshot := eventSnapshot(event)
	_ = service.CreateOperationLog(nil, user.ID, "update", event.Title, map[string]interface{}{
		"before": beforeSnapshot,
		"after":  afterSnapshot,
	})
	participantIDs := collectParticipantIDs(event.Participants)
	_ = service.CreateChangeNotifications(event, participantIDs)

	Success(c, buildEventResponse(event, user.ID))
}

// DeleteEvent 删除日程并生成操作日志与通知。
func (e EventController) DeleteEvent(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	var event model.Event
	if err := model.DB.Preload("Participants").First(&event, c.Param("id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, 40401, "资源不存在")
			return
		}
		Error(c, 50000, "服务器内部错误")
		return
	}
	if event.UserID != user.ID {
		Error(c, 40301, "无权限")
		return
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
		Error(c, 50000, "服务器内部错误")
		return
	}

	_ = service.CreateChangeNotifications(event, participantIDs)
	Success(c, gin.H{"deleted": true})
}

// parseRFC3339 解析 RFC3339 时间字符串。
func parseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

// isValidEventType 校验日程类型枚举。
func isValidEventType(value string) bool {
	return value == "work" || value == "life" || value == "growth"
}

// hasParticipant 判断用户是否为参与人。
func hasParticipant(list []model.EventParticipant, userID uint) bool {
	for _, participant := range list {
		if participant.UserID == userID {
			return true
		}
	}
	return false
}

// collectParticipantIDs 提取参与人 ID 并去重。
func collectParticipantIDs(list []model.EventParticipant) []uint {
	ids := make([]uint, 0, len(list))
	for _, participant := range list {
		ids = append(ids, participant.UserID)
	}
	return uniqueUintList(ids)
}

// buildEventResponse 构建带权限标识的日程响应。
func buildEventResponse(event model.Event, viewerID uint) gin.H {
	isCreator := event.UserID == viewerID
	isCollaboration := !isCreator && hasParticipant(event.Participants, viewerID)
	return gin.H{
		"id":               event.ID,
		"user_id":          event.UserID,
		"title":            event.Title,
		"type":             event.Type,
		"start_time":       event.StartTime,
		"end_time":         event.EndTime,
		"location":         event.Location,
		"description":      event.Description,
		"created_at":       event.CreatedAt,
		"updated_at":       event.UpdatedAt,
		"is_creator":       isCreator,
		"is_collaboration": isCollaboration,
		"creator":          event.Creator,
		"participants":     event.Participants,
	}
}

// eventSnapshot 输出日程快照用于操作日志。
func eventSnapshot(event model.Event) map[string]interface{} {
	return map[string]interface{}{
		"title":           event.Title,
		"type":            event.Type,
		"start_time":      event.StartTime,
		"end_time":        event.EndTime,
		"location":        event.Location,
		"description":     event.Description,
		"participant_ids": collectParticipantIDs(event.Participants),
	}
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
