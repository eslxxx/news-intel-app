# News Intel App

新闻情报聚合应用，支持多源新闻采集、AI翻译/摘要、多渠道推送。

## 功能特性

- 多源新闻采集（RSS、Hacker News、GitHub Trending等）
- AI 翻译和摘要（OpenAI）
- 多渠道推送（邮箱、ntfy）
- HTML 邮件模板编辑器
- 定时任务调度
- Web 管理后台

## 快速开始

### Docker 部署（推荐）

1. 克隆项目并配置环境变量：

```bash
cp .env.example .env
# 编辑 .env 填入你的 OpenAI API Key
```

2. 启动服务：

```bash
docker-compose up -d
```

3. 访问 http://localhost:5555

### 本地开发

**后端：**

```bash
cd backend
go mod tidy
go run ./cmd/server
```

**前端：**

```bash
cd frontend
npm install
npm run dev
```

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| PORT | 服务端口 | 5555 |
| OPENAI_API_KEY | OpenAI API密钥 | - |
| OPENAI_BASE_URL | OpenAI API地址 | https://api.openai.com/v1 |
| OPENAI_MODEL | 使用的模型 | gpt-4o-mini |
| DB_PATH | 数据库路径 | ./data/news.db |

### 默认新闻源

- Hacker News
- TechCrunch
- The Verge
- Ars Technica
- MIT Tech Review
- AI News
- GitHub Trending
- BBC News

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/news | 获取新闻列表 |
| POST | /api/news/collect | 触发新闻采集 |
| POST | /api/news/process | 触发AI处理 |
| GET | /api/sources | 获取新闻源 |
| GET | /api/channels | 获取推送渠道 |
| GET | /api/tasks | 获取推送任务 |
| GET | /api/templates | 获取邮件模板 |
| GET | /api/ai/config | 获取AI配置 |
| GET | /api/stats | 获取统计数据 |

## 技术栈

- **后端**: Go + Fiber + SQLite
- **前端**: React + Vite + Ant Design
- **AI**: OpenAI API
- **部署**: Docker

## License

MIT
