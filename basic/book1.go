/* 패키지/모듈/기본 구조 */
// 모든 GO 코드는 패키지(package)안에 있음.
// 실행 파일의 엔트리포인트는 package main + func main().
// 외부 패키지는 go mod init <모듈명> -> go get으로 관리
package main

import (
	"fmt" // 표준 출력
	"time" // 시간
)

func main() {
	// 현재 시간을 가져옵니다.
	currentTime := time.Now()

	// 현재 시간을 출력합니다.
	fmt.Println("현재 시간:", currentTime)

	// 현재 시간의 연도, 월, 일, 시, 분, 초를 출력합니다.
	fmt.Printf("연도: %d\n", currentTime.Year())
	fmt.Printf("월: %d\n", currentTime.Month())
	fmt.Printf("일: %d\n", currentTime.Day())
	fmt.Printf("시: %d\n", currentTime.Hour())
	fmt.Printf("분: %d\n", currentTime.Minute())
	fmt.Printf("초: %d\n", currentTime.Second())
}

// 대문자로 시작하는 이름은 export(외부 공개), 소문자는 패키지 내부에서만 사용 가능

/* 기본 자료형과 리터럴 */
// 불리언: bool (true, false)
// 정수: int, int8/16/32/64, 부호 없는 uint/uint8/16/32/64, rune(유니코드 코드포인트, 실제는 int32)
// 실수: float32, float64
// 복소수: complex64, complex128 (ex. complex(1, 2))
// 문자열: string  (불변, UTF-8). 백틱(`) 사용 시 raw string literal (이스케이프 처리 없음)
var a int = 42
b := 3.14                // 타입 추론(float64)
var c rune = '한'         // rune
s := "hello\nworld"
raw := `C:\path\no\escape`
// 제로값: 선언만 하고 초기화 안 하면 자동 기본값
// 숫자(0), bool(false), string(""), 포인터/슬라이스/맵/채널/함수/인터페이스(nil)

/* 변수 선언과 상수, iota */
var x int // 제로값 0
var y, z = 1, "go" // 다중 할당
w := 10 // 짧은 선언 (함수 내부에서만)
const Pi = 3.14159 // 상수는 컴파일타임 값

// iota: 연속 상수 만들 때 유용
// iota는 const 블록이 시작될 때 0부터 시작.
// 블록 안에서 줄이 하나 내려갈 때마다 1씩 증가.
// 같은 줄에 여러 개를 써도 같은 값.
const 블록이 끝나면 초기화됨(다음 블록에서 다시 0부터 시작).
const (
	_ = iota // 0 버리기
	KB = 1 << (10 * iota) // 1 << 10
	MB = 1 << (10 * iota) // 1 << 20
	GB = 1 << (10 * iota) // 1 << 30
	TB = 1 << (10 * iota) // 1 << 40
)
// 타입 있는 상수 vs 무타입 상수(untyped). 무타입 상수는 문맥에 따라 형 변환 없이 다양한 타입에 맞춰짐

/* 제어문: if/for/switch/defer */
// if (초기화문 가능)
if n := len(s); n > 0 {
	fmt.Println("문자열 길이:", n)
} else {
	fmt.Println("빈 문자열")
}

// for (세 가지 형태만 존재, while은 없음)
for i := 0; i < 3; i++ {
	fmt.Println("반복:", i) // 전통적
}
for i < 10 {
	i++
	fmt.Println("반복:", i) // while 스타일
}
for {
	fmt.Println("무한 반복") // 무한 loop
}

// range
// range의 반복 변수는 복사본임. 루프 변수 주소를 캠쳐할 때 주의 (클로저, 고루틴에서 흔한 버그임)
nums := []int{1, 2, 3, 4, 5}
for idx, v := range nums {
	fmt.Println("인덱스:", idx, "값:", v)
}

