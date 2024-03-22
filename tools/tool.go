package tools

import (
	"fmt"
	"time"
	"unicode"
)

func CurrTime() string {
	return time.Now().Format(time.TimeOnly)
}

func CurrDateTime() string {
	return time.Now().Format(time.DateTime)
}

func FormatDuration(elapsed int64) string {
	days := elapsed / (24 * 60 * 60)
	hours := (elapsed / (60 * 60)) % 24
	minutes := (elapsed / 60) % 60
	seconds := elapsed % 60
	if days == 0 && hours != 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	if days == 0 && hours == 0 && minutes != 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	if days == 0 && hours == 0 && minutes == 0 && seconds != 0 {
		return fmt.Sprintf("%ds", seconds)
	}

	return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Percent[T any](m map[string]map[string]T, key string) string {
	total := 0
	cnt := 0
	for k, v := range m {
		total += len(v)
		if k == key {
			cnt = len(v)
		}
	}
	return fmt.Sprintf("%d/%d", cnt, total)
}

func Contains[T comparable](s []T, v T) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func RemoveUnprintable(s string) string {
	var result []rune
	for _, r := range s {
		if unicode.IsPrint(r) {
			result = append(result, r)
		}
	}
	return string(result)
}
