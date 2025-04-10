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
		return fmt.Sprintf("%dm%ds", minutes, seconds)
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

func GetTimeElapse(givenTimeStr string) string {
	// 定义时间格式
	timeFormat := "2006-01-02T15:04:05+08:00"

	// 解析给定的时间
	givenTime, err := time.ParseInLocation(timeFormat, givenTimeStr, time.Local)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return ""
	}

	// 获取当前时间
	currentTime := time.Now().Local()
	fmt.Printf("%v | %v\n", currentTime, givenTime)
	// 计算两个时间之间的差
	diff := currentTime.Sub(givenTime)

	// 根据差值输出结果
	if diff.Hours() > 1 {
		return fmt.Sprintf("%d小时以前", int(diff.Hours()))
	} else {
		return fmt.Sprintf("%d分钟以前", int(diff.Minutes()))
	}
}
