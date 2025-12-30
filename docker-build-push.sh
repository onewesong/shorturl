#!/usr/bin/env bash
set -euo pipefail

IMAGE_REPO="${IMAGE_REPO:-asffda/shorturl}"
VERSION="${1:-}"

usage() {
  cat <<'EOF'
用法：
  scripts/docker-build-push.sh [version]

行为：
  - 无 version：build + push <IMAGE_REPO>:latest
  - 有 version：build + push <IMAGE_REPO>:latest 以及 <IMAGE_REPO>:<version>

环境变量：
  IMAGE_REPO   镜像仓库名（默认 asffda/shorturl）

示例：
  IMAGE_REPO=asffda/shorturl scripts/docker-build-push.sh
  IMAGE_REPO=asffda/shorturl scripts/docker-build-push.sh v1.2.3
EOF
}

if [[ "${VERSION:-}" == "-h" || "${VERSION:-}" == "--help" ]]; then
  usage
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "未找到 docker，请先安装 Docker。" >&2
  exit 1
fi

if [[ ! -f "Dockerfile" ]]; then
  echo "当前目录未找到 Dockerfile，请在仓库根目录运行。" >&2
  exit 1
fi

tags=("${IMAGE_REPO}:latest")
if [[ -n "${VERSION}" ]]; then
  tags+=("${IMAGE_REPO}:${VERSION}")
fi

echo "构建镜像：${tags[*]}"
build_args=()
for t in "${tags[@]}"; do
  build_args+=( -t "$t" )
done

docker build "${build_args[@]}" .

for t in "${tags[@]}"; do
  echo "推送镜像：${t}"
  docker push "$t"
done
