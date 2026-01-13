import React, { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, Switch, message, Popconfirm, Space, InputNumber } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SendOutlined } from '@ant-design/icons';
import { getChannels, createChannel, updateChannel, deleteChannel, testChannel } from '../api';

const ChannelsPage: React.FC = () => {
  const [channels, setChannels] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [channelType, setChannelType] = useState('email');
  const [form] = Form.useForm();

  const fetchChannels = async () => {
    setLoading(true);
    try {
      const res = await getChannels();
      setChannels(res.data || []);
    } catch {
      message.error('获取失败');
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchChannels();
  }, []);

  const handleSubmit = async (values: any) => {
    try {
      const config: any = {};
      if (values.type === 'email') {
        config.smtp_host = values.smtp_host;
        config.smtp_port = values.smtp_port;
        config.username = values.username;
        config.password = values.password;
        config.from_address = values.from_address;
        config.from_name = values.from_name;
        config.to_addresses = values.to_addresses;
      } else if (values.type === 'ntfy') {
        config.server_url = values.server_url || 'https://ntfy.sh';
        config.topic = values.topic;
        config.token = values.token;
      }

      const data = {
        name: values.name,
        type: values.type,
        config: JSON.stringify(config),
        enabled: values.enabled,
      };

      if (editingId) {
        await updateChannel(editingId, data);
        message.success('更新成功');
      } else {
        await createChannel(data);
        message.success('创建成功');
      }
      setModalOpen(false);
      form.resetFields();
      setEditingId(null);
      fetchChannels();
    } catch {
      message.error('操作失败');
    }
  };

  const handleEdit = (record: any) => {
    setEditingId(record.id);
    const config = JSON.parse(record.config || '{}');
    setChannelType(record.type);
    form.setFieldsValue({
      ...record,
      ...config,
    });
    setModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteChannel(id);
      message.success('删除成功');
      fetchChannels();
    } catch {
      message.error('删除失败');
    }
  };

  const handleTest = async (id: string) => {
    try {
      await testChannel(id);
      message.success('测试消息已发送');
    } catch (e: any) {
      message.error(e.response?.data?.error || '测试失败');
    }
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '类型', dataIndex: 'type', key: 'type' },
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
          <Button type="link" icon={<SendOutlined />} onClick={() => handleTest(record.id)}>测试</Button>
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
        <h2>推送渠道</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setEditingId(null); setChannelType('email'); setModalOpen(true); }}>
          添加渠道
        </Button>
      </div>

      <Table columns={columns} dataSource={channels} rowKey="id" loading={loading} />

      <Modal
        title={editingId ? '编辑渠道' : '添加渠道'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit} initialValues={{ type: 'email', enabled: true }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select onChange={(v) => setChannelType(v)} options={[
              { value: 'email', label: '邮箱' },
              { value: 'ntfy', label: 'ntfy' },
            ]} />
          </Form.Item>

          {channelType === 'email' && (
            <>
              <Form.Item name="smtp_host" label="SMTP服务器" rules={[{ required: true }]}>
                <Input placeholder="smtp.gmail.com" />
              </Form.Item>
              <Form.Item name="smtp_port" label="SMTP端口" rules={[{ required: true }]}>
                <InputNumber style={{ width: '100%' }} placeholder="587" />
              </Form.Item>
              <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="password" label="密码" rules={[{ required: true }]}>
                <Input.Password />
              </Form.Item>
              <Form.Item name="from_address" label="发件人邮箱" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="from_name" label="发件人名称">
                <Input placeholder="News Intel" />
              </Form.Item>
              <Form.Item name="to_addresses" label="收件人邮箱(逗号分隔)" rules={[{ required: true }]}>
                <Input placeholder="user1@example.com,user2@example.com" />
              </Form.Item>
            </>
          )}

          {channelType === 'ntfy' && (
            <>
              <Form.Item name="server_url" label="服务器URL">
                <Input placeholder="https://ntfy.sh" />
              </Form.Item>
              <Form.Item name="topic" label="Topic" rules={[{ required: true }]}>
                <Input placeholder="my-news-topic" />
              </Form.Item>
              <Form.Item name="token" label="Token(可选)">
                <Input.Password />
              </Form.Item>
            </>
          )}

          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ChannelsPage;