// switch (fallthrough 없음이 기본)
"""
[기본틀]
값 스위치:
switch <값> {
case 값1, 값2:
    ...
default:
    ...
}

타입 스위치:
switch <변수>.(type) {
case 타입1:
    ...
case 타입2:
    ...
default:
    ...
}

"""
// 값 스위치
"""
[동작 원리]
day := time.Now().Weekday()
	time.Now() → 현재 시각을 가져옴.
	.Weekday() → 요일(time.Weekday 타입, 예: time.Monday, time.Saturday 등)을 반환.
	이렇게 반환된 값을 day 변수에 저장.

switch day { ... }
	day 값에 따라 분기 실행.

case time.Saturday, time.Sunday: day가 토요일 또는 일요일이면 이 블록 실행 → "weekend" 출력.
default: 위 case에 해당하지 않으면 "weekday" 출력.

Go의 switch는 자동 break라서, 조건이 맞으면 그 블록 실행 후 바로 빠져나옵니다.
한 줄에 초기화문(day := ...)과 조건변수(day)를 같이 쓸 수 있습니다.
"""

switch day := time.Now().Weekday(); day {
case time.Saturday, time.Sunday:
	fmt.Println("weekend")
default:
	fmt.Println("weekday")
}

// 타입 스위치
"""
[동작 원리]
var any any = 3.14
	any는 Go 1.18에서 추가된 interface{}의 별칭.
	모든 타입을 담을 수 있음.
	지금은 3.14(float64 타입)를 담음.

	switch v := any.(type) {
	case int:
		any 안에 값이 int 타입이면 실행.
	case float64:
		any 안에 값이 float64 타입이면 실행 → 여기서 매칭됨.
		v에는 타입이 확정된 값이 들어옴 (float64로).
	default:
		위 조건에 없는 타입일 때 실행. %T 포맷은 변수의 타입 이름을 출력.

case int: any 안에 값이 int 타입이면 실행.
case float64: any 안에 값이 float64 타입이면 실행 → 여기서 매칭됨.

v에는 타입이 확정된 값이 들어옴 (float64로).

default: 위 조건에 없는 타입일 때 실행. %T 포맷은 변수의 타입 이름을 출력.

타입 스위치는 **타입 단언(type assertion)**의 확장판.
	v, ok := any.(float64) // 한 타입만 확인
	대신 여러 타입을 한 번에 검사 가능.
"""
var any any = 3.14
switch v := any.(type) {
case int:
	fmt.Println("int", v)
case float64:
	fmt.Println("float64", v)
default:
	fmt.Printf("unknown %T\n", v)
}

// defer / panic / recover (리소스 정리)
f, _ := os.Create("out.txt")
defer f.Close() // 함수 끝날 때 실행

/* 컬렉션: 배열/슬라이스/맵 */
// 배열: 고정 크기, 값 타입
var arr [3]int = [3]int{1, 2, 3}

// 슬라이스: 가변 길이, 공유되는 내부 배열
// make{[]Type, len, cap}으로 생성, append로 확장.
// 슬라이스는 얕은 복사. 공유로 인한 사이드이펙트에 주의해야 함.
// 독립본이 필요한 경우 copy(dst, src) 사용 
s := make([]int, 0, 5) // 길이 0, 용량 5인 슬라이스 생성
s = append(s, 1, 2, 3)  // 값 추가
t := s[1:3]            // 슬라이스 t는 s의 1번 인덱스부터 2번 인덱스까지
u := append(s, 99) // cap 넘치면 새로운 배열 할당

// 맵: 해시맵 구현: 키 K -> 값 v 매핑.
// 참조 타입: 변수에 할당하면 “헤더(포인터)”만 복사되고, 같은 내부 데이터를 바라봄.
// 순서가 없음: 반복 순서는 매번 달라질 수 있음(의도적으로 랜덤).

// make로 생성 (권장)
m := make(map[string]int)          // 빈 맵
m2 := make(map[string]int, 1000)   // 용량 힌트(대략적인 초기 버킷 수)

// 리터럴
m3 := map[string]int{"a": 1, "b": 2}

// 제로값(nil map)
var mNil map[string]int            // nil 맵(읽기만 가능), 쓰기 불가 (panic)

