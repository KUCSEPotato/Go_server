package main

import "fmt"

// var 변수명 변수타입 = 값
// 변수명 변수타입 := 값 // 변수의 타입 제한적
// 변수명 := 값
// var (v1 = 10
//       v2 = 20
//       v3 = 30
// )

func main() {
	// var 선언
	// Go는 변수 선언만 하고 사용하지 않으면 에러 발생
	var i1 int = 10
	var s1 string = "potato"

	fmt.Println(i1)
	fmt.Println(s1)

	// 타입 생략 가능
	var i2 = 10
	var s2 = "boiledpotato"

	fmt.Println(i2)
	fmt.Println(s2)

	// := 를 이용한 변수 선언
	i3 := 10
	s3 := "mashedpotato"

	fmt.Println(i3)
	fmt.Println(s3)

	// 다수의 변수를 동시에 선언
	var i4, i5, i6 int = 10, 20, 30
	s4, s5, s6 := "potato1", "potato2", "potato3"

	fmt.Println(i4, i5, i6)
	fmt.Println(s4, s5, s6)

	// var ()를 이용한 변수 선언
	var (
		i7         int    = 10
		i8         int    = 20
		i9         int    = 30
		s7, s8, s9 string = "potato1", "potato2", "potato3"
	)

	fmt.Println(i7, i8, i9)
	fmt.Println(s7, s8, s9)
}
