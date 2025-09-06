# 사물함 예약 시스템 부하 테스트

1000명 동시 접속을 시뮬레이션하는 부하 테스트 스크립트입니다.

## 🚀 빠른 시작

### 1. 자동 실행 (권장)
```bash
./run_load_test.sh
```

### 2. 수동 실행
```bash
# 의존성 설치
pip install aiohttp asyncpg

# 테스트 실행
python3 load_test.py
```

## 🗄️ 데이터베이스 관리

테스트 스크립트는 **자동으로 1000명의 테스트 사용자를 생성**하고, **테스트 완료 후 DB를 원래 상태로 복원**합니다.

### 자동 DB 관리 기능:
1. **기존 데이터 백업** - 테스트 전 현재 DB 상태 저장
2. **테스트 사용자 생성** - student_id: 20231000 ~ 20232999
3. **테스트 실행** - 생성된 사용자로 부하 테스트
4. **데이터 정리** - 테스트 데이터 삭제 및 원본 복원

### DB 연결 설정:
환경변수로 DB 설정을 변경할 수 있습니다:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=locker_db
export DB_USER=postgres
export DB_PASSWORD=password

python3 load_test.py
```

## ⚙️ 설정

`load_test_config.py` 파일에서 테스트 설정을 변경할 수 있습니다:

```python
# 서버 설정
BASE_URL = "http://localhost:3000"

# 부하 테스트 설정
TOTAL_USERS = 1000        # 총 사용자 수
CONCURRENT_USERS = 50     # 동시 접속 사용자 수
CONFIRMATION_RATE = 0.5   # 사물함 확정 확률

# 타임아웃 설정
REQUEST_TIMEOUT = 30      # 요청 타임아웃 (초)
```

## 📊 테스트 시나리오

각 가상 사용자는 다음 순서로 작업을 수행합니다:

1. **로그인** - 랜덤한 사용자 정보로 로그인
2. **사물함 목록 조회** - 사용 가능한 사물함 확인
3. **사물함 선점** - 랜덤한 사물함을 선택하여 hold
4. **대기** - 0.5~2초 랜덤 대기 (실제 사용자 행동 시뮬레이션)
5. **확정/포기** - 50% 확률로 확정하거나 포기
6. **내 사물함 조회** - 확정한 경우 내 사물함 정보 조회

## 📈 결과 분석

테스트 완료 후 다음 정보를 제공합니다:

### 전체 통계
- 총 요청 수
- 성공/실패 요청 수
- 성공률 (%)
- 평균 처리량 (requests/second)

### 응답 시간 통계
- 평균, 중앙값, 최소/최대값
- 95th 백분위수

### 엔드포인트별 통계
- 각 API 엔드포인트별 성공률과 평균 응답 시간

### 오류 분석
- 실패한 요청들의 상세 정보
- 오류 유형별 발생 빈도

## 📋 결과 파일

`load_test_results.json` 파일에 상세한 테스트 결과가 저장됩니다:

```json
{
  "config": { ... },
  "summary": { ... },
  "detailed_results": [ ... ]
}
```

## 🎯 성능 목표

### 권장 성능 지표
- **응답 시간**: 평균 < 1초, 95th < 3초
- **성공률**: > 95%
- **처리량**: > 100 requests/second
- **동시 접속**: 100명 이상 안정적 처리

### 병목 지점 분석
- DB 연결 풀 크기
- Redis 연결 수
- 네트워크 대역폭
- 서버 CPU/메모리 사용률

## 🔧 문제 해결

### 연결 오류가 많이 발생하는 경우
```python
# load_test_config.py에서 동시 접속 수 줄이기
CONCURRENT_USERS = 25
MAX_CONNECTIONS = 100
```

### 타임아웃이 많이 발생하는 경우
```python
# 타임아웃 시간 늘리기
REQUEST_TIMEOUT = 60
TOTAL_TIMEOUT = 120
```

### 서버 부하가 너무 높은 경우
```python
# 배치 간 대기 시간 늘리기
BATCH_DELAY = 2.0
USER_THINK_TIME_MIN = 1.0
USER_THINK_TIME_MAX = 3.0
```

## 📝 로그 분석

테스트 실행 중 다음과 같은 로그가 출력됩니다:

```
User 42: Starting scenario
User 42: Login successful
User 42: Got 15 lockers
User 42: Attempting to hold locker 105
User 42: Successfully held locker 105
User 42: Confirmed locker 105
```

## 🚨 주의사항

1. **서버 준비**: 테스트 전에 서버와 DB, Redis가 실행 중인지 확인
2. **DB 권한**: 테스트 사용자는 DB에 CREATE, INSERT, DELETE 권한이 필요
3. **네트워크**: 로컬 네트워크에서 테스트할 것을 권장
4. **리소스**: 테스트 머신의 CPU/메모리 사용률 모니터링
5. **안전성**: 테스트 완료 시 자동으로 DB가 원래 상태로 복원됨

### DB 없이 테스트하기:
asyncpg가 설치되지 않았거나 DB 연결에 실패해도 테스트는 계속 진행됩니다.
이 경우 기존 DB의 사용자로 테스트가 수행됩니다.

## 🔄 반복 테스트

동일한 조건에서 여러 번 테스트하여 일관성을 확인하세요:

```bash
# 3번 연속 실행
for i in {1..3}; do
    echo "=== Test Run $i ==="
    python3 load_test.py
    sleep 30  # 서버 안정화 대기
done
```
