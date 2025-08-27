package main

/*
기본적으로 채널은 버퍼를 사용하지 않고 생성됩니다.
버퍼를 이용하지 않을 경우에 채널에 데이터를 바로 보내지 ㅇ낳는 경우 오류가 발생합니다.

make() 함수에 인자를 추가하여 버퍼를 생성할 수 있습니다.
버퍼를 이용할 때는 채널에 데이터를 수신 후 다른 고루틴에서 데이터를 수신하지 ㅇ낳아도 다른 작업을 수행할 수 있습니다.
*/

import "fmt"

func main() {
	c := make(chan int)
	c <- 1           // 수신 루틴이 없으므로 데드락 발생
	fmt.Println(<-c) // comment해도 데드락 (별도의 고루틴이 없이 때문)

	ch := make(chan int, 1) // buffer size 1
	// 수신자가 없더라도 보낼 수 있다.
	ch <- 1
	fmt.Println(<-ch)
}
