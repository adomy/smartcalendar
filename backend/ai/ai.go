// Package ai 实现基于大模型的意图识别与解析逻辑。
package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"smartcalendar/config"
	"smartcalendar/model"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
)

// Proposal 表示经过模型解析后的结构化日程意图。
type Proposal struct {
	Action              string     `json:"action"`
	Title               string     `json:"title"`
	Type                string     `json:"type"`
	StartTime           *time.Time `json:"start_time"`
	EndTime             *time.Time `json:"end_time"`
	Location            string     `json:"location"`
	Description         string     `json:"description"`
	ParticipantKeywords []string   `json:"participant_keywords"`
	ParticipantIDs      []uint     `json:"participant_ids"`
	EventID             *uint64    `json:"event_id"`
	TargetTime          *time.Time `json:"target_time"`
	TargetKeywords      []string   `json:"target_keywords"`
}

// ParseResult 表示模型解析后返回给业务层的结果。
type ParseResult struct {
	Intent      string
	Proposal    Proposal
	NeedConfirm bool
	Result      string
}

// AIService 负责调用 Ark 模型进行意图识别与结构化解析。
type AIService struct {
	cfg       config.AppConfig
	model     *ark.ChatModel
	modelOnce sync.Once
	modelErr  error
	pending   sync.Map
}

// NewAIService 初始化 AI 服务并保留模型配置。
func NewAIService(cfg config.AppConfig) *AIService {
	return &AIService{cfg: cfg}
}

// ParseMessage 将用户输入发送给大模型并解析意图与字段。
func (a *AIService) ParseMessage(message string) (ParseResult, error) {
	ctx := context.Background()
	chatModel, err := a.getModel(ctx)
	if err != nil {
		return ParseResult{}, err
	}

	now := time.Now().Format(time.RFC3339)
	systemPrompt := buildSystemPrompt()
	userPrompt := fmt.Sprintf("当前时间：%s\n用户输入：%s", now, message)

	resp, err := chatModel.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: userPrompt},
	})
	if err != nil {
		return ParseResult{}, err
	}

	intent, err := parseIntent(resp.Content)
	if err != nil {
		return ParseResult{Intent: "unknown", NeedConfirm: false, Result: "意图解析失败，请换一种表达"}, nil
	}

	proposal := buildProposal(intent)
	result := formatResult(proposal)
	return result, nil
}

// StoreProposal 缓存待确认的 Proposal，返回 confirm_id。
func (a *AIService) StoreProposal(proposal Proposal) string {
	confirmID := fmt.Sprintf("c_%d", time.Now().UnixNano())
	a.pending.Store(confirmID, proposal)
	return confirmID
}

// ConsumeProposal 取出并删除缓存的 Proposal。
func (a *AIService) ConsumeProposal(confirmID string) (Proposal, bool) {
	value, ok := a.pending.Load(confirmID)
	if !ok {
		return Proposal{}, false
	}
	a.pending.Delete(confirmID)
	return value.(Proposal), true
}

// getModel 构建或复用 Ark ChatModel。
func (a *AIService) getModel(ctx context.Context) (*ark.ChatModel, error) {
	a.modelOnce.Do(func() {
		if a.cfg.ArkModelID == "" {
			a.modelErr = errors.New("ARK_MODEL_ID 未配置")
			return
		}
		if a.cfg.ArkAPIKey == "" && (a.cfg.ArkAccessKey == "" || a.cfg.ArkSecretKey == "") {
			a.modelErr = errors.New("ARK_API_KEY 或 ARK_ACCESS_KEY/ARK_SECRET_KEY 未配置")
			return
		}
		a.model, a.modelErr = ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey:    a.cfg.ArkAPIKey,
			AccessKey: a.cfg.ArkAccessKey,
			SecretKey: a.cfg.ArkSecretKey,
			Model:     a.cfg.ArkModelID,
			BaseURL:   a.cfg.ArkBaseURL,
			Region:    a.cfg.ArkRegion,
		})
	})
	return a.model, a.modelErr
}

// intentPayload 为大模型输出的 JSON 结构。
type intentPayload struct {
	Action              string   `json:"action"`
	Title               string   `json:"title"`
	Type                string   `json:"type"`
	StartTime           string   `json:"start_time"`
	EndTime             string   `json:"end_time"`
	Location            string   `json:"location"`
	Description         string   `json:"description"`
	ParticipantKeywords []string `json:"participant_keywords"`
	EventID             string   `json:"event_id"`
	TargetTime          string   `json:"target_time"`
	TargetKeywords      []string `json:"target_keywords"`
}

// buildSystemPrompt 约束大模型输出为 JSON。
func buildSystemPrompt() string {
	return strings.TrimSpace(`你是智能日程助手，请根据用户输入识别意图并输出JSON。
如果没有结束时间，默认开始时间后1小时；如果没有开始时间，默认结束时间前1小时。
只输出JSON，不要输出任何解释或markdown。
JSON字段:
action: "create" | "update" | "delete" | "unknown"
title: 日程标题
type: "work" | "life" | "growth"
start_time: RFC3339
end_time: RFC3339
location: 地点
description: 描述
participant_keywords: 参与人关键词数组
event_id: 如果用户指定了具体ID则填写
target_time: 需要修改/删除的原日程时间(RFC3339)
target_keywords: 用于匹配原日程的关键词数组
如果缺失信息，请留空字符串或空数组。`)
}

