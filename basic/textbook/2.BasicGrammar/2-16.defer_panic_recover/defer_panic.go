/*
defer:
	함수가 종료될 때까지 실행을 미루는 기능
	try catch finally 문의 finally 블록과 유사
panic:
	실행 시점에서 현재 함수의 defer() 함수를 실행 후 리턴
	상위로 계속 리턴하여 에러를 출력하고 종료

	defer:
		- defer 키워드는 함수가 종료되기 직전에 지정한 함수를 실행하도록 예약합니다.
		- 여러 개의 defer가 있을 경우, 나중에 작성된 defer가 먼저 실행됩니다(LIFO 순서).
		- 주로 파일 닫기, 리소스 해제, unlock 등 정리 작업에 사용됩니다.
		- 예시:
			func example() {
				defer fmt.Println("함수 종료 직전 실행")
				fmt.Println("함수 본문 실행")
			}
			// 출력: 함수 본문 실행
			//      함수 종료 직전 실행

	panic:
		- panic 함수는 실행 중 치명적인 오류가 발생했을 때 프로그램의 실행을 중단시키고, 현재 함수의 defer 함수들을 모두 실행한 후 호출 스택을 따라 상위 함수로 에러를 전파합니다.
		- 최상위까지 전파되면 프로그램이 비정상적으로 종료됩니다.
		- panic은 주로 복구할 수 없는 상황(예: 배열 인덱스 초과, nil 포인터 참조 등)에 사용합니다.
		- 예시:
			func example() {
				defer fmt.Println("defer 실행")
				panic("치명적 오류 발생")
			}
			// 출력: defer 실행
			//      panic:
*/

package main

import "os"

func main() {
	f, err := os.Open("1.txt")
	if err != nil {
		// error가 발생한 시점에 종료
		panic(err) // 파일 열기 실패 시 panic 발생
	}

	// main 마지막에 파일 close 실행
	defer f.Close()

	// 파일 읽기
	bytes := make([]byte, 1024)
	f.Read(bytes)
	println(len(bytes))
}

/*
os.Open("1.txt")에서 파일이 존재하지 않거나, 권한이 없거나, 경로가 잘못된 경우
해당 파일이 없으면 err에 에러가 할당되고, panic(err)이 호출되어 프로그램이 종료됩니다.

f.Read(bytes)에서 파일을 읽는 도중 에러가 발생할 경우
하지만 현재 코드에서는 Read의 반환값 중 에러를 체크하지 않고 있습니다.
만약 읽기 도중 에러가 발생해도, 코드상에서는 에러를 처리하지 않고 넘어갑니다.
즉, 실제로 프로그램이 종료되는 error 조건은 파일이 없거나 열 수 없을 때입니다.
파일 읽기 에러는 현재 코드에서는 무시되고 있습니다.
*/
