package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"smartcalendar/config"

	"github.com/google/uuid"
)

type speechQueryResult struct {
	Result struct {
		Text string `json:"text"`
	} `json:"result"`
}

func SubmitSpeechTask(cfg config.AppConfig, fileURL string, userID string) (string, error) {
	if cfg.SpeechApiKey == "" || cfg.SpeechResourceID == "" {
		return "", errors.New("语音识别配置缺失")
	}
	taskID := uuid.NewString()
	payload := map[string]any{
		"user": map[string]any{
			"uid": userID,
		},
		"audio": map[string]any{
			"url":      fileURL,
			"language": cfg.SpeechLanguage,
			"format":   cfg.SpeechFormat,
			"codec":    cfg.SpeechCodec,
			"rate":     cfg.SpeechRate,
			"bits":     cfg.SpeechBits,
			"channel":  cfg.SpeechChannel,
		},
		"request": map[string]any{
			"model_name":  cfg.SpeechModelName,
			"enable_itn":  true,
			"enable_punc": true,
		},
	}
	if cfg.SpeechModelVersion != "" {
		payload["request"].(map[string]any)["model_version"] = cfg.SpeechModelVersion
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(cfg.SpeechBaseURL, "/") + "/submit"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", cfg.SpeechApiKey)
	req.Header.Set("X-Api-Resource-Id", cfg.SpeechResourceID)
	req.Header.Set("X-Api-Request-Id", taskID)
	req.Header.Set("X-Api-Sequence", "-1")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	statusCode := resp.Header.Get("X-Api-Status-Code")
	if statusCode != "20000000" {
		message := resp.Header.Get("X-Api-Message")
		if message == "" {
			message = "语音识别提交失败"
		}
		return "", errors.New(message)
	}
	return taskID, nil
}

func QuerySpeechTask(cfg config.AppConfig, taskID string) (string, string, error) {
	if cfg.SpeechApiKey == "" || cfg.SpeechResourceID == "" {
		return "", "", errors.New("语音识别配置缺失")
	}
	body := []byte("{}")
	endpoint := strings.TrimRight(cfg.SpeechBaseURL, "/") + "/query"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", cfg.SpeechApiKey)
	req.Header.Set("X-Api-Resource-Id", cfg.SpeechResourceID)
	req.Header.Set("X-Api-Request-Id", taskID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	statusCode := resp.Header.Get("X-Api-Status-Code")
	if statusCode == "20000001" || statusCode == "20000002" {
		return "processing", "", nil
	}
	if statusCode != "20000000" {
		message := resp.Header.Get("X-Api-Message")
		if message == "" {
			message = "语音识别查询失败"
		}
		return "", "", errors.New(message)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	var result speechQueryResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", err
	}
	return "done", result.Result.Text, nil
}
