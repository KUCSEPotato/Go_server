package main

import "fmt"

func main() {
	ch := make(chan int, 2)

	// 채널에 송신
	ch <- 1
	ch <- 2

	// 채널 닫기
	close(ch)

	// 채널에서 수신
	fmt.Println(<-ch)
	fmt.Println(<-ch)

	// 방법1
	// 채널이 닫힌 것을 감지할 때까지 계속 수신
	if _, success := <-ch; !success {
		fmt.Println("no more values.")
	}

	// 방법2
	// 위 표현과 동일한 채널 range 문
	for i := range ch {
		fmt.Println(i)
	}
}
