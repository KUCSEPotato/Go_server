package main

/*
Go has error interface
*/

// Go는 error를 함께 리턴하는 경우가 많고, error를 nil 체크해서 오류를 처리할 수 있음.
import (
	"log"
	"os"
)

type error interface {
	Error() string
}

// 사용자 정의 에러 타입
type MyError struct {
	msg string
}

func (e MyError) Error() string {
	return e.msg
}

func otherFunc() (int, error) {
	// 예시: 에러 발생
	return 0, MyError{msg: "custom error"}
}

func main() {
	f, err := os.Open("C:\\temp\\1.txt")
	if err != nil {
		log.Fatal(err.Error()) // log.Fatal() 함수는 에러 메시지를 출력하고 프로그램을 종료
	}
	println(f.Name(), "file opened") // 파일이 정상적으로 열렸을 때 출력

	// 사용자 정의 에러를 잡을 때
	_, err = otherFunc()
	if err == nil {
		println("No error occurred")
	} else {
		switch e := err.(type) {
		case MyError:
			log.Print("Log me error: ", e.Error())
		default:
			log.Fatal(err.Error())
		}
	}
}
