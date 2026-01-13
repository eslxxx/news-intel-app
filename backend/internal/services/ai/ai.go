package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"news-intel-app/internal/database"
	"news-intel-app/internal/models"

	openai "github.com/sashabaranov/go-openai"
)

type AIService struct {
	client *openai.Client
	config *models.AIConfig
}

func New(apiKey, baseURL, model string) *AIService {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	return &AIService{
		client: openai.NewClientWithConfig(config),
		config: &models.AIConfig{
			Model:      model,
			TargetLang: "zh-CN",
		},
	}
}

// LoadConfig 从数据库加载AI配置
func (s *AIService) LoadConfig() error {
	row := database.DB.QueryRow("SELECT provider, api_key, base_url, model, enable_trans, enable_summary, enable_filter, target_lang FROM ai_configs LIMIT 1")
	
	var cfg models.AIConfig
	err := row.Scan(&cfg.Provider, &cfg.APIKey, &cfg.BaseURL, &cfg.Model, &cfg.EnableTrans, &cfg.EnableSummary, &cfg.EnableFilter, &cfg.TargetLang)
	if err != nil {
		return err
	}

	s.config = &cfg
	
	// 重新初始化client
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}
	s.client = openai.NewClientWithConfig(config)
	
	return nil
}

// Translate 翻译文本
func (s *AIService) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	var prompt string
	switch targetLang {
	case "zh-ug":
		// 双语翻译：中文 + 维吾尔语
		prompt = fmt.Sprintf(`将以下文本翻译成中文和维吾尔语两种语言，格式如下：
【中文】翻译内容
【ئۇيغۇرچە】翻译内容

只返回翻译结果，不要添加任何其他解释：

%s`, text)
	case "ug":
		prompt = fmt.Sprintf("将以下文本翻译成维吾尔语(Uyghur)，只返回翻译结果，不要添加任何解释：\n\n%s", text)
	default:
		prompt = fmt.Sprintf("将以下文本翻译成%s，只返回翻译结果，不要添加任何解释：\n\n%s", targetLang, text)
	}

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Temperature: 0.3,
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from AI")
}

// Summarize 生成摘要（支持多语言）
func (s *AIService) Summarize(text string, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	var prompt string
	switch targetLang {
	case "zh-ug":
		// 双语摘要：中文 + 维吾尔语
		prompt = fmt.Sprintf(`为以下新闻生成双语摘要，格式如下：
【中文】中文摘要（不超过100字）
【ئۇيغۇرچە】维吾尔语摘要（不超过100字）

只返回摘要内容，不要添加任何其他解释：

%s`, text)
	case "ug":
		prompt = fmt.Sprintf("为以下新闻生成一个简洁的维吾尔语(Uyghur)摘要（不超过100字），只返回摘要内容：\n\n%s", text)
	default:
		prompt = fmt.Sprintf("为以下新闻生成一个简洁的中文摘要（不超过100字）：\n\n%s", text)
	}

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Temperature: 0.5,
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from AI")
}

// FilterNews 筛选新闻（判断是否值得推送）
func (s *AIService) FilterNews(news *models.News) (bool, error) {
	prompt := fmt.Sprintf(`判断以下新闻是否有价值推送给用户。
标题: %s
内容: %s

请返回JSON格式: {"valuable": true/false, "reason": "原因"}`, news.Title, news.Content)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Temperature: 0.3,
		},
	)
	if err != nil {
		return true, err
	}

	if len(resp.Choices) > 0 {
		var result struct {
			Valuable bool   `json:"valuable"`
			Reason   string `json:"reason"`
		}
		if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
			return true, nil // 解析失败默认保留
		}
		return result.Valuable, nil
	}
	return true, nil
}

// ProcessNews 处理新闻（翻译+摘要）
func (s *AIService) ProcessNews(news *models.News) error {
	// 翻译标题（支持双语：中文+维语）
	if news.Title != "" {
		transTitle, err := s.Translate(news.Title, s.config.TargetLang)
		if err != nil {
			log.Printf("Failed to translate title: %v", err)
		} else {
			news.TransTitle = transTitle
		}
	}

	// 生成摘要（使用相同的目标语言，支持双语）
	content := news.Content
	if content == "" {
		content = news.Title
	}
	summary, err := s.Summarize(content, s.config.TargetLang)
	if err != nil {
		log.Printf("Failed to summarize: %v", err)
	} else {
		news.TransSummary = summary
	}

	news.Translated = true
	return nil
}

