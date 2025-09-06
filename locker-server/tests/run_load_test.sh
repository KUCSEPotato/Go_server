#!/bin/bash

# 부하 테스트 실행 스크립트

echo "🚀 Setting up load test environment..."

# Python 가상환경 생성 (이미 있으면 스킵)
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# 가상환경 활성화
echo "Activating virtual environment..."
source venv/bin/activate

# 필요한 패키지 설치
echo "Installing required packages..."
pip install --upgrade pip
pip install aiohttp asyncpg

echo "✅ Environment setup complete!"
echo ""
echo "🔥 Starting load test..."
echo "Target: http://localhost:3000"
echo "Users: 1000 (50 concurrent)"
echo ""

# 부하 테스트 실행
python3 load_test.py

echo ""
echo "🏁 Load test completed!"
