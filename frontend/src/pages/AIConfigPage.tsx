import React, { useEffect, useState } from 'react';
import { Card, Form, Input, Select, Switch, Button, message, Divider, Space, AutoComplete } from 'antd';
import { getAIConfig, saveAIConfig, translateText, summarizeText } from '../api';

const AIConfigPage: React.FC = () => {
  const [form] = Form.useForm();
  const [testForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [testResult, setTestResult] = useState('');
  const [testing, setTesting] = useState(false);

  const fetchConfig = async () => {
    setLoading(true);
    try {
      const res = await getAIConfig();
      form.setFieldsValue(res.data);
    } catch {
      message.error('获取配置失败');
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchConfig();
  }, []);

  const handleSave = async (values: any) => {
    try {
      await saveAIConfig(values);
      message.success('保存成功');
    } catch {
      message.error('保存失败');
    }
  };

  const handleTestTranslate = async () => {
    const text = testForm.getFieldValue('test_text');
    if (!text) {
      message.warning('请输入测试文本');
      return;
    }
    setTesting(true);
    try {
      const res = await translateText(text);
      setTestResult(res.data.result);
    } catch (e: any) {
      message.error(e.response?.data?.error || '翻译失败');
    }
    setTesting(false);
  };

  const handleTestSummarize = async () => {
    const text = testForm.getFieldValue('test_text');
    if (!text) {
      message.warning('请输入测试文本');
      return;
    }
    setTesting(true);
    try {
      const res = await summarizeText(text);
      setTestResult(res.data.result);
    } catch (e: any) {
      message.error(e.response?.data?.error || '摘要失败');
    }
    setTesting(false);
  };

  return (
    <div>
      <div className="page-header">
        <h2>AI 设置</h2>
      </div>

      <div style={{ display: 'flex', gap: 24 }}>
        <Card title="AI 配置" style={{ flex: 1 }} loading={loading}>
          <Form form={form} layout="vertical" onFinish={handleSave} initialValues={{
            provider: 'openai',
            model: 'gpt-4o-mini',
            enable_trans: true,
            enable_summary: true,
            target_lang: 'zh-CN',
          }}>
            <Form.Item name="provider" label="AI 服务商">
              <Select options={[
                { value: 'openai', label: 'OpenAI' },
                { value: 'claude', label: 'Claude' },
                { value: 'ollama', label: 'Ollama (本地)' },
              ]} />
            </Form.Item>
            <Form.Item name="api_key" label="API Key" rules={[{ required: true }]}>
              <Input.Password placeholder="sk-..." />
            </Form.Item>
            <Form.Item name="base_url" label="Base URL" extra="留空使用默认地址">
              <Input placeholder="https://api.openai.com/v1" />
            </Form.Item>
            <Form.Item name="model" label="模型" extra="可选择预设或直接输入第三方平台的模型名称">
              <AutoComplete
                placeholder="选择或输入模型名称，如 gpt-4o-mini"
                options={[
                  { value: 'gpt-4o-mini', label: 'GPT-4o Mini' },
                  { value: 'gpt-4o', label: 'GPT-4o' },
                  { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
                  { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' },
                  { value: 'claude-3-5-sonnet-20241022', label: 'Claude 3.5 Sonnet' },
                  { value: 'claude-3-opus-20240229', label: 'Claude 3 Opus' },
                  { value: 'deepseek-chat', label: 'DeepSeek Chat' },
                  { value: 'deepseek-reasoner', label: 'DeepSeek Reasoner' },
                  { value: 'qwen-turbo', label: '通义千问 Turbo' },
                  { value: 'qwen-plus', label: '通义千问 Plus' },
                  { value: 'glm-4', label: 'GLM-4' },
                  { value: 'moonshot-v1-8k', label: 'Moonshot v1' },
                ]}
                filterOption={(inputValue, option) =>
                  option!.value.toLowerCase().indexOf(inputValue.toLowerCase()) !== -1
                }
              />
            </Form.Item>
            <Form.Item name="target_lang" label="翻译目标语言">
              <Select options={[
                { value: 'zh-CN', label: '简体中文' },
                { value: 'ug', label: 'ئۇيغۇرچە (维吾尔语)' },
                { value: 'zh-ug', label: '中文 + 维吾尔语 (双语)' },
                { value: 'zh-TW', label: '繁体中文' },
                { value: 'en', label: 'English' },
                { value: 'ja', label: '日本語' },
                { value: 'ko', label: '한국어 (韩语)' },
                { value: 'ar', label: 'العربية (阿拉伯语)' },
              ]} />
            </Form.Item>

            <Divider>功能开关</Divider>

            <Form.Item name="enable_trans" label="启用翻译" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item name="enable_summary" label="启用摘要" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item name="enable_filter" label="启用智能筛选" valuePropName="checked" extra="AI自动过滤低价值新闻">
              <Switch />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit">保存配置</Button>
            </Form.Item>
          </Form>
        </Card>

        <Card title="功能测试" style={{ flex: 1 }}>
          <Form form={testForm} layout="vertical">
            <Form.Item name="test_text" label="测试文本">
              <Input.TextArea rows={6} placeholder="输入要翻译或摘要的文本..." />
            </Form.Item>
            <Form.Item>
              <Space>
                <Button onClick={handleTestTranslate} loading={testing}>测试翻译</Button>
                <Button onClick={handleTestSummarize} loading={testing}>测试摘要</Button>
              </Space>
            </Form.Item>
          </Form>

          {testResult && (
            <div style={{ marginTop: 16 }}>
              <div style={{ fontWeight: 600, marginBottom: 8 }}>结果:</div>
              <div style={{ background: '#f5f5f5', padding: 16, borderRadius: 8, whiteSpace: 'pre-wrap' }}>
                {testResult}
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
};

export default AIConfigPage;
