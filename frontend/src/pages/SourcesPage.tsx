import React, { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, Switch, message, Popconfirm, Space } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { getSources, createSource, updateSource, deleteSource } from '../api';

const SourcesPage: React.FC = () => {
  const [sources, setSources] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form] = Form.useForm();

  const fetchSources = async () => {
    setLoading(true);
    try {
      const res = await getSources();
      setSources(res.data || []);
    } catch {
      message.error('获取失败');
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchSources();
  }, []);

  const handleSubmit = async (values: any) => {
    try {
      if (editingId) {
        await updateSource(editingId, values);
        message.success('更新成功');
      } else {
        await createSource(values);
        message.success('创建成功');
      }
      setModalOpen(false);
      form.resetFields();
      setEditingId(null);
      fetchSources();
    } catch {
      message.error('操作失败');
    }
  };

  const handleEdit = (record: any) => {
    setEditingId(record.id);
    form.setFieldsValue(record);
    setModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteSource(id);
      message.success('删除成功');
      fetchSources();
    } catch {
      message.error('删除失败');
    }
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '类型', dataIndex: 'type', key: 'type' },
    { title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true },
    { title: '分类', dataIndex: 'category', key: 'category' },
    { title: '间隔(分钟)', dataIndex: 'interval', key: 'interval' },
    {
      title: '启用',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (v: boolean) => <Switch checked={v} disabled />,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          <Popconfirm title="确定删除?" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between' }}>
        <h2>新闻源管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setEditingId(null); setModalOpen(true); }}>
          添加新闻源
        </Button>
      </div>

      <Table columns={columns} dataSource={sources} rowKey="id" loading={loading} />

      <Modal
        title={editingId ? '编辑新闻源' : '添加新闻源'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit} initialValues={{ type: 'rss', category: 'tech', enabled: true, interval: 60 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'rss', label: 'RSS' }, { value: 'api', label: 'API' }]} />
          </Form.Item>
          <Form.Item name="url" label="URL" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="category" label="分类">
            <Select options={[
              { value: 'tech', label: '科技' },
              { value: 'ai', label: 'AI' },
              { value: 'github', label: 'GitHub' },
              { value: 'international', label: '国际' },
              { value: 'trending', label: '热门' },
            ]} />
          </Form.Item>
          <Form.Item name="interval" label="采集间隔(分钟)">
            <Input type="number" />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SourcesPage;
