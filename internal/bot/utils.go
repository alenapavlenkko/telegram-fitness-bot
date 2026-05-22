package bot

import (
	"strconv"
	"strings"
)

// ParseAdminIDs преобразует строку admin ids в []int64
func ParseAdminIDs(ids string) []int64 {

	var result []int64

	if ids == "" {
		return result
	}

	for _, s := range strings.Split(ids, ",") {

		id, err := strconv.ParseInt(
			strings.TrimSpace(s),
			10,
			64,
		)

		if err == nil {
			result = append(result, id)
		}
	}

	return result
}

// escapeMarkdown экранирует markdown символы
func escapeMarkdown(text string) string {

	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"`", "\\`",
	)

	return replacer.Replace(text)
}
