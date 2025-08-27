package main

import (
	"fmt"
	"runtime"
	"sync"
)

type counter struct {
	i  int64      // share data
	mu sync.Mutex // protect shared data
}

// increase counter
func (c *counter) increase() {
	c.mu.Lock()   // i 값을 변경하는 부분(임계 영역)을 뮤텍스로 잠금
	c.i++         // 공유 데이터 변경
	c.mu.Unlock() // i 값을 변경 완료한 후 뮤텍스 잠금 해제
}

// print counter value
func (c *counter) display() {
	fmt.Println(c.i)
}

func main() {
	// 모든 CPU 코어 사용
	runtime.GOMAXPROCS(runtime.NumCPU())

	c := counter{i: 0}     // couter 생성
	wg := sync.WaitGroup{} // WaitGroup 생성

	// 10개의 goroutine을 생성하여 counter 값을 증가시킴
	for j := 0; j < 10; j++ {
		wg.Add(1) // WaitGroup 카운터 증가
		go func() {
			defer wg.Done() // goroutine 종료 시 WaitGroup 카운터 감소
			for i := 0; i < 1000; i++ {
				c.increase()
				if err := recover(); err != nil {
					fmt.Println("Recovered from error:", err)
				}
			}
		}()
	}

	wg.Wait()   // 모든 goroutine이 종료될 때까지 대기
	c.display() // 최종 counter 값 출력
}
