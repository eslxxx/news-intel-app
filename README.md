# News Intel App

新闻情报聚合应用，支持多源新闻采集、AI翻译/摘要、多渠道推送。

## 功能特性

- 多源新闻采集（支持任意 RSS 源）
- AI 翻译和摘要（兼容 OpenAI API）
- 多渠道推送（邮箱、ntfy）
- HTML 邮件模板编辑器
- 定时任务调度
- Web 管理后台

## 快速部署

### 方式一：Docker Compose（推荐）

**1. 克隆项目**

```bash
git clone https://github.com/eslxxx/news-intel-app.git
cd news-intel-app
```

**2. 配置环境变量**

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的配置：

```env
OPENAI_API_KEY=your-api-key
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o-mini
```

**3. 启动服务**

```bash
docker compose up -d
```

**4. 访问应用**

打开浏览器访问 http://localhost:5555

---

### 方式二：Docker 拉取镜像

如果你不想从源码构建，可以直接拉取预构建镜像：

```bash
# 拉取镜像
docker pull ghcr.io/eslxxx/news-intel-app:latest

# 运行容器
docker run -d \
  --name news-intel-app \
  -p 5555:5555 \
  -v $(pwd)/data:/app/data \
  -e OPENAI_API_KEY=your-api-key \
  -e OPENAI_BASE_URL=https://api.openai.com/v1 \
  -e OPENAI_MODEL=gpt-4o-mini \
  ghcr.io/eslxxx/news-intel-app:latest
```

> 注意：首次使用需要在 Web 界面添加新闻源

---

### 方式三：本地开发

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
| OPENAI_API_KEY | OpenAI API 密钥 | - |
| OPENAI_BASE_URL | OpenAI API 地址（支持兼容接口） | https://api.openai.com/v1 |
| OPENAI_MODEL | 使用的模型 | gpt-4o-mini |
| DB_PATH | 数据库路径 | ./data/news.db |

### 添加新闻源

首次启动后，进入 Web 管理后台 → 新闻源管理，添加你想要订阅的 RSS 源。

**支持的源类型：**
- RSS/Atom 订阅源
- 任意支持 RSS 输出的网站

**示例 RSS 源：**

| 名称 | URL | 分类 |
|------|-----|------|
| Hacker News | https://hnrss.org/frontpage | tech |
| TechCrunch | https://techcrunch.com/feed/ | tech |
| The Verge | https://www.theverge.com/rss/index.xml | tech |
| BBC News | https://feeds.bbci.co.uk/news/world/rss.xml | international |

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/news | 获取新闻列表 |
| POST | /api/news/collect | 触发新闻采集 |
| POST | /api/news/process | 触发 AI 处理 |
| GET | /api/sources | 获取新闻源 |
| POST | /api/sources | 添加新闻源 |
| GET | /api/channels | 获取推送渠道 |
| GET | /api/tasks | 获取推送任务 |
| GET | /api/templates | 获取邮件模板 |
| GET | /api/ai/config | 获取 AI 配置 |
| GET | /api/stats | 获取统计数据 |

## 技术栈

- **后端**: Go + Fiber + SQLite
- **前端**: React + Vite + Ant Design
- **AI**: OpenAI API（支持兼容接口）
- **部署**: Docker

## 数据持久化

应用数据存储在 `./data` 目录：
- `news.db` - SQLite 数据库（新闻、配置等）

建议定期备份此目录。

## 常见问题

**Q: 如何使用国内 AI 服务？**

修改 `OPENAI_BASE_URL` 为兼容 OpenAI API 的服务地址，如：
- 阿里云通义千问
- 智谱 AI
- 其他兼容接口

**Q: 容器无法启动？**

检查端口 5555 是否被占用，或修改 `docker-compose.yml` 中的端口映射。

**Q: 如何更新应用？**

```bash
git pull
docker compose down
docker compose build --no-cache
docker compose up -d
```

## License

MIT
