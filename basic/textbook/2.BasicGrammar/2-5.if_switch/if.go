/*
if 문의 기본 구조는 다른 언어와 동일합니다.

if condition {

} else if condition {

} else {

}
*/
package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix()) // rand seed 는 매번 다른 값을 주지 않으면 결과가 동일함
	i := rand.Intn(15)

	fmt.Println(i)

	if i < 10 {
		fmt.Println("Less than 10")
	} else if i < 20 {
		fmt.Println("Less than 20")
	} else {
		fmt.Println("20 or more")
	}

	// ------------ if condition ---------------
	str := "Hello World!"
	if err := ioutil.WriteFile("test.txt", []byte(str), 0644); err != nil {
		fmt.Println("Error:", err)
	}
}
