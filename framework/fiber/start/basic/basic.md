# basic
- routing: 애플리케이션의 특정 엔드포인트에 대한 클라이언트 요청에 어떻게 응답할지 결정하는 것을 가리킴.
    - uri (또는 경로)와 http request method
    - 각 경로에는 경로가 일치할 때 실행되는 여러 개의 핸들러 함수가 존재
    - 기본적인 경로 정의는 아래와 같음.
    ``` go
    // Funtion signature
    app.Method(path string, ...func(*fiber.Ctx) error)
    ```
        - app: Fiber instance
        - Method: HTTP method
        - path: virtual path of server
        - func(*fiber.Ctx) error: 경로가 일치할 때 실행되는 컨텍스트를 포함하는 콜백합수
## 추가적인 예시 코드
``` go
// Respond with "Hello, World!" on root path, "/"
app.Get("/", func(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
})

// GET http://localhost:8080/hello%20world
app.Get("/:value", func(c *fiber.Ctx) error {
	return c.SendString("value: " + c.Params("value"))
	// => Get request with value: hello world
})

// GET http://localhost:3000/john
app.Get("/:name?", func(c *fiber.Ctx) error {
	if c.Params("name") != "" {
		return c.SendString("Hello " + c.Params("name"))
		// => Hello john
	}
	return c.SendString("Where is john?")
})

// GET http://localhost:3000/api/user/john
app.Get("/api/*", func(c *fiber.Ctx) error {
	return c.SendString("API path: " + c.Params("*"))
	// => API path: user/john
})
```

# static file
- image, css, js 파일과 같은 정적 파일을 제공하려면 함수 핸들러를 파일이나 디렉토리 문자열로 바꿔야 한다.

``` go
app.Static(prefix, root string, config ...Static)
```

- 다음 코드는 ./public 이라는 이름의 디렉토리에 있는 파일을 제공하는 예시입니다.
``` go
app := fiber.New()
app.Static("/", "./public")
app.Listen(":3000")
```
- http://localhost:3000/hello.html
- http://localhost:3000/js/jquery.js
- http://localhost:3000/css/style.css