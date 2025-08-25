package main

import (
	"fmt"
	"time"
)

func main() {
	timeStr := "2025-10-15 10:15:00"
	parseTime, err := time.Parse("2004-10-15 10:15:00", timeStr)

	if err != nil {
		fmt.Println("Error parsing time:", err)
	}

	fmt.Println(parseTime)
}
