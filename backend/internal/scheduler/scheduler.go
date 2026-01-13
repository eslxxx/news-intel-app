package scheduler

import (
	"log"

	"news-intel-app/internal/database"
	"news-intel-app/internal/models"
	"news-intel-app/internal/services/ai"
	"news-intel-app/internal/services/collector"
	"news-intel-app/internal/services/pusher"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron      *cron.Cron
	collector *collector.Collector
	ai        *ai.AIService
	pusher    *pusher.Pusher
}

func New(col *collector.Collector, aiSvc *ai.AIService, push *pusher.Pusher) *Scheduler {
	return &Scheduler{
		cron:      cron.New(),
		collector: col,
		ai:        aiSvc,
		pusher:    push,
	}
}

func (s *Scheduler) Start() {
	// 每30分钟采集一次新闻，采集后立即翻译
	s.cron.AddFunc("*/30 * * * *", func() {
		s.CollectAndTranslate()
	})

	// 加载推送任务
	s.loadPushTasks()

	s.cron.Start()
	log.Println("Scheduler started")
}

// CollectAndTranslate 采集新闻并立即翻译
func (s *Scheduler) CollectAndTranslate() {
	log.Println("Scheduled: Collecting news...")
	newNews, err := s.collector.CollectAll()
	if err != nil {
		log.Printf("Scheduled collect error: %v", err)
		return
	}

	if len(newNews) == 0 {
		log.Println("No new news to translate")
		return
	}

	log.Printf("Translating %d new news...", len(newNews))
	if err := s.ai.ProcessAndMoveToReading(newNews); err != nil {
		log.Printf("Scheduled AI process error: %v", err)
	}

	// 翻译完成后检查自动推送
	if err := s.pusher.CheckAndAutoPush(); err != nil {
		log.Printf("Auto push check error: %v", err)
	}
}

func (s *Scheduler) loadPushTasks() {
	rows, err := database.DB.Query("SELECT id, name, cron_expr, channel_id, template_id, categories, enabled FROM push_tasks WHERE enabled = 1")
	if err != nil {
		log.Printf("Failed to load push tasks: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t models.PushTask
		if err := rows.Scan(&t.ID, &t.Name, &t.CronExpr, &t.ChannelID, &t.TemplateID, &t.Categories, &t.Enabled); err != nil {
			continue
		}

		task := t // 创建副本
		_, err := s.cron.AddFunc(task.CronExpr, func() {
			log.Printf("Executing push task: %s", task.Name)
			if err := s.pusher.ExecutePushTask(&task); err != nil {
				log.Printf("Push task error: %v", err)
			}
		})
		if err != nil {
			log.Printf("Failed to add cron job for task %s: %v", t.Name, err)
		}
	}
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
