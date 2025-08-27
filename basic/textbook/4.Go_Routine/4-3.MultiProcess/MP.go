package main

import (
	"fmt"
	"runtime"
)

/*
NumCPU() → 하드웨어 물리 CPU 코어 수 (불변)
GOMAXPROCS(n) → Go 런타임이 동시에 사용할 수 있는 논리 CPU 개수 (동적으로 변경 가능)
따라서 동일 소스코드 내부에서도 GOMAXPROCS 값은 여러 번 바꿀 수 있음.
*/
func main() {
	// 모든 CPU 코어 사용
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("NumCPU():", runtime.NumCPU())        // 물리적 CPU 수
	fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0)) // 현재 사용 중인 CPU 수

	// 4개의 CPU 코어만 사용
	runtime.GOMAXPROCS(4)
	fmt.Println("NumCPU():", runtime.NumCPU())        // 여전히 물리적 CPU 수
	fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0)) // 현재 사용 중인 CPU 수 (4로 바뀜)
}
