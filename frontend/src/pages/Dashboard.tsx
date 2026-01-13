import React, { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, List, Tag, Button, message, Spin } from 'antd';
import { ReloadOutlined, RobotOutlined, GlobalOutlined, SendOutlined, FileTextOutlined } from '@ant-design/icons';
import { getStats, getNews, triggerCollect, triggerProcess } from '../api';
import dayjs from 'dayjs';

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<any>({});
  const [news, setNews] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [statsRes, newsRes] = await Promise.all([getStats(), getNews({ limit: 10 })]);
      setStats(statsRes.data);
      setNews(newsRes.data.data || []);
    } catch (e) {
      console.error(e);
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleCollect = async () => {
    try {
      await triggerCollect();
      message.success('采集任务已启动');
    } catch {
      message.error('启动失败');
    }
  };

  const handleProcess = async () => {
    try {
      await triggerProcess();
      message.success('AI处理任务已启动');
    } catch {
      message.error('启动失败');
    }
  };

  const categoryColors: Record<string, string> = {
    tech: 'blue',
    ai: 'purple',
    github: 'green',
    international: 'orange',
    trending: 'red',
  };

  if (loading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  }

  return (
    <div>
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>仪表盘</h2>
        <div>
          <Button icon={<ReloadOutlined />} onClick={handleCollect} style={{ marginRight: 8 }}>
            立即采集
          </Button>
          <Button icon={<RobotOutlined />} onClick={handleProcess}>
            AI处理
          </Button>
        </div>
      </div>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card className="stat-card">
            <Statistic
              title="总新闻数"
              value={stats.total_news || 0}
              prefix={<FileTextOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="stat-card">
            <Statistic
              title="今日新闻"
              value={stats.today_news || 0}
              prefix={<GlobalOutlined />}
              valueStyle={{ color: '#667eea' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="stat-card">
            <Statistic
              title="活跃新闻源"
              value={stats.sources_count || 0}
              prefix={<GlobalOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="stat-card">
            <Statistic
              title="推送渠道"
              value={stats.channels_count || 0}
              prefix={<SendOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={16}>
          <Card title="最新新闻" extra={<a href="/news">查看全部</a>}>
            <List
              itemLayout="vertical"
              dataSource={news}
              renderItem={(item: any) => (
                <List.Item key={item.id}>
                  <div className="news-title">
                    <a href={item.url} target="_blank" rel="noopener noreferrer">
                      {item.trans_title || item.title}
                    </a>
                  </div>
                  <div className="news-meta">
                    <Tag color={categoryColors[item.category] || 'default'}>{item.category}</Tag>
                    <span>{item.source}</span>
                    <span style={{ marginLeft: 12 }}>{dayjs(item.created_at).format('MM-DD HH:mm')}</span>
                  </div>
                  {item.trans_summary && <div className="news-summary">{item.trans_summary}</div>}
                </List.Item>
              )}
            />
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="分类统计">
            {stats.by_category && Object.entries(stats.by_category).map(([cat, count]) => (
              <div key={cat} style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                <Tag color={categoryColors[cat] || 'default'}>{cat}</Tag>
                <span>{count as number} 篇</span>
              </div>
            ))}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
