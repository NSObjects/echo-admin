import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Form,
  Input,
  Modal,
  message,
  Space,
  Switch,
  Table,
  Tag,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import { listConfigs, type SystemConfig, upsertConfig } from '@/services/admin';

type ConfigFormValues = {
  key: string;
  name: string;
  value: string;
  public: boolean;
};

const Configs: React.FC = () => {
  const access = useAccess();
  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<SystemConfig>();
  const [form] = Form.useForm<ConfigFormValues>();

  const loadData = async () => {
    setLoading(true);
    try {
      setConfigs(await listConfigs());
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

  const openCreate = () => {
    setEditing(undefined);
    form.resetFields();
    form.setFieldsValue({ public: false, value: '' });
    setModalOpen(true);
  };

  const openEdit = (record: SystemConfig) => {
    setEditing(record);
    form.setFieldsValue(record);
    setModalOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    await upsertConfig(values.key, {
      name: values.name,
      value: values.value,
      public: values.public,
    });
    message.success(editing ? '配置已更新' : '配置已创建');
    setModalOpen(false);
    await loadData();
  };

  const columns: ColumnsType<SystemConfig> = [
    { title: '键', dataIndex: 'key' },
    { title: '名称', dataIndex: 'name' },
    { title: '值', dataIndex: 'value', ellipsis: true },
    {
      title: '公开',
      dataIndex: 'public',
      render: (value: boolean) => (
        <Tag color={value ? 'green' : 'default'}>{value ? '是' : '否'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) =>
        access.canConfigUpdate ? (
          <Button type="link" onClick={() => openEdit(record)}>
            编辑
          </Button>
        ) : null,
    },
  ];

  return (
    <PageContainer title="系统配置">
      <Table<SystemConfig>
        rowKey="key"
        columns={columns}
        dataSource={configs}
        loading={loading}
        pagination={false}
        title={() => (
          <Space>
            {access.canConfigUpdate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={openCreate}
              >
                新增配置
              </Button>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title={editing ? '编辑配置' : '新增配置'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form<ConfigFormValues> form={form} layout="vertical">
          <Form.Item
            label="键"
            name="key"
            rules={[{ required: true, message: '请输入配置键' }]}
          >
            <Input disabled={Boolean(editing)} maxLength={120} />
          </Form.Item>
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入配置名称' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
          <Form.Item label="值" name="value">
            <Input.TextArea maxLength={4000} rows={4} />
          </Form.Item>
          <Form.Item label="公开" name="public" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default Configs;
