package main

/*
채널은 송신/수신이 모두 가능한 상태로 생성되지만, 송신/수신 하나만 가능한 채널을 생성할 수도 있습니다.

송신은 chan<- 으로, 수신은 <-chan 으로 표현합니다. 송신 채널에서 수신을 하거나, 수신 채널에서 송신을 하면 오류가 발생합니다.
*/

import "fmt"

func main() {
	ch := make(chan string, 1) // buffer size 1
	sendChan(ch)

	if err := recover(); err != nil {
		fmt.Println("Error occurred:", err)
	}

	receiveChan(ch)
	if err := recover(); err != nil {
		fmt.Println("Error occurred:", err)
	}
}

func sendChan(ch chan<- string) {
	ch <- "potato"
	// x := <-ch // receive operation not allowed, error occurs
}

func receiveChan(ch <-chan string) {
	data := <-ch
	fmt.Println(data)
}
