/*
golang의 switch 문은 break가 없어도 case 문 조건이 성립하는 시점에 종료됩니다.
또한 case 조건에 여러 개의 값을 동시에 등록 할 수 있습니다.
*/
package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix()) // rand seed 는 매번 다른 값을 주지 않으면 결과가 동일함
	i := rand.Intn(15)

	fmt.Println(i)

	switch {
	case 0, 1:
		fmt.Println("A")
	case 2, 3, 4:
		fmt.Println("B")
	case 5:
		fmt.Println("C")
	default:
		fmt.Println("D")
	}
}
