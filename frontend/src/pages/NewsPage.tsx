import React, { useEffect, useState } from 'react';
import { Card, List, Tag, Select, Button, Pagination, message, Popconfirm, Empty, Spin } from 'antd';
import { ReloadOutlined, DeleteOutlined } from '@ant-design/icons';
import { getNews, deleteNews, triggerCollect } from '../api';
import dayjs from 'dayjs';

const NewsPage: React.FC = () => {
  const [news, setNews] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [category, setCategory] = useState<string>('');
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const fetchNews = async () => {
    setLoading(true);
    try {
      const res = await getNews({
        category: category || undefined,
        limit: pageSize,
        offset: (page - 1) * pageSize,
      });
      setNews(res.data.data || []);
      setTotal(res.data.total || 0);
    } catch (e) {
      message.error('获取新闻失败');
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchNews();
  }, [category, page]);

  const handleDelete = async (id: string) => {
    try {
      await deleteNews(id);
      message.success('删除成功');
      fetchNews();
    } catch {
      message.error('删除失败');
    }
  };

  const handleCollect = async () => {
    try {
      await triggerCollect();
      message.success('采集任务已启动，请稍后刷新');
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

  const categories = [
    { value: '', label: '全部分类' },
    { value: 'tech', label: '科技' },
    { value: 'ai', label: 'AI' },
    { value: 'github', label: 'GitHub' },
    { value: 'international', label: '国际' },
    { value: 'trending', label: '热门' },
  ];

  return (
    <div>
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>新闻列表</h2>
        <div>
          <Select
            style={{ width: 120, marginRight: 8 }}
            value={category}
            onChange={setCategory}
            options={categories}
          />
          <Button icon={<ReloadOutlined />} onClick={handleCollect}>
            立即采集
          </Button>
        </div>
      </div>

      <Spin spinning={loading}>
        {news.length === 0 ? (
          <Empty description="暂无新闻，点击立即采集获取" />
        ) : (
          <>
            <List
              grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 3 }}
              dataSource={news}
              renderItem={(item: any) => (
                <List.Item>
                  <Card
                    className="news-card"
                    actions={[
                      <a href={item.url} target="_blank" rel="noopener noreferrer" key="view">
                        查看原文
                      </a>,
                      <Popconfirm title="确定删除?" onConfirm={() => handleDelete(item.id)} key="delete">
                        <DeleteOutlined />
                      </Popconfirm>,
                    ]}
                  >
                    <div className="news-title">
                      <a href={item.url} target="_blank" rel="noopener noreferrer">
                        {item.trans_title || item.title}
                      </a>
                    </div>
                    <div className="news-meta">
                      <Tag color={categoryColors[item.category] || 'default'}>{item.category}</Tag>
                      <span>{item.source}</span>
                      <span style={{ marginLeft: 8 }}>{dayjs(item.created_at).format('MM-DD HH:mm')}</span>
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

export default NewsPage;
