#!/bin/sh

exec 2>&1
set -e

print_banner() {
  msg="$1"
  side=30
  content=" $msg "
  content_len=${#content}
  line_len=$((side * 2 + content_len))

  line=$(printf '*%.0s' $(seq 1 "$line_len"))
  side_eq=$(printf '*%.0s' $(seq 1 "$side"))

  printf "%s\n%s%s%s\n%s\n" "$line" "$side_eq" "$content" "$side_eq" "$line"
}

print_banner "Starting Sandbox Server..."

# 启动后台健康检查进程
(
  while true; do
    if sh /sandbox/bootstrap/healthcheck.sh; then
      print_banner "Sandbox Server Ready!"
      break
    else
      sleep 1
    fi
  done
)&

# 启动沙箱 HTTP 服务
# 使用 Deno 运行沙箱服务器(使用 vendor 目录中的依赖)
deno run \
  --allow-net \
  --allow-read \
  --allow-write \
  --allow-env \
  --import-map=/sandbox/vendor/import_map.json \
  /sandbox/bootstrap/sandbox_server.ts
