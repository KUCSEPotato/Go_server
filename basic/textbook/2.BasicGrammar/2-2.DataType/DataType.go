package main

import "fmt"

func main() {
	/* 부호가 있는 정수형에는 int, int8, int32, int64가 있음
	   부호가 없는 정수형에는 uint, uint8, uint32, uint64가 있음 */
	var i int = 1
	var s string = "string"

	fmt.Println(i)
	fmt.Println(s)

	// 실수형에는 float32, float64가 있음
	// 복소수도 가능: complex64, complex128
	var f float32 = 3.14
	var c complex64 = 1 + 2i

	fmt.Println(f)
	fmt.Println(c)

	// 문자열에는 string
	/* ""로 둘러 쌓인 상태로 선언
	복수 라인으로 선언할 수 없음
	특수 문자는 이스케이프 문자를 이용해서 처리
	''로 둘러쌓인 문장은 이스케이프 문자열을 해석하지 않고 처리
	아래의 str3은 출력시 개행하지 않음*/
	var str string = "string"
	var str2 string = "B\nB"
	var str3 string = `A\nA`

	fmt.Println(str)
	fmt.Println(str2)
	fmt.Println(str3)

	// 불린
	var tr bool = true
	var fal bool = false

	fmt.Println(tr)
	fmt.Println(fal)

	// 기타: byte, rune
	var b byte = 'a'
	var r rune = '한'

	fmt.Println(b)
	fmt.Println(r)
}
