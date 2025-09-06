#!/usr/bin/env python3
"""
Race Condition 테스트 - 간단 버전
"""

import asyncio
import aiohttp
import time
import sys
import os

# 현재 디렉토리를 sys.path에 추가
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from load_test_config import BASE_URL, TOTAL_USERS
from load_test import TestDataManager

async def main():
    print("🧪 Race Condition Test Starting...")
    
    # 데이터 매니저 초기화
    data_manager = TestDataManager()
    
    try:
        # 데이터베이스 연결 및 초기화
        await data_manager.connect_db()
        
        # 데이터베이스를 깨끗한 상태로 초기화
        await data_manager.reset_database()
        
        # 기존 데이터 백업 및 테스트 데이터 생성
        await data_manager.backup_existing_data()
        
        # 점유되지 않은 기존 사물함 중 하나를 선택
        if data_manager.db_pool:
            async with data_manager.db_pool.acquire() as conn:
                # 점유되지 않은 사물함 찾기
                available_lockers = await conn.fetch(
                    "SELECT locker_id FROM locker_info WHERE owner IS NULL ORDER BY locker_id LIMIT 1"
                )
                if available_lockers:
                    target_locker = available_lockers[0]['locker_id']
                    print(f"🎯 Target locker: {target_locker} (existing available locker)")
                else:
                    # 사용 가능한 사물함이 없으면 새로 생성
                    await data_manager.create_test_lockers(1)
                    if data_manager.created_lockers:
                        target_locker = data_manager.created_lockers[0]
                        print(f"🎯 Target locker: {target_locker} (newly created)")
                    else:
                        print("❌ Failed to find or create test locker")
                        return
        else:
            print("❌ Database not connected")
            return
        
        # 테스트 사용자들 - 10명으로 확장
        test_users = [
            ("20231234", "홍길동", "01012345678"),
            ("20231235", "김철수", "01087654321"),
            ("TEST001", "테스트유저1", "01011111111"),
            ("TEST002", "테스트유저2", "01022222222"),
            ("TEST003", "테스트유저3", "01033333333"),
            ("TEST004", "테스트유저4", "01044444444"),
            ("TEST005", "테스트유저5", "01055555555"),
            ("TEST006", "테스트유저6", "01066666666"),
            ("TEST007", "테스트유저7", "01077777777"),
            ("TEST008", "테스트유저8", "01088888888"),
        ]
        
        # 테스트 사용자들을 데이터베이스에 등록
        if data_manager.db_pool:
            async with data_manager.db_pool.acquire() as conn:
                for student_id, name, phone in test_users[2:]:  # 첫 2명은 이미 존재
                    try:
                        await conn.execute(
                            "INSERT INTO users (student_id, name, phone_number) VALUES ($1, $2, $3) ON CONFLICT (student_id) DO NOTHING",
                            student_id, name, phone
                        )
                    except Exception as e:
                        print(f"⚠️ Failed to create user {student_id}: {e}")
                print(f"✅ Test users prepared")
        
        async def test_user(session, student_id, name, phone, user_id):
            # 로그인
            login_data = {"student_id": student_id, "name": name, "phone_number": phone}
            async with session.post(f"{BASE_URL}/api/v1/auth/login", json=login_data) as resp:
                if resp.status != 200:
                    print(f"User {user_id}: Login failed")
                    return False
                token = (await resp.json()).get("access_token")
            
            # 사물함 점유 시도 (더 정확한 타이밍)
            headers = {"Authorization": f"Bearer {token}"}
            
            # 모든 사용자가 동시에 시작하도록 약간의 동기화
            print(f"User {user_id}: Ready to compete...")
            
            start_time = time.time()
            async with session.post(f"{BASE_URL}/api/v1/lockers/{target_locker}/hold", headers=headers) as resp:
                end_time = time.time()
                response_time = end_time - start_time
                response_text = await resp.text()
                
                if resp.status in [200, 201]:  # 200과 201 모두 성공
                    print(f"🏆 User {user_id}: SUCCESS! Request at {end_time:.6f}, took {response_time:.3f}s")
                    return True
                else:
                    print(f"❌ User {user_id}: FAILED! Status {resp.status}, Request at {end_time:.6f}, took {response_time:.3f}s")
                    print(f"   Response: {response_text}")
                    return False
        
        # 동시 실행 - 10명이 경쟁
        async with aiohttp.ClientSession() as session:
            tasks = []
            for i in range(10):  # 10명이 경쟁
                user_data = test_users[i]
                task = test_user(session, user_data[0], user_data[1], user_data[2], i)
                tasks.append(task)
            
            print(f"🚀 Starting race with {len(tasks)} users for locker {target_locker}...")
            print("⚡ All users attempting to hold the same locker simultaneously!")
            
            # 모든 태스크를 동시에 실행
            results = await asyncio.gather(*tasks)
            
            # 결과 분석
            winners = []
            for i, result in enumerate(results):
                if result:  # 성공한 경우
                    winners.append(i)
            
            winner_count = len(winners)
            print(f"\n📊 RACE CONDITION TEST RESULTS:")
            print(f"   🏆 Winners: {winner_count}")
            print(f"   ❌ Failures: {10 - winner_count}")
            
            if winners:
                winner_list = ", ".join([f"User {i}" for i in winners])
                print(f"   🥇 Winner(s): {winner_list}")
            
            if winner_count == 1:
                print("   ✅ PERFECT: Race condition handled correctly!")
                print("   ✅ Exactly 1 user got the locker, 9 users got 409 errors!")
            elif winner_count == 0:
                print("   ⚠️  No winners - all failed!")
            else:
                print("   🚨 CRITICAL: Multiple winners - race condition NOT handled!")
                print("   🚨 This indicates a serious concurrency bug!")
    
    except Exception as e:
        print(f"💥 Test failed: {e}")
    
    finally:
        # 정리 (올바른 순서로)
        try:
            # 1. 먼저 테스트 사용자들 정리
            if hasattr(data_manager, 'db_pool') and data_manager.db_pool:
                # 테스트 사용자들 삭제
                test_user_ids = [user[0] for user in test_users[2:]]  # 원본 2명 제외
                if test_user_ids:
                    # 먼저 locker_info에서 owner 참조 제거
                    await data_manager.db_pool.execute(
                        "UPDATE locker_info SET owner = NULL WHERE owner = ANY($1)",
                        test_user_ids
                    )
                    # 할당 기록 삭제
                    await data_manager.db_pool.execute(
                        "DELETE FROM locker_assignments WHERE student_id = ANY($1)",
                        test_user_ids
                    )
                    # 토큰 삭제
                    await data_manager.db_pool.execute(
                        "DELETE FROM auth_refresh_tokens WHERE student_id = ANY($1)",
                        test_user_ids
                    )
                    # 사용자 삭제
                    await data_manager.db_pool.execute(
                        "DELETE FROM users WHERE student_id = ANY($1)",
                        test_user_ids
                    )
                    print(f"🗑️ Cleaned up {len(test_user_ids)} test users")
                
                # 기존 사물함 할당 상태 초기화
                await data_manager.db_pool.execute("UPDATE locker_info SET owner = NULL WHERE owner IS NOT NULL")
                await data_manager.db_pool.execute("DELETE FROM locker_assignments")
                print(f"🗑️ Cleaned up all locker assignments")
            
            # 2. 그 다음 일반 정리
            await data_manager.cleanup_test_data()
        except Exception as cleanup_error:
            print(f"⚠️ Manual cleanup error: {cleanup_error}")
            # 최후의 수단으로 개별 정리 시도
            try:
                if hasattr(data_manager, 'db_pool') and data_manager.db_pool:
                    # 모든 점유 상태 초기화
                    await data_manager.db_pool.execute("UPDATE locker_info SET owner = NULL WHERE owner IS NOT NULL")
                    await data_manager.db_pool.execute("DELETE FROM locker_assignments")
                    print("🧹 Force cleaned all locker states")
            except Exception as force_error:
                print(f"⚠️ Force cleanup failed: {force_error}")
        
        await data_manager.close_db()
        print("🧹 Cleanup completed")

if __name__ == "__main__":
    asyncio.run(main())
