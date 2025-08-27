package main

// 채널은 데이터를 주고 받을 수 있는 통로입니다.
// make() 함수를 사용하여 채널을 생성하고, <- 연산자를 사용하여 데이터를 전송합니다.
// 주로 고루틴에서 함수들 사이에 데이터를 주고 받는데 이용됩니다.

import "fmt"

func main() {
	// create string channel
	ch := make(chan string)

	// send value to channel
	go func() {
		ch <- "potato"
	}()

	// receive value from channel
	value := <-ch
	fmt.Println("Received from channel:", value)
}
