## 전체 코드

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	done1 := make(chan bool)
	done2 := make(chan bool)

	go run1(done1)
	go run2(done2)

EXIT:
	for {
		select {
		case <-done1:
			fmt.Println("done run1")
		case <-done2:
			fmt.Println("done run2")
			break EXIT
		}
	}
}

func run1(done chan bool) {
	time.Sleep(1 * time.Second)
	done <- true
}

func run2(done chan bool) {
	time.Sleep(2 * time.Second)
	done <- true
}
```

---

## 동작 흐름 설명

### 1. 채널 생성

```go
done1 := make(chan bool)
done2 := make(chan bool)
```

* `done1`, `done2`는 두 개의 신호 채널.
* 고루틴이 일을 마치면 `true`를 보내서 "끝났어!"라는 신호를 준다.

---

### 2. 고루틴 실행

```go
go run1(done1)
go run2(done2)
```

* `run1`과 `run2`가 각각 별도 고루틴에서 실행된다.
* `run1`은 1초 뒤에 `done1 <- true`를 보냄.
* `run2`는 2초 뒤에 `done2 <- true`를 보냄.

---

### 3. 메인 루프와 `select`

```go
EXIT:
	for {
		select {
		case <-done1:
			fmt.Println("done run1")
		case <-done2:
			fmt.Println("done run2")
			break EXIT
		}
	}
```

* `select`는 여러 채널 중 **준비된(ready) 채널 하나를 랜덤 선택**해서 처리한다.

#### 실행 타이밍

* 약 1초 뒤 → `run1`이 끝나고 `done1 <- true` 실행 → `case <-done1` 실행 → `"done run1"` 출력
* 약 2초 뒤 → `run2`가 끝나고 `done2 <- true` 실행 → `case <-done2` 실행 → `"done run2"` 출력 → `break EXIT`로 루프 탈출

---

### 4. 라벨이 붙은 `break`

```go
break EXIT
```

* 일반적인 `break`는 `select`만 빠져나오지만,
* 라벨을 붙이면(`EXIT:`) 해당 라벨이 걸린 `for`까지 탈출한다.
* 즉, `done2`가 오면 **루프 전체를 종료**한다.

---

## 실행 결과 (예상)

실행하면 대략 이런 출력이 나온다:

```
done run1
done run2
```

* 순서는 항상 위와 같음 (`run1`이 1초, `run2`가 2초).
* `done2`를 받고 나면 `for` 루프가 끝나서 `main()` 종료.

---

## 핵심 포인트

1. **채널**: 고루틴 간 동기화 신호를 전달하는 도구.
2. **select**: 여러 채널을 동시에 기다리면서, 준비된 쪽만 처리.
3. **라벨 break**: 중첩 루프나 select에서 깔끔하게 빠져나올 때 사용.
4. 이 구조는 “특정 이벤트가 발생하면 루프를 종료”하는 데 자주 쓰임.

---

정리하면, 이 코드는 **두 고루틴(run1, run2)을 실행해두고, 그 중 run2가 끝날 때까지 기다리다가 프로그램을 종료**하는 예제예요.
