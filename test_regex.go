package main

import (
	"fmt"
	"regexp"
)

func main() {
	// 更宽松的正则表达式测试
	testCases := []string{
		`.*\[.*(\d+).*\].*\b(\S+).*\b(ALLOW|DENY)\b.*`,  // 非常宽松的匹配
	}

	// 模拟 ufw status numbered 输出的一行
	testLine := "[ 1] 22/tcp                     ALLOW IN    Anywhere                  "

	for i, regexStr := range testCases {
		regex := regexp.MustCompile(regexStr)
		matches := regex.FindStringSubmatch(testLine)
		fmt.Printf("测试正则[%d]: %s\n", i, regexStr)
		fmt.Printf("  测试行: '%s'\n", testLine)
		if len(matches) > 0 {
			fmt.Printf("  匹配成功，分组: %v\n", matches)
		} else {
			fmt.Printf("  匹配失败\n")
		}
		fmt.Println()
	}

	// 直接匹配行中的数字和端口
	numRegex := regexp.MustCompile(`\[(\s*\d+\s*)\]`)
	numMatches := numRegex.FindStringSubmatch(testLine)
	fmt.Printf("单独匹配编号: %v\n", numMatches)

	portRegex := regexp.MustCompile(`\b(\S+)/tcp`)
	portMatches := portRegex.FindStringSubmatch(testLine)
	fmt.Printf("单独匹配端口: %v\n", portMatches)

	actionRegex := regexp.MustCompile(`\b(ALLOW|DENY)\b`)
	actionMatches := actionRegex.FindStringSubmatch(testLine)
	fmt.Printf("单独匹配动作: %v\n", actionMatches)
}
