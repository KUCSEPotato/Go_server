# main 패키지를 특별하게 인식
- 실행 시작 점으로 인식
- 공유 라이브러리를 만들때는 main 으로 만들면 안됨
# import 키워드로 다른 패키지의 함수를 불러올 수 있음
- ${GOROOT}/pkg는 표준 패키지
- ${GOPATH}/pkg는 3rd 파티 패키지
# 패키지내 함수, 구조체, 인터페이스의 영역
- 첫문자를 대문자로 시작하면 public
- 첫문자를 소문자로 시작하면 private

# init 함수
- 패키지를 로드할 때 자동으로 실행
``` go
package main

var pop map[string]string

func init() {
    // 패키지 로드시 map 초기화
    pop = make(map[string]string)
}
```
- init() 함수만을 호출할 때 사용
``` go
package main
import _ "other/xlib"
```

# alias 사용
``` go
import {
    mongo "other/monge/db"
    mysql "other/mysql/db"
} 

func main() {
    mondb := mongo.Get()
    mydb := mysql.Get()
    // ...
}