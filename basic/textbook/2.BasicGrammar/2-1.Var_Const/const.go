// 상수는 const 예약어를 이용하여 변수와 동일하게 선언할 수 있음
package main

import "fmt"

const i int = 1
const s string = "constant"

// itoa가 0이 할당되고, 순서대로 증가
// A: 0, B: 1, C: 2, D: 3
const (
	A = iota
	B
	C
	D
)

func main() {
	fmt.Println(i)
	fmt.Println(s)

	fmt.Println(A, B, C, D)
}
