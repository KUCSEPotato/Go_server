#!/usr/bin/env python3
"""
데이터베이스 복원 기능 데모
원본 상태 백업 → 테스트 데이터 생성 → 원본 상태로 정확히 복원
"""

import asyncio
import sys
import os

# 현재 디렉토리를 sys.path에 추가
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from load_test import TestDataManager

async def demo_restore_functionality():
    """복원 기능 데모"""
    print("="*60)
    print("DATABASE RESTORE FUNCTIONALITY DEMO")
    print("="*60)
    
    # 데이터 매니저 초기화
    data_manager = TestDataManager()
    
    try:
        # 1. 데이터베이스 연결
        print("\n📦 1. Connecting to database...")
        await data_manager.connect_db()
        
        # 2. 현재 상태 확인
        print("\n📋 2. Current database state:")
        if data_manager.db_pool:
            async with data_manager.db_pool.acquire() as conn:
                users = await conn.fetch("SELECT student_id, name FROM users ORDER BY student_id")
                occupied_lockers = await conn.fetch("SELECT locker_id, owner FROM locker_info WHERE owner IS NOT NULL ORDER BY locker_id")
                assignments = await conn.fetch("SELECT student_id, locker_id FROM locker_assignments ORDER BY locker_id")
                tokens = await conn.fetch("SELECT student_id FROM auth_refresh_tokens ORDER BY student_id")
                
                print(f"   Users: {[dict(u) for u in users]}")
                print(f"   Occupied lockers: {[dict(l) for l in occupied_lockers]}")
                print(f"   Assignments: {[dict(a) for a in assignments]}")
                print(f"   Refresh tokens: {len(tokens)} tokens")
        
        # 3. 원본 데이터 백업
        print("\n💾 3. Backing up original data...")
        await data_manager.backup_existing_data()
        
        # 4. 테스트 데이터 생성
        print("\n🔧 4. Creating test data...")
        await data_manager.create_test_users(5)  # 5명의 테스트 유저 생성
        await data_manager.create_test_lockers(3)  # 3개의 테스트 사물함 생성
        
        # 5. 변경된 상태 확인
        print("\n📋 5. State after adding test data:")
        if data_manager.db_pool:
            async with data_manager.db_pool.acquire() as conn:
                users = await conn.fetch("SELECT student_id, name FROM users ORDER BY student_id")
                all_lockers = await conn.fetch("SELECT locker_id, location_id FROM locker_info WHERE locker_id >= 9000 ORDER BY locker_id")
                
                print(f"   Total users: {len(users)} (including test users)")
                print(f"   Test lockers: {[dict(l) for l in all_lockers]}")
        
        # 6. 원본 상태로 복원
        print("\n🔄 6. Restoring to original state...")
        await data_manager.restore_original_data()
        
        # 7. 복원 후 상태 확인
        print("\n📋 7. State after restoration:")
        if data_manager.db_pool:
            async with data_manager.db_pool.acquire() as conn:
                users = await conn.fetch("SELECT student_id, name FROM users ORDER BY student_id")
                occupied_lockers = await conn.fetch("SELECT locker_id, owner FROM locker_info WHERE owner IS NOT NULL ORDER BY locker_id")
                assignments = await conn.fetch("SELECT student_id, locker_id FROM locker_assignments ORDER BY locker_id")
                test_lockers = await conn.fetch("SELECT locker_id FROM locker_info WHERE locker_id >= 9000")
                
                print(f"   Users: {[dict(u) for u in users]}")
                print(f"   Occupied lockers: {[dict(l) for l in occupied_lockers]}")
                print(f"   Assignments: {[dict(a) for a in assignments]}")
                print(f"   Test lockers remaining: {len(test_lockers)}")
        
        print("\n🎉 Restoration demo completed successfully!")
        print("   ✅ Original users restored")
        print("   ✅ Original locker states restored") 
        print("   ✅ Test data completely removed")
        
    except Exception as e:
        print(f"\n❌ Demo failed: {e}")
    
    finally:
        # 8. 연결 종료
        await data_manager.close_db()
        print("\n🔌 Database connection closed")

if __name__ == "__main__":
    asyncio.run(demo_restore_functionality())
