Go 언어(Golang)에서 **함수**는 매우 중요한 개념이며, 특히 다음과 같은 특징들을 중심으로 잘 설계되어 있습니다:

* ✅ 일급 함수 (First-class functions)
* ✅ 다중 반환값 (Multiple return values)
* ✅ 가변 인자 (Variadic Parameters, 특히 슬라이스와 함께 사용)
* ✅ 익명 함수 (Anonymous functions)와 클로저(Closures)

이제 각 요소를 하나씩 **정리 + 예시 코드 + 개념 해설**로 아주 자세히 설명해드릴게요.

---

## 1. 📌 일급 함수 (First-Class Functions)

Go에서는 함수도 **값처럼 다룰 수 있는 일급 객체**입니다. 즉,

* 함수를 변수에 할당 가능
* 함수를 인자로 전달 가능
* 함수를 반환값으로 전달 가능

### 🧠 개념

> "함수 자체를 데이터처럼 다룰 수 있다"는 것이 핵심입니다.

### ✅ 예시

```go
package main

import "fmt"

func add(a, b int) int {
	return a + b
}

func calc(op func(int, int) int, x, y int) int {
	return op(x, y)
}

func main() {
	f := add           // 함수 자체를 변수에 할당
	result := calc(f, 3, 4) // 함수 인자로 전달
	fmt.Println(result)     // 7 출력
}
```

---

## 2. 📌 다중 반환값 (Multiple Return Values)

Go 함수는 **두 개 이상의 값을 동시에 반환**할 수 있습니다. 이건 에러 처리에도 자주 쓰입니다.

### ✅ 예시

```go
func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("0으로 나눌 수 없습니다")
	}
	return a / b, nil
}

func main() {
	quotient, err := divide(10, 2)
	if err != nil {
		fmt.Println("에러 발생:", err)
	} else {
		fmt.Println("결과:", quotient)
	}
}
```

### 🧠 실용성

* `result, err := someFunc()` 패턴은 Go에서 매우 일반적
* `_`(언더스코어)로 불필요한 값 무시 가능

---

## 3. 📌 가변 인자 함수 (Variadic Functions) — 주로 슬라이스로 다룸

함수 인자 중 **마지막 인자를 가변 개수(slice처럼)** 받을 수 있습니다.

### ✅ 예시

```go
func sum(nums ...int) int {
	total := 0
	for _, num := range nums {
		total += num
	}
	return total
}

func main() {
	fmt.Println(sum(1, 2, 3))         // 6
	fmt.Println(sum(5, 10, 15, 20))   // 50

	values := []int{100, 200, 300}
	fmt.Println(sum(values...))       // 600 (슬라이스도 펼쳐서 전달 가능)
}
```

### 🧠 포인트

* `...int` : 여러 개 인자 받음
* 슬라이스를 `...`으로 **펼쳐서 호출** 가능 → `sum(values...)`

---

## 4. 📌 익명 함수와 클로저 (Anonymous Function & Closure)

### 4.1. 익명 함수

> 이름 없이 즉석에서 정의하는 함수입니다. 변수에 저장하거나 바로 실행 가능

```go
func main() {
	// 변수에 저장
	add := func(a, b int) int {
		return a + b
	}
	fmt.Println(add(2, 3)) // 5

	// 즉시 실행
	func(msg string) {
		fmt.Println("Hello,", msg)
	}("Gopher")
}
```

---

### 4.2. 클로저 (Closure)

> 함수가 **자신이 선언된 환경의 변수에 접근**할 수 있는 특성

```go
func counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	next := counter()
	fmt.Println(next()) // 1
	fmt.Println(next()) // 2
	fmt.Println(next()) // 3 (내부 상태 유지)
}
```

### 🧠 활용

* 상태 유지 함수 (스코프 밖에서도 변수 `count`가 살아 있음)
* 커링(Currying) 등 함수형 프로그래밍 스타일도 가능

---

## ✨ 요약 정리

| 기능    | 설명                          | 예시                        |
| ----- | --------------------------- | ------------------------- |
| 일급 함수 | 함수는 값처럼 다룰 수 있음             | 변수 할당, 함수 전달, 반환          |
| 다중 반환 | 여러 값을 동시에 리턴 가능             | `return a, nil`           |
| 가변 인자 | 마지막 인자를 `...T` 형식으로 여러 개 받음 | `sum(...int)`             |
| 익명 함수 | 이름 없이 함수 정의                 | `func(x int) int { ... }` |
| 클로저   | 외부 변수 상태를 기억하는 함수           | 내부 변수 유지 counter          |

---

필요하다면 각 개념에 대해 더 깊이 들어가거나 실전 예제를 같이 만들어볼 수도 있어요. 추가로 `defer`, `panic/recover`, `method receiver` 등도 관심 있으신가요?
