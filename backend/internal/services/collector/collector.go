package collector

import (
	"context"
	"log"
	"time"

	"news-intel-app/internal/database"
	"news-intel-app/internal/models"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
)

// Collector 新闻采集器
type Collector struct {
	parser *gofeed.Parser
}

func New() *Collector {
	return &Collector{
		parser: gofeed.NewParser(),
	}
}

// CollectRSS 采集RSS源
func (c *Collector) CollectRSS(source *models.NewsSource) ([]models.News, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	feed, err := c.parser.ParseURLWithContext(source.URL, ctx)
	if err != nil {
		return nil, err
	}

	var news []models.News
	for _, item := range feed.Items {
		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		// 只采集今天的新闻
		if time.Since(publishedAt) > 24*time.Hour {
			continue
		}

		imageURL := ""
		if item.Image != nil {
			imageURL = item.Image.URL
		}

		author := ""
		if item.Author != nil {
			author = item.Author.Name
		}

		n := models.News{
			ID:          uuid.New().String(),
			Title:       item.Title,
			Content:     item.Description,
			URL:         item.Link,
			Source:      source.Name,
			Category:    source.Category,
			ImageURL:    imageURL,
			Author:      author,
			PublishedAt: publishedAt,
			CreatedAt:   time.Now(),
		}
		news = append(news, n)
	}

	return news, nil
}

// SaveNews 保存新闻到数据库，返回新保存的新闻列表
func (c *Collector) SaveNews(news []models.News) ([]models.News, error) {
	stmt, err := database.DB.Prepare(`
		INSERT OR IGNORE INTO news (id, title, content, summary, url, source, category, image_url, author, published_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var savedNews []models.News
	for _, n := range news {
		result, err := stmt.Exec(n.ID, n.Title, n.Content, n.Summary, n.URL, n.Source, n.Category, n.ImageURL, n.Author, n.PublishedAt, n.CreatedAt)
		if err != nil {
			log.Printf("Failed to save news: %v", err)
			continue
		}
		// 检查是否真的插入了（不是重复的）
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			savedNews = append(savedNews, n)
		}
	}

	return savedNews, nil
}

// CollectAll 采集所有启用的新闻源，返回新采集的新闻
func (c *Collector) CollectAll() ([]models.News, error) {
	rows, err := database.DB.Query("SELECT id, name, type, url, category FROM news_sources WHERE enabled = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allNewNews []models.News
	for rows.Next() {
		var source models.NewsSource
		if err := rows.Scan(&source.ID, &source.Name, &source.Type, &source.URL, &source.Category); err != nil {
			continue
		}

		switch source.Type {
		case "rss":
			news, err := c.CollectRSS(&source)
			if err != nil {
				log.Printf("Failed to collect from %s: %v", source.Name, err)
				continue
			}
			savedNews, err := c.SaveNews(news)
			if err != nil {
				log.Printf("Failed to save news from %s: %v", source.Name, err)
			}
			allNewNews = append(allNewNews, savedNews...)
			log.Printf("Collected %d news from %s", len(savedNews), source.Name)
		}
	}

	return allNewNews, nil
}

// GetDefaultSources 获取默认新闻源
func GetDefaultSources() []models.NewsSource {
	return []models.NewsSource{
		{ID: uuid.New().String(), Name: "Hacker News", Type: "rss", URL: "https://hnrss.org/frontpage", Category: "tech", Enabled: true, Interval: 30},
		{ID: uuid.New().String(), Name: "TechCrunch", Type: "rss", URL: "https://techcrunch.com/feed/", Category: "tech", Enabled: true, Interval: 60},
		{ID: uuid.New().String(), Name: "The Verge", Type: "rss", URL: "https://www.theverge.com/rss/index.xml", Category: "tech", Enabled: true, Interval: 60},
		{ID: uuid.New().String(), Name: "Ars Technica", Type: "rss", URL: "https://feeds.arstechnica.com/arstechnica/index", Category: "tech", Enabled: true, Interval: 60},
		{ID: uuid.New().String(), Name: "MIT Tech Review", Type: "rss", URL: "https://www.technologyreview.com/feed/", Category: "ai", Enabled: true, Interval: 60},
		{ID: uuid.New().String(), Name: "AI News", Type: "rss", URL: "https://www.artificialintelligence-news.com/feed/", Category: "ai", Enabled: true, Interval: 60},
		{ID: uuid.New().String(), Name: "GitHub Trending", Type: "rss", URL: "https://ossinsight.io/blog/rss.xml", Category: "github", Enabled: true, Interval: 120},
		{ID: uuid.New().String(), Name: "Al Jazeera World", Type: "rss", URL: "https://www.aljazeera.com/xml/rss/all.xml", Category: "international", Enabled: true, Interval: 60},
		{ID: uuid.New().String(), Name: "BBC News", Type: "rss", URL: "https://feeds.bbci.co.uk/news/world/rss.xml", Category: "international", Enabled: true, Interval: 60},
	}
}

// InitDefaultSources 初始化默认新闻源
func InitDefaultSources() error {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM news_sources").Scan(&count)
	if count > 0 {
		return nil
	}

	sources := GetDefaultSources()
	stmt, err := database.DB.Prepare(`
		INSERT INTO news_sources (id, name, type, url, category, enabled, interval_mins) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range sources {
		stmt.Exec(s.ID, s.Name, s.Type, s.URL, s.Category, s.Enabled, s.Interval)
	}

	log.Println("Default news sources initialized")
	return nil
}
