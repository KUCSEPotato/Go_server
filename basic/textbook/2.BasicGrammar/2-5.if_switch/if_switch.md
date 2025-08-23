# Golang `if` 문 정리

## 1. 기본 구조
Go의 `if` 문은 다른 언어(C, Java 등)와 비슷하지만 **조건에 괄호가 필요 없음**.

```go
if condition {
    // 실행 코드
} else if otherCondition {
    // 실행 코드
} else {
    // 실행 코드
}
````

---

## 2. 특징

* 조건문 괄호 `()` 불필요
* `if` 문 안에서 **변수 선언 가능**

  * 이 변수는 해당 `if` 블록 내부에서만 유효
* 주로 에러 처리에 활용됨

---

## 3. 예제 1: 기본 if-else

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

func main() {
    rand.Seed(time.Now().Unix()) // 매번 다른 난수 생성
    i := rand.Intn(15)           // 0~14 사이 난수

    fmt.Println("Random number:", i)

    if i < 10 {
        fmt.Println("Less than 10")
    } else if i < 20 {
        fmt.Println("Less than 20")
    } else {
        fmt.Println("20 or more")
    }
}
```

---

## 4. 예제 2: if 안에서 변수 선언

```go
package main

import (
    "fmt"
    "io/ioutil"
)

func main() {
    str := "Hello World!"

    // if 안에서 변수 선언 (에러 체크)
    if err := ioutil.WriteFile("test.txt", []byte(str), 0644); err != nil {
        fmt.Println("Error:", err)
    }
}
```

### 동작 설명

* `"Hello World!"` 문자열을 `test.txt` 파일에 씀
* `WriteFile`이 반환하는 에러(`err`)를 if 문 안에서 선언하고 체크
* 정상 동작 시 아무것도 출력되지 않고 파일만 생성됨
* 오류 발생 시 `"Error: ..." ` 메시지 출력

---

## 5. 정리

* Go의 `if` 문은 **간결**하고, **조건문 내부에서 변수 선언 가능**
* 파일 I/O 같은 에러 처리에서 자주 활용됨

```
