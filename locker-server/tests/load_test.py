#!/usr/bin/env python3
"""
사물함 예약 시스템 부하 테스트 스크립트
1000명 동시 접속 시뮬레이션
"""

import asyncio
import aiohttp
import time
import json
import random
from dataclasses import dataclass
from typing import List, Dict, Optional
import statistics
import asyncpg
import sys
import os

# 현재 디렉토리를 sys.path에 추가
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# 설정 파일 임포트
try:
    from load_test_config import *
except ImportError:
    # 기본 설정값
    BASE_URL = "http://localhost:3000"
    TOTAL_USERS = 1000
    CONCURRENT_USERS = 50
    BATCH_DELAY = 1.0
    REQUEST_TIMEOUT = 30
    TOTAL_TIMEOUT = 60
    MAX_CONNECTIONS = 200
    MAX_CONNECTIONS_PER_HOST = 200
    CONFIRMATION_RATE = 0.5
    USER_THINK_TIME_MIN = 0.5
    USER_THINK_TIME_MAX = 2.0
    SAVE_DETAILED_RESULTS = True
    RESULTS_FILENAME = "load_test_results.json"

# DB 연결 설정 (환경변수에서 읽기)
import os
DB_HOST = os.getenv('DB_HOST', 'localhost')
DB_PORT = os.getenv('DB_PORT', '5432')
DB_NAME = os.getenv('DB_NAME', 'locker')
DB_USER = os.getenv('DB_USER', 'locker')
DB_PASSWORD = os.getenv('DB_PASSWORD', 'secure_password_2024')

@dataclass
class TestResult:
    success: bool
    status_code: int
    response_time: float
    endpoint: str
    error_message: Optional[str] = None

