## errer message
- no required module provides package github.com/gofiber/fiber/v2: go.mod file not found in current directory or any parent directory
    - go module (go.mod)가 없어서 Fiber 패키지를 불러올 수 없다는 의미.
    - Go 1.16 이상에서는 기본적으로 모듈 로드를 사용하기 때문에, 프로젝트 시작 시 모듈 초기화 필요

## 해결 방법
1. 프로젝트 폴더로 진입: cd [folder]
2. Go module 초기화: go mod init [module 이름 (보통 프로젝트명이나 깃헙 레포 주소)]
    - 실행 후 go.mod 파일이 생성
3. fiber install: go get github.com/gofiber/fiber/v2
    - or go get github.com/gofiber/fiber/v2@latest
    - go.sum 파일도 생기고, fiber 라이브러리가 의존성에 추가됨.
4. 코드 실행: go run [main file name]

