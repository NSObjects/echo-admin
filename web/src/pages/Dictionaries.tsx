import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  message,
  Space,
  Switch,
  Table,
  Tag,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  addDictionaryItem,
  createDictionary,
  type Dictionary,
  type DictionaryItem,
  listDictionaries,
  updateDictionaryItem,
} from '@/services/admin';

type DictionaryFormValues = {
  code: string;
  name: string;
};

type ItemFormValues = {
  label: string;
  value: string;
  sort: number;
  active: boolean;
};

type ItemTarget = {
  code: string;
  item?: DictionaryItem;
};

const Dictionaries: React.FC = () => {
  const access = useAccess();
  const [dictionaries, setDictionaries] = useState<Dictionary[]>([]);
  const [loading, setLoading] = useState(false);
  const [dictionaryModalOpen, setDictionaryModalOpen] = useState(false);
  const [itemTarget, setItemTarget] = useState<ItemTarget>();
  const [dictionaryForm] = Form.useForm<DictionaryFormValues>();
  const [itemForm] = Form.useForm<ItemFormValues>();

  const loadData = async () => {
    setLoading(true);
    try {
      setDictionaries(await listDictionaries());
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

  const openDictionaryModal = () => {
    dictionaryForm.resetFields();
    setDictionaryModalOpen(true);
  };

  const submitDictionary = async () => {
    const values = await dictionaryForm.validateFields();
    await createDictionary(values);
    message.success('字典已创建');
    setDictionaryModalOpen(false);
    await loadData();
  };

  const openItemModal = (code: string, item?: DictionaryItem) => {
    setItemTarget({ code, item });
    itemForm.resetFields();
    itemForm.setFieldsValue(
      item ?? { label: '', value: '', sort: 100, active: true },
    );
  };

  const submitItem = async () => {
    const target = itemTarget;
    if (!target) return;
    const values = await itemForm.validateFields();
    if (target.item) {
      await updateDictionaryItem(target.code, target.item.id, values);
      message.success('字典项已更新');
    } else {
      await addDictionaryItem(target.code, values);
      message.success('字典项已创建');
    }
    setItemTarget(undefined);
    await loadData();
  };

  const itemColumns: ColumnsType<DictionaryItem> = [
    { title: '标签', dataIndex: 'label' },
    { title: '值', dataIndex: 'value' },
    { title: '排序', dataIndex: 'sort', width: 88 },
    {
      title: '状态',
      dataIndex: 'active',
      render: (active: boolean) => (
        <Tag color={active ? 'green' : 'default'}>
          {active ? '启用' : '停用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, item) =>
        access.canDictUpdate ? (
          <Button
            type="link"
            onClick={() => openItemModal(itemTarget?.code ?? '', item)}
          >
            编辑
          </Button>
        ) : null,
    },
  ];

  const columns: ColumnsType<Dictionary> = [
    { title: '编码', dataIndex: 'code' },
    { title: '名称', dataIndex: 'name' },
    {
      title: '字典项',
      dataIndex: 'items',
      render: (items: DictionaryItem[]) => `${items.length} 项`,
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) =>
        access.canDictCreate ? (
          <Button type="link" onClick={() => openItemModal(record.code)}>
            新增字典项
          </Button>
        ) : null,
    },
  ];

  return (
    <PageContainer title="数据字典">
      <Table<Dictionary>
        rowKey="code"
        columns={columns}
        dataSource={dictionaries}
        loading={loading}
        pagination={false}
        expandable={{
          expandedRowRender: (record) => (
            <Table<DictionaryItem>
              rowKey="id"
              columns={itemColumns.map((column) =>
                column.key === 'actions'
                  ? {
                      ...column,
                      render: (_, item: DictionaryItem) =>
                        access.canDictUpdate ? (
                          <Button
                            type="link"
                            onClick={() => openItemModal(record.code, item)}
                          >
                            编辑
                          </Button>
                        ) : null,
                    }
                  : column,
              )}
              dataSource={record.items}
              pagination={false}
              size="small"
            />
          ),
        }}
        title={() => (
          <Space>
            {access.canDictCreate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={openDictionaryModal}
              >
                新增字典
              </Button>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title="新增字典"
        open={dictionaryModalOpen}
        onOk={() => void submitDictionary()}
        onCancel={() => setDictionaryModalOpen(false)}
        destroyOnHidden
      >
        <Form<DictionaryFormValues> form={dictionaryForm} layout="vertical">
          <Form.Item
            label="编码"
            name="code"
            rules={[{ required: true, message: '请输入字典编码' }]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入字典名称' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title={itemTarget?.item ? '编辑字典项' : '新增字典项'}
        open={Boolean(itemTarget)}
        onOk={() => void submitItem()}
        onCancel={() => setItemTarget(undefined)}
        destroyOnHidden
      >
        <Form<ItemFormValues> form={itemForm} layout="vertical">
          <Form.Item
            label="标签"
            name="label"
            rules={[{ required: true, message: '请输入标签' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
          <Form.Item
            label="值"
            name="value"
            rules={[{ required: true, message: '请输入值' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
          <Form.Item label="排序" name="sort">
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="启用" name="active" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default Dictionaries;