// ProcessUnprocessedNews 处理未处理的新闻
func (s *AIService) ProcessUnprocessedNews(limit int) error {
	rows, err := database.DB.Query(`
		SELECT id, title, content FROM news 
		WHERE translated = 0 AND is_filtered = 0 
		ORDER BY created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var news models.News
		if err := rows.Scan(&news.ID, &news.Title, &news.Content); err != nil {
			continue
		}

		if err := s.ProcessNews(&news); err != nil {
			log.Printf("Failed to process news %s: %v", news.ID, err)
			continue
		}

		// 更新数据库，同时移入阅读窗口
		_, err := database.DB.Exec(`
			UPDATE news SET translated = 1, trans_title = ?, trans_summary = ?, in_reading = 1, reading_at = ? WHERE id = ?
		`, news.TransTitle, news.TransSummary, time.Now(), news.ID)
		if err != nil {
			log.Printf("Failed to update news %s: %v", news.ID, err)
		}
	}

	return nil
}

// ProcessAndMoveToReading 处理新闻列表并移入阅读窗口
func (s *AIService) ProcessAndMoveToReading(newsList []models.News) error {
	for _, news := range newsList {
		if err := s.ProcessNews(&news); err != nil {
			log.Printf("Failed to process news %s: %v", news.ID, err)
			continue
		}

		// 更新数据库，同时移入阅读窗口
		_, err := database.DB.Exec(`
			UPDATE news SET translated = 1, trans_title = ?, trans_summary = ?, in_reading = 1, reading_at = ? WHERE id = ?
		`, news.TransTitle, news.TransSummary, time.Now(), news.ID)
		if err != nil {
			log.Printf("Failed to update news %s: %v", news.ID, err)
		} else {
			log.Printf("Translated and moved to reading: %s", news.Title)
		}
	}

	return nil
}

// GenerateEmailTemplate 根据用户描述生成邮件模板
func (s *AIService) GenerateEmailTemplate(description string, currentTemplate string) (string, error) {
	var prompt string
	
	if currentTemplate != "" {
		// 修改现有模板
		prompt = fmt.Sprintf(`你是一个专业的邮件模板设计师。用户希望修改现有的邮件模板。

用户需求：%s

当前模板：
%s

请根据用户需求修改模板。要求：
1. 生成完整的 HTML 邮件模板
2. 必须包含以下 Go 模板变量（保持原样不变）：
   - {{.Date}} 日期
   - {{.Count}} 新闻数量
   - {{.Generated}} 生成时间
   - {{range .News}}...{{end}} 遍历新闻
   - {{.Title}} 原标题
   - {{.TransTitle}} 翻译后标题
   - {{.TransSummary}} 翻译后摘要
   - {{.URL}} 链接
   - {{.Source}} 来源
   - {{.Category}} 分类
3. 样式要美观、现代、响应式
4. 只返回 HTML 代码，不要任何解释

直接输出完整的 HTML 模板：`, description, currentTemplate)
	} else {
		// 创建新模板
		prompt = fmt.Sprintf(`你是一个专业的邮件模板设计师。请根据用户的描述创建一个新闻邮件模板。

用户需求：%s

要求：
1. 生成完整的 HTML 邮件模板
2. 必须包含以下 Go 模板变量：
   - {{.Date}} 日期
   - {{.Count}} 新闻数量
   - {{.Generated}} 生成时间
   - {{range .News}}...{{end}} 遍历新闻列表
   - 在循环内使用：{{.Title}}、{{.TransTitle}}、{{.TransSummary}}、{{.URL}}、{{.Source}}、{{.Category}}
3. 使用 {{if .TransTitle}}{{.TransTitle}}{{else}}{{.Title}}{{end}} 来优先显示翻译标题
4. 样式要美观、现代、响应式
5. 颜色搭配协调，排版清晰
6. 只返回 HTML 代码，不要任何解释

直接输出完整的 HTML 模板：`, description)
	}

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Temperature: 0.7,
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		// 清理可能的 markdown 代码块标记
		content = cleanHTMLResponse(content)
		return content, nil
	}
	return "", fmt.Errorf("no response from AI")
}

// cleanHTMLResponse 清理 AI 返回的 HTML 代码
func cleanHTMLResponse(content string) string {
	// 移除 markdown 代码块标记
	if len(content) > 7 && content[:7] == "```html" {
		content = content[7:]
	} else if len(content) > 3 && content[:3] == "```" {
		content = content[3:]
	}
	if len(content) > 3 && content[len(content)-3:] == "```" {
		content = content[:len(content)-3]
	}
	// 去除首尾空白
	for len(content) > 0 && (content[0] == '\n' || content[0] == '\r' || content[0] == ' ') {
		content = content[1:]
	}
	for len(content) > 0 && (content[len(content)-1] == '\n' || content[len(content)-1] == '\r' || content[len(content)-1] == ' ') {
		content = content[:len(content)-1]
	}
	return content
}

// SetAutoPushCallback 设置自动推送回调
var AutoPushCallback func() error

// TriggerAutoPushCheck 触发自动推送检查
func TriggerAutoPushCheck() {
	if AutoPushCallback != nil {
		go func() {
			if err := AutoPushCallback(); err != nil {
				log.Printf("Auto push check error: %v", err)
			}
		}()
	}
}
