#!/bin/bash

# 启动应用程序
echo "Starting application..."
cd /code/darwinia-network/degov-apps/backend
./bin/server &
SERVER_PID=$!

# 等待服务器启动
sleep 3

# 测试用户注册
echo "Testing user registration..."
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "mutation { register(input: { username: \"testuser\", email: \"test@example.com\", password: \"password123\" }) { user { id username email } message } }"}' \
  "http://localhost:1096/graphql"

echo -e "\n\nTesting get users..."
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "query { users { id username email } }"}' \
  "http://localhost:1096/graphql"

# 清理
echo -e "\n\nStopping server..."
kill $SERVER_PID
