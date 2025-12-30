# shorturl（极简自用）

## 启动（Docker）

1. 修改 `docker-compose.yml` 里的 `ADMIN_PASSWORD` / `SESSION_SECRET`
2. 启动：

```bash
docker compose up --build
```

- 后台：`http://localhost:<PUBLISHED_PORT>/admin/`（默认 `8080`）
- 访问短链：`http://localhost:<PUBLISHED_PORT>/<code>`

如果本机 `8080` 端口已占用，可以这样启动：

```bash
PUBLISHED_PORT=38080 docker compose up --build
```

## 本地开发（可选）

- 构建后台前端：`npm --prefix web install && npm --prefix web run build`

## 环境变量

- `PORT`：默认 `8080`
- `DB_PATH`：默认 `./data/shorturl.db`
- `ADMIN_USERNAME`：默认 `admin`（仅首次初始化生效）
- `ADMIN_PASSWORD`：首次启动数据库为空时必填
- `SESSION_SECRET`：用于登录 Cookie 签名；建议设置一个随机长字符串
- `COOKIE_SECURE`：`true/false`，https 部署时建议 `true`
