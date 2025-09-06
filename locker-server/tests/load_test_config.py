# 부하 테스트 설정 파일

# 서버 설정
BASE_URL = "http://localhost:3000"

# 데이터베이스 설정 (환경변수로 오버라이드 가능)
import os

DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = int(os.getenv("DB_PORT", "5432"))
DB_NAME = os.getenv("DB_NAME", "locker")
DB_USER = os.getenv("DB_USER", "locker")
DB_PASSWORD = os.getenv("DB_PASSWORD", "secure_password_2024")

# 부하 테스트 설정 - 대규모 부하 테스트
TOTAL_USERS = 1000        # 총 가상 사용자 수
CONCURRENT_USERS = 300    # 동시 접속 사용자 수
BATCH_DELAY = 0.0         # 배치 간 대기 시간 없음 (완전 동시)

# 테스트 환경 설정
TEST_LOCKERS_COUNT = 150  # 테스트용 추가 사물함 150개

# 타임아웃 설정
REQUEST_TIMEOUT = 30      # 개별 요청 타임아웃 (초)
TOTAL_TIMEOUT = 120       # 전체 세션 타임아웃 (초) - 대규모 테스트용

# 연결 설정
MAX_CONNECTIONS = 500     # 최대 연결 수 - 대규모 테스트용
MAX_CONNECTIONS_PER_HOST = 500  # 호스트당 최대 연결 수

# 시나리오 설정
CONFIRMATION_RATE = 1.0   # 확정 비율 (100% 확정)
USER_THINK_TIME_MIN = 0.1 # 사용자 대기 시간 최소 (초)
USER_THINK_TIME_MAX = 0.5 # 사용자 대기 시간 최대 (초)

# 결과 저장 설정
SAVE_DETAILED_RESULTS = True
RESULTS_FILENAME = "load_test_results.json"

# 결과 저장
SAVE_DETAILED_RESULTS = True  # 상세 결과를 파일로 저장할지 여부
RESULTS_FILENAME = "load_test_results.json"