m := make(map[string]int)
m["a"] = 10            // 쓰기
x := m["a"]            // 읽기 (없으면 0)
delete(m, "a")         // 삭제 (키 없으면 무시)
ln := len(m)           // 현재 원소 수

v, ok := m["k"]        // ok == false면 '부재'
if !ok {
    // 키가 없었음 (v는 값 타입의 제로값)
}
// 순서 비결정: 정렬이 필요하면 키를 슬라이스로 뽑아 정렬 후 사용
for k, v := range m {
    fmt.Println(k, v)
}

/* 함수와 클로저, 가변 인자, 다중 반환 */
"""
함수는 값이다: 변수/인자/반환에 자유롭게 쓰기.

클로저는 변수 자체를 캡쳐한다: 루프 변수 캡쳐 버그 조심(인자 복사 or 지역변수 복사).

가변 인자는 마지막 하나만 가능, 내부에선 []T.
[]int → ...any는 직접 변환 필요.

다중 반환은 Go의 일상: (값, error) 관용구, 불필요 값은 _로 버리기.

이름 붙은 반환값 + defer로 반환값을 마지막에 조정할 수 있으나, 가독성을 우선.
"""

// 다중 반환
// 같은 타입의 매개변수는 마지막에 한 번만 타입을 씁니다(a, b int).
func div(a, b int) (int, error) {
	if b == 0 { return 0, fmt.Errorf("divide by zero") }
	return a / b, nil
}

// 가변 인자
// 함수 내부에서 nums는 []int 슬라이스
// 빈 호출도 허용 (sum() -> 0)
func sum(nums ...int) int {
	total := 0
	for _, n := range nums { total += n }
	return total
}

// 클로저 (상태 캡쳐)
// 클로저 = 바깥 스코프의 변수를 캡쳐하는 함수 값입니다. 캡쳐된 변수는 함수가 살아있는 동안 유지
func counter() func() int {
	i := 0
	return func() int { i++; return i }
}
// 잘못된 예시
for i := 0; i < 3; i++ {
	go func() {
		fmt.Println(i) // (X) 반복변수 i를 “공유 변수”로 캡쳐
	}()
}
// 바른 예시
for i := 0; i < 3; i++ {
	go func(i int) {
		fmt.Println(i) // 각 고루틴마다 복사된 i
	}(i)
}

// 이름 붙은 반환값도 가능
// 함수 인자는 값 복사가 기본. 큰 구조를 수정하려면 포인터를 넘겨야 함
func area(w, h float64) (a float64) {
	a = w * h
	return // a가 암묵적으로 반환
}

/* 구조체(Struct)와 메서드(Method), 임베딩(Embedding) */
"""
구조체: 여러 필드를 묶어 하나의 타입을 정의.

메서드: 특정 타입(주로 구조체)에 연결된 함수. 값 리시버 / 포인터 리시버 선택 중요.
 - 값 리시버: 읽기 전용 느낌(수정은 원본에 반영되지 않음).
 - 포인터 리시버: 원본 수정 가능, 복사비용 절약.

임베딩(Embedding): 구조체 안에 다른 구조체를 필드명 없이 포함시켜 해당 필드/메서드를 직접 사용 가능.
"""

package main

import (
	"fmt"
	"time"
)

type User struct {
	ID   int       `json:"id"`   // 태그(JSON 직렬화 등)
	Name string    `json:"name"`
	When time.Time `json:"joined_at"`
}

// 값 리시버 메서드
func (u User) Display() string {
	return fmt.Sprintf("%d:%s", u.ID, u.Name)
}

// 포인터 리시버 메서드
func (u *User) Rename(n string) {
	u.Name = n
}

// 구조체 임베딩
type Audited struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Post struct {
	Audited
	Title string
}

