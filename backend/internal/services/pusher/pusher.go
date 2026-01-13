package pusher

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"news-intel-app/internal/database"
	"news-intel-app/internal/models"

	"gopkg.in/gomail.v2"
)

type Pusher struct{}

func New() *Pusher {
	return &Pusher{}
}

// SendEmail 发送邮件
func (p *Pusher) SendEmail(config *models.EmailConfig, subject, htmlContent string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", config.FromName, config.FromAddress))
	
	toAddresses := strings.Split(config.ToAddresses, ",")
	m.SetHeader("To", toAddresses...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlContent)

	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.Username, config.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}

// SendNtfy 发送ntfy通知
func (p *Pusher) SendNtfy(config *models.NtfyConfig, title, message string) error {
	url := fmt.Sprintf("%s/%s", config.ServerURL, config.Topic)
	
	req, err := http.NewRequest("POST", url, strings.NewReader(message))
	if err != nil {
		return err
	}

	req.Header.Set("Title", title)
	req.Header.Set("Priority", "default")
	if config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+config.Token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ntfy error: %s", string(body))
	}

	return nil
}

// RenderTemplate 渲染邮件模板
func (p *Pusher) RenderTemplate(tmplContent string, news []models.News) (string, error) {
	tmpl, err := template.New("email").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"News":      news,
		"Date":      time.Now().Format("2006-01-02"),
		"Count":     len(news),
		"Generated": time.Now().Format("2006-01-02 15:04:05"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecutePushTask 执行推送任务 - 从阅读窗口取未推送的新闻
func (p *Pusher) ExecutePushTask(task *models.PushTask) error {
	// 获取渠道配置
	var channel models.PushChannel
	err := database.DB.QueryRow("SELECT id, name, type, config FROM push_channels WHERE id = ?", task.ChannelID).
		Scan(&channel.ID, &channel.Name, &channel.Type, &channel.Config)
	if err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}

	// 从阅读窗口获取未推送的新闻
	categories := strings.Split(task.Categories, ",")
	placeholders := make([]string, len(categories))
	args := make([]interface{}, len(categories))
	for i, c := range categories {
		placeholders[i] = "?"
		args[i] = strings.TrimSpace(c)
	}

	query := fmt.Sprintf(`
		SELECT id, title, content, summary, url, source, category, image_url, trans_title, trans_summary 
		FROM news 
		WHERE in_reading = 1 AND pushed = 0 AND translated = 1 AND category IN (%s)
		ORDER BY reading_at DESC LIMIT 20
	`, strings.Join(placeholders, ","))

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var news []models.News
	var newsIDs []string
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
		newsIDs = append(newsIDs, n.ID)
	}

	if len(news) == 0 {
		log.Println("No unpushed news in reading window")
		return nil
	}

	log.Printf("Pushing %d news from reading window...", len(news))

	var pushErr error
	switch channel.Type {
	case "email":
		pushErr = p.pushEmail(&channel, task, news)
	case "ntfy":
		pushErr = p.pushNtfy(&channel, news)
	}

	// 推送成功后标记为已推送
	if pushErr == nil {
		for _, id := range newsIDs {
			database.DB.Exec("UPDATE news SET pushed = 1, pushed_at = ? WHERE id = ?", time.Now(), id)
		}
		log.Printf("Marked %d news as pushed", len(newsIDs))
	}

	return pushErr
}

func (p *Pusher) pushEmail(channel *models.PushChannel, task *models.PushTask, news []models.News) error {
	var config models.EmailConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return err
	}

	// 获取模板
	var tmpl models.EmailTemplate
	err := database.DB.QueryRow("SELECT content, subject FROM email_templates WHERE id = ?", task.TemplateID).
		Scan(&tmpl.Content, &tmpl.Subject)
	if err != nil {
		// 使用默认模板
		tmpl.Content = GetDefaultEmailTemplate()
		tmpl.Subject = fmt.Sprintf("新闻情报日报 - %s", time.Now().Format("2006-01-02"))
	}

	htmlContent, err := p.RenderTemplate(tmpl.Content, news)
	if err != nil {
		return err
	}

	return p.SendEmail(&config, tmpl.Subject, htmlContent)
}

func (p *Pusher) pushNtfy(channel *models.PushChannel, news []models.News) error {
	var config models.NtfyConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return err
	}

	// 构建摘要消息（ntfy 支持 Markdown）
	var sb strings.Builder
	maxNews := 10
	if len(news) < maxNews {
		maxNews = len(news)
	}

	for i := 0; i < maxNews; i++ {
		n := news[i]
		title := n.TransTitle
		if title == "" {
			title = n.Title
		}
		summary := n.TransSummary
		if summary == "" {
			summary = n.Summary
		}
		// 截断摘要
		if len(summary) > 100 {
			summary = summary[:100] + "..."
		}

		sb.WriteString(fmt.Sprintf("**%d. %s**\n", i+1, title))
		if summary != "" {
			sb.WriteString(fmt.Sprintf("%s\n", summary))
		}
		sb.WriteString(fmt.Sprintf("[查看原文](%s)\n\n", n.URL))
	}

	if len(news) > maxNews {
		sb.WriteString(fmt.Sprintf("...还有 %d 条新闻", len(news)-maxNews))
	}

	// 发送单条汇总消息
	mainTitle := fmt.Sprintf("新闻日报 - 共 %d 条", len(news))
	return p.SendNtfyMarkdown(&config, mainTitle, sb.String())
}

