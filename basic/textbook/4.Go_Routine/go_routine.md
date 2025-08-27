# Go Routine?
- Go 루틴은 Go에서 관리해주는 경량 쓰레드(lightweight thread)이다.
    - 다시 말해, 운영체제의 쓰레드가 아니라 Go 런타임이 관리하는 m:n 스케줄링 기반의 실행 단위 이다.
    - 생성 비용이 매우 적기 때문에, 일반적인 쓰레드보다 훨씬 많은 수를 동시에 실행할 수 있다.
    ``` go
    package main

    import (
        "fmt"
        "time"
    )

    func printNumbers() {
        for i := 1; i <= 5; i++ {
            fmt.Println(i)
            time.Sleep(500 * time.Millisecond)
        }
    }

    func main() {
        // 고루틴 실행
        go printNumbers()

        // main 고루틴
        fmt.Println("main 함수 실행 중")
        time.Sleep(3 * time.Second) // main이 종료되면 모든 고루틴도 종료됨
    }
    ```
# 특징
- 경량성: 운영체제 쓰레드보다 메모리 사용량과 생성 비용이 훨씬 적다
- m:n 스케줄링: go 런타임이 여러 고루틴을 운영체제 쓰레드 위에서 자동으로 스케줄링한다.
- 비동기 실행: go 키워드로 실행된 함수는 즉시 반환되고, 해당 함수는 별도의 고루틴에서 실행됨.

# 채널?
- 고루틴끼리 데이터를 안전하게 주고 받기 위해서 채널(channel)을 사용한다.
    ``` go
    package main

    import "fmt"

    func sum(a int, b int, ch chan int) {
        ch <- a+b // 채널로 값 전송
    }

    func main() {
        ch := make(chan int)

        go sum(3,4,ch)

        res := <- ch // 채널에서 값 수신
        fmt.Println("sum:", res)
    }
    ```

# 주의할 점
- main 고루틴이 끝나면 모든 고루틴이 함께 종료됩니다. 따라서 sync.WaitGroup이나 채널을 통해 동기화가 필요합니다.
- 공유 자원을 다룰 때는 race condition을 피해야 하므로, mutex 또는 채널로 조율해야 합니다.

# 보충: 런타임이란?
## 답변: gpt
좋은 질문이에요 👍 Go의 **런타임(runtime)** 이라는 말은 단순히 "프로그램 실행 시간"을 의미하는 게 아니라, **Go 프로그램이 실행될 때 자동으로 함께 동작하는 Go 전용의 시스템 소프트웨어**를 뜻해요.

---

## 1. 런타임(Runtime)의 일반적 의미

프로그래밍 언어에서 **런타임**은 보통 다음을 포함합니다:

* **언어가 돌아가기 위해 필요한 지원 코드** (ex: 메모리 관리, 스케줄러, 가비지 컬렉터)
* **컴파일된 코드와 운영체제(OS) 사이의 중간 계층**
* 즉, "내가 작성한 코드 + 자동으로 링크된 언어 런타임 라이브러리"가 합쳐져서 실행됩니다.

---

## 2. Go 런타임의 역할

Go는 **고루틴, 채널, 가비지 컬렉션** 같은 기능을 제공하는데, 이건 OS가 직접 지원하지 않아요.
따라서 Go는 **런타임 라이브러리**를 통해 이를 구현합니다.

### Go 런타임이 하는 일

1. **고루틴 관리**

   * `go func()`로 만든 고루틴을 생성, 실행, 스케줄링
   * M\:N 스케줄러를 사용해서 수많은 고루틴을 적은 수의 OS 스레드 위에서 실행

2. **스케줄링**

   * 어떤 고루틴을 언제 실행할지 정해줌
   * preemptive(선점형) 방식으로 동작 → 긴 고루틴이 CPU 독점 못하게 막음

3. **메모리 관리**

   * **가비지 컬렉션(GC)**: 사용하지 않는 메모리 자동 해제
   * 힙/스택 메모리 관리

4. **네트워킹 및 시스템 호출**

   * 비동기 I/O 지원 (`net/http`, `os` 패키지 등이 런타임 기능과 맞물려 동작)

5. **에러/패닉 처리**

   * `panic`, `recover` 같은 기능 제공

---

## 3. 비유

* **OS(운영체제)**: 컴퓨터 하드웨어 위에서 프로그램들을 관리
* **Go 런타임**: OS 위에서 돌아가지만, **고루틴·채널 같은 Go 언어만의 기능**을 대신 처리하는 작은 "미니 운영체제" 같은 존재

---

## 4. 코드로 보는 런타임 예시

아래처럼 `go` 키워드로 고루틴을 만들면, OS가 직접 고루틴을 관리하는 게 아니라 **Go 런타임 스케줄러**가 실행 위치를 조율합니다.

```go
package main

import (
	"fmt"
	"time"
)

func worker(id int) {
	for i := 0; i < 3; i++ {
		fmt.Printf("Worker %d: %d\n", id, i)
		time.Sleep(time.Millisecond * 500)
	}
}

func main() {
	for i := 1; i <= 3; i++ {
		go worker(i) // 런타임 스케줄러가 관리하는 고루틴
	}

	time.Sleep(2 * time.Second) // main 종료 방지
}
```

출력 순서는 **예측 불가**해요. 왜냐하면 Go 런타임 스케줄러가 CPU 시간 배분을 결정하기 때문이에요.

---

👉 정리하면, **Go 런타임은 고루틴, 채널, 가비지 컬렉션, 스케줄러 같은 언어 레벨 기능을 OS 위에서 대신 관리해주는 작은 운영체제 같은 역할**을 하는 거예요.

# go scheduling
https://syntaxsugar.tistory.com/entry/GoGolang-Scheduler
