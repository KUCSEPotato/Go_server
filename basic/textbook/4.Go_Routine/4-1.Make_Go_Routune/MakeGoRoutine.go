// go 키워드를 이용하여 함수를 호출하면 해당 함수는 별도의 Go 루틴으로 실행된다.
package main

import (
	"fmt"
	"time"
)

func say(s string) {
	for i := 0; i < 10; i++ {
		fmt.Println(s, "***", i)
	}
}

func main() {
	// 함수를 동기적으로 실행
	say("Hello")

	// 함수를 비동기적으로 실행
	go say("Go Routine1")
	go say("Go Routine2")
	go say("Go Routine3")

	// 메인 함수가 종료되지 않도록 대기
	time.Sleep(3 * time.Second)
}
