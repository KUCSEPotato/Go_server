- Go를 이용한 테스트를 진행할 때 기존 파일명에 _test.go로 생성하면 테스트 파일이 됩니다.
``` go
package main

import "testing"

func Test[함수명](t *testing.T) {
    함수명()
}
```