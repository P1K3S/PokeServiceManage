package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	output := `Status: active

     To                         Action      From
     --                         ------      ----
[ 1] 22/tcp                     ALLOW IN    Anywhere                  
[ 2] 62500:62501/tcp            ALLOW IN    Anywhere                  
[ 3] 9301/tcp                   ALLOW IN    Anywhere                  
[ 4] 9602/tcp                   ALLOW IN    Anywhere                  
[ 5] 9102/tcp                   ALLOW IN    Anywhere                  
[ 6] 9297/tcp                   ALLOW IN    Anywhere                  
[ 7] 9601/tcp                   ALLOW IN    Anywhere                  
[ 8] 9202/tcp                   ALLOW IN    Anywhere                  
[ 9] 9702/tcp                   ALLOW IN    Anywhere                  
[10] 9711/tcp                   ALLOW IN    Anywhere                  
[11] 9712/tcp                   ALLOW IN    Anywhere                  
[12] 9295/tcp                   ALLOW IN    Anywhere                  
[13] 9294/tcp                   ALLOW IN    Anywhere                  
[14] 9201/tcp                   DENY IN     Anywhere                  
[15] 9299/tcp                   DENY IN     Anywhere                  
[16] 9302/tcp                   DENY IN     Anywhere                  
[17] 9703/tcp                   ALLOW IN    Anywhere                  
[18] 9232/tcp                   ALLOW IN    Anywhere                  
[19] 9999/tcp                   ALLOW IN    Anywhere                  
`

	regex := regexp.MustCompile(`\[\s*(\d+)\s*\]\s+(\S+)\s+(ALLOW|DENY)`)

	lines := strings.Split(output, "\n")
	matched := 0
	for i, line := range lines {
		matches := regex.FindStringSubmatch(line)
		if len(matches) == 4 {
			fmt.Printf("行[%d]: num=%s, portSpec=%s, action=%s\n", i, matches[1], matches[2], matches[3])
			matched++
		}
	}
	fmt.Printf("\n共匹配 %d 条规则\n", matched)
}
