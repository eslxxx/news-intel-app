package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return err
	}

	// 创建表
	if err := createTables(); err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	tables := `
	-- 新闻表
	CREATE TABLE IF NOT EXISTS news (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT,
		summary TEXT,
		url TEXT UNIQUE,
		source TEXT,
		category TEXT,
		image_url TEXT,
		author TEXT,
		published_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		translated INTEGER DEFAULT 0,
		trans_title TEXT,
		trans_content TEXT,
		trans_summary TEXT,
		is_filtered INTEGER DEFAULT 0,
		tags TEXT,
		in_reading INTEGER DEFAULT 0,
		reading_at DATETIME,
		pushed INTEGER DEFAULT 0,
		pushed_at DATETIME
	);

	-- 新闻源表
	CREATE TABLE IF NOT EXISTS news_sources (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		url TEXT,
		category TEXT,
		enabled INTEGER DEFAULT 1,
		interval_mins INTEGER DEFAULT 60,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 推送渠道表
	CREATE TABLE IF NOT EXISTS push_channels (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		config TEXT,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 邮件模板表
	CREATE TABLE IF NOT EXISTS email_templates (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		subject TEXT,
		content TEXT,
		is_default INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- AI配置表
	CREATE TABLE IF NOT EXISTS ai_configs (
		id TEXT PRIMARY KEY,
		provider TEXT NOT NULL,
		api_key TEXT,
		base_url TEXT,
		model TEXT,
		enable_trans INTEGER DEFAULT 1,
		enable_summary INTEGER DEFAULT 1,
		enable_filter INTEGER DEFAULT 0,
		target_lang TEXT DEFAULT 'zh-CN',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 推送任务表
	CREATE TABLE IF NOT EXISTS push_tasks (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		cron_expr TEXT,
		channel_id TEXT,
		template_id TEXT,
		categories TEXT,
		enabled INTEGER DEFAULT 1,
		last_run_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 系统设置表
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);

	-- 创建索引
	CREATE INDEX IF NOT EXISTS idx_news_source ON news(source);
	CREATE INDEX IF NOT EXISTS idx_news_category ON news(category);
	CREATE INDEX IF NOT EXISTS idx_news_created ON news(created_at);
	CREATE INDEX IF NOT EXISTS idx_news_published ON news(published_at);
	CREATE INDEX IF NOT EXISTS idx_news_reading ON news(in_reading);
	CREATE INDEX IF NOT EXISTS idx_news_translated ON news(translated);
	`

	_, err := DB.Exec(tables)
	return err
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