// SendNtfyMarkdown 发送 ntfy Markdown 格式通知
func (p *Pusher) SendNtfyMarkdown(config *models.NtfyConfig, title, message string) error {
	url := fmt.Sprintf("%s/%s", config.ServerURL, config.Topic)

	req, err := http.NewRequest("POST", url, strings.NewReader(message))
	if err != nil {
		return err
	}

	req.Header.Set("Title", title)
	req.Header.Set("Priority", "default")
	req.Header.Set("Markdown", "yes") // 启用 Markdown 支持
	if config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+config.Token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ntfy error: %s", string(body))
	}

	return nil
}

// GetDefaultEmailTemplate 获取默认邮件模板
func GetDefaultEmailTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 680px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .header p { margin: 10px 0 0; opacity: 0.9; }
        .content { padding: 20px; }
        .news-item { border-bottom: 1px solid #eee; padding: 20px 0; }
        .news-item:last-child { border-bottom: none; }
        .news-title { font-size: 18px; font-weight: 600; color: #333; margin: 0 0 10px; }
        .news-title a { color: #667eea; text-decoration: none; }
        .news-title a:hover { text-decoration: underline; }
        .news-meta { font-size: 12px; color: #999; margin-bottom: 10px; }
        .news-meta span { margin-right: 15px; }
        .news-summary { color: #666; line-height: 1.8; margin: 0; }
        .category-tag { display: inline-block; background: #f0f0f0; padding: 2px 8px; border-radius: 4px; font-size: 11px; color: #666; }
        .footer { background: #fafafa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        .uyghur-text { direction: rtl; text-align: right; font-family: 'UKIJ Tuz Tom', 'UKIJ Tuz', Arial, sans-serif; }
        .bilingual { background: #f9f9f9; padding: 12px; border-radius: 8px; margin-top: 10px; }
        .bilingual .zh { margin-bottom: 8px; padding-bottom: 8px; border-bottom: 1px dashed #ddd; }
        .bilingual .ug { direction: rtl; text-align: right; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>新闻情报日报</h1>
            <p>{{.Date}} · 共 {{.Count}} 条新闻</p>
        </div>
        <div class="content">
            {{range .News}}
            <div class="news-item">
                <h2 class="news-title">
                    <a href="{{.URL}}" target="_blank">{{if .TransTitle}}{{.TransTitle}}{{else}}{{.Title}}{{end}}</a>
                </h2>
                <div class="news-meta">
                    <span class="category-tag">{{.Category}}</span>
                    <span>来源: {{.Source}}</span>
                </div>
                <div class="news-summary">{{if .TransSummary}}{{.TransSummary}}{{else}}{{.Summary}}{{end}}</div>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>由 News Intel App 自动生成于 {{.Generated}}</p>
        </div>
    </div>
</body>
</html>`
}

// GetBilingualEmailTemplate 获取双语邮件模板（中文+维吾尔语）
func GetBilingualEmailTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 720px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .header .subtitle { margin: 10px 0 0; opacity: 0.9; font-size: 14px; }
        .header .ug-title { direction: rtl; font-size: 20px; margin-top: 8px; }
        .content { padding: 20px; }
        .news-item { border-bottom: 1px solid #eee; padding: 24px 0; }
        .news-item:last-child { border-bottom: none; }
        .news-title { font-size: 18px; font-weight: 600; margin: 0 0 12px; }
        .news-title a { color: #667eea; text-decoration: none; }
        .news-title a:hover { text-decoration: underline; }
        .news-meta { font-size: 12px; color: #999; margin-bottom: 12px; }
        .news-meta span { margin-right: 15px; }
        .category-tag { display: inline-block; background: #667eea; color: white; padding: 2px 10px; border-radius: 4px; font-size: 11px; }
        .bilingual-content { background: #fafafa; border-radius: 8px; overflow: hidden; }
        .lang-section { padding: 15px; }
        .lang-section.zh { border-bottom: 1px solid #eee; }
        .lang-section.ug { direction: rtl; text-align: right; background: #f5f5f5; }
        .lang-label { font-size: 11px; color: #999; margin-bottom: 6px; text-transform: uppercase; }
        .lang-text { color: #555; line-height: 1.8; margin: 0; }
        .footer { background: #fafafa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        .footer .ug { direction: rtl; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>新闻情报日报</h1>
            <div class="ug-title">خەۋەر ئۇچۇرلىرى كۈندىلىك</div>
            <p class="subtitle">{{.Date}} · 共 {{.Count}} 条新闻 · جەمئى {{.Count}} خەۋەر</p>
        </div>
        <div class="content">
            {{range .News}}
            <div class="news-item">
                <h2 class="news-title">
                    <a href="{{.URL}}" target="_blank">{{.Title}}</a>
                </h2>
                <div class="news-meta">
                    <span class="category-tag">{{.Category}}</span>
                    <span>来源/مەنبە: {{.Source}}</span>
                </div>
                <div class="bilingual-content">
                    <div class="lang-section zh">
                        <div class="lang-label">中文</div>
                        <p class="lang-text">{{if .TransSummary}}{{.TransSummary}}{{else}}{{.Summary}}{{end}}</p>
                    </div>
                    <div class="lang-section ug">
                        <div class="lang-label">ئۇيغۇرچە</div>
                        <p class="lang-text">{{if .TransTitle}}{{.TransTitle}}{{else}}{{.Title}}{{end}}</p>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>由 News Intel App 自动生成于 {{.Generated}}</p>
            <p class="ug">News Intel App تەرىپىدىن ئاپتوماتىك ھاسىل قىلىندى</p>
        </div>
    </div>
</body>
</html>`
}

// GetAutoPushConfig 获取自动打包推送配置
func (p *Pusher) GetAutoPushConfig() (enabled bool, threshold int, channelID, templateID string) {
	var enabledStr, thresholdStr string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'auto_push_enabled'").Scan(&enabledStr)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'auto_push_threshold'").Scan(&thresholdStr)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'auto_push_channel_id'").Scan(&channelID)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'auto_push_template_id'").Scan(&templateID)

	enabled = enabledStr == "1"
	threshold = 6
	if thresholdStr != "" {
		fmt.Sscanf(thresholdStr, "%d", &threshold)
	}
	return
}

// CheckAndAutoPush 检查并触发自动打包推送
func (p *Pusher) CheckAndAutoPush() error {
	enabled, threshold, channelID, templateID := p.GetAutoPushConfig()
	if !enabled || channelID == "" {
		return nil
	}

	// 获取等待推送的新闻数量
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM news WHERE in_reading = 1 AND pushed = 0 AND translated = 1").Scan(&count)

	if count < threshold {
		log.Printf("Auto push: waiting for more news (%d/%d)", count, threshold)
		return nil
	}

	log.Printf("Auto push: threshold reached (%d/%d), triggering push...", count, threshold)

	// 获取渠道配置
	var channel models.PushChannel
	err := database.DB.QueryRow("SELECT id, name, type, config FROM push_channels WHERE id = ?", channelID).
		Scan(&channel.ID, &channel.Name, &channel.Type, &channel.Config)
	if err != nil {
		return fmt.Errorf("auto push channel not found: %w", err)
	}

	// 获取待推送的新闻（取 threshold 条）
	rows, err := database.DB.Query(`
		SELECT id, title, content, summary, url, source, category, image_url, trans_title, trans_summary 
		FROM news 
		WHERE in_reading = 1 AND pushed = 0 AND translated = 1
		ORDER BY reading_at ASC LIMIT ?
	`, threshold)
	if err != nil {
		return err
	}
	defer rows.Close()

	var news []models.News
	var newsIDs []string
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
		newsIDs = append(newsIDs, n.ID)
	}

	if len(news) == 0 {
		return nil
	}

	log.Printf("Auto pushing %d news...", len(news))

	var pushErr error
	switch channel.Type {
	case "email":
		// 获取模板
		var tmplContent, tmplSubject string
		if templateID != "" {
			database.DB.QueryRow("SELECT content, subject FROM email_templates WHERE id = ?", templateID).
				Scan(&tmplContent, &tmplSubject)
		}
		if tmplContent == "" {
			tmplContent = GetDefaultEmailTemplate()
			tmplSubject = fmt.Sprintf("新闻情报 - %s", time.Now().Format("2006-01-02 15:04"))
		}

		htmlContent, err := p.RenderTemplate(tmplContent, news)
		if err != nil {
			return err
		}

		var config models.EmailConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return err
		}
		pushErr = p.SendEmail(&config, tmplSubject, htmlContent)

	case "ntfy":
		pushErr = p.pushNtfy(&channel, news)
	}

	// 推送成功后标记为已推送
	if pushErr == nil {
		for _, id := range newsIDs {
			database.DB.Exec("UPDATE news SET pushed = 1, pushed_at = ? WHERE id = ?", time.Now(), id)
		}
		log.Printf("Auto push completed: %d news pushed", len(newsIDs))
	} else {
		log.Printf("Auto push failed: %v", pushErr)
	}

	return pushErr
}

// GetPendingPushCount 获取等待推送的新闻数量
func (p *Pusher) GetPendingPushCount() int {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM news WHERE in_reading = 1 AND pushed = 0 AND translated = 1").Scan(&count)
	return count
}
