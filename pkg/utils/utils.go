package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

func MsOrS(t, time string) (string, error) {
	num, err := strconv.ParseInt(time, 10, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return "", fmt.Errorf("time should be a number")
	}
	if t == "EX" {
		// isnum?
		return strconv.Itoa(int(num) * 1000), nil
	} else if t == "PX" {
		// millionsec to sec
		return strconv.Itoa(int(num)), nil
	}
	return "", fmt.Errorf("invalid SET command")
}

func IsNumeric(s string) bool {
	re := regexp.MustCompile(`^-?\d+$`) // 匹配可选的负号 + 数字
	return re.MatchString(s)
}

func Btoi(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
