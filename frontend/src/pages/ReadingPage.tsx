import React, { useEffect, useState } from 'react';
import { Card, List, Tag, Select, Button, Pagination, message, Popconfirm, Empty, Spin, Badge, Space, Tooltip } from 'antd';
import { DeleteOutlined, ClearOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { getReadingNews, removeFromReading, clearPushedNews } from '../api';
import dayjs from 'dayjs';

const ReadingPage: React.FC = () => {
  const [news, setNews] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [unpushedCount, setUnpushedCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [category, setCategory] = useState<string>('');
  const [pushedFilter, setPushedFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const fetchNews = async () => {
    setLoading(true);
    try {
      const res = await getReadingNews({
        category: category || undefined,
        pushed: pushedFilter,
        limit: pageSize,
        offset: (page - 1) * pageSize,
      });
      setNews(res.data.data || []);
      setTotal(res.data.total || 0);
      setUnpushedCount(res.data.unpushed_count || 0);
    } catch (e) {
      message.error('获取阅读列表失败');
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchNews();
  }, [category, pushedFilter, page]);

  const handleRemove = async (id: string) => {
    try {
      await removeFromReading(id);
      message.success('已移出阅读窗口');
      fetchNews();
    } catch {
      message.error('操作失败');
    }
  };

  const handleClearPushed = async () => {
    try {
      await clearPushedNews();
      message.success('已清空已推送的新闻');
      fetchNews();
    } catch {
      message.error('操作失败');
    }
  };

  const categoryColors: Record<string, string> = {
    tech: 'blue',
    ai: 'purple',
    github: 'green',
    international: 'orange',
    trending: 'red',
  };

  const categories = [
    { value: '', label: '全部分类' },
    { value: 'tech', label: '科技' },
    { value: 'ai', label: 'AI' },
    { value: 'github', label: 'GitHub' },
    { value: 'international', label: '国际' },
  ];

  const pushedOptions = [
    { value: 'all', label: '全部状态' },
    { value: 'no', label: '待推送' },
    { value: 'yes', label: '已推送' },
  ];

  return (
    <div>
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 12 }}>
        <div>
          <h2 style={{ display: 'inline', marginRight: 16 }}>阅读窗口</h2>
          <Badge count={unpushedCount} style={{ backgroundColor: '#52c41a' }} title="待推送新闻">
            <Tag color="green">待推送</Tag>
          </Badge>
        </div>
        <Space wrap>
          <Select
            style={{ width: 120 }}
            value={category}
            onChange={(v) => { setCategory(v); setPage(1); }}
            options={categories}
          />
          <Select
            style={{ width: 120 }}
            value={pushedFilter}
            onChange={(v) => { setPushedFilter(v); setPage(1); }}
            options={pushedOptions}
          />
          <Popconfirm title="确定清空所有已推送的新闻?" onConfirm={handleClearPushed}>
            <Button icon={<ClearOutlined />} danger>清空已推送</Button>
          </Popconfirm>
        </Space>
      </div>

      <div style={{ marginBottom: 16, padding: 16, background: '#f6ffed', borderRadius: 8, border: '1px solid #b7eb8f' }}>
        <strong>说明：</strong>新闻采集后会自动翻译并加入阅读窗口。推送任务会从这里取未推送的新闻发送，推送后标记为已推送。
      </div>

      <Spin spinning={loading}>
        {news.length === 0 ? (
          <Empty description="暂无新闻，等待采集翻译..." />
        ) : (
          <>
            <List
              grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 3 }}
              dataSource={news}
              renderItem={(item: any) => (
                <List.Item>
                  <Card
                    className="news-card"
                    style={{ borderLeft: item.pushed ? '3px solid #ccc' : '3px solid #52c41a' }}
                    actions={[
                      <a href={item.url} target="_blank" rel="noopener noreferrer" key="view">
                        查看原文
                      </a>,
                      <Popconfirm title="移出阅读窗口?" onConfirm={() => handleRemove(item.id)} key="remove">
                        <DeleteOutlined />
                      </Popconfirm>,
                    ]}
                  >
                    <div className="news-title">
                      <a href={item.url} target="_blank" rel="noopener noreferrer">
                        {item.trans_title || item.title}
                      </a>
                      {item.pushed && (
                        <Tooltip title={`已推送于 ${dayjs(item.pushed_at).format('MM-DD HH:mm')}`}>
                          <CheckCircleOutlined style={{ color: '#999', marginLeft: 8 }} />
                        </Tooltip>
                      )}
                    </div>
                    <div className="news-meta">
                      <Tag color={categoryColors[item.category] || 'default'}>{item.category}</Tag>
                      <span>{item.source}</span>
                      <span style={{ marginLeft: 8 }}>{dayjs(item.reading_at).format('MM-DD HH:mm')}</span>
                      {!item.pushed && <Tag color="green" style={{ marginLeft: 8 }}>待推送</Tag>}
                    </div>
                    <div className="news-summary" style={{ marginTop: 8 }}>
                      {item.trans_summary || item.summary || '暂无摘要'}
                    </div>
                  </Card>
                </List.Item>
              )}
            />
            <div style={{ textAlign: 'center', marginTop: 24 }}>
              <Pagination
                current={page}
                pageSize={pageSize}
                total={total}
                onChange={setPage}
                showSizeChanger={false}
              />
            </div>
          </>
        )}
      </Spin>
    </div>
  );
};

export default ReadingPage;
