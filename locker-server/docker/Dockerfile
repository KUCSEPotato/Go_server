# syntax=docker/dockerfile:1

############################
# 1) Build stage
############################
FROM golang:1.23 AS builder

# 빌드에 필요한 최소 도구만
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 의존성 캐시 최적화
COPY go.mod go.sum ./
RUN go mod download

# 소스 복사
COPY . .

# 민감 ENV(예: DB_URL, JWT 등) 정의 금지!
# 빌드 관련 고정값만 사용
ENV CGO_ENABLED=0 GOOS=linux

# 메인 패키지 경로가 리포지토리 루트(./)라고 가정
# (만약 ./cmd/server 가 메인이라면 끝 경로를 ./cmd/server 로 바꿔주세요)
RUN go build -trimpath -ldflags="-s -w" -o /bin/locker-server ./cmd/server/main.go

############################
# 2) Runtime stage
############################
FROM alpine:3.20

# 런타임 필수 패키지
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g '' app

USER app
WORKDIR /app

# 실행 바이너리만 복사
COPY --from=builder /bin/locker-server ./locker-server

# 앱 기본 포트(실행 시점에 compose에서 매핑)
EXPOSE 3000

CMD ["./locker-server"]