package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"news-intel-app/internal/database"
	"news-intel-app/internal/models"
	"news-intel-app/internal/services/ai"
	"news-intel-app/internal/services/collector"
	"news-intel-app/internal/services/pusher"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	collector *collector.Collector
	ai        *ai.AIService
	pusher    *pusher.Pusher
}

func NewHandler(col *collector.Collector, aiSvc *ai.AIService, push *pusher.Pusher) *Handler {
	return &Handler{
		collector: col,
		ai:        aiSvc,
		pusher:    push,
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	// 新闻相关
	api.Get("/news", h.GetNews)
	api.Get("/news/:id", h.GetNewsDetail)
	api.Delete("/news/:id", h.DeleteNews)
	api.Post("/news/collect", h.TriggerCollect)
	api.Post("/news/process", h.TriggerProcess)

	// 阅读窗口
	api.Get("/reading", h.GetReadingNews)
	api.Post("/reading/:id/add", h.AddToReading)
	api.Post("/reading/:id/remove", h.RemoveFromReading)
	api.Post("/reading/clear-pushed", h.ClearPushedNews)

	// 新闻源
	api.Get("/sources", h.GetSources)
	api.Post("/sources", h.CreateSource)
	api.Put("/sources/:id", h.UpdateSource)
	api.Delete("/sources/:id", h.DeleteSource)

	// 推送渠道
	api.Get("/channels", h.GetChannels)
	api.Post("/channels", h.CreateChannel)
	api.Put("/channels/:id", h.UpdateChannel)
	api.Delete("/channels/:id", h.DeleteChannel)
	api.Post("/channels/:id/test", h.TestChannel)

	// 推送任务
	api.Get("/tasks", h.GetTasks)
	api.Post("/tasks", h.CreateTask)
	api.Put("/tasks/:id", h.UpdateTask)
	api.Delete("/tasks/:id", h.DeleteTask)
	api.Post("/tasks/:id/run", h.RunTask)

	// 自动打包推送
	api.Get("/auto-push/config", h.GetAutoPushConfig)
	api.Post("/auto-push/config", h.SaveAutoPushConfig)
	api.Get("/auto-push/status", h.GetAutoPushStatus)

	// 邮件模板
	api.Get("/templates", h.GetTemplates)
	api.Post("/templates", h.CreateTemplate)
	api.Put("/templates/:id", h.UpdateTemplate)
	api.Delete("/templates/:id", h.DeleteTemplate)
	api.Post("/templates/preview", h.PreviewTemplate)
	api.Post("/templates/ai-generate", h.AIGenerateTemplate)

	// AI配置
	api.Get("/ai/config", h.GetAIConfig)
	api.Post("/ai/config", h.SaveAIConfig)
	api.Post("/ai/translate", h.TranslateText)
	api.Post("/ai/summarize", h.SummarizeText)

	// 统计
	api.Get("/stats", h.GetStats)
}

// ========== 新闻相关 ==========

func (h *Handler) GetNews(c *fiber.Ctx) error {
	category := c.Query("category")
	source := c.Query("source")
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	query := `SELECT id, title, content, summary, url, source, category, image_url, author, 
		published_at, created_at, translated, trans_title, trans_content, trans_summary, is_filtered, tags 
		FROM news WHERE is_filtered = 0`
	args := []interface{}{}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if source != "" {
		query += " AND source = ?"
		args = append(args, source)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var news []models.News
	for rows.Next() {
		var n models.News
		var publishedAt, createdAt sql.NullTime
		var tags, transTitle, transContent, transSummary, content, summary, imageURL, author sql.NullString
		err := rows.Scan(&n.ID, &n.Title, &content, &summary, &n.URL, &n.Source, &n.Category,
			&imageURL, &author, &publishedAt, &createdAt, &n.Translated, &transTitle, &transContent, &transSummary, &n.IsFiltered, &tags)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		if publishedAt.Valid {
			n.PublishedAt = publishedAt.Time
		}
		if createdAt.Valid {
			n.CreatedAt = createdAt.Time
		}
		if tags.Valid {
			n.Tags = tags.String
		}
		if transTitle.Valid {
			n.TransTitle = transTitle.String
		}
		if transContent.Valid {
			n.TransContent = transContent.String
		}
		if transSummary.Valid {
			n.TransSummary = transSummary.String
		}
		if content.Valid {
			n.Content = content.String
		}
		if summary.Valid {
			n.Summary = summary.String
		}
		if imageURL.Valid {
			n.ImageURL = imageURL.String
		}
		if author.Valid {
			n.Author = author.String
		}
		news = append(news, n)
	}

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM news WHERE is_filtered = 0"
	if category != "" {
		countQuery += " AND category = '" + category + "'"
	}
	database.DB.QueryRow(countQuery).Scan(&total)

	return c.JSON(fiber.Map{
		"data":  news,
		"total": total,
	})
}

func (h *Handler) GetNewsDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	var n models.News
	var publishedAt, createdAt sql.NullTime
	
	err := database.DB.QueryRow(`
		SELECT id, title, content, summary, url, source, category, image_url, author,
		published_at, created_at, translated, trans_title, trans_content, trans_summary, is_filtered, tags
		FROM news WHERE id = ?
	`, id).Scan(&n.ID, &n.Title, &n.Content, &n.Summary, &n.URL, &n.Source, &n.Category,
		&n.ImageURL, &n.Author, &publishedAt, &createdAt, &n.Translated, &n.TransTitle, &n.TransContent, &n.TransSummary, &n.IsFiltered, &n.Tags)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "News not found"})
	}

	if publishedAt.Valid {
		n.PublishedAt = publishedAt.Time
	}
	if createdAt.Valid {
		n.CreatedAt = createdAt.Time
	}

	return c.JSON(n)
}

func (h *Handler) DeleteNews(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := database.DB.Exec("DELETE FROM news WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) TriggerCollect(c *fiber.Ctx) error {
	go func() {
		newNews, err := h.collector.CollectAll()
		if err != nil {
			log.Printf("Collect error: %v", err)
			return
		}
		// 采集后立即翻译
		if len(newNews) > 0 {
			log.Printf("Translating %d new news...", len(newNews))
			if err := h.ai.ProcessAndMoveToReading(newNews); err != nil {
				log.Printf("Translate error: %v", err)
			}
		}
	}()
	return c.JSON(fiber.Map{"message": "Collection and translation started"})
}

func (h *Handler) TriggerProcess(c *fiber.Ctx) error {
	go func() {
		if err := h.ai.ProcessUnprocessedNews(10); err != nil {
			log.Printf("Process error: %v", err)
		}
	}()
	return c.JSON(fiber.Map{"message": "Processing started"})
}

// ========== 阅读窗口相关 ==========

func (h *Handler) GetReadingNews(c *fiber.Ctx) error {
	category := c.Query("category")
	pushed := c.Query("pushed") // "all", "yes", "no"
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	query := `SELECT id, title, content, summary, url, source, category, image_url, author, 
		published_at, created_at, translated, trans_title, trans_content, trans_summary, 
		is_filtered, tags, in_reading, reading_at, pushed, pushed_at 
		FROM news WHERE in_reading = 1`
	args := []interface{}{}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if pushed == "yes" {
		query += " AND pushed = 1"
	} else if pushed == "no" {
		query += " AND pushed = 0"
	}

	query += " ORDER BY reading_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var news []models.News
	for rows.Next() {
		var n models.News
		var publishedAt, createdAt, readingAt, pushedAt sql.NullTime
		var tags, transTitle, transContent, transSummary, content, summary, imageURL, author sql.NullString
		err := rows.Scan(&n.ID, &n.Title, &content, &summary, &n.URL, &n.Source, &n.Category,
			&imageURL, &author, &publishedAt, &createdAt, &n.Translated, &transTitle, &transContent, &transSummary,
			&n.IsFiltered, &tags, &n.InReading, &readingAt, &n.Pushed, &pushedAt)
		if err != nil {
			log.Printf("Scan reading news error: %v", err)
			continue
		}
		if publishedAt.Valid {
			n.PublishedAt = publishedAt.Time
		}
		if createdAt.Valid {
			n.CreatedAt = createdAt.Time
		}
		if readingAt.Valid {
			n.ReadingAt = readingAt.Time
		}
		if pushedAt.Valid {
			n.PushedAt = pushedAt.Time
		}
		if tags.Valid {
			n.Tags = tags.String
		}
		if transTitle.Valid {
			n.TransTitle = transTitle.String
		}
		if transContent.Valid {
			n.TransContent = transContent.String
		}
		if transSummary.Valid {
			n.TransSummary = transSummary.String
		}
		if content.Valid {
			n.Content = content.String
		}
		if summary.Valid {
			n.Summary = summary.String
		}
		if imageURL.Valid {
			n.ImageURL = imageURL.String
		}
		if author.Valid {
			n.Author = author.String
		}
		news = append(news, n)
	}

	// 获取总数
	var total, unpushedCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM news WHERE in_reading = 1").Scan(&total)
	database.DB.QueryRow("SELECT COUNT(*) FROM news WHERE in_reading = 1 AND pushed = 0").Scan(&unpushedCount)

	return c.JSON(fiber.Map{
		"data":           news,
		"total":          total,
		"unpushed_count": unpushedCount,
	})
}

func (h *Handler) AddToReading(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := database.DB.Exec("UPDATE news SET in_reading = 1, reading_at = ? WHERE id = ?", time.Now(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) RemoveFromReading(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := database.DB.Exec("UPDATE news SET in_reading = 0 WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) ClearPushedNews(c *fiber.Ctx) error {
	_, err := database.DB.Exec("UPDATE news SET in_reading = 0 WHERE in_reading = 1 AND pushed = 1")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Cleared pushed news from reading"})
}

// ========== 新闻源相关 ==========

func (h *Handler) GetSources(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT id, name, type, url, category, enabled, interval_mins, created_at FROM news_sources ORDER BY created_at DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var sources []models.NewsSource
	for rows.Next() {
		var s models.NewsSource
		rows.Scan(&s.ID, &s.Name, &s.Type, &s.URL, &s.Category, &s.Enabled, &s.Interval, &s.CreatedAt)
		sources = append(sources, s)
	}

	return c.JSON(sources)
}

func (h *Handler) CreateSource(c *fiber.Ctx) error {
	var s models.NewsSource
	if err := c.BodyParser(&s); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	s.ID = uuid.New().String()
	s.CreatedAt = time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO news_sources (id, name, type, url, category, enabled, interval_mins, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, s.ID, s.Name, s.Type, s.URL, s.Category, s.Enabled, s.Interval, s.CreatedAt)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(s)
}

func (h *Handler) UpdateSource(c *fiber.Ctx) error {
	id := c.Params("id")
	var s models.NewsSource
	if err := c.BodyParser(&s); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	_, err := database.DB.Exec(`
		UPDATE news_sources SET name = ?, type = ?, url = ?, category = ?, enabled = ?, interval_mins = ?, updated_at = ?
		WHERE id = ?
	`, s.Name, s.Type, s.URL, s.Category, s.Enabled, s.Interval, time.Now(), id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) DeleteSource(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := database.DB.Exec("DELETE FROM news_sources WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

// ========== 推送渠道相关 ==========

func (h *Handler) GetChannels(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT id, name, type, config, enabled, created_at FROM push_channels ORDER BY created_at DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var channels []models.PushChannel
	for rows.Next() {
		var ch models.PushChannel
		rows.Scan(&ch.ID, &ch.Name, &ch.Type, &ch.Config, &ch.Enabled, &ch.CreatedAt)
		channels = append(channels, ch)
	}

	return c.JSON(channels)
}

func (h *Handler) CreateChannel(c *fiber.Ctx) error {
	var ch models.PushChannel
	if err := c.BodyParser(&ch); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ch.ID = uuid.New().String()
	ch.CreatedAt = time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO push_channels (id, name, type, config, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, ch.ID, ch.Name, ch.Type, ch.Config, ch.Enabled, ch.CreatedAt)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(ch)
}

func (h *Handler) UpdateChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	var ch models.PushChannel
	if err := c.BodyParser(&ch); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	_, err := database.DB.Exec(`
		UPDATE push_channels SET name = ?, type = ?, config = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`, ch.Name, ch.Type, ch.Config, ch.Enabled, time.Now(), id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) DeleteChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := database.DB.Exec("DELETE FROM push_channels WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) TestChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var ch models.PushChannel
	err := database.DB.QueryRow("SELECT id, name, type, config FROM push_channels WHERE id = ?", id).
		Scan(&ch.ID, &ch.Name, &ch.Type, &ch.Config)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Channel not found"})
	}

	switch ch.Type {
	case "email":
		var config models.EmailConfig
		json.Unmarshal([]byte(ch.Config), &config)
		err = h.pusher.SendEmail(&config, "测试邮件 - News Intel", "<h1>测试成功!</h1><p>您的邮件配置正确。</p>")
	case "ntfy":
		var config models.NtfyConfig
		json.Unmarshal([]byte(ch.Config), &config)
		err = h.pusher.SendNtfy(&config, "测试通知", "您的ntfy配置正确!")
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Test sent successfully"})
}

// ========== 推送任务相关 ==========

func (h *Handler) GetTasks(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT id, name, cron_expr, channel_id, template_id, categories, enabled, last_run_at, created_at FROM push_tasks ORDER BY created_at DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var tasks []models.PushTask
	for rows.Next() {
		var t models.PushTask
		var lastRunAt sql.NullTime
		rows.Scan(&t.ID, &t.Name, &t.CronExpr, &t.ChannelID, &t.TemplateID, &t.Categories, &t.Enabled, &lastRunAt, &t.CreatedAt)
		if lastRunAt.Valid {
			t.LastRunAt = lastRunAt.Time
		}
		tasks = append(tasks, t)
	}

	return c.JSON(tasks)
}

func (h *Handler) CreateTask(c *fiber.Ctx) error {
	var t models.PushTask
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO push_tasks (id, name, cron_expr, channel_id, template_id, categories, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, t.ID, t.Name, t.CronExpr, t.ChannelID, t.TemplateID, t.Categories, t.Enabled, t.CreatedAt)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(t)
}

func (h *Handler) UpdateTask(c *fiber.Ctx) error {
	id := c.Params("id")
	var t models.PushTask
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	_, err := database.DB.Exec(`
		UPDATE push_tasks SET name = ?, cron_expr = ?, channel_id = ?, template_id = ?, categories = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`, t.Name, t.CronExpr, t.ChannelID, t.TemplateID, t.Categories, t.Enabled, time.Now(), id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) DeleteTask(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := database.DB.Exec("DELETE FROM push_tasks WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) RunTask(c *fiber.Ctx) error {
	id := c.Params("id")

	var t models.PushTask
	err := database.DB.QueryRow("SELECT id, name, cron_expr, channel_id, template_id, categories, enabled FROM push_tasks WHERE id = ?", id).
		Scan(&t.ID, &t.Name, &t.CronExpr, &t.ChannelID, &t.TemplateID, &t.Categories, &t.Enabled)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Task not found"})
	}

	go func() {
		if err := h.pusher.ExecutePushTask(&t); err != nil {
			log.Printf("Push task error: %v", err)
		}
		database.DB.Exec("UPDATE push_tasks SET last_run_at = ? WHERE id = ?", time.Now(), t.ID)
	}()

	return c.JSON(fiber.Map{"message": "Task started"})
}

// ========== 邮件模板相关 ==========

func (h *Handler) GetTemplates(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT id, name, subject, content, is_default, created_at FROM email_templates ORDER BY created_at DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var templates []models.EmailTemplate
	for rows.Next() {
		var t models.EmailTemplate
		rows.Scan(&t.ID, &t.Name, &t.Subject, &t.Content, &t.IsDefault, &t.CreatedAt)
		templates = append(templates, t)
	}

	return c.JSON(templates)
}

func (h *Handler) CreateTemplate(c *fiber.Ctx) error {
	var t models.EmailTemplate
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO email_templates (id, name, subject, content, is_default, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, t.ID, t.Name, t.Subject, t.Content, t.IsDefault, t.CreatedAt)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(t)
}

func (h *Handler) UpdateTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	var t models.EmailTemplate
	if err := c.BodyParser(&t); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	_, err := database.DB.Exec(`
		UPDATE email_templates SET name = ?, subject = ?, content = ?, is_default = ?, updated_at = ?
		WHERE id = ?
	`, t.Name, t.Subject, t.Content, t.IsDefault, time.Now(), id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) DeleteTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := database.DB.Exec("DELETE FROM email_templates WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) PreviewTemplate(c *fiber.Ctx) error {
	var req struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 从阅读窗口获取真实新闻数据用于预览
	rows, err := database.DB.Query(`
		SELECT id, title, content, summary, url, source, category, image_url, trans_title, trans_summary 
		FROM news WHERE in_reading = 1 AND translated = 1 
		ORDER BY reading_at DESC LIMIT 5
	`)
	
	var news []models.News
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var n models.News
			var transTitle, transSummary, content, summary, imageURL sql.NullString
			if err := rows.Scan(&n.ID, &n.Title, &content, &summary, &n.URL, &n.Source, &n.Category, &imageURL, &transTitle, &transSummary); err != nil {
				continue
			}
			if transTitle.Valid {
				n.TransTitle = transTitle.String
			}
			if transSummary.Valid {
				n.TransSummary = transSummary.String
			}
			if content.Valid {
				n.Content = content.String
			}
			if summary.Valid {
				n.Summary = summary.String
			}
			if imageURL.Valid {
				n.ImageURL = imageURL.String
			}
			news = append(news, n)
		}
	}

	// 如果没有真实数据，使用示例数据
	if len(news) == 0 {
		news = []models.News{
			{Title: "Example: OpenAI announces GPT-5", TransTitle: "示例：OpenAI 发布 GPT-5", Source: "Hacker News", Category: "tech", URL: "https://example.com/1", TransSummary: "这是一条示例新闻的摘要内容，用于模板预览。实际推送时会使用阅读窗口中的真实新闻。"},
			{Title: "Example: Apple reveals new AI features", TransTitle: "示例：苹果发布新 AI 功能", Source: "TechCrunch", Category: "ai", URL: "https://example.com/2", TransSummary: "这是另一条示例新闻的摘要内容。请先采集并翻译新闻后再预览模板。"},
		}
	}

	html, err := h.pusher.RenderTemplate(req.Content, news)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"html": html, "news_count": len(news)})
}

// AIGenerateTemplate AI生成邮件模板
func (h *Handler) AIGenerateTemplate(c *fiber.Ctx) error {
	var req struct {
		Description     string `json:"description"`
		CurrentTemplate string `json:"current_template"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "请输入模板设计需求"})
	}

	template, err := h.ai.GenerateEmailTemplate(req.Description, req.CurrentTemplate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "AI 生成失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{"template": template})
}

// ========== AI配置相关 ==========

func (h *Handler) GetAIConfig(c *fiber.Ctx) error {
	var cfg models.AIConfig
	err := database.DB.QueryRow(`
		SELECT id, provider, api_key, base_url, model, enable_trans, enable_summary, enable_filter, target_lang 
		FROM ai_configs LIMIT 1
	`).Scan(&cfg.ID, &cfg.Provider, &cfg.APIKey, &cfg.BaseURL, &cfg.Model, &cfg.EnableTrans, &cfg.EnableSummary, &cfg.EnableFilter, &cfg.TargetLang)

	if err != nil {
		// 返回默认配置
		return c.JSON(models.AIConfig{
			Provider:     "openai",
			Model:        "gpt-4o-mini",
			EnableTrans:  true,
			EnableSummary: true,
			TargetLang:   "zh-CN",
		})
	}

	return c.JSON(cfg)
}

func (h *Handler) SaveAIConfig(c *fiber.Ctx) error {
	var cfg models.AIConfig
	if err := c.BodyParser(&cfg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 先删除旧配置
	database.DB.Exec("DELETE FROM ai_configs")

	cfg.ID = uuid.New().String()
	_, err := database.DB.Exec(`
		INSERT INTO ai_configs (id, provider, api_key, base_url, model, enable_trans, enable_summary, enable_filter, target_lang)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cfg.ID, cfg.Provider, cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.EnableTrans, cfg.EnableSummary, cfg.EnableFilter, cfg.TargetLang)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 重新加载AI配置
	h.ai.LoadConfig()

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) TranslateText(c *fiber.Ctx) error {
	var req struct {
		Text       string `json:"text"`
		TargetLang string `json:"target_lang"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if req.TargetLang == "" {
		req.TargetLang = "zh-CN"
	}

	result, err := h.ai.Translate(req.Text, req.TargetLang)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"result": result})
}

func (h *Handler) SummarizeText(c *fiber.Ctx) error {
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	result, err := h.ai.Summarize(req.Text)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"result": result})
}

// ========== 统计 ==========

func (h *Handler) GetStats(c *fiber.Ctx) error {
	var totalNews, todayNews, sourcesCount, channelsCount int

	database.DB.QueryRow("SELECT COUNT(*) FROM news").Scan(&totalNews)
	database.DB.QueryRow("SELECT COUNT(*) FROM news WHERE created_at > datetime('now', '-1 day')").Scan(&todayNews)
	database.DB.QueryRow("SELECT COUNT(*) FROM news_sources WHERE enabled = 1").Scan(&sourcesCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM push_channels WHERE enabled = 1").Scan(&channelsCount)

	// 按分类统计
	rows, _ := database.DB.Query("SELECT category, COUNT(*) as count FROM news GROUP BY category")
	defer rows.Close()

	categoryStats := make(map[string]int)
	for rows.Next() {
		var cat string
		var count int
		rows.Scan(&cat, &count)
		categoryStats[cat] = count
	}

	return c.JSON(fiber.Map{
		"total_news":     totalNews,
		"today_news":     todayNews,
		"sources_count":  sourcesCount,
		"channels_count": channelsCount,
		"by_category":    categoryStats,
	})
}

// ========== 自动打包推送 ==========

func (h *Handler) GetAutoPushConfig(c *fiber.Ctx) error {
	enabled, threshold, channelID, templateID := h.pusher.GetAutoPushConfig()
	pendingCount := h.pusher.GetPendingPushCount()

	return c.JSON(fiber.Map{
		"enabled":       enabled,
		"threshold":     threshold,
		"channel_id":    channelID,
		"template_id":   templateID,
		"pending_count": pendingCount,
	})
}

func (h *Handler) SaveAutoPushConfig(c *fiber.Ctx) error {
	var req struct {
		Enabled    bool   `json:"enabled"`
		Threshold  int    `json:"threshold"`
		ChannelID  string `json:"channel_id"`
		TemplateID string `json:"template_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	enabledStr := "0"
	if req.Enabled {
		enabledStr = "1"
	}
	if req.Threshold < 1 {
		req.Threshold = 6
	}

	database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('auto_push_enabled', ?)", enabledStr)
	database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('auto_push_threshold', ?)", fmt.Sprintf("%d", req.Threshold))
	database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('auto_push_channel_id', ?)", req.ChannelID)
	database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('auto_push_template_id', ?)", req.TemplateID)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) GetAutoPushStatus(c *fiber.Ctx) error {
	enabled, threshold, _, _ := h.pusher.GetAutoPushConfig()
	pendingCount := h.pusher.GetPendingPushCount()

	return c.JSON(fiber.Map{
		"enabled":       enabled,
		"threshold":     threshold,
		"pending_count": pendingCount,
		"ready":         pendingCount >= threshold,
	})
}
