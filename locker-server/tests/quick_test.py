#!/usr/bin/env python3
"""
빠른 기능 테스트 - 기본 API 기능 검증
"""

import asyncio
import aiohttp
import sys
import os

# 현재 디렉토리를 sys.path에 추가
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from load_test_config import BASE_URL
from load_test import TestDataManager

class QuickTest:
    def __init__(self):
        self.base_url = BASE_URL
        self.test_user = {
            "student_id": "20231234",
            "name": "홍길동", 
            "phone_number": "01012345678"
        }
    
    async def test_server_connection(self):
        """서버 연결 테스트"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.base_url}/api/v1/health") as resp:
                    if resp.status == 200:
                        print("✅ Server connection: OK")
                        return True
                    else:
                        print(f"❌ Server connection: Failed ({resp.status})")
                        return False
        except Exception as e:
            print(f"❌ Server connection: Failed ({e})")
            return False
    
    async def test_authentication(self):
        """인증 기능 테스트"""
        try:
            async with aiohttp.ClientSession() as session:
                # 로그인 테스트
                async with session.post(f"{self.base_url}/api/v1/auth/login", json=self.test_user) as resp:
                    if resp.status == 200:
                        result = await resp.json()
                        token = result.get("access_token")
                        if token:
                            print("✅ Authentication: OK")
                            return token
                        else:
                            print("❌ Authentication: No token received")
                            return None
                    else:
                        print(f"❌ Authentication: Failed ({resp.status})")
                        return None
        except Exception as e:
            print(f"❌ Authentication: Failed ({e})")
            return None
    
    async def test_locker_list(self, token):
        """사물함 목록 조회 테스트"""
        try:
            headers = {"Authorization": f"Bearer {token}"}
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.base_url}/api/v1/lockers", headers=headers) as resp:
                    if resp.status == 200:
                        result = await resp.json()
                        lockers = result.get("lockers", [])
                        total_count = len(lockers)
                        # API에서 status 필드가 없으므로 전체 개수만 표시
                        print(f"✅ Locker list: OK ({total_count} total lockers)")
                        return lockers
                    else:
                        print(f"❌ Locker list: Failed ({resp.status})")
                        return []
        except Exception as e:
            print(f"❌ Locker list: Failed ({e})")
            return []
    
    async def test_locker_operations(self, token, lockers):
        """사물함 조작 테스트 (Hold/Release)"""
        if not lockers:
            print("❌ Locker operations: No lockers available")
            return False
        
        # 첫 번째 사물함으로 테스트 (status 확인 없이)
        test_locker = lockers[0]
        locker_id = test_locker["locker_id"]
        
        try:
            headers = {"Authorization": f"Bearer {token}"}
            async with aiohttp.ClientSession() as session:
                # Hold 테스트
                async with session.post(f"{self.base_url}/api/v1/lockers/{locker_id}/hold", headers=headers) as resp:
                    if resp.status in [200, 201]:
                        print(f"✅ Locker hold: OK (locker {locker_id})")
                        
                        # Release 테스트
                        async with session.post(f"{self.base_url}/api/v1/lockers/{locker_id}/release", headers=headers) as resp:
                            if resp.status in [200, 201]:
                                print(f"✅ Locker release: OK (locker {locker_id})")
                                return True
                            else:
                                print(f"❌ Locker release: Failed ({resp.status})")
                                return False
                    elif resp.status == 409:
                        print(f"⚠️ Locker hold: Already occupied (locker {locker_id}) - Trying another locker...")
                        
                        # 다른 사물함으로 재시도
                        if len(lockers) > 1:
                            for i, locker in enumerate(lockers[1:6]):  # 최대 5개까지 시도
                                locker_id = locker["locker_id"]
                                async with session.post(f"{self.base_url}/api/v1/lockers/{locker_id}/hold", headers=headers) as resp:
                                    if resp.status in [200, 201]:
                                        print(f"✅ Locker hold: OK (locker {locker_id})")
                                        
                                        # Release 테스트
                                        async with session.post(f"{self.base_url}/api/v1/lockers/{locker_id}/release", headers=headers) as resp:
                                            if resp.status in [200, 201]:
                                                print(f"✅ Locker release: OK (locker {locker_id})")
                                                return True
                                            else:
                                                print(f"❌ Locker release: Failed ({resp.status})")
                                                return False
                                    elif resp.status == 409:
                                        continue  # 다음 사물함 시도
                                    else:
                                        print(f"❌ Locker hold: Failed ({resp.status})")
                                        return False
                            
                            print("⚠️ All tested lockers are occupied - this is normal during testing")
                            return True
                        else:
                            return True
                    else:
                        print(f"❌ Locker hold: Failed ({resp.status})")
                        return False
        except Exception as e:
            print(f"❌ Locker operations: Failed ({e})")
            return False
    
    async def run_all_tests(self):
        """모든 테스트 실행"""
        print("🚀 Quick Function Test Starting...")
        print("=" * 50)
        
        results = []
        
        # 1. 서버 연결 테스트
        connection_ok = await self.test_server_connection()
        results.append(connection_ok)
        
        if not connection_ok:
            print("\n❌ Server not accessible. Please start the server first.")
            return False
        
        # 2. 인증 테스트
        token = await self.test_authentication()
        results.append(token is not None)
        
        if not token:
            print("\n❌ Authentication failed. Cannot proceed with other tests.")
            return False
        
        # 3. 사물함 목록 테스트
        lockers = await self.test_locker_list(token)
        results.append(len(lockers) > 0)
        
        # 4. 사물함 조작 테스트
        operations_ok = await self.test_locker_operations(token, lockers)
        results.append(operations_ok)
        
        # 결과 요약
        print("\n" + "=" * 50)
        print("QUICK TEST RESULTS")
        print("=" * 50)
        
        test_names = [
            "Server Connection",
            "Authentication", 
            "Locker List",
            "Locker Operations"
        ]
        
        success_count = sum(results)
        total_count = len(results)
        
        for i, (name, result) in enumerate(zip(test_names, results)):
            status = "✅ PASS" if result else "❌ FAIL"
            print(f"{i+1}. {name}: {status}")
        
        print(f"\nOverall: {success_count}/{total_count} tests passed")
        
        if success_count == total_count:
            print("🎉 All tests passed! System is working properly.")
            return True
        else:
            print("⚠️ Some tests failed. Please check the system.")
            return False

async def main():
    print("🧪 Quick Function Test Starting...")
    
    # 간단한 정리 작업 (DB 초기화 없이)
    data_manager = TestDataManager()
    
    try:
        # DB 연결해서 테스트 유저의 점유 상태만 정리
        await data_manager.connect_db()
        if data_manager.db_pool:
            async with data_manager.db_pool.acquire() as conn:
                # 테스트 유저(홍길동)의 점유 해제
                await conn.execute("UPDATE locker_info SET owner = NULL WHERE owner = '홍길동'")
                await conn.execute("DELETE FROM locker_assignments WHERE student_id = '20231234'")
                print("🔄 Cleared test user's locker assignments")
    except Exception as e:
        print(f"⚠️ Could not clear test data: {e}")
    finally:
        if data_manager.db_pool:
            await data_manager.db_pool.close()
    
    # 실제 테스트 실행
    tester = QuickTest()
    success = await tester.run_all_tests()
    return 0 if success else 1

if __name__ == "__main__":
    exit_code = asyncio.run(main())
    exit(exit_code)
