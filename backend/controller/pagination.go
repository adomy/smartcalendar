package controller

import "strconv"

// parsePage 解析分页页码。
func parsePage(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	page, err := strconv.Atoi(value)
	if err != nil || page <= 0 {
		return defaultValue
	}
	return page
}

// parsePageSize 解析分页大小并限制最大值。
func parsePageSize(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	size, err := strconv.Atoi(value)
	if err != nil || size <= 0 || size > 100 {
		return defaultValue
	}
	return size
}
