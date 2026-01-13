package main

import (
	"log"

	"news-intel-app/internal/api"
	"news-intel-app/internal/config"
	"news-intel-app/internal/database"
	"news-intel-app/internal/scheduler"
	"news-intel-app/internal/services/ai"
	"news-intel-app/internal/services/collector"
	"news-intel-app/internal/services/pusher"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()

	// 初始化默认新闻源
	if err := collector.InitDefaultSources(); err != nil {
		log.Printf("Failed to init default sources: %v", err)
	}

	// 初始化服务
	col := collector.New()
	aiSvc := ai.New(cfg.OpenAIKey, cfg.OpenAIBase, cfg.OpenAIModel)
	
	// 尝试从数据库加载 AI 配置（优先使用数据库配置）
	if err := aiSvc.LoadConfig(); err != nil {
		log.Printf("No AI config in database, using env vars: %v", err)
	} else {
		log.Println("AI config loaded from database")
	}
	
	push := pusher.New()

	// 初始化定时任务
	sched := scheduler.New(col, aiSvc, push)
	sched.Start()
	defer sched.Stop()

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB
	})

	// 中间件
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// 静态文件 (前端)
	app.Static("/", "./frontend/dist")

	// API 路由
	handler := api.NewHandler(col, aiSvc, push)
	handler.RegisterRoutes(app)

	// SPA fallback
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./frontend/dist/index.html")
	})

	// 启动时执行一次采集和翻译
	go func() {
		log.Println("Initial news collection...")
		newNews, err := col.CollectAll()
		if err != nil {
			log.Printf("Initial collect error: %v", err)
			return
		}
		if len(newNews) > 0 {
			log.Printf("Translating %d new news...", len(newNews))
			if err := aiSvc.ProcessAndMoveToReading(newNews); err != nil {
				log.Printf("Initial translate error: %v", err)
			}
		}
	}()

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