class TestDataManager:
    """테스트 데이터 관리 클래스"""
    
    def __init__(self):
        self.db_pool = None
        self.created_users = []
        self.created_lockers = []  # 새로 생성한 사물함 목록
        self.original_locker_assignments = []
        self.original_locker_info = []
        self.original_refresh_tokens = []
    
    async def connect_db(self):
        """DB 연결"""
        try:
            import asyncpg
            self.db_pool = await asyncpg.create_pool(
                host=DB_HOST,
                port=int(DB_PORT),
                database=DB_NAME,
                user=DB_USER,
                password=DB_PASSWORD,
                min_size=1,
                max_size=5
            )
            print("✅ Database connected successfully")
            return True
        except ImportError:
            print("❌ asyncpg not installed. DB operations will be skipped.")
            print("   Install with: pip install asyncpg")
            return False
        except Exception as e:
            print(f"❌ Database connection failed: {e}")
            return False
    
    async def backup_existing_data(self):
        """기존 데이터 백업"""
        if not self.db_pool:
            return
            
        try:
            async with self.db_pool.acquire() as conn:
                # 기존 locker_assignments 백업
                assignments = await conn.fetch("SELECT * FROM locker_assignments")
                self.original_locker_assignments = [dict(row) for row in assignments]
                
                # 기존 locker_info 백업 (owner가 있는 것들만)
                lockers = await conn.fetch("SELECT * FROM locker_info WHERE owner IS NOT NULL")
                self.original_locker_info = [dict(row) for row in lockers]
                
                # 기존 refresh_tokens 백업
                tokens = await conn.fetch("SELECT * FROM auth_refresh_tokens")
                self.original_refresh_tokens = [dict(row) for row in tokens]
                
                print(f"✅ Backed up existing data:")
                print(f"   - {len(self.original_locker_assignments)} locker assignments")
                print(f"   - {len(self.original_locker_info)} owned lockers")
                print(f"   - {len(self.original_refresh_tokens)} refresh tokens")
                
        except Exception as e:
            print(f"❌ Failed to backup existing data: {e}")
    
    async def reset_database(self):
        """데이터베이스를 깨끗한 상태로 초기화"""
        if not self.db_pool:
            print("⚠️ DB not connected. Cannot reset database.")
            return
            
        try:
            async with self.db_pool.acquire() as conn:
                async with conn.transaction():
                    print("🔄 Resetting database to clean state...")
                    
                    # 1. 모든 locker_assignments 삭제 (임시 점유 상태)
                    await conn.execute("DELETE FROM locker_assignments")
                    print("   ✅ Cleared all locker assignments")
                    
                    # 2. 모든 locker_info의 owner를 NULL로 설정 (점유 해제)
                    result = await conn.execute("UPDATE locker_info SET owner = NULL WHERE owner IS NOT NULL")
                    freed_count = int(result.split()[-1])
                    print(f"   ✅ Freed {freed_count} occupied lockers")
                    
                    # 3. 테스트 사용자들의 refresh token 삭제 (TEST로 시작하는 student_id)
                    await conn.execute("DELETE FROM auth_refresh_tokens WHERE student_id LIKE 'TEST%'")
                    
                    # 4. 테스트 사용자 삭제 (TEST로 시작하는 student_id)
                    deleted_users = await conn.execute("DELETE FROM users WHERE student_id LIKE 'TEST%'")
                    test_user_count = int(deleted_users.split()[-1]) if deleted_users.split()[-1].isdigit() else 0
                    print(f"   ✅ Deleted {test_user_count} test users")
                    
                    # 5. 테스트 사물함 삭제 (ID 9000 이상)
                    deleted_lockers = await conn.execute("DELETE FROM locker_info WHERE locker_id >= 9000")
                    test_locker_count = int(deleted_lockers.split()[-1]) if deleted_lockers.split()[-1].isdigit() else 0
                    print(f"   ✅ Deleted {test_locker_count} test lockers")
                    
                    print("🎉 Database reset completed successfully!")
                    
        except Exception as e:
            print(f"❌ Failed to reset database: {e}")
        
        # 6. Redis 캐시 정리
        await self.clear_redis_cache()
    
    async def clear_redis_cache(self):
        """Redis 캐시의 사물함 관련 데이터 정리"""
        try:
            import aioredis
            
            # Redis 연결 설정
            redis_host = os.getenv('REDIS_HOST', 'localhost')
            redis_port = os.getenv('REDIS_PORT', '6379')
            redis_password = os.getenv('REDIS_PASSWORD', '')
            
            # Redis 연결
            if redis_password:
                redis_url = f"redis://:{redis_password}@{redis_host}:{redis_port}"
            else:
                redis_url = f"redis://{redis_host}:{redis_port}"
            
            redis = aioredis.from_url(redis_url)
            
            # 사물함 관련 키들 찾기
            locker_keys = await redis.keys("locker:hold:*")
            
            if locker_keys:
                # 모든 사물함 hold 키 삭제
                await redis.delete(*locker_keys)
                print(f"   ✅ Cleared {len(locker_keys)} Redis cache entries")
            else:
                print("   ✅ No Redis cache entries to clear")
                
            await redis.close()
            
        except ImportError:
            print("   ⚠️ aioredis not installed, skipping Redis cache cleanup")
        except Exception as e:
            print(f"   ⚠️ Failed to clear Redis cache: {e}")
    
    async def create_test_users(self, num_users: int) -> List[Dict]:
        """테스트용 임시 사용자 생성 - 원본 사용자 기반으로 확장"""
        users_info = []
        
        if not self.db_pool:
            print("⚠️ DB not connected. Using existing users for test.")
            return users_info
            
        try:
            async with self.db_pool.acquire() as conn:
                # 원본 사용자들 (홍길동, 김철수)
                original_users = [
                    {"student_id": "20231234", "name": "홍길동", "phone_number": "01012345678"},
                    {"student_id": "20231235", "name": "김철수", "phone_number": "01087654321"}
                ]
                
                # 원본 사용자들을 먼저 users_info에 추가
                users_info.extend(original_users)
                
                # 추가 테스트 사용자 생성 (필요한 경우)
                if num_users > 2:
                    print(f"🔄 Creating {num_users - 2} temporary test users...")
                    
                    users_data = []
                    for i in range(2, num_users):  # 3번째부터 생성
                        student_id = f"TEST{i:04d}"  # TEST0002, TEST0003, ... 형식으로 생성
                        name = f"임시사용자{i:04d}"
                        phone = f"010{random.randint(10000000,99999999):08d}"
                        
                        user_info = {
                            "student_id": student_id,
                            "name": name,
                            "phone_number": phone
                        }
                        
                        users_data.append((student_id, name, phone))
                        users_info.append(user_info)
                        self.created_users.append(student_id)
                    
                    # 배치 삽입 (임시 사용자만)
                    if users_data:
                        await conn.executemany(
                            "INSERT INTO users (student_id, name, phone_number, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())",
                            users_data
                        )
                        print(f"✅ Created {len(users_data)} temporary users for testing")
                
                print(f"✅ Ready with {len(users_info)} users (2 original + {len(users_info)-2} temporary)")
                    
        except Exception as e:
            print(f"❌ Failed to create test users: {e}")
            # 오류 발생 시 최소한 원본 사용자들은 유지
            users_info = [
                {"student_id": "20231234", "name": "홍길동", "phone_number": "01012345678"},
                {"student_id": "20231235", "name": "김철수", "phone_number": "01087654321"}
            ]
        
        return users_info
    
    async def create_test_lockers(self, num_lockers: int = 300):
        """테스트용 사물함 생성"""
        if not self.db_pool:
            print("⚠️ DB not connected. Skipping locker creation.")
            return
            
        try:
            async with self.db_pool.acquire() as conn:
                # 기존 사물함 ID 조회 (중복 방지)
                existing_ids = await conn.fetch("SELECT locker_id FROM locker_info")
                existing_id_set = {row['locker_id'] for row in existing_ids}
                
                # 기존 위치 조회
                locations = await conn.fetch("SELECT location_id, name FROM locker_locations")
                if not locations:
                    print("❌ No locker locations found. Cannot create test lockers.")
                    return
                
                print(f"🔄 Creating {num_lockers} test lockers...")
                
                lockers_data = []
                for i in range(num_lockers):
                    # 중복되지 않는 locker_id 생성 (9000번대 사용)
                    locker_id = 9000 + i
                    while locker_id in existing_id_set:
                        locker_id += 1
                    
                    # 위치를 순환하여 할당
                    location = locations[i % len(locations)]
                    location_id = location['location_id']
                    
                    locker_data = (locker_id, location_id)
                    lockers_data.append(locker_data)
                    self.created_lockers.append(locker_id)
                    existing_id_set.add(locker_id)
                
                # 배치 삽입
                await conn.executemany(
                    "INSERT INTO locker_info (locker_id, location_id) VALUES ($1, $2)",
                    lockers_data
                )
                
                print(f"✅ Created {len(lockers_data)} test lockers (ID: {min(self.created_lockers)} ~ {max(self.created_lockers)})")
                
                # 위치별 분포 출력
                location_counts = {}
                for location in locations:
                    count = sum(1 for i in range(num_lockers) if i % len(locations) == location['location_id'] - 1)
                    if count > 0:
                        location_counts[location['name']] = count
                
                for name, count in location_counts.items():
                    print(f"   {name}: {count}개")
                
        except Exception as e:
            print(f"❌ Failed to create test lockers: {e}")
            # 오류 발생 시 생성된 사물함 목록 초기화
            self.created_lockers = []
    
    async def cleanup_test_data(self):
        """테스트 데이터 정리 - 임시 사용자 완전 삭제, 원본 사용자는 보호"""
        if not self.db_pool:
            return
            
        try:
            async with self.db_pool.acquire() as conn:
                async with conn.transaction():
                    print("🔄 Cleaning up test data...")
                    
                    # 1. 임시 사용자들과 관련된 모든 데이터 삭제
                    if self.created_users:
                        # 임시 사용자들의 refresh token 삭제
                        await conn.execute(
                            "DELETE FROM auth_refresh_tokens WHERE student_id = ANY($1)",
                            self.created_users
                        )
                        
                        # 임시 사용자들의 사물함 할당 기록 삭제
                        await conn.execute(
                            "DELETE FROM locker_assignments WHERE student_id = ANY($1)",
                            self.created_users
                        )
                        
                        # 임시 사용자들이 소유한 사물함 해제
                        await conn.execute(
                            "UPDATE locker_info SET owner = NULL WHERE owner = ANY($1)",
                            self.created_users
                        )
                        
                        # 임시 사용자들 삭제
                        await conn.execute(
                            "DELETE FROM users WHERE student_id = ANY($1)",
                            self.created_users
                        )
                        
                        print(f"🗑️ Deleted {len(self.created_users)} temporary users")
                    
                    # 2. 테스트로 생성된 사물함 삭제
                    if self.created_lockers:
                        locker_ids = self.created_lockers
                        await conn.execute(
                            "DELETE FROM locker_info WHERE locker_id = ANY($1)",
                            locker_ids
                        )
                        print(f"🗑️ Deleted {len(locker_ids)} test lockers")
                    
                    # 3. 원본 사용자들의 상태 초기화 (깨끗한 상태로)
                    await conn.execute(
                        "DELETE FROM auth_refresh_tokens WHERE student_id IN ('20231234', '20231235')"
                    )
                    await conn.execute(
                        "DELETE FROM locker_assignments WHERE student_id IN ('20231234', '20231235')"
                    )
                    await conn.execute(
                        "UPDATE locker_info SET owner = NULL WHERE owner IN ('20231234', '20231235')"
                    )
                    
                    # 4. 모든 사물함의 점유 상태 완전 초기화 (혹시 남아있는 다른 점유 상태도 정리)
                    all_owned_result = await conn.fetch("SELECT COUNT(*) as count FROM locker_info WHERE owner IS NOT NULL")
                    owned_count = all_owned_result[0]['count'] if all_owned_result else 0
                    
                    if owned_count > 0:
                        await conn.execute("UPDATE locker_info SET owner = NULL WHERE owner IS NOT NULL")
                        print(f"🧹 Released {owned_count} additional occupied lockers")
                    
                    # 5. 모든 사물함 할당 기록 정리 (테스트 중 생성된 할당 기록들)
                    assignment_result = await conn.fetch("SELECT COUNT(*) as count FROM locker_assignments")
                    assignment_count = assignment_result[0]['count'] if assignment_result else 0
                    
                    if assignment_count > 0:
                        await conn.execute("DELETE FROM locker_assignments")
                        print(f"🧹 Cleared {assignment_count} locker assignment records")
                    
                    # 6. 모든 refresh token 정리 (테스트 중 생성된 토큰들)
                    token_result = await conn.fetch("SELECT COUNT(*) as count FROM auth_refresh_tokens")
                    token_count = token_result[0]['count'] if token_result else 0
                    
                    if token_count > 0:
                        await conn.execute("DELETE FROM auth_refresh_tokens")
                        print(f"🧹 Cleared {token_count} refresh tokens")
                    
                    print("✅ Database restored to original clean state")
                    print("   - Original users (홍길동, 김철수) preserved")
                    print("   - All temporary test data removed")
                    print("   - All locker occupancy states reset")
                    
        except Exception as e:
            print(f"❌ Failed to cleanup test data: {e}")
    
    async def restore_original_data(self):
        """원본 데이터베이스 상태로 정확히 복원"""
        if not self.db_pool:
            return
            
        try:
            async with self.db_pool.acquire() as conn:
                async with conn.transaction():
                    print("🔄 Restoring to original database state...")
                    
                    # 1. 모든 테스트 데이터 삭제
                    await conn.execute("DELETE FROM auth_refresh_tokens")
                    await conn.execute("DELETE FROM locker_assignments")
                    await conn.execute("UPDATE locker_info SET owner = NULL")
                    await conn.execute("DELETE FROM users WHERE student_id LIKE 'TEST%'")
                    await conn.execute("DELETE FROM locker_info WHERE locker_id >= 9000")
                    
                    # 2. 원본 사용자 복원
                    if hasattr(self, 'original_users') and self.original_users:
                        for user in self.original_users:
                            await conn.execute(
                                "INSERT INTO users (student_id, name, phone_number) VALUES ($1, $2, $3) ON CONFLICT (student_id) DO UPDATE SET name = $2, phone_number = $3",
                                user['student_id'], user['name'], user['phone_number']
                            )
                        print(f"✅ Restored {len(self.original_users)} original users")
                    
                    # 3. 원본 사물함 점유 상태 복원
                    if hasattr(self, 'original_locker_info') and self.original_locker_info:
                        for locker in self.original_locker_info:
                            await conn.execute(
                                "UPDATE locker_info SET owner = $1 WHERE locker_id = $2",
                                locker['owner'], locker['locker_id']
                            )
                        print(f"✅ Restored {len(self.original_locker_info)} original locker occupancies")
                    
                    # 4. 원본 사물함 할당 기록 복원
                    if hasattr(self, 'original_locker_assignments') and self.original_locker_assignments:
                        for assignment in self.original_locker_assignments:
                            await conn.execute(
                                "INSERT INTO locker_assignments (student_id, locker_id, assigned_at) VALUES ($1, $2, $3)",
                                assignment['student_id'], assignment['locker_id'], assignment['assigned_at']
                            )
                        print(f"✅ Restored {len(self.original_locker_assignments)} original locker assignments")
                    
                    # 5. 원본 refresh token 복원
                    if hasattr(self, 'original_refresh_tokens') and self.original_refresh_tokens:
                        for token in self.original_refresh_tokens:
                            await conn.execute(
                                "INSERT INTO auth_refresh_tokens (student_id, refresh_token, expires_at) VALUES ($1, $2, $3)",
                                token['student_id'], token['refresh_token'], token['expires_at']
                            )
                        print(f"✅ Restored {len(self.original_refresh_tokens)} original refresh tokens")
                    
                    print("🎉 Database restored to exact original state!")
                    
        except Exception as e:
            print(f"❌ Failed to restore original data: {e}")
        finally:
            # 생성된 사용자 목록 초기화
            self.created_users = []
            self.created_lockers = []
    
    async def close_db(self):
        """DB 연결 종료"""
        if self.db_pool:
            await self.db_pool.close()
            print("✅ Database connection closed")

