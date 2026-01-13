package models

import "time"

// News 新闻模型
type News struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`      // 来源: rss, twitter, github, hackernews
	Category    string    `json:"category"`    // 分类: tech, ai, international, trending
	ImageURL    string    `json:"image_url"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	Translated  bool      `json:"translated"`
	TransTitle  string    `json:"trans_title"`   // 翻译后标题
	TransContent string   `json:"trans_content"` // 翻译后内容
	TransSummary string   `json:"trans_summary"` // 翻译后摘要
	IsFiltered  bool      `json:"is_filtered"`   // 是否被AI筛选掉
	Tags        string    `json:"tags"`          // 标签，逗号分隔
	InReading   bool      `json:"in_reading"`    // 是否在阅读窗口
	ReadingAt   time.Time `json:"reading_at"`    // 加入阅读窗口时间
	Pushed      bool      `json:"pushed"`        // 是否已推送
	PushedAt    time.Time `json:"pushed_at"`     // 推送时间
}

// NewsSource 新闻源配置
type NewsSource struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`       // rss, api, scraper
	URL       string    `json:"url"`
	Category  string    `json:"category"`
	Enabled   bool      `json:"enabled"`
	Interval  int       `json:"interval"`   // 采集间隔(分钟)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PushChannel 推送渠道配置
type PushChannel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`       // email, ntfy, webhook
	Config    string    `json:"config"`     // JSON配置
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmailConfig 邮箱配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	FromAddress  string `json:"from_address"`
	FromName     string `json:"from_name"`
	ToAddresses  string `json:"to_addresses"` // 逗号分隔
}

// NtfyConfig ntfy配置
type NtfyConfig struct {
	ServerURL string `json:"server_url"`
	Topic     string `json:"topic"`
	Token     string `json:"token"`
}

// EmailTemplate 邮件模板
type EmailTemplate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Content   string    `json:"content"`     // HTML内容
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AIConfig AI配置
type AIConfig struct {
	ID           string `json:"id"`
	Provider     string `json:"provider"`      // openai, claude, ollama
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url"`
	Model        string `json:"model"`
	EnableTrans  bool   `json:"enable_trans"`  // 启用翻译
	EnableSummary bool  `json:"enable_summary"` // 启用摘要
	EnableFilter bool   `json:"enable_filter"`  // 启用筛选
	TargetLang   string `json:"target_lang"`   // 目标语言
}

// PushTask 推送任务
type PushTask struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CronExpr    string    `json:"cron_expr"`    // cron表达式
	ChannelID   string    `json:"channel_id"`
	TemplateID  string    `json:"template_id"`
	Categories  string    `json:"categories"`   // 推送的分类，逗号分隔
	Enabled     bool      `json:"enabled"`
	LastRunAt   time.Time `json:"last_run_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Settings 系统设置
type Settings struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
