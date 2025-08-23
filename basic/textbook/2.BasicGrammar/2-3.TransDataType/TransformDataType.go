/*
	데이터 타입 변환은 타입(값) 형식으로 변경할 수 있습니다.

Atoi, ParseInt 함수가 대표적이고, 내부 동작은 아래와 같습니다.

### 1. `strconv.Atoi`

* **이름 뜻**: ASCII to Integer
* **기본 용도**: 문자열을 **10진수 int**로 바꿔줌
* 내부적으로는 사실 `ParseInt(s, 10, 0)`을 호출해요.

  - `base = 10` (십진수)
  - `bitsize = 0` → 자동으로 `int` 크기에 맞춤 (32bit/64bit 플랫폼에 따라 다름)

즉, `Atoi("123")`는 단순히 `ParseInt("123", 10, 0)`의 **얇은 포장(wrapper)** 함수예요.
그래서 편하게 "문자열 → int"만 하고 싶을 때 쓰는 거예요.

---

### 2. `strconv.ParseInt`

* 훨씬 **범용적**이에요.
* 사용자가 **진법(base)** 과 **비트 크기(bitsize)** 를 지정할 수 있어요.

  - `base` : 2, 8, 10, 16 등 원하는 진법
  - `bitsize` : 0, 8, 16, 32, 64 → `int8`, `int16`, `int32`, `int64`로 변환
  - 반환 타입: 항상 **int64**
    → 원하는 크기로 다시 변환해서 써야 할 수 있음

예시:
```
strconv.ParseInt("1010", 2, 8)   // 2진수 "1010"을 int8로 해석
strconv.ParseInt("7B", 16, 32)   // 16진수 "7B"를 int32 범위로 해석
```

### 3. 정리
* `Atoi` = "간단하게 문자열을 int로"
* `ParseInt` = "진법/비트 크기까지 세밀하게 조절하고 싶을 때"
*/
package main

import (
	"fmt"
	"reflect"
	"strconv"
)

func main() {
	strInt := "100"
	intStr, _ := strconv.Atoi(strInt)
	fmt.Println(intStr, reflect.TypeOf(intStr)) // 100 string

	i, err := strconv.Atoi(strInt)
	fmt.Println(i, reflect.TypeOf(i), err) // 100 int <nil>

	strInt = "987654321"
	i8, err := strconv.ParseInt(strInt, 0, 8)
	i16, err := strconv.ParseInt(strInt, 0, 16)
	i32, err := strconv.ParseInt(strInt, 0, 32)
	i64, err := strconv.ParseInt(strInt, 0, 64)

	fmt.Println(i8, reflect.TypeOf(i8), err)   // 127 int64 <nil>
	fmt.Println(i16, reflect.TypeOf(i16), err) // 32767 int64 <nil>
	fmt.Println(i32, reflect.TypeOf(i32), err) // 987654321 int64 <nil>
	fmt.Println(i64, reflect.TypeOf(i64), err) // 40926266145 int64 <nil>
}
