import React, { useEffect, useState, useCallback } from 'react';
import { Button, Card, Form, Input, message, Popconfirm, Space, List, Spin, Alert, Modal, Drawer } from 'antd';
import { PlusOutlined, DeleteOutlined, ReloadOutlined, RobotOutlined, SaveOutlined, SendOutlined } from '@ant-design/icons';
import Editor from '@monaco-editor/react';
import { getTemplates, createTemplate, updateTemplate, deleteTemplate, previewTemplate, aiGenerateTemplate } from '../api';
import debounce from 'lodash/debounce';

const { TextArea } = Input;

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
  
  // AI 生成相关状态
  const [aiDrawerOpen, setAiDrawerOpen] = useState(false);
  const [aiPrompt, setAiPrompt] = useState('');
  const [aiGenerating, setAiGenerating] = useState(false);
  const [aiGeneratedTemplate, setAiGeneratedTemplate] = useState('');
  const [aiPreviewHtml, setAiPreviewHtml] = useState('');
  const [aiPreviewLoading, setAiPreviewLoading] = useState(false);
  const [chatHistory, setChatHistory] = useState<{role: 'user' | 'ai', content: string}[]>([]);

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

  // AI 生成模板预览
  const previewAiTemplate = async (templateContent: string) => {
    if (!templateContent) return;
    setAiPreviewLoading(true);
    try {
      const res = await previewTemplate(templateContent);
      setAiPreviewHtml(res.data.html);
    } catch {
      setAiPreviewHtml('<div style="padding:20px;color:#f00;">模板语法错误</div>');
    }
    setAiPreviewLoading(false);
  };

  useEffect(() => {
    fetchTemplates();
    debouncedPreview(content);
  }, []);

  useEffect(() => {
    debouncedPreview(content);
  }, [content]);

  // 当 AI 生成模板时自动预览
  useEffect(() => {
    if (aiGeneratedTemplate) {
      previewAiTemplate(aiGeneratedTemplate);
    }
  }, [aiGeneratedTemplate]);

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

  // AI 生成模板
  const handleAiGenerate = async () => {
    if (!aiPrompt.trim()) {
      message.warning('请输入模板设计需求');
      return;
    }

    setAiGenerating(true);
    setChatHistory(prev => [...prev, { role: 'user', content: aiPrompt }]);

    try {
      const res = await aiGenerateTemplate(aiPrompt, aiGeneratedTemplate || undefined);
      const generatedTemplate = res.data.template;
      setAiGeneratedTemplate(generatedTemplate);
      setChatHistory(prev => [...prev, { role: 'ai', content: '已根据您的需求生成模板，请查看右侧预览效果。' }]);
      setAiPrompt('');
    } catch (err: any) {
      message.error(err.response?.data?.error || 'AI 生成失败');
      setChatHistory(prev => [...prev, { role: 'ai', content: '生成失败，请重试。' }]);
    }
    setAiGenerating(false);
  };

  // 应用 AI 生成的模板到编辑器
  const handleApplyAiTemplate = () => {
    if (!aiGeneratedTemplate) {
      message.warning('还没有生成模板');
      return;
    }
    setContent(aiGeneratedTemplate);
    setAiDrawerOpen(false);
    message.success('模板已应用到编辑器');
  };

  // 一键保存 AI 生成的模板
  const handleSaveAiTemplate = async () => {
    if (!aiGeneratedTemplate) {
      message.warning('还没有生成模板');
      return;
    }

    Modal.confirm({
      title: '保存模板',
      content: (
        <Form layout="vertical" id="saveAiTemplateForm">
          <Form.Item label="模板名称" required>
            <Input id="aiTemplateName" placeholder="请输入模板名称" />
          </Form.Item>
          <Form.Item label="邮件主题">
            <Input id="aiTemplateSubject" placeholder="新闻日报 - {{.Date}}" />
          </Form.Item>
        </Form>
      ),
      onOk: async () => {
        const nameInput = document.getElementById('aiTemplateName') as HTMLInputElement;
        const subjectInput = document.getElementById('aiTemplateSubject') as HTMLInputElement;
        
        if (!nameInput?.value) {
          message.error('请输入模板名称');
          throw new Error('名称不能为空');
        }

        try {
          await createTemplate({
            name: nameInput.value,
            subject: subjectInput?.value || '新闻日报 - {{.Date}}',
            content: aiGeneratedTemplate,
          });
          message.success('模板保存成功');
          fetchTemplates();
          setAiDrawerOpen(false);
          // 重置 AI 对话
          setAiGeneratedTemplate('');
          setAiPreviewHtml('');
          setChatHistory([]);
        } catch {
          message.error('保存失败');
          throw new Error('保存失败');
        }
      },
    });
  };

  // 打开 AI 助手
  const openAiDrawer = () => {
    setAiDrawerOpen(true);
    setAiGeneratedTemplate('');
    setAiPreviewHtml('');
    setChatHistory([]);
    setAiPrompt('');
  };

  return (
    <div>
      <div className="page-header">
        <h2>邮件模板编辑器</h2>
        <Button 
          type="primary" 
          icon={<RobotOutlined />} 
          onClick={openAiDrawer}
          style={{ marginLeft: 16 }}
        >
          AI 创建模板
        </Button>
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

      {/* AI 生成模板抽屉 */}
      <Drawer
        title={
          <Space>
            <RobotOutlined />
            <span>AI 模板设计助手</span>
          </Space>
        }
        placement="right"
        width="80%"
        open={aiDrawerOpen}
        onClose={() => setAiDrawerOpen(false)}
        extra={
          <Space>
            <Button onClick={handleApplyAiTemplate} disabled={!aiGeneratedTemplate}>
              应用到编辑器
            </Button>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSaveAiTemplate} disabled={!aiGeneratedTemplate}>
              保存模板
            </Button>
          </Space>
        }
      >
        <div style={{ display: 'flex', gap: 16, height: 'calc(100vh - 150px)' }}>
          {/* 左侧对话区域 */}
          <div style={{ width: 400, display: 'flex', flexDirection: 'column' }}>
            <Alert
              message="使用说明"
              description="描述你想要的邮件模板样式，AI 会为你生成。如果不满意可以继续提出修改意见。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
            
            {/* 对话历史 */}
            <div style={{ 
              flex: 1, 
              overflow: 'auto', 
              border: '1px solid #f0f0f0', 
              borderRadius: 8, 
              padding: 16,
              marginBottom: 16,
              background: '#fafafa'
            }}>
              {chatHistory.length === 0 ? (
                <div style={{ color: '#999', textAlign: 'center', padding: 40 }}>
                  <RobotOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                  <p>告诉我你想要什么样的邮件模板</p>
                  <p style={{ fontSize: 12 }}>例如：深色主题、简约风格、卡片式布局...</p>
                </div>
              ) : (
                chatHistory.map((msg, idx) => (
                  <div 
                    key={idx} 
                    style={{ 
                      marginBottom: 12,
                      display: 'flex',
                      justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start'
                    }}
                  >
                    <div style={{
                      maxWidth: '85%',
                      padding: '8px 12px',
                      borderRadius: 8,
                      background: msg.role === 'user' ? '#1890ff' : '#fff',
                      color: msg.role === 'user' ? '#fff' : '#333',
                      boxShadow: '0 1px 2px rgba(0,0,0,0.1)'
                    }}>
                      {msg.content}
                    </div>
                  </div>
                ))
              )}
              {aiGenerating && (
                <div style={{ display: 'flex', justifyContent: 'flex-start' }}>
                  <div style={{ padding: '8px 12px', background: '#fff', borderRadius: 8 }}>
                    <Spin size="small" /> 正在生成...
                  </div>
                </div>
              )}
            </div>

            {/* 输入区域 */}
            <div style={{ display: 'flex', gap: 8 }}>
              <TextArea
                value={aiPrompt}
                onChange={(e) => setAiPrompt(e.target.value)}
                placeholder="描述你想要的模板样式，例如：暗色主题科技风格，使用蓝紫色渐变..."
                autoSize={{ minRows: 2, maxRows: 4 }}
                onPressEnter={(e) => {
                  if (!e.shiftKey) {
                    e.preventDefault();
                    handleAiGenerate();
                  }
                }}
                disabled={aiGenerating}
              />
              <Button 
                type="primary" 
                icon={<SendOutlined />} 
                onClick={handleAiGenerate}
                loading={aiGenerating}
                style={{ height: 'auto' }}
              >
                发送
              </Button>
            </div>
          </div>

          {/* 右侧预览区域 */}
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
            <div style={{ marginBottom: 8, fontWeight: 500 }}>
              实时预览
            </div>
            <div style={{ 
              flex: 1, 
              border: '1px solid #f0f0f0', 
              borderRadius: 8, 
              overflow: 'hidden',
              background: '#f5f5f5',
              position: 'relative'
            }}>
              {aiPreviewLoading && (
                <div style={{ 
                  position: 'absolute', 
                  top: '50%', 
                  left: '50%', 
                  transform: 'translate(-50%, -50%)', 
                  zIndex: 10,
                  background: 'rgba(255,255,255,0.9)',
                  padding: 20,
                  borderRadius: 8
                }}>
                  <Spin />
                </div>
              )}
              {aiGeneratedTemplate ? (
                <iframe 
                  srcDoc={aiPreviewHtml} 
                  style={{ width: '100%', height: '100%', border: 'none', background: '#fff' }} 
                  title="AI预览"
                />
              ) : (
                <div style={{ 
                  height: '100%', 
                  display: 'flex', 
                  alignItems: 'center', 
                  justifyContent: 'center',
                  color: '#999'
                }}>
                  <div style={{ textAlign: 'center' }}>
                    <RobotOutlined style={{ fontSize: 64, marginBottom: 16, opacity: 0.3 }} />
                    <p>AI 生成的模板预览将显示在这里</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </Drawer>
    </div>
  );
};

export default TemplatesPage;
