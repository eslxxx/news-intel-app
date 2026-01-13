import React, { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, Switch, message, Popconfirm, Space, Tag, Card, InputNumber, Alert, Badge, Divider } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { getTasks, createTask, updateTask, deleteTask, runTask, getChannels, getTemplates, getAutoPushConfig, saveAutoPushConfig } from '../api';
import dayjs from 'dayjs';

const TasksPage: React.FC = () => {
  const [tasks, setTasks] = useState<any[]>([]);
  const [channels, setChannels] = useState<any[]>([]);
  const [templates, setTemplates] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form] = Form.useForm();

  // 自动打包推送配置
  const [autoPushConfig, setAutoPushConfig] = useState({
    enabled: false,
    threshold: 6,
    channel_id: '',
    template_id: '',
    pending_count: 0,
  });
  const [autoPushForm] = Form.useForm();

  const fetchData = async () => {
    setLoading(true);
    try {
      const [tasksRes, channelsRes, templatesRes, autoPushRes] = await Promise.all([
        getTasks(),
        getChannels(),
        getTemplates(),
        getAutoPushConfig(),
      ]);
      setTasks(tasksRes.data || []);
      setChannels(channelsRes.data || []);
      setTemplates(templatesRes.data || []);
      setAutoPushConfig(autoPushRes.data);
      autoPushForm.setFieldsValue(autoPushRes.data);
    } catch {
      message.error('获取失败');
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchData();
    // 定时刷新等待数量
    const interval = setInterval(async () => {
      try {
        const res = await getAutoPushConfig();
        setAutoPushConfig(res.data);
      } catch {}
    }, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleSubmit = async (values: any) => {
    try {
      const data = {
        ...values,
        categories: Array.isArray(values.categories) ? values.categories.join(',') : values.categories,
      };

      if (editingId) {
        await updateTask(editingId, data);
        message.success('更新成功');
      } else {
        await createTask(data);
        message.success('创建成功');
      }
      setModalOpen(false);
      form.resetFields();
      setEditingId(null);
      fetchData();
    } catch {
      message.error('操作失败');
    }
  };

  const handleEdit = (record: any) => {
    setEditingId(record.id);
    form.setFieldsValue({
      ...record,
      categories: record.categories ? record.categories.split(',') : [],
    });
    setModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteTask(id);
      message.success('删除成功');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const handleRun = async (id: string) => {
    try {
      await runTask(id);
      message.success('任务已启动');
    } catch {
      message.error('启动失败');
    }
  };

  const handleSaveAutoPush = async () => {
    try {
      const values = await autoPushForm.validateFields();
      await saveAutoPushConfig(values);
      message.success('自动推送配置已保存');
      fetchData();
    } catch {
      message.error('保存失败');
    }
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: 'Cron表达式', dataIndex: 'cron_expr', key: 'cron_expr' },
    {
      title: '分类',
      dataIndex: 'categories',
      key: 'categories',
      render: (v: string) => v?.split(',').map(c => <Tag key={c}>{c}</Tag>),
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (v: boolean) => <Switch checked={v} disabled />,
    },
    {
      title: '上次运行',
      dataIndex: 'last_run_at',
      key: 'last_run_at',
      render: (v: string) => v ? dayjs(v).format('MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        <Space>
          <Button type="link" icon={<PlayCircleOutlined />} onClick={() => handleRun(record.id)}>运行</Button>
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          <Popconfirm title="确定删除?" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const cronPresets = [
    { value: '0 8 * * *', label: '每天早上8点' },
    { value: '0 9,18 * * *', label: '每天9点和18点' },
    { value: '0 * * * *', label: '每小时' },
    { value: '0 8 * * 1', label: '每周一早上8点' },
  ];

  return (
    <div>
      <div className="page-header">
        <h2>推送任务</h2>
      </div>

      {/* 自动打包推送配置 */}
      <Card 
        title={
          <Space>
            <ThunderboltOutlined style={{ color: autoPushConfig.enabled ? '#52c41a' : '#999' }} />
            <span>自动打包推送</span>
            {autoPushConfig.enabled && (
              <Badge 
                count={`${autoPushConfig.pending_count}/${autoPushConfig.threshold}`} 
                style={{ backgroundColor: autoPushConfig.pending_count >= autoPushConfig.threshold ? '#52c41a' : '#1890ff' }}
              />
            )}
          </Space>
        }
        style={{ marginBottom: 24 }}
        size="small"
      >
        <Alert
          message="自动打包推送说明"
          description={`当阅读窗口中待推送的新闻数量达到设定阈值时，系统会自动将这批新闻打包推送。当前等待：${autoPushConfig.pending_count} 条`}
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Form form={autoPushForm} layout="inline" style={{ flexWrap: 'wrap', gap: 8 }}>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch checkedChildren="开" unCheckedChildren="关" />
          </Form.Item>
          <Form.Item name="threshold" label="触发数量">
            <InputNumber min={1} max={50} style={{ width: 80 }} />
          </Form.Item>
          <Form.Item name="channel_id" label="推送渠道" style={{ minWidth: 200 }}>
            <Select 
              options={channels.map(c => ({ value: c.id, label: `${c.name} (${c.type})` }))} 
              placeholder="选择渠道" 
              allowClear 
            />
          </Form.Item>
          <Form.Item name="template_id" label="邮件模板" style={{ minWidth: 180 }}>
            <Select 
              options={templates.map(t => ({ value: t.id, label: t.name }))} 
              placeholder="选择模板" 
              allowClear 
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={handleSaveAutoPush}>保存配置</Button>
          </Form.Item>
        </Form>
      </Card>

      <Divider />

      {/* 定时推送任务 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h3 style={{ margin: 0 }}>定时推送任务</h3>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setEditingId(null); setModalOpen(true); }}>
          添加任务
        </Button>
      </div>

      <Table columns={columns} dataSource={tasks} rowKey="id" loading={loading} />

      <Modal
        title={editingId ? '编辑任务' : '添加任务'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit} initialValues={{ enabled: true }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="每日新闻推送" />
          </Form.Item>
          <Form.Item name="cron_expr" label="Cron表达式" rules={[{ required: true }]} extra="例: 0 8 * * * 表示每天早上8点">
            <Select options={cronPresets} allowClear showSearch placeholder="选择或输入" mode="tags" maxCount={1} />
          </Form.Item>
          <Form.Item name="channel_id" label="推送渠道" rules={[{ required: true }]}>
            <Select options={channels.map(c => ({ value: c.id, label: c.name }))} placeholder="选择推送渠道" />
          </Form.Item>
          <Form.Item name="template_id" label="邮件模板">
            <Select options={templates.map(t => ({ value: t.id, label: t.name }))} placeholder="选择模板(可选)" allowClear />
          </Form.Item>
          <Form.Item name="categories" label="推送分类" rules={[{ required: true }]}>
            <Select mode="multiple" options={[
              { value: 'tech', label: '科技' },
              { value: 'ai', label: 'AI' },
              { value: 'github', label: 'GitHub' },
              { value: 'international', label: '国际' },
              { value: 'trending', label: '热门' },
            ]} placeholder="选择要推送的分类" />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TasksPage;
