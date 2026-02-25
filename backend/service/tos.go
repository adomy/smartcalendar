package service

import (
	"context"
	"errors"
	"io"
	"net/url"
	"path"
	"strings"

	"smartcalendar/config"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func UploadToTOS(ctx context.Context, cfg config.AppConfig, objectKey string, reader io.Reader) (string, error) {
	if cfg.TOSEndpoint == "" || cfg.TOSRegion == "" || cfg.TOSBucket == "" || cfg.TOSAccessKey == "" || cfg.TOSSecretKey == "" {
		return "", errors.New("TOS 配置缺失")
	}

	client, err := tos.NewClientV2(cfg.TOSEndpoint, tos.WithRegion(cfg.TOSRegion), tos.WithCredentials(tos.NewStaticCredentials(cfg.TOSAccessKey, cfg.TOSSecretKey)))
	if err != nil {
		return "", err
	}
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: cfg.TOSBucket,
			Key:    objectKey,
		},
		Content: reader,
	})
	if err != nil {
		return "", err
	}
	return buildTOSPublicURL(cfg, objectKey)
}

func buildTOSPublicURL(cfg config.AppConfig, objectKey string) (string, error) {
	if cfg.TOSPublicBaseURL != "" {
		return strings.TrimRight(cfg.TOSPublicBaseURL, "/") + "/" + objectKey, nil
	}
	parsed, err := url.Parse(cfg.TOSEndpoint)
	if err != nil || parsed.Host == "" {
		return "", errors.New("TOS endpoint 无效")
	}
	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "https"
	}
	host := cfg.TOSBucket + "." + parsed.Host
	return scheme + "://" + host + path.Join("/", objectKey), nil
}