func main() {
	u := User{ID: 1, Name: "Alice", When: time.Now()}
	fmt.Println(u.Display())

	u.Rename("Bob")
	fmt.Println(u.Display())

	p := Post{
		Audited: Audited{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Title: "First Post",
	}
	fmt.Println(p.Title, p.CreatedAt)
}

/* 인터페이스(Interface)와 타입 단언(Type Assertion) */
"""
Go 인터페이스는 암묵적으로 구현(implements 키워드 없음).
 - 어떤 타입이 인터페이스 메서드를 모두 구현하면 자동으로 인터페이스 만족.

빈 인터페이스(any)는 모든 타입을 담을 수 있음.
타입 단언(type assertion): 인터페이스 변수 안의 실제 타입 꺼내기.
타입 스위치(type switch): 여러 타입 분기 처리.
"""

package main

import (
	"fmt"
)

type Stringer interface {
	String() string
}

type Point struct{ X, Y int }

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

func main() {
	var s Stringer = Point{1, 2}
	fmt.Println(s.String())

	// 빈 인터페이스(any)
	var x any = "hello"
	if str, ok := x.(string); ok {
		fmt.Println("string:", str)
	}

	// 타입 스위치
	switch v := x.(type) {
	case string:
		fmt.Println("string type:", v)
	case int:
		fmt.Println("int type:", v)
	default:
		fmt.Println("unknown type")
	}
}

/* 포인터(Pointer) */
"""
& : 주소 연산자
* : 역참조(주소가 가리키는 값 접근)

포인터 연산(+,-)은 불가.
구조체 메서드에서 포인터 리시버를 사용하면 수정이 원본에 반영됨.
"""

package main

import "fmt"

func main() {
	n := 10
	p := &n       // n의 주소
	fmt.Println(*p) // 역참조 → 10

	*p = 20
	fmt.Println(n) // 20

	type Counter struct{ n int }
	c := Counter{n: 5}
	cp := &c
	cp.n++
	fmt.Println(c.n) // 6
}

/* 에러 처리(Error Handling) */
"""
Go는 예외 대신 error 값을 반환.
 - (값, error) 관용구
 - errors.Is / errors.As로 오류 판별
 - wrapping: fmt.Errorf("...: %w", err)

에러 무시는 지양. 항상 err != nil 검사.
"""

package main

import (
	"errors"
	"fmt"
	"os"
)

func readFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func main() {
	data, err := readFile("test.txt")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("파일 없음")
		} else {
			fmt.Println("기타 에러:", err)
		}
		return
	}
	fmt.Println(string(data))
}

/* 제네릭(Generic) */
"""
Go 1.18+에서 지원.
 - 타입 매개변수 [T any]
 - 제약(constraint) 지정 가능.
"""

package main

import "fmt"

type Number interface {
	~int | ~float64
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func main() {
	fmt.Println(Max(3, 5))
	fmt.Println(Max(2.5, 1.8))
}
/* 동시성(Concurrency) - goroutine, channel, select */
"""
goroutine: go 키워드로 함수 비동기 실행.
channel: goroutine 간 통신.
select: 다중 채널 대기.

공유 자원 접근 시 sync 패키지(Mutex, WaitGroup) 사용.
"""

package main

import (
	"fmt"
	"time"
)

func worker(ch chan string) {
	time.Sleep(1 * time.Second)
	ch <- "done"
}

func main() {
	ch := make(chan string)
	go worker(ch)

	select {
	case msg := <-ch:
		fmt.Println("received:", msg)
	case <-time.After(2 * time.Second):
		fmt.Println("timeout")
	}
}

/* new vs make */
"""
new(T): 제로값으로 초기화된 *T 반환.
make: 슬라이스, 맵, 채널 초기화.
"""

package main

import "fmt"

func main() {
	p := new(int)
	fmt.Println(*p) // 0
	*p = 42
	fmt.Println(*p) // 42

	s := make([]int, 0, 5)
	fmt.Println(s, len(s), cap(s))

	m := make(map[string]int)
	m["a"] = 1
	fmt.Println(m)

	c := make(chan int, 2)
	c <- 10
	fmt.Println(<-c)
}

/* 테스트, 포맷, 정적 분석 */
"""
테스트: _test.go 파일 + go test
포맷: go fmt ./...
정적분석: go vet ./...
"""

// mathx.go
package mathx

func Add(a, b int) int {
	return a + b
}

// mathx_test.go
package mathx

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Fatalf("want 5, got %d", got)
	}
}

