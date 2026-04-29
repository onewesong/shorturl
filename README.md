<div align="center">

# shorturl

一个轻量的短链服务，提供 Go API、React 管理后台、SQLite 持久化，以及面向增长场景的访问分析能力。

[![CI](https://github.com/onewesong/shorturl/actions/workflows/ci.yml/badge.svg)](https://github.com/onewesong/shorturl/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![Admin](https://img.shields.io/badge/Admin-React%2018%20%2B%20Vite-646CFF?logo=react)](./web/admin)
[![SQLite](https://img.shields.io/badge/DB-SQLite-003B57?logo=sqlite)](https://www.sqlite.org/index.html)
[![Docker](https://img.shields.io/badge/Docker-GHCR-2496ED?logo=docker)](https://github.com/onewesong/shorturl/pkgs/container/shorturl)
[![Analytics](https://img.shields.io/badge/Analytics-IP%20%2F%20UA%20%2F%20Trend-E67E22)](#访问分析)

</div>

`shorturl` 适合做内部营销跳转、渠道投放跟踪、短链管理和访问数据回看。当前版本已经支持短链访问明细、来源域名、客户端识别和近 7/30 天访问曲线。

## 功能特性

- Go 1.24 + Gin HTTP API
- React 18 + Vite 管理后台
- SQLite 持久化，单文件部署简单
- 管理后台账号密码登录
- 短链列表、新建、编辑、启用 / 禁用
- `GET /:code` 短链跳转
- 访问分析：来源 IP 脱敏、Referer、客户端、设备、系统、访问曲线
- Docker Compose 本地运行
- GitHub Actions CI

## 项目结构

- [cmd/shorturl](./cmd/shorturl): 服务入口
- [internal/config](./internal/config): 环境变量加载
- [internal/httpapi](./internal/httpapi): 管理 API、静态托管、公开跳转路由
- [internal/links](./internal/links): 短链领域服务和分析类型
- [internal/store/sqlite](./internal/store/sqlite): SQLite 仓储、管理员鉴权、访问分析存储
- [web/admin](./web/admin): React 管理后台
- [api_test](./api_test): httpyac 示例请求

## 快速开始

1. 复制环境变量模板：

```bash
cp .env.example .env
```

2. 至少确认以下变量：

- `ADMIN_USERNAME`
- `ADMIN_PASSWORD`
- `SESSION_SECRET`

3. 安装并构建管理后台：

```bash
make admin-install
make admin-build
```

4. 启动服务：

```bash
make run
```

默认地址：

- 健康检查: `http://localhost:8080/healthz`
- 管理后台: `http://localhost:8080/admin`
- 短链访问: `http://localhost:8080/<code>`

## Docker Compose

如果你想直接用容器启动：

```bash
docker compose up -d --build
```

说明：

- 服务监听宿主机 `${PUBLISHED_PORT:-38080}`
- SQLite 数据库存放在宿主机 `./data`
- 管理后台构建产物会在镜像构建阶段自动打包

GitHub Actions 镜像发布规则：

- push tag 如 `v1.2.3`: 发布 `v1.2.3`、`1.2`、`1`、`latest`
- 普通分支 push / PR: 只运行测试和镜像构建校验，不推送到 GHCR

发布示例：

```bash
git tag v0.1.0
git push origin v0.1.0
```

## 环境变量

- `HOST`: 服务监听地址，默认 `0.0.0.0`
- `PORT`: 服务端口，默认 `8080`
- `DB_PATH`: SQLite 文件路径，默认 `./data/shorturl.db`
- `ADMIN_USERNAME`: 数据库为空时初始化管理员用户名，默认 `admin`
- `ADMIN_PASSWORD`: 数据库为空时初始化管理员密码
- `SESSION_SECRET`: 登录 Cookie 签名密钥，建议使用随机长字符串
- `COOKIE_SECURE`: `true/false`，HTTPS 部署时建议设置为 `true`
- `ADMIN_STATIC_DIR`: 管理后台构建产物路径，默认 `./web/admin/dist`
- `PUBLISHED_PORT`: Docker Compose 对外暴露端口，默认 `38080`

示例：

```bash
HOST=0.0.0.0 \
PORT=8080 \
DB_PATH=./data/shorturl.db \
ADMIN_USERNAME=admin \
ADMIN_PASSWORD=change-me \
SESSION_SECRET=change-me-too \
make run
```

## 访问分析

当前管理后台已支持以下分析能力：

- 最近 7 天 / 30 天访问曲线
- 窗口点击数
- 独立 IP 数
- 最近访问时间
- 来源域名分布
- 客户端分布
- 最近访问明细

访问明细当前会记录：

- 来源 IP 脱敏值
- Referer 和 Referer Host
- User-Agent
- 客户端名称
- 客户端类型
- 设备类型
- 操作系统
- 访问时间

## 管理 API

认证相关：

- `POST /admin/api/v1/auth/login`
- `POST /admin/api/v1/auth/logout`
- `GET /admin/api/v1/auth/session`

短链管理：

- `GET /admin/api/v1/links`
- `POST /admin/api/v1/links`
- `PUT /admin/api/v1/links/:id`
- `DELETE /admin/api/v1/links/:id`

访问分析：

- `GET /admin/api/v1/links/:id/analytics?days=7`

统一返回格式：

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

## 本地开发

常用命令：

- `make fmt`: 格式化 Go 代码
- `make tidy`: 整理 Go 依赖
- `make test`: 运行 Go 测试
- `make build`: 构建后端二进制到 `bin/shorturl`
- `make run`: 本地启动服务并自动加载 `.env`
- `make admin-install`: 安装前端依赖
- `make admin-dev`: 启动前端开发服务器
- `make admin-test`: 运行前端测试
- `make admin-build`: 构建管理后台静态资源
- `make docker-build`: 构建 Docker 镜像
- `make docker-up`: 通过 Docker Compose 启动
- `make docker-down`: 停止 Docker Compose
- `make docker-logs`: 查看容器日志

前端开发时，Vite 默认把 `/admin/api` 代理到 `http://localhost:${PORT:-8080}`。

## 验证

可以直接使用 [api_test](./api_test) 目录下的示例请求：

- `admin.http`: 登录、会话、退出
- `links.http`: 列表、新建、更新
- `redirect.http`: 短链跳转

也可以直接执行：

```bash
go test ./...
cd web/admin && npm test && npm run build
```
