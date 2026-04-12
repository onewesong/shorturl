# shorturl

一个按 ForgeBase 架构整理过的短链项目，内置：

- Go 1.24 + Gin API 服务
- React 18 + Vite 管理后台
- SQLite 持久化
- Docker Compose 本地运行
- GitHub Actions CI

当前项目保留了短链核心能力：

- `GET /:code` 短链跳转
- 管理后台账号密码登录
- 短链列表、新建、编辑、启用/禁用、点击统计

## 目录结构

- [cmd/shorturl](./cmd/shorturl)：服务入口
- [internal/config](./internal/config)：环境变量加载
- [internal/httpapi](./internal/httpapi)：管理 API、静态托管、公开跳转路由
- [internal/links](./internal/links)：短链领域服务和类型
- [internal/store/sqlite](./internal/store/sqlite)：SQLite 仓储和管理员鉴权
- [web/admin](./web/admin)：React 管理后台
- [api_test](./api_test)：httpyac 示例请求

## 快速开始

1. 复制环境变量模板：

```bash
cp .env.example .env
```

2. 至少配置以下变量：

- `ADMIN_USERNAME`
- `ADMIN_PASSWORD`
- `SESSION_SECRET`

3. 启动服务：

```bash
make run
```

默认地址：

- API 健康检查：`http://localhost:8080/healthz`
- Admin：`http://localhost:8080/admin`
- 短链访问：`http://localhost:8080/<code>`

## 本地开发

常用命令：

- `make fmt`：格式化 Go 代码
- `make tidy`：整理 Go 依赖
- `make test`：运行 Go 测试
- `make build`：构建后端二进制到 `bin/shorturl`
- `make admin-install`：安装前端依赖
- `make admin-dev`：启动前端开发服务器
- `make admin-test`：运行前端测试
- `make admin-build`：构建管理后台静态资源
- `make docker-build`：构建 Docker 镜像
- `make docker-up`：通过 Docker Compose 启动
- `make docker-down`：停止 Docker Compose
- `make docker-logs`：查看容器日志

前端开发时，Vite 默认把 `/admin/api` 代理到 `http://localhost:${PORT:-8080}`。

## 环境变量

- `HOST`：服务监听地址，默认 `0.0.0.0`
- `PORT`：服务端口，默认 `8080`
- `DB_PATH`：SQLite 文件路径，默认 `./data/shorturl.db`
- `ADMIN_USERNAME`：数据库为空时初始化管理员用户名，默认 `admin`
- `ADMIN_PASSWORD`：数据库为空时初始化管理员密码，必填
- `SESSION_SECRET`：登录 Cookie 签名密钥，建议设置随机长字符串
- `COOKIE_SECURE`：`true/false`，HTTPS 部署时建议 `true`
- `ADMIN_STATIC_DIR`：React 管理后台构建产物路径，默认 `./web/admin/dist`

Docker Compose 额外使用：

- `PUBLISHED_PORT`：宿主机暴露端口，默认 `38080`

## 管理 API

- `POST /admin/api/v1/auth/login`
- `POST /admin/api/v1/auth/logout`
- `GET /admin/api/v1/auth/session`
- `GET /admin/api/v1/links`
- `POST /admin/api/v1/links`
- `PUT /admin/api/v1/links/:id`

接口统一返回：

```json
{
  "success": true,
  "data": {}
}
```

失败时：

```json
{
  "success": false,
  "error": "invalid_request"
}
```

## Docker Compose

```bash
docker compose up -d --build
```

说明：

- 服务监听宿主机 `${PUBLISHED_PORT:-38080}`
- SQLite 数据库存放在宿主机 `./data`
- 管理后台构建产物在镜像构建阶段自动打包

## 验证

可以直接使用 `api_test/*.http`：

- `admin.http`：登录、会话、退出
- `links.http`：列表、新建、更新
- `redirect.http`：短链跳转
