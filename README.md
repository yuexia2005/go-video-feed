# 视频 Feed 流后端服务

基于 Go + Gin + GORM + MySQL + Redis + Docker 构建的短视频后端服务，支持用户认证、视频发布与详情、Feed 流、热榜、评论、点赞等功能，并采用JWT鉴权与token失效控制，引入redis缓存，具备完善的高可用防护和容器化部署能力。

## ✨ 功能特性

- 用户注册/登录（JWT 认证）
- 视频上传（文件保存至 `./uploads`）
- Feed 流（游标分页 + Redis ZSet 缓存）
- 热榜（基于点赞数排序）
- 评论系统（发表、分页、删除）
- 点赞/取消点赞
- 删除视频（同步清理数据库、Redis、本地文件）

## 🛠 高可用亮点

- **限流**：令牌桶中间件，保护系统免受突发流量冲击。
- **防击穿**：singleflight 合并并发查询，保护数据库。
- **熔断降级**：使用 gobreaker 对 Redis 操作熔断，故障时降级到数据库。
- **优雅停机**：捕获 SIGTERM，等待现有请求完成再退出。
- **健康检查**：`/health` 接口配合 Docker `healthcheck`。
- **容器化**：Docker Compose 一键启动，`restart: always` 自动重启。

## 📦 技术栈

- **语言**：Go 1.21
- **Web 框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL 8.0
- **缓存**：Redis 7.0
- **容器化**：Docker + Docker Compose

## 🚀 快速启动

### 前置条件

- Go 1.21+
- Docker & Docker Compose（可选）

### 本地运行

```bash
git clone https://github.com/yuexia2005/studygo
cd video_feed

# 安装依赖
go mod tidy
# 修改数据库连接配置（默认为 config 或环境变量，见下文）
# 启动前请确保 MySQL 和 Redis 已启动且配置正确

# 运行
make build
make run   # 或 go run main.go

### Docker 一键部署（推荐）
make docker
该命令会自动构建镜像并启动三个容器：video_app、video_mysql、video_redis。
运行后访问 http://localhost:8085（前端页面需单独打开 videofeed.html 文件）。
停止服务：make docker-down

##环境变量
变量名	    说明	     默认值
DB_HOST	    MySQL主机	localhost
DB_PORT	    MySQL端口	3306
DB_USER	    MySQL用户名	root
DB_PASSWORD	MySQL密码	123456
DB_NAME	    数据库名	video_feed
REDIS_ADDR	Redis 地址	localhost:6379


##目录结构
.
├── main.go                 # 程序入口，优雅停机逻辑
├── go.mod / go.sum
├── Makefile                # 快捷命令
├── Dockerfile
├── docker-compose.yml
├── README.md
├── models/                 # 数据模型（User, Video, Like, Comment）
├── controllers/            # 业务控制器（Feed, Hot, Comment, Like, Video）
├── routes/                 # 路由注册与中间件
├── middleware/             # 限流、认证等中间件
├── utils/                  # JWT 工具函数
├── uploads/                # 上传的视频文件存储目录（自动创建）
└── videofeed.html          # 简易前端演示页面（可选）

##API 示例

##注册
bash
POST /register
Content-Type: application/json

{
    "username": "testuser",
    "password": "123456"
}

##登录
bash
POST /login
Content-Type: application/json

{
    "username": "testuser",
    "password": "123456"
}
##响应：
{"token": "eyJhbGciOiJIUzI1NiIsInR5..."}

##上传视频（需 token）
POST /api/video/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

title: 视频标题
description: 视频描述
video: <video_file>
##响应：
{"video_id": 1}

##获取 Feed 流（需 token）
GET /api/feed?limit=5&last_id=0
Authorization: Bearer <token>
##响应：
{"list": [...]}


##点赞/取消点赞
bash
POST /api/video/1/like
Authorization: Bearer <token>
##响应：
{"liked": true, "like_count": 1}

##获取热榜
bash
GET /api/hot?limit=10
Authorization: Bearer <token>
响应：
{"list": [...]}

更多接口（评论、删除视频等）请参阅源代码中的路由定义。

##🤝 贡献
欢迎提交 Issue 或 Pull Request。
