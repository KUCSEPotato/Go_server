# New
- New method는 새로운 App instance를 만들 때 사용한다.
- config parameter를 사용해서 custom이 가능하다. 사용하지 않을 경우에는 default 설정이 사용된다.
``` go
// signature
func New(config ...Config) *App

// Example
// Default config
app := fiber.New()
```

# config
- 위에서 언급한 config에 대한 설명입니다.
``` go
// Custom config
app := fiber.New(fiber.Config{
    Prefork:       true,
    CaseSensitive: true,
    StrictRouting: true,
    ServerHeader:  "Fiber",
    AppName: "Test App v1.0.1",
})
```

## config field
| Property                         | Type                | Description                                                                        | Default               |
| -------------------------------- | ------------------- | ---------------------------------------------------------------------------------- | --------------------- |
| **AppName**                      | string              | 앱 이름을 설정합니다.                                                                       | `""`                  |
| **BodyLimit**                    | int                 | 요청 바디의 최대 크기를 설정합니다. 초과 시 `413 Request Entity Too Large` 응답을 보냅니다.                 | `4 * 1024 * 1024`     |
| **CaseSensitive**                | bool                | `/Foo`와 `/foo`를 서로 다르게 취급할지 여부.                                                    | `false`               |
| **ColorScheme**                  | Colors              | 커스텀 컬러 스킴 정의 (시작 메시지, 라우트 리스트, 일부 미들웨어 출력에 사용).                                    | `DefaultColors`       |
| **CompressedFileSuffix**         | string              | 압축된 파일 저장 시 붙일 접미사.                                                                | `.fiber.gz`           |
| **Concurrency**                  | int                 | 최대 동시 연결 수.                                                                        | `256 * 1024`          |
| **DisableDefaultContentType**    | bool                | 기본 `Content-Type` 헤더를 응답에 포함하지 않도록 설정.                                             | `false`               |
| **DisableDefaultDate**           | bool                | 기본 `Date` 헤더를 응답에 포함하지 않도록 설정.                                                     | `false`               |
| **DisableHeaderNormalizing**     | bool                | 기본적으로 헤더 이름을 정규화(`conteNT-tYPE → Content-Type`)하는 기능을 끔.                           | `false`               |
| **DisableKeepalive**             | bool                | keep-alive 연결을 비활성화, 첫 응답 후 연결 종료.                                                 | `false`               |
| **DisablePreParseMultipartForm** | bool                | 멀티파트 폼 데이터를 사전 파싱하지 않음. 서버가 멀티파트를 바이너리 blob으로 다루거나 직접 파싱하고 싶을 때 유용.                | `false`               |
| **DisableStartupMessage**        | bool                | 서버 시작 시 출력되는 디버그 메시지를 숨김.                                                          | `false`               |
| **ETag**                         | bool                | ETag 헤더 생성 여부. CRC-32 기반 약(weak) ETag 생성이 기본.                                      | `false`               |
| **EnableIPValidation**           | bool                | `c.IP()`와 `c.IPs()` 반환 시 IP 유효성 검사. 성능에 약간의 비용이 있으므로 신뢰된 프록시 뒤에 있다면 끄는 것을 권장.      | `false`               |
| **EnablePrintRoutes**            | bool                | 서버 시작 시 등록된 라우트(method, path, handler 등) 출력.                                       | `false`               |
| **EnableSplittingOnParsers**     | bool                | 쿼리/바디/헤더 파라미터를 콤마(`,`)로 분리해 배열처럼 파싱. 예: `/api?foo=bar,baz → foo[]=bar, foo[]=baz`. | `false`               |
| **EnableTrustedProxyCheck**      | bool                | 신뢰된 프록시만 `c.IP()`, `c.Protocol()`, `c.Hostname()` 등에 반영. `TrustedProxies` 목록을 참조.  | `false`               |
| **ErrorHandler**                 | ErrorHandler        | 핸들러에서 반환된 에러를 처리하는 전역 에러 핸들러.                                                      | `DefaultErrorHandler` |
| **GETOnly**                      | bool                | GET 외의 모든 요청 거부. 단순 GET 서버에서 DoS 방어에 유용.                                           | `false`               |
| **IdleTimeout**                  | time.Duration       | keep-alive에서 다음 요청을 기다리는 최대 시간. `0`이면 `ReadTimeout` 값을 사용.                         | `nil`                 |
| **Immutable**                    | bool                | `Ctx`에서 반환된 값들을 불변으로 보장. 기본은 핸들러 반환 시까지만 유효.                                       | `false`               |
| **JSONDecoder**                  | utils.JSONUnmarshal | JSON 디코딩 시 사용할 라이브러리 지정.                                                           | `json.Unmarshal`      |
| **JSONEncoder**                  | utils.JSONMarshal   | JSON 인코딩 시 사용할 라이브러리 지정.                                                           | `json.Marshal`        |
| **Network**                      | string              | 네트워크 타입: `"tcp"`, `"tcp4"`, `"tcp6"`. Prefork 모드에선 `tcp4`, `tcp6`만 가능.             | `NetworkTCP4`         |
| **PassLocalsToViews**            | bool                | `c.Locals`에 저장한 값을 템플릿 렌더러로 전달.                                                    | `false`               |
| **Prefork**                      | bool                | `SO_REUSEPORT` 기반 멀티 프로세스 리스닝. Docker에서 Prefork 모드 실행 시 shell을 통해 앱을 실행해야 함.       | `false`               |
| **ProxyHeader**                  | string              | `c.IP()`를 특정 헤더(`X-Forwarded-*`) 값으로 반환하도록 지정.                                     | `""`                  |
| **ReadBufferSize**               | int                 | 요청 읽기 버퍼 크기(헤더 크기 제한). 큰 헤더/쿠키 전송 시 증가 필요.                                         | `4096`                |
| **ReadTimeout**                  | time.Duration       | 전체 요청(바디 포함)을 읽는 데 허용할 최대 시간.                                                      | `nil`                 |
| **RequestMethods**               | \[]string           | 허용할 HTTP 메서드 집합 지정.                                                                | `DefaultMethods`      |
| **ServerHeader**                 | string              | 응답 헤더에 `Server: ...` 값을 설정.                                                        | `""`                  |
| **StreamRequestBody**            | bool                | 요청 바디를 스트리밍 모드로 읽음. 큰 요청은 바디를 모두 읽기 전에 핸들러 실행 시작.                                  | `false`               |
| **StrictRouting**                | bool                | `/foo`와 `/foo/`를 서로 다른 경로로 취급.                                                     | `false`               |
| **TrustedProxies**               | \[]string           | 신뢰할 프록시 IP/대역 목록. `EnableTrustedProxyCheck` 활성화 시 참조됨.                             | `[]string{}`          |
| **UnescapePath**                 | bool                | URL 인코딩된 경로 문자를 복원 후 라우팅.                                                          | `false`               |
| **Views**                        | Views               | Fiber 템플릿 엔진 인터페이스.                                                                | `nil`                 |
| **ViewsLayout**                  | string              | 전역 템플릿 레이아웃 파일. `Render()`에서 오버라이드 가능.                                             | `""`                  |
| **WriteBufferSize**              | int                 | 응답 쓰기 버퍼 크기.                                                                       | `4096`                |
| **WriteTimeout**                 | time.Duration       | 응답 쓰기 제한 시간.                                                                       | `nil`                 |
| **XMLEncoder**                   | utils.XMLMarshal    | XML 인코딩 시 사용할 라이브러리 지정.                                                            | `xml.Marshal`         |

