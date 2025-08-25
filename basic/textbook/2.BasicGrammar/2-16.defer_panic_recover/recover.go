/*
recover:
  - panic에 의해 중단된 프로그램 흐름을 정상적으로 복구할 때 사용하는 함수입니다.
  - recover는 반드시 defer 함수 안에서 호출되어야 하며, panic이 발생한 경우 panic 값을 반환하고, 그렇지 않으면 nil을 반환합니다.
  - recover를 사용하면 프로그램이 비정상적으로 종료되는 것을 방지할 수 있습니다.
*/
package main

import (
	"fmt"
	"os"
)

func main() {
	// 잘못된 파일명을 넣음
	openFile("Invalid.txt")

	// recover에 의해 프로그램 흐름이 정상적으로 복구됨
	// println("Done") 문장이 실행됨
}

func openFile(filename string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in openFile:", r)
		}
	}()

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 파일을 정상적으로 읽는 코드
}