// parseIntent 解析模型输出的 JSON 并映射为 intentPayload。
func parseIntent(content string) (intentPayload, error) {
	raw := extractJSON(content)
	if raw == "" {
		return intentPayload{}, errors.New("empty")
	}
	var payload intentPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return intentPayload{}, err
	}
	payload.Action = strings.ToLower(strings.TrimSpace(payload.Action))
	return payload, nil
}

// buildProposal 将 intentPayload 转换为 Proposal。
func buildProposal(payload intentPayload) Proposal {
	startTime := parseTime(payload.StartTime)
	endTime := parseTime(payload.EndTime)
	targetTime := parseTime(payload.TargetTime)
	participantKeywords := normalizeKeywords(payload.ParticipantKeywords)
	targetKeywords := normalizeKeywords(payload.TargetKeywords)
	var eventID *uint64
	if payload.EventID == "" {
		eventID = nil
	} else {
		eventIDVal, _ := strconv.ParseUint(payload.EventID, 10, 64)
		eventID = &eventIDVal
	}
	proposal := Proposal{
		Action:              payload.Action,
		Title:               strings.TrimSpace(payload.Title),
		Type:                strings.TrimSpace(payload.Type),
		StartTime:           startTime,
		EndTime:             endTime,
		Location:            strings.TrimSpace(payload.Location),
		Description:         strings.TrimSpace(payload.Description),
		ParticipantKeywords: participantKeywords,
		ParticipantIDs:      resolveParticipants(participantKeywords),
		EventID:             eventID,
		TargetTime:          targetTime,
		TargetKeywords:      targetKeywords,
	}
	return proposal
}

// formatResult 将 Proposal 转换为前端可展示的确认提示。
func formatResult(proposal Proposal) ParseResult {
	switch proposal.Action {
	case "create":
		if proposal.Title == "" || proposal.StartTime == nil || proposal.EndTime == nil {
			return ParseResult{Intent: "create", NeedConfirm: false, Result: "创建日程需要标题和时间，请补充"}
		}
		text := fmt.Sprintf("识别到创建日程：%s %s-%s", proposal.Title, proposal.StartTime.Format("2006-01-02 15:04"), proposal.EndTime.Format("15:04"))
		if proposal.Location != "" {
			text += "（地点：" + proposal.Location + "）"
		}
		return ParseResult{Intent: "create", NeedConfirm: true, Proposal: proposal, Result: text + "。是否确认创建？"}
	case "update":
		if proposal.EventID == nil && proposal.TargetTime == nil && len(proposal.TargetKeywords) == 0 {
			return ParseResult{Intent: "update", NeedConfirm: false, Result: "修改日程需要提供原日程时间或关键词"}
		}
		text := "识别到修改日程请求"
		if proposal.Title != "" {
			text += "，标题改为：" + proposal.Title
		}
		if proposal.StartTime != nil {
			text += "，开始时间：" + proposal.StartTime.Format(time.RFC3339)
		}
		if proposal.EndTime != nil {
			text += "，结束时间：" + proposal.EndTime.Format(time.RFC3339)
		}
		return ParseResult{Intent: "update", NeedConfirm: true, Proposal: proposal, Result: text + "。是否确认修改？"}
	case "delete":
		if proposal.EventID == nil && proposal.TargetTime == nil && len(proposal.TargetKeywords) == 0 {
			return ParseResult{Intent: "delete", NeedConfirm: false, Result: "删除日程需要提供原日程时间或关键词"}
		}
		return ParseResult{Intent: "delete", NeedConfirm: true, Proposal: proposal, Result: "识别到删除日程请求。是否确认删除？"}
	default:
		return ParseResult{Intent: "unknown", NeedConfirm: false, Result: "暂仅支持创建、修改、删除日程"}
	}
}

func extractJSON(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimPrefix(trimmed, "json")
		trimmed = strings.TrimSpace(trimmed)
		trimmed = strings.TrimSuffix(trimmed, "```")
	}
	return strings.TrimSpace(trimmed)
}

func parseTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}

func normalizeKeywords(input []string) []string {
	var output []string
	for _, item := range input {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		output = append(output, trimmed)
	}
	if len(output) == 0 {
		return nil
	}
	return output
}

func resolveParticipants(keywords []string) []uint {
	if len(keywords) == 0 {
		return nil
	}
	var ids []uint
	for _, keyword := range keywords {
		var users []model.User
		if err := model.DB.Where("nickname LIKE ?", "%"+keyword+"%").Limit(3).Find(&users).Error; err != nil {
			continue
		}
		for _, user := range users {
			ids = append(ids, user.ID)
		}
	}
	return uniqueUintList(ids)
}

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
