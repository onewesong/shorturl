# shorturl（极简自用）

## 启动（Docker Compose：本地构建）

1. 准备环境变量（推荐在根目录创建 `.env`，`docker compose` 会自动读取；该文件已在 `.gitignore` 中忽略）：

```ini
# 对外暴露端口（宿主机端口）
PUBLISHED_PORT=38080

# 首次初始化管理员（数据库为空时必填 ADMIN_PASSWORD）
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change-me

# Cookie 签名密钥（建议随机长字符串；不设置会导致重启后登录失效）
SESSION_SECRET=change-me-too

# 容器内数据库路径（配合 volumes: ./data:/data 使用）
DB_PATH=/data/shorturl.db
```

2. 启动（本地构建镜像）：

```bash
docker compose up --build
```

- 后台：`http://localhost:<PUBLISHED_PORT>/admin/login`（`PUBLISHED_PORT` 默认 `38080`）
- 访问短链：`http://localhost:<PUBLISHED_PORT>/<code>`

如果本机端口有冲突，可以临时覆盖：

```bash
PUBLISHED_PORT=38081 docker compose up --build
```

## 启动（Docker Hub 镜像）

你已发布镜像：`asffda/shorturl:latest`。

最小可用（建议带数据卷持久化）：

```bash
docker pull asffda/shorturl:latest

docker run --rm -p 38080:8080 \
  -v "$(pwd)/data:/data" \
  -e ADMIN_USERNAME=admin \
  -e ADMIN_PASSWORD=change-me \
  -e SESSION_SECRET=change-me-too \
  -e DB_PATH=/data/shorturl.db \
  asffda/shorturl:latest
```

## 环境变量

应用支持的环境变量：

- `HOST`：默认 `0.0.0.0`
- `PORT`：默认 `8080`
- `DB_PATH`：默认 `./data/shorturl.db`（容器里通常用 `/data/shorturl.db` 并挂载数据卷）
- `ADMIN_USERNAME`：默认 `admin`（仅首次初始化生效；数据库已有用户时不会再改）
- `ADMIN_PASSWORD`：数据库为空时必填，用于初始化管理员
- `SESSION_SECRET`：用于登录 Cookie 签名；建议设置随机长字符串（不设置会导致重启后登录失效）
- `COOKIE_SECURE`：`true/false`，https 部署时建议 `true`

Docker Compose 额外使用：

- `PUBLISHED_PORT`：宿主机暴露端口（映射到容器 `8080`）
