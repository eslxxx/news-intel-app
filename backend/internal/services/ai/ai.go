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

// Summarize 生成摘要
func (s *AIService) Summarize(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	prompt := fmt.Sprintf("为以下新闻生成一个简洁的中文摘要（不超过100字）：\n\n%s", text)

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
	// 翻译标题
	if news.Title != "" {
		transTitle, err := s.Translate(news.Title, s.config.TargetLang)
		if err != nil {
			log.Printf("Failed to translate title: %v", err)
		} else {
			news.TransTitle = transTitle
		}
	}

	// 生成摘要
	content := news.Content
	if content == "" {
		content = news.Title
	}
	summary, err := s.Summarize(content)
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