# NewError
- NewError는 새로운 HTTPError instance를 만든다.
    - optinal message를 함꼐 포함시킬 수 있다.
``` go
// Signature
func NewError(code int, message ...string) *Error

// Example
app.Get("/", func(c *fiber.Ctx) error {
    return fiber.NewError(782, "Custom error message")
})
```

# IsChild
- fiber.IsChild() 는 현재 실행 중인 Go 프로세스가 Prefork로 생성된 자식 프로세스인지 여부를 알려준다.
    - 반환 값: true → Prefork 자식 프로세스에서 실행 중임, false → 부모 프로세스이거나 Prefork를 쓰지 않는 일반 실행
    - Prefork: Config{ Prefork: true } 옵션을 켜면 Prefork 모드로 실행
        - SO_REUSEPORT 소켓 옵션을 써서 하나의 포트에 여러 프로세스를 동시에 바인딩하는 방식
        - Fiber는 부모 프로세스(Parent) 하나와, 요청을 실제로 처리하는 여러 개의 자식 프로세스(Child) 를 띄운다.
- 필요한 이유?
    - Prefork 모드에선 보통: 
        - 부모 프로세스: 실제 요청을 처리하지 않고, 자식들을 포크(fork)해서 관리만 함.
        - 자식 프로세스들: 요청을 처리하는 서버 역할.
    - 따라서, 어떤 초기화 로직은 부모에서만 실행하거나, 어떤 자원은 자식에서만 실행해야 할 때가 있음. 이럴 때 IsChild()를 조건문에 써서 분기합.
``` go
package main

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
)

func main() {
    app := fiber.New(fiber.Config{
        Prefork: true, // Prefork 모드 활성화
    })

    if fiber.IsChild() {
        // Prefork 자식 프로세스일 때만 실행
        fmt.Println("Child process: start workers or DB connections")
    } else {
        // 부모 프로세스일 때만 실행
        fmt.Println("Parent process: setup metrics or logging")
    }

    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello, Fiber Prefork!")
    })

    app.Listen(":3000")
}
```