// 배열은 길이가 고정된 형태로 선언되고, 다차원 배열을 이용할 수 있습니다.

package main

import "fmt"

// 배열 선언
// var 변수명 변수타입
// var 배열명 = [길이]타입{초기값1, 초기값2, ...}
func main() {
	// 배열 선언
	var intArray [3]int
	intArray[0] = 1
	intArray[1] = 2
	intArray[2] = 3

	// 선언과 동시에 초기화
	var stringArrayWithInit = [3]string{"Hello", "World", "!"}
	for idx := range stringArrayWithInit {
		fmt.Printf("%s\n", stringArrayWithInit[idx])
	}

	// 멀티플 배열
	var intMultipleArray = [2][3]int{
		{1, 2, 3},
		{4, 5, 6},
	}

	for x := range intMultipleArray {
		for y := range intMultipleArray[x] {
			fmt.Println(intMultipleArray[x][y])
		}
	}
}
