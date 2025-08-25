package main

import "fmt"

/*
fmt.Print()는 인수를 문자열로 출력한다.
fmt.Println()은 인수 사이에 공백을 넣고 마지막으로 개행 문자 \n 을 출력한다.
fmt.Printf()는 %d(숫자) 또는 %s(문자열)과 같은 형식을 지정하여 인수를 출력할 수 있다.
*/

func main() {
	for idx := 1; idx <= 5; idx++ {
		fmt.Printf("Iteration: %d\n", idx)
	}

	for idx := 1; idx <= 5; {
		fmt.Printf("Iteration: %d\n", idx)
		idx++
	}

	idx := 1
	for idx <= 5 {
		fmt.Printf("Iteration: %d\n", idx)
		idx++
	}

	idx = 1 // 앞서 초기 선언 된 변수의 경우 :=가 아닌 =로 값을 변경해야 한다.
	for true {
		fmt.Printf("Iteration: %d\n", idx)
		idx++
		if idx > 5 { // loop 탈출 조건
			break
		}
	}
}
