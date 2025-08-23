package main

import "fmt"

func main() {
	x, y := 10, 3 // 1010, 0011

	// 산술
	fmt.Println(x + y) // 13
	fmt.Println(x - y) // 7
	fmt.Println(x * y) // 30
	fmt.Println(x / y) // 3
	fmt.Println(x % y) // 1

	// 비교
	fmt.Println(x == y) // false
	fmt.Println(x > y)  // true

	// 논리
	fmt.Println(x > 5 && y < 5) // true
	fmt.Println(!(x > y))       // false

	// 비트
	fmt.Println(x & y)  // 2 (1010 & 0011 = 0010)
	fmt.Println(x | y)  // 11 (1010 | 0011 = 1011)
	fmt.Println(x ^ y)  // 9 (1010 ^ 0011 = 1001)
	fmt.Println(x << 1) // 20 (1010 << 1 = 10100)
	fmt.Println(x >> 1) // 5  (1010 >> 1 = 0101)
}
