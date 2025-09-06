#!/usr/bin/env python3
"""
디버깅용 로그인 테스트 스크립트
실제 load_test.py에서 사용하는 로직과 동일하게 테스트
"""

import asyncio
import asyncpg
import aiohttp
import json
import random

# DB 연결 설정
DB_HOST = 'localhost'
DB_PORT = 5432
DB_NAME = 'locker'
DB_USER = 'locker'
DB_PASSWORD = 'secure_password_2024'

async def create_test_user():
    """테스트 사용자 1명 생성"""
    conn = await asyncpg.connect(
        host=DB_HOST,
        port=DB_PORT,
        database=DB_NAME,
        user=DB_USER,
        password=DB_PASSWORD
    )
    
    student_id = "TEST9999"
    name = "디버그사용자"
    phone = f"010{random.randint(10000000,99999999):08d}"
    
    try:
        # 기존 사용자 삭제 (있다면)
        await conn.execute("DELETE FROM users WHERE student_id = $1", student_id)
        
        # 새 사용자 생성
        await conn.execute(
            "INSERT INTO users (student_id, name, phone_number, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())",
            student_id, name, phone
        )
        print(f"✅ Created test user: {student_id} - {name} ({phone})")
        
        # 생성 확인
        user = await conn.fetchrow("SELECT * FROM users WHERE student_id = $1", student_id)
        if user:
            print(f"✅ User confirmed in DB: {dict(user)}")
        else:
            print("❌ User not found in DB after creation")
            
        await conn.close()
        return {"student_id": student_id, "name": name, "phone_number": phone}
        
    except Exception as e:
        print(f"❌ Failed to create user: {e}")
        await conn.close()
        return None

async def test_login(user_info):
    """생성된 사용자로 로그인 테스트"""
    if not user_info:
        return False
        
    login_data = {
        "student_id": user_info["student_id"],
        "name": user_info["name"],
        "phone_number": user_info["phone_number"]
    }
    
    print(f"🔄 Attempting login with data: {json.dumps(login_data, ensure_ascii=False)}")
    
    async with aiohttp.ClientSession() as session:
        try:
            async with session.post(
                "http://localhost:3000/api/v1/auth/login",
                json=login_data,
                timeout=aiohttp.ClientTimeout(total=10)
            ) as response:
                response_text = await response.text()
                print(f"📊 Response Status: {response.status}")
                print(f"📊 Response Headers: {dict(response.headers)}")
                print(f"📊 Response Body: {response_text}")
                
                if response.status == 200:
                    print("✅ Login SUCCESS!")
                    return True
                else:
                    print(f"❌ Login FAILED with status {response.status}")
                    return False
                    
        except Exception as e:
            print(f"❌ Login request failed: {e}")
            return False

async def cleanup_test_user(student_id):
    """테스트 사용자 정리"""
    conn = await asyncpg.connect(
        host=DB_HOST,
        port=DB_PORT,
        database=DB_NAME,
        user=DB_USER,
        password=DB_PASSWORD
    )
    
    try:
        # 먼저 관련된 refresh token 삭제
        await conn.execute("DELETE FROM auth_refresh_tokens WHERE student_id = $1", student_id)
        # 그 다음 사용자 삭제
        await conn.execute("DELETE FROM users WHERE student_id = $1", student_id)
        print(f"🧹 Cleaned up test user: {student_id}")
    except Exception as e:
        print(f"⚠️ Cleanup failed: {e}")
    finally:
        await conn.close()

async def main():
    print("🔍 Starting debug login test...")
    
    # 1. 테스트 사용자 생성
    user_info = await create_test_user()
    if not user_info:
        print("❌ Cannot proceed without test user")
        return
    
    # 2. 로그인 테스트
    success = await test_login(user_info)
    
    # 3. 정리
    await cleanup_test_user(user_info["student_id"])
    
    print(f"\n🎯 Final Result: {'SUCCESS' if success else 'FAILED'}")

if __name__ == "__main__":
    asyncio.run(main())
