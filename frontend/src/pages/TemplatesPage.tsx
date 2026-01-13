import React, { useEffect, useState, useCallback } from 'react';
import { Button, Card, Form, Input, message, Popconfirm, Space, List, Spin, Alert } from 'antd';
import { PlusOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import Editor from '@monaco-editor/react';
import { getTemplates, createTemplate, updateTemplate, deleteTemplate, previewTemplate } from '../api';
import debounce from 'lodash/debounce';

const defaultTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f5f5f5; padding: 20px; margin: 0; }
        .container { max-width: 680px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .header p { margin: 10px 0 0; opacity: 0.9; }
        .content { padding: 20px; }
        .news-item { border-bottom: 1px solid #eee; padding: 20px 0; }
        .news-item:last-child { border-bottom: none; }
        .news-title { font-size: 18px; font-weight: 600; margin: 0 0 8px; }
        .news-title a { color: #667eea; text-decoration: none; }
        .news-title a:hover { text-decoration: underline; }
        .news-meta { font-size: 12px; color: #999; margin-bottom: 10px; }
        .news-meta .tag { background: #f0f0f0; padding: 2px 8px; border-radius: 4px; margin-right: 8px; }
        .news-summary { color: #555; line-height: 1.8; }
        .footer { background: #fafafa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
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
                    <span class="tag">{{.Category}}</span>
                    <span>来源: {{.Source}}</span>
                </div>
                <p class="news-summary">{{if .TransSummary}}{{.TransSummary}}{{else}}{{.Summary}}{{end}}</p>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>由 News Intel App 自动生成于 {{.Generated}}</p>
        </div>
    </div>
</body>
</html>`;

const TemplatesPage: React.FC = () => {
  const [templates, setTemplates] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [content, setContent] = useState(defaultTemplate);
  const [previewHtml, setPreviewHtml] = useState('');
  const [previewLoading, setPreviewLoading] = useState(false);
  const [newsCount, setNewsCount] = useState(0);
  const [form] = Form.useForm();

  const fetchTemplates = async () => {
    setLoading(true);
    try {
      const res = await getTemplates();
      setTemplates(res.data || []);
    } catch {
      message.error('获取失败');
    }
    setLoading(false);
  };

  // 防抖预览
  const debouncedPreview = useCallback(
    debounce(async (templateContent: string) => {
      if (!templateContent) return;
      setPreviewLoading(true);
      try {
        const res = await previewTemplate(templateContent);
        setPreviewHtml(res.data.html);
        setNewsCount(res.data.news_count || 0);
      } catch {
        setPreviewHtml('<div style="padding:20px;color:#999;">模板语法错误，请检查</div>');
      }
      setPreviewLoading(false);
    }, 500),
    []
  );

  useEffect(() => {
    fetchTemplates();
    debouncedPreview(content);
  }, []);

  useEffect(() => {
    debouncedPreview(content);
  }, [content]);

  const handleSelect = (item: any) => {
    setSelectedId(item.id);
    setContent(item.content);
    form.setFieldsValue({ name: item.name, subject: item.subject });
  };

  const handleNew = () => {
    setSelectedId(null);
    setContent(defaultTemplate);
    form.resetFields();
  };

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      const data = { ...values, content };

      if (selectedId) {
        await updateTemplate(selectedId, data);
        message.success('更新成功');
      } else {
        const res = await createTemplate(data);
        setSelectedId(res.data.id);
        message.success('创建成功');
      }
      fetchTemplates();
    } catch {
      message.error('保存失败');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteTemplate(id);
      message.success('删除成功');
      if (selectedId === id) {
        handleNew();
      }
      fetchTemplates();
    } catch {
      message.error('删除失败');
    }
  };

  const handleRefreshPreview = () => {
    debouncedPreview(content);
  };

  return (
    <div>
      <div className="page-header">
        <h2>邮件模板编辑器</h2>
      </div>

      <Alert 
        message="模板变量说明" 
        description={
          <span>
            <code>{'{{.Date}}'}</code> 日期 | 
            <code>{'{{.Count}}'}</code> 新闻数量 | 
            <code>{'{{.Generated}}'}</code> 生成时间 | 
            <code>{'{{range .News}}...{{end}}'}</code> 遍历新闻 | 
            <code>{'{{.Title}}'}</code> 原标题 | 
            <code>{'{{.TransTitle}}'}</code> 翻译标题 | 
            <code>{'{{.TransSummary}}'}</code> 翻译摘要 | 
            <code>{'{{.URL}}'}</code> 链接 | 
            <code>{'{{.Source}}'}</code> 来源 | 
            <code>{'{{.Category}}'}</code> 分类
          </span>
        }
        type="info" 
        showIcon 
        style={{ marginBottom: 16 }}
      />

      <div style={{ display: 'flex', gap: 16, height: 600 }}>
        {/* 左侧模板列表 */}
        <Card 
          title="模板列表" 
          style={{ width: 220, flexShrink: 0, height: '100%', display: 'flex', flexDirection: 'column' }} 
          bodyStyle={{ padding: 8, flex: 1, overflow: 'auto' }} 
          extra={<Button type="link" size="small" icon={<PlusOutlined />} onClick={handleNew}>新建</Button>}
        >
          <List
            size="small"
            loading={loading}
            dataSource={templates}
            renderItem={(item: any) => (
              <List.Item
                style={{ 
                  cursor: 'pointer', 
                  background: selectedId === item.id ? '#e6f7ff' : 'transparent', 
                  padding: '8px 12px', 
                  borderRadius: 4,
                  marginBottom: 4
                }}
                onClick={() => handleSelect(item)}
                actions={[
                  <Popconfirm key="del" title="确定删除?" onConfirm={(e) => { e?.stopPropagation(); handleDelete(item.id); }}>
                    <Button type="link" danger size="small" icon={<DeleteOutlined />} onClick={(e) => e.stopPropagation()} />
                  </Popconfirm>
                ]}
              >
                <span style={{ fontSize: 13 }}>{item.name}</span>
              </List.Item>
            )}
          />
        </Card>

        {/* 中间编辑器 */}
        <Card 
          title="HTML 编辑" 
          style={{ flex: 1, height: '100%', display: 'flex', flexDirection: 'column' }} 
          bodyStyle={{ padding: 0, flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }} 
          extra={
            <Space>
              <Button type="primary" onClick={handleSave}>保存模板</Button>
            </Space>
          }
        >
          <Form form={form} layout="inline" style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0' }}>
            <Form.Item name="name" label="名称" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
              <Input style={{ width: 150 }} placeholder="模板名称" />
            </Form.Item>
            <Form.Item name="subject" label="邮件主题" style={{ marginBottom: 0 }}>
              <Input style={{ width: 250 }} placeholder="新闻日报 - {{.Date}}" />
            </Form.Item>
          </Form>
          <div style={{ flex: 1, overflow: 'hidden' }}>
            <Editor
              height="100%"
              language="html"
              value={content}
            onChange={(v) => setContent(v || '')}
            options={{
              minimap: { enabled: false },
              fontSize: 13,
              wordWrap: 'on',
              lineNumbers: 'on',
              scrollBeyondLastLine: false,
            }}
          />
          </div>
        </Card>

        {/* 右侧实时预览 */}
        <Card 
          title={`实时预览 (${newsCount} 条新闻)`} 
          style={{ flex: 1, height: '100%', display: 'flex', flexDirection: 'column' }} 
          bodyStyle={{ padding: 0, flex: 1, overflow: 'hidden', position: 'relative' }}
          extra={
            <Button icon={<ReloadOutlined />} size="small" onClick={handleRefreshPreview} loading={previewLoading}>
              刷新
            </Button>
          }
        >
          {previewLoading && (
            <div style={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)', zIndex: 10, background: 'rgba(255,255,255,0.8)', padding: 20, borderRadius: 8 }}>
              <Spin />
            </div>
          )}
          <div style={{ width: '100%', height: '100%', overflow: 'auto', background: '#f5f5f5' }}>
            <iframe 
              srcDoc={previewHtml} 
              style={{ width: '100%', minHeight: '100%', border: 'none', display: 'block', background: '#fff' }} 
              title="预览"
            />
          </div>
        </Card>
      </div>
    </div>
  );
};

export default TemplatesPage;