class LoadTester:
    def __init__(self, base_url: str = "http://localhost:3000"):
        self.base_url = base_url
        self.results: List[TestResult] = []
        self.data_manager = TestDataManager()
        self.test_users = []  # 생성된 테스트 사용자 정보 저장
        self.ownership_verifications = {
            "hold_verifications": {"correct": 0, "incorrect": 0, "failed": 0},
            "confirm_verifications": {"correct": 0, "incorrect": 0, "failed": 0}
        }
        self.locker_competition_stats = {
            "hold_attempts": 0,
            "hold_successes": 0,
            "ownership_verified": 0,
            "confirm_attempts": 0,
            "confirm_successes": 0
        }
        
    async def login_user(self, session: aiohttp.ClientSession, user_id: int) -> Optional[Dict]:
        """사용자 로그인"""
        # 생성된 테스트 사용자 정보 사용
        if user_id < len(self.test_users):
            user_info = self.test_users[user_id]
            login_data = {
                "student_id": user_info["student_id"],
                "name": user_info["name"],
                "phone_number": user_info["phone_number"]
            }
        else:
            # 폴백: 랜덤 데이터 (DB 연결 실패 시)
            login_data = {
                "student_id": f"2023{1000 + user_id:04d}",
                "name": f"테스트유저{user_id}",
                "phone_number": f"010{random.randint(10000000,99999999):08d}"  # 하이픈 없는 형식
            }
        
        start_time = time.time()
        try:
            async with session.post(
                f"{self.base_url}/api/v1/auth/login",
                json=login_data,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                response_time = time.time() - start_time
                
                if response.status == 200:
                    data = await response.json()
                    self.results.append(TestResult(
                        success=True,
                        status_code=response.status,
                        response_time=response_time,
                        endpoint="POST /auth/login"
                    ))
                    return data
                else:
                    self.results.append(TestResult(
                        success=False,
                        status_code=response.status,
                        response_time=response_time,
                        endpoint="POST /auth/login",
                        error_message=f"Login failed: {response.status}"
                    ))
                    return None
                    
        except Exception as e:
            response_time = time.time() - start_time
            self.results.append(TestResult(
                success=False,
                status_code=0,
                response_time=response_time,
                endpoint="POST /auth/login",
                error_message=str(e)
            ))
            return None

    async def get_lockers(self, session: aiohttp.ClientSession, access_token: str) -> Optional[List]:
        """사물함 목록 조회"""
        headers = {"Authorization": f"Bearer {access_token}"}
        
        start_time = time.time()
        try:
            async with session.get(
                f"{self.base_url}/api/v1/lockers",
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                response_time = time.time() - start_time
                
                if response.status == 200:
                    data = await response.json()
                    self.results.append(TestResult(
                        success=True,
                        status_code=response.status,
                        response_time=response_time,
                        endpoint="GET /lockers"
                    ))
                    return data.get("lockers", [])
                else:
                    self.results.append(TestResult(
                        success=False,
                        status_code=response.status,
                        response_time=response_time,
                        endpoint="GET /lockers",
                        error_message=f"Get lockers failed: {response.status}"
                    ))
                    return None
                    
        except Exception as e:
            response_time = time.time() - start_time
            self.results.append(TestResult(
                success=False,
                status_code=0,
                response_time=response_time,
                endpoint="GET /lockers",
                error_message=str(e)
            ))
            return None

    async def hold_locker(self, session: aiohttp.ClientSession, access_token: str, locker_id: int) -> bool:
        """사물함 선점"""
        headers = {"Authorization": f"Bearer {access_token}"}
        
        # 통계 업데이트
        self.locker_competition_stats["hold_attempts"] += 1
        
        start_time = time.time()
        try:
            async with session.post(
                f"{self.base_url}/api/v1/lockers/{locker_id}/hold",
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                response_time = time.time() - start_time
                
                success = response.status == 201
                if success:
                    self.locker_competition_stats["hold_successes"] += 1
                
                self.results.append(TestResult(
                    success=success,
                    status_code=response.status,
                    response_time=response_time,
                    endpoint=f"POST /lockers/{locker_id}/hold",
                    error_message=None if success else f"Hold failed: {response.status}"
                ))
                return success
                
        except Exception as e:
            response_time = time.time() - start_time
            self.results.append(TestResult(
                success=False,
                status_code=0,
                response_time=response_time,
                endpoint=f"POST /lockers/{locker_id}/hold",
                error_message=str(e)
            ))
            return False

    async def confirm_locker(self, session: aiohttp.ClientSession, access_token: str, locker_id: int) -> bool:
        """사물함 확정"""
        headers = {"Authorization": f"Bearer {access_token}"}
        
        # 통계 업데이트
        self.locker_competition_stats["confirm_attempts"] += 1
        
        start_time = time.time()
        try:
            async with session.post(
                f"{self.base_url}/api/v1/lockers/{locker_id}/confirm",
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                response_time = time.time() - start_time
                
                success = response.status == 200
                if success:
                    self.locker_competition_stats["confirm_successes"] += 1
                    
                self.results.append(TestResult(
                    success=success,
                    status_code=response.status,
                    response_time=response_time,
                    endpoint=f"POST /lockers/{locker_id}/confirm",
                    error_message=None if success else f"Confirm failed: {response.status}"
                ))
                return success
                
        except Exception as e:
            response_time = time.time() - start_time
            self.results.append(TestResult(
                success=False,
                status_code=0,
                response_time=response_time,
                endpoint=f"POST /lockers/{locker_id}/confirm",
                error_message=str(e)
            ))
            return False

    async def get_my_locker(self, session: aiohttp.ClientSession, access_token: str) -> bool:
        """내 사물함 조회"""
        headers = {"Authorization": f"Bearer {access_token}"}
        
        start_time = time.time()
        try:
            async with session.get(
                f"{self.base_url}/api/v1/lockers/me",
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                response_time = time.time() - start_time
                
                success = response.status == 200
                self.results.append(TestResult(
                    success=success,
                    status_code=response.status,
                    response_time=response_time,
                    endpoint="GET /lockers/me",
                    error_message=None if success else f"Get my locker failed: {response.status}"
                ))
                return success
                
        except Exception as e:
            response_time = time.time() - start_time
            self.results.append(TestResult(
                success=False,
                status_code=0,
                response_time=response_time,
                endpoint="GET /lockers/me",
                error_message=str(e)
            ))
            return False

    async def verify_locker_ownership(self, session: aiohttp.ClientSession, access_token: str, locker_id: int, expected_student_id: str):
        """사물함 소유권 검증 - /lockers/me API 사용"""
        try:
            start_time = time.time()
            headers = {"Authorization": f"Bearer {access_token}"}
            
            # 내 사물함 조회로 실제 소유권 확인 (올바른 API 경로 사용)
            async with session.get(f"{self.base_url}/api/v1/lockers/me", headers=headers) as response:
                response_time = time.time() - start_time
                
                if response.status == 200:
                    data = await response.json()
                    my_locker = data.get("locker")  # 단일 사물함 객체 또는 null
                    
                    self.results.append(TestResult(
                        success=True,
                        status_code=response.status,
                        response_time=response_time,
                        endpoint=f"GET /lockers/me (verify {locker_id})",
                        error_message=None
                    ))
                    
                    if my_locker is not None:
                        # GetMyLocker 핸들러의 LockerResponse 구조에 맞게 수정
                        actual_locker_id = my_locker.get("locker_id")  # "id" 대신 "locker_id"
                        actual_owner = my_locker.get("owner")          # "user_id" 대신 "owner"
                        
                        # 예상한 사물함 ID와 일치하는지 확인
                        is_correct_locker = actual_locker_id == locker_id
                        is_correct_owner = actual_owner == expected_student_id
                        
                        return is_correct_locker and is_correct_owner, actual_owner, actual_locker_id
                    else:
                        # 사물함을 소유하지 않음
                        return False, None, None
                else:
                    self.results.append(TestResult(
                        success=False,
                        status_code=response.status,
                        response_time=response_time,
                        endpoint=f"GET /lockers/me (verify {locker_id})",
                        error_message=f"Verification failed: {response.status}"
                    ))
                    return False, None, None
                    
        except Exception as e:
            response_time = time.time() - start_time
            self.results.append(TestResult(
                success=False,
                status_code=0,
                response_time=response_time,
                endpoint=f"GET /lockers/me (verify {locker_id})",
                error_message=str(e)
            ))
            return False, None, None

    async def user_scenario(self, session: aiohttp.ClientSession, user_id: int):
        """단일 사용자 시나리오"""
        print(f"User {user_id}: Starting scenario")
        
        # 1. 로그인
        auth_data = await self.login_user(session, user_id)
        if not auth_data:
            print(f"User {user_id}: Login failed")
            return
            
        access_token = auth_data.get("access_token")
        if not access_token:
            print(f"User {user_id}: No access token")
            return
            
        print(f"User {user_id}: Login successful")
        
        # 2. 사물함 목록 조회
        lockers = await self.get_lockers(session, access_token)
        if not lockers:
            print(f"User {user_id}: Failed to get lockers")
            return
            
        print(f"User {user_id}: Got {len(lockers)} lockers")
        
        # 3. 랜덤한 사물함 선점 시도
        available_lockers = [l for l in lockers if l.get("owner") is None]
        if not available_lockers:
            print(f"User {user_id}: No available lockers")
            return
            
        target_locker = random.choice(available_lockers)
        locker_id = target_locker["locker_id"]
        
        print(f"User {user_id}: Attempting to hold locker {locker_id}")
        
        # 4. 사물함 선점
        hold_success = await self.hold_locker(session, access_token, locker_id)
        if not hold_success:
            print(f"User {user_id}: Failed to hold locker {locker_id}")
            return
            
        print(f"User {user_id}: Successfully held locker {locker_id}")
        
        # 5. 잠시 대기 (실제 사용자 행동 시뮬레이션)
        await asyncio.sleep(random.uniform(USER_THINK_TIME_MIN, USER_THINK_TIME_MAX))
        
        # 6. 설정된 확률로 확정 또는 포기
        if random.random() < CONFIRMATION_RATE:
            # 확정 (Hold → Confirm)
            confirm_success = await self.confirm_locker(session, access_token, locker_id)
            if confirm_success:
                print(f"User {user_id}: Confirmed locker {locker_id}")
                
                # 6.1 확정 후 소유권 검증
                if hasattr(self, 'test_users') and self.test_users and user_id < len(self.test_users):
                    expected_student_id = self.test_users[user_id]["student_id"]
                    is_correct_owner, actual_owner, actual_locker_id = await self.verify_locker_ownership(session, access_token, locker_id, expected_student_id)
                    
                    if is_correct_owner is not False and actual_owner is not None:
                        if is_correct_owner:
                            print(f"User {user_id}: ✅ Final ownership confirmed for locker {actual_locker_id} (expected: {locker_id})")
                            self.ownership_verifications["hold_verifications"]["correct"] += 1
                            self.locker_competition_stats["ownership_verified"] += 1
                        else:
                            print(f"User {user_id}: ❌ Ownership mismatch - expected locker {locker_id}, but owns {actual_locker_id}")
                            self.ownership_verifications["hold_verifications"]["incorrect"] += 1
                    else:
                        print(f"User {user_id}: ⚠️ Could not verify ownership for locker {locker_id}")
                        self.ownership_verifications["hold_verifications"]["failed"] += 1
                
                # 7. 내 사물함 조회
                await self.get_my_locker(session, access_token)
            else:
                print(f"User {user_id}: Failed to confirm locker {locker_id}")
        else:
            print(f"User {user_id}: Decided not to confirm locker {locker_id}")

    async def run_load_test(self, num_users: int = 1000, concurrent_users: int = 100):
        """부하 테스트 실행"""
        print(f"Starting load test with {num_users} users ({concurrent_users} concurrent)")
        
        connector = aiohttp.TCPConnector(limit=MAX_CONNECTIONS, limit_per_host=MAX_CONNECTIONS_PER_HOST)
        timeout = aiohttp.ClientTimeout(total=TOTAL_TIMEOUT)
        
        async with aiohttp.ClientSession(connector=connector, timeout=timeout) as session:
            # 사용자들을 배치로 나누어 실행
            for batch_start in range(0, num_users, concurrent_users):
                batch_end = min(batch_start + concurrent_users, num_users)
                batch_size = batch_end - batch_start
                
                print(f"Running batch {batch_start//concurrent_users + 1}: users {batch_start}-{batch_end-1}")
                
                # 현재 배치의 사용자들을 동시에 실행
                tasks = []
                for user_id in range(batch_start, batch_end):
                    task = asyncio.create_task(self.user_scenario(session, user_id))
                    tasks.append(task)
                
                # 배치 완료 대기
                await asyncio.gather(*tasks, return_exceptions=True)
                
                print(f"Batch {batch_start//concurrent_users + 1} completed")
                
                # 배치 간 잠시 대기 (서버 부하 조절)
                if batch_end < num_users:
                    await asyncio.sleep(BATCH_DELAY)

    def print_results(self):
        """테스트 결과 출력"""
        if not self.results:
            print("No test results to analyze")
            return
            
        print("\n" + "="*80)
        print("LOAD TEST RESULTS")
        print("="*80)
        
        # 전체 통계
        total_requests = len(self.results)
        successful_requests = sum(1 for r in self.results if r.success)
        failed_requests = total_requests - successful_requests
        success_rate = (successful_requests / total_requests) * 100 if total_requests > 0 else 0
        
        print(f"Total Requests: {total_requests}")
        print(f"Successful: {successful_requests}")
        print(f"Failed: {failed_requests}")
        print(f"Success Rate: {success_rate:.2f}%")
        
        # 응답 시간 통계
        response_times = [r.response_time for r in self.results if r.success]
        if response_times:
            print(f"\nResponse Time Statistics:")
            print(f"  Average: {statistics.mean(response_times):.3f}s")
            print(f"  Median: {statistics.median(response_times):.3f}s")
            print(f"  Min: {min(response_times):.3f}s")
            print(f"  Max: {max(response_times):.3f}s")
            print(f"  95th Percentile: {sorted(response_times)[int(len(response_times) * 0.95)]:.3f}s")
        
        # 엔드포인트별 통계
        endpoint_stats = {}
        for result in self.results:
            endpoint = result.endpoint
            if endpoint not in endpoint_stats:
                endpoint_stats[endpoint] = {"total": 0, "success": 0, "times": []}
            
            endpoint_stats[endpoint]["total"] += 1
            if result.success:
                endpoint_stats[endpoint]["success"] += 1
                endpoint_stats[endpoint]["times"].append(result.response_time)
        
        print(f"\nEndpoint Statistics:")
        print("-" * 80)
        for endpoint, stats in endpoint_stats.items():
            success_rate = (stats["success"] / stats["total"]) * 100
            avg_time = statistics.mean(stats["times"]) if stats["times"] else 0
            print(f"{endpoint:30} | {stats['success']:4}/{stats['total']:4} ({success_rate:5.1f}%) | Avg: {avg_time:.3f}s")
        
        # 실패한 요청들
        failed_results = [r for r in self.results if not r.success]
        if failed_results:
            print(f"\nFailed Requests ({len(failed_results)}):")
            print("-" * 80)
            error_counts = {}
            for result in failed_results:
                key = f"{result.endpoint} - {result.status_code}"
                error_counts[key] = error_counts.get(key, 0) + 1
            
            for error, count in sorted(error_counts.items(), key=lambda x: x[1], reverse=True):
                print(f"{error:50} | {count:4} times")
        
        # 소유권 검증 통계
        if any(v["correct"] + v["incorrect"] + v["failed"] > 0 for v in self.ownership_verifications.values()):
            print(f"\nOwnership Verification Statistics:")
            print("-" * 80)
            
            # Hold 후 검증
            hold_stats = self.ownership_verifications["hold_verifications"]
            hold_total = hold_stats["correct"] + hold_stats["incorrect"] + hold_stats["failed"]
            if hold_total > 0:
                hold_success_rate = (hold_stats["correct"] / hold_total) * 100
                print(f"After Hold    | {hold_stats['correct']:4}/{hold_total:4} ({hold_success_rate:5.1f}%) | Incorrect: {hold_stats['incorrect']:2} | Failed: {hold_stats['failed']:2}")
            
            # Confirm 후 검증  
            confirm_stats = self.ownership_verifications["confirm_verifications"]
            confirm_total = confirm_stats["correct"] + confirm_stats["incorrect"] + confirm_stats["failed"]
            if confirm_total > 0:
                confirm_success_rate = (confirm_stats["correct"] / confirm_total) * 100
                print(f"After Confirm | {confirm_stats['correct']:4}/{confirm_total:4} ({confirm_success_rate:5.1f}%) | Incorrect: {confirm_stats['incorrect']:2} | Failed: {confirm_stats['failed']:2}")
        
        # 사물함 경쟁 통계
        comp_stats = self.locker_competition_stats
        if comp_stats["hold_attempts"] > 0:
            print(f"\nLocker Competition Statistics:")
            print("-" * 80)
            hold_success_rate = (comp_stats["hold_successes"] / comp_stats["hold_attempts"]) * 100
            ownership_rate = (comp_stats["ownership_verified"] / comp_stats["hold_successes"]) * 100 if comp_stats["hold_successes"] > 0 else 0
            confirm_success_rate = (comp_stats["confirm_successes"] / comp_stats["confirm_attempts"]) * 100 if comp_stats["confirm_attempts"] > 0 else 0
            
            print(f"Hold Attempts    | {comp_stats['hold_attempts']:4} total")
            print(f"Hold Successes   | {comp_stats['hold_successes']:4}/{comp_stats['hold_attempts']:4} ({hold_success_rate:5.1f}%)")
            print(f"Ownership Verified| {comp_stats['ownership_verified']:4}/{comp_stats['hold_successes']:4} ({ownership_rate:5.1f}%)")
            print(f"Confirm Attempts | {comp_stats['confirm_attempts']:4} total")
            print(f"Confirm Successes| {comp_stats['confirm_successes']:4}/{comp_stats['confirm_attempts']:4} ({confirm_success_rate:5.1f}%)")

async def main():
    """메인 함수"""
    print("="*80)
    print("LOCKER RESERVATION SYSTEM - LOAD TEST")
    print("="*80)
    print(f"Target: {BASE_URL}")
    print(f"Total Users: {TOTAL_USERS}")
    print(f"Concurrent Users: {CONCURRENT_USERS}")
    print(f"Confirmation Rate: {CONFIRMATION_RATE*100:.0f}%")
    print("="*80)
    
    # 부하 테스트 실행
    tester = LoadTester(BASE_URL)
    start_time = time.time()
    
    try:
        # 1. DB 연결 및 데이터베이스 초기화
        print("\n📦 Setting up test environment...")
        db_connected = await tester.data_manager.connect_db()
        
        if db_connected:
            # 데이터베이스를 깨끗한 상태로 초기화
            await tester.data_manager.reset_database()
            
            # 기존 데이터 백업 및 테스트 데이터 생성
            await tester.data_manager.backup_existing_data()
            users_info = await tester.data_manager.create_test_users(TOTAL_USERS)
            await tester.data_manager.create_test_lockers(TEST_LOCKERS_COUNT)  # 설정에서 가져온 개수
            tester.test_users = users_info  # 생성된 사용자 정보를 LoadTester에 전달
        
        print(f"\n🚀 Starting load test...")
        
        # 2. 부하 테스트 실행
        await tester.run_load_test(TOTAL_USERS, CONCURRENT_USERS)
        
    except KeyboardInterrupt:
        print("\n⚠️ Test interrupted by user")
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
    finally:
        # 3. 테스트 데이터 정리 (항상 실행)
        print("\n🧹 Cleaning up test data...")
        await tester.data_manager.cleanup_test_data()
        await tester.data_manager.close_db()
    
    end_time = time.time()
    total_time = end_time - start_time
    
    print(f"\n✅ Test completed in {total_time:.2f} seconds")
    if tester.results:
        print(f"📊 Average throughput: {len(tester.results)/total_time:.2f} requests/second")
    
    # 결과 출력
    tester.print_results()
    
    # 상세 결과를 파일로 저장
    if SAVE_DETAILED_RESULTS and tester.results:
        results_data = {
            "config": {
                "base_url": BASE_URL,
                "total_users": TOTAL_USERS,
                "concurrent_users": CONCURRENT_USERS,
                "confirmation_rate": CONFIRMATION_RATE,
                "test_duration": total_time,
                "db_connected": db_connected
            },
            "summary": {
                "total_requests": len(tester.results),
                "successful_requests": sum(1 for r in tester.results if r.success),
                "failed_requests": sum(1 for r in tester.results if not r.success),
                "average_throughput": len(tester.results)/total_time if total_time > 0 else 0
            },
            "detailed_results": [
                {
                    "success": r.success,
                    "status_code": r.status_code,
                    "response_time": r.response_time,
                    "endpoint": r.endpoint,
                    "error_message": r.error_message
                } for r in tester.results
            ]
        }
        
        with open(RESULTS_FILENAME, 'w', encoding='utf-8') as f:
            json.dump(results_data, f, indent=2, ensure_ascii=False)
        
        print(f"\n📄 Detailed results saved to: {RESULTS_FILENAME}")

if __name__ == "__main__":
    asyncio.run(main())
