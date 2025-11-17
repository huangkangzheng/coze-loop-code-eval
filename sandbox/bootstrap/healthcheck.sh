#!/bin/sh

# 健康检查: 通过 HTTP 请求检查沙箱服务是否正常运行
# 检查 /health 端点返回 200 状态码

set -e

# 使用 curl 检查健康状态
if command -v curl > /dev/null 2>&1; then
  response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null || echo "000")
  if [ "$response" = "200" ]; then
    exit 0
  else
    exit 1
  fi
else
  # 如果没有 curl,使用 wget
  if wget -q --spider http://localhost:8080/health 2>/dev/null; then
    exit 0
  else
    exit 1
  fi
fi
