#!/bin/bash
# 构建 erc20-scanner 基础镜像
# 基于 alpine:latest，预装 ca-certificates tzdata wget，避免每次容器启动都 apk add
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE_NAME="erc20-scanner-base:latest"

echo "Building ${IMAGE_NAME} ..."
docker build -t "${IMAGE_NAME}" "${SCRIPT_DIR}"

echo ""
echo "Done. Image built: ${IMAGE_NAME}"
echo ""

# 如果需要在其他机器上部署，用 save/load 分发镜像（不需要 registry）
echo "# 本机直接用 docker-compose up -d 即可"
echo ""
echo "# 如果要把镜像分发到其他机器（二选一）："
echo "#   方式1: save/load（离线，适合测试环境）"
echo "  docker save ${IMAGE_NAME} | gzip > erc20-scanner-base.tar.gz"
echo "  # 拷贝到目标机器后: docker load < erc20-scanner-base.tar.gz"
echo ""
echo "#   方式2: 推送到 registry（在线，适合生产环境）"
echo "  docker tag ${IMAGE_NAME} your-registry/erc20-scanner-base:latest"
echo "  docker push your-registry/erc20-scanner-base:latest"
