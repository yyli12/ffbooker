package booker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	inputTimeMaxTime = 5
)

func InputTime() (*time.Time, error) {
	for i := 0; i < inputTimeMaxTime; i++ {
		fmt.Println("Input time slot you want to book:")
		fmt.Printf("format: YYYY-MM-DD HH:SS, e.g. %s\n", time.Now().Format("2006-01-02 15:04"))

		reader := bufio.NewReader(os.Stdin)
		timeString, readErr := reader.ReadString('\n')
		if readErr != nil {
			continue
		}
		timeString = strings.TrimSpace(timeString) + ":00 +0800"
		classTime, parseErr := time.Parse("2006-01-02 15:04:05 -0700", timeString)
		if parseErr == nil {
			return &classTime, nil
		} else {
			fmt.Println(timeString)
			fmt.Println("Time format error", parseErr)
		}
	}
	return nil, fmt.Errorf("input time error")
}
