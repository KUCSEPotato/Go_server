## 1. Fiber란?

* **Go 언어 기반의 웹 프레임워크**
* \[Express.js]\(Node.js 프레임워크)에서 영감을 받아, 문법과 구조가 Express랑 매우 비슷해요.
* 내부적으로는 **fasthttp**라는 초고속 HTTP 엔진을 기반으로 동작합니다.
  → fasthttp는 Go의 기본 `net/http`보다 훨씬 빠른 성능을 보여줍니다.
* **특징**: 가볍고, 빠르고, 직관적임 → “Go 버전의 Express.js”라고 보면 됩니다.

---

## 2. Fiber의 장점

1. **고성능**

   * fasthttp 기반 → Go 웹 프레임워크 중 성능 최상위권.
   * 초당 수십만 요청 처리 가능.

2. **직관적인 API**

   * Express.js와 거의 똑같은 스타일이라, JS/TS 경험이 있다면 바로 적응 가능.
   * 라우팅, 미들웨어, 컨텍스트 사용이 매우 간단.

3. **미들웨어 지원**

   * 로깅, 인증(JWT, 세션), CORS, 압축, 정적 파일 서빙 등 다 지원.
   * 필요하면 커스텀 미들웨어도 쉽게 추가 가능.

4. **경량성**

   * 불필요한 기능 없이 필요한 것만 빠르게 구현 가능.

---

## 3. Fiber 기본 예제

```go
package main

import (
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// GET 요청
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Fiber 🚀")
	})

	// POST 요청 (ex: 사물함 신청)
	app.Post("/locker/apply", func(c *fiber.Ctx) error {
		// 신청자 정보 파싱
		type Request struct {
			UserID  string `json:"user_id"`
			LockerID int    `json:"locker_id"`
		}
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request",
			})
		}

		// TODO: DB/Redis에 선착순 예약 로직 작성
		return c.JSON(fiber.Map{
			"message": "Locker applied",
			"user":    req.UserID,
			"locker":  req.LockerID,
		})
	})

	app.Listen(":3000")
}
```

---

## 4. 사물함 선착순 예약 시 Fiber + Go의 강점

* **고루틴** 덕분에 동시에 수많은 요청을 처리할 수 있음.
* **Fiber**가 `fasthttp` 기반이라 요청 처리 속도가 매우 빠름.
* **실시간 선착순 처리**는 DB Lock이나 Redis를 활용해서 “첫 번째 요청만 성공 → 나머지는 실패” 같은 로직을 구현하기 좋음.

  * 예: Redis의 `SETNX` 명령어로 locker\_id를 key로 설정 → 이미 존재하면 실패.

---

## 5. Fiber를 쓸 때 고려할 점

* Go 생태계에서는 Fiber 말고도 `Gin` 같은 프레임워크가 많이 쓰임.

  * Gin → 표준 라이브러리 친화적, 안정적, 커뮤니티 크다.
  * Fiber → 성능 최상위, Express.js처럼 문법 단순.
* **실시간성**(선착순 보장) 자체는 프레임워크보다는 **DB/Redis의 원자적 연산**과 **락 관리**가 핵심이에요.

👉 정리:
Fiber는 **Express.js와 비슷한 문법, fasthttp 기반 초고속 성능** 덕분에 Go에서 빠르고 직관적으로 서버를 만들 때 좋은 선택이에요.
사물함 선착순 예약 시스템에서는 Fiber가 빠른 요청 처리를 담당하고, 선착순 보장은 Redis/DB 트랜잭션으로 구현하면 돼요.
