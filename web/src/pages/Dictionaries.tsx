import {
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  PlusOutlined,
  ReloadOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  message,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Upload,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  addDictionaryItem,
  createDictionary,
  type Dictionary,
  type DictionaryBundle,
  type DictionaryItem,
  deleteDictionary,
  deleteDictionaryItem,
  exportDictionaries,
  importDictionaries,
  listDictionaries,
  updateDictionary,
  updateDictionaryItem,
} from '@/services/admin';

type DictionaryFormValues = {
  code: string;
  name: string;
};

type ItemFormValues = {
  parent_id?: number;
  label: string;
  value: string;
  extend?: string;
  sort: number;
  active: boolean;
};

type ItemTarget = {
  code: string;
  dictionary: Dictionary;
  item?: DictionaryItem;
  parentID?: number;
};

const Dictionaries: React.FC = () => {
  const access = useAccess();
  const [dictionaries, setDictionaries] = useState<Dictionary[]>([]);
  const [loading, setLoading] = useState(false);
  const [dictionaryModalOpen, setDictionaryModalOpen] = useState(false);
  const [editingDictionary, setEditingDictionary] = useState<Dictionary>();
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

  const openDictionaryModal = (dictionary?: Dictionary) => {
    setEditingDictionary(dictionary);
    dictionaryForm.resetFields();
    if (dictionary) {
      dictionaryForm.setFieldsValue({
        code: dictionary.code,
        name: dictionary.name,
      });
    }
    setDictionaryModalOpen(true);
  };

  const submitDictionary = async () => {
    const values = await dictionaryForm.validateFields();
    if (editingDictionary) {
      await updateDictionary(editingDictionary.code, { name: values.name });
      message.success('字典已更新');
    } else {
      await createDictionary(values);
      message.success('字典已创建');
    }
    setDictionaryModalOpen(false);
    setEditingDictionary(undefined);
    await loadData();
  };

  const openItemModal = (
    dictionary: Dictionary,
    item?: DictionaryItem,
    parentID?: number,
  ) => {
    setItemTarget({ code: dictionary.code, dictionary, item, parentID });
    itemForm.resetFields();
    itemForm.setFieldsValue(
      item ?? {
        parent_id: parentID,
        label: '',
        value: '',
        extend: '',
        sort: 100,
        active: true,
      },
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

  const removeDictionary = async (record: Dictionary) => {
    await deleteDictionary(record.code);
    message.success('字典已删除');
    await loadData();
  };

  const removeDictionaryItem = async (code: string, item: DictionaryItem) => {
    await deleteDictionaryItem(code, item.id);
    message.success('字典项已删除');
    await loadData();
  };

  const downloadDictionaries = async () => {
    const blob = await exportDictionaries();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = 'dictionaries.json';
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const importDictionaryFile = async (file: File) => {
    let bundle: DictionaryBundle;
    try {
      bundle = JSON.parse(await file.text()) as DictionaryBundle;
    } catch {
      message.error('字典文件不是有效JSON');
      return Upload.LIST_IGNORE;
    }
    await importDictionaries(bundle);
    message.success('字典已导入');
    await loadData();
    return Upload.LIST_IGNORE;
  };

  const itemColumns = (dictionary: Dictionary): ColumnsType<DictionaryItem> => [
    { title: '标签', dataIndex: 'label' },
    { title: '值', dataIndex: 'value' },
    {
      title: '扩展',
      dataIndex: 'extend',
      render: (extend?: string) => extend || '-',
    },
    { title: '层级', dataIndex: 'level', width: 80 },
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
      render: (_, item) => (
        <Space>
          {access.canDictCreate ? (
            <Button
              type="link"
              icon={<PlusOutlined />}
              onClick={() => openItemModal(dictionary, undefined, item.id)}
            >
              新增子项
            </Button>
          ) : null}
          {access.canDictUpdate ? (
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => openItemModal(dictionary, item)}
            >
              编辑
            </Button>
          ) : null}
          {access.canDictDelete ? (
            <Popconfirm
              title="删除字典项"
              description={`确认删除 ${item.label}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeDictionaryItem(dictionary.code, item)}
            >
              <Button danger type="link" icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          ) : null}
        </Space>
      ),
    },
  ];

  const columns: ColumnsType<Dictionary> = [
    { title: '编码', dataIndex: 'code' },
    { title: '名称', dataIndex: 'name' },
    {
      title: '字典项',
      dataIndex: 'items',
      render: (items: DictionaryItem[]) => `${countItems(items)} 项`,
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          {access.canDictCreate ? (
            <Button type="link" onClick={() => openItemModal(record)}>
              新增字典项
            </Button>
          ) : null}
          {access.canDictUpdate ? (
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => openDictionaryModal(record)}
            >
              编辑
            </Button>
          ) : null}
          {access.canDictDelete ? (
            <Popconfirm
              title="删除字典"
              description={`确认删除 ${record.name}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeDictionary(record)}
            >
              <Button danger type="link" icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          ) : null}
        </Space>
      ),
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
              columns={itemColumns(record)}
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
                onClick={() => openDictionaryModal()}
              >
                新增字典
              </Button>
            ) : null}
            {access.canDictRead ? (
              <Button
                icon={<DownloadOutlined />}
                onClick={() => void downloadDictionaries()}
              >
                导出
              </Button>
            ) : null}
            {access.canDictCreate ? (
              <Upload
                accept="application/json"
                showUploadList={false}
                beforeUpload={(file) => importDictionaryFile(file)}
              >
                <Button icon={<UploadOutlined />}>导入</Button>
              </Upload>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title={editingDictionary ? '编辑字典' : '新增字典'}
        open={dictionaryModalOpen}
        onOk={() => void submitDictionary()}
        onCancel={() => {
          setDictionaryModalOpen(false);
          setEditingDictionary(undefined);
        }}
        destroyOnHidden
      >
        <Form<DictionaryFormValues> form={dictionaryForm} layout="vertical">
          <Form.Item
            label="编码"
            name="code"
            rules={[{ required: true, message: '请输入字典编码' }]}
          >
            <Input disabled={Boolean(editingDictionary)} maxLength={80} />
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
          <Form.Item label="父级" name="parent_id">
            <Select
              allowClear
              options={parentOptions(itemTarget).map((item) => ({
                label: item.label,
                value: item.id,
              }))}
            />
          </Form.Item>
          <Form.Item
            label="标签"
            name="label"
            rules={[{ required: true, message: '请输入标签' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
          <Form.Item label="扩展值" name="extend" rules={[{ max: 4000 }]}>
            <Input.TextArea maxLength={4000} rows={3} />
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

type ItemOption = {
  id: number;
  label: string;
  path: string;
};

const parentOptions = (target?: ItemTarget): ItemOption[] => {
  if (!target) {
    return [];
  }
  const editingID = target.item?.id;
  return flattenItems(target.dictionary.items).filter((item) => {
    if (!editingID) {
      return true;
    }
    if (item.id === editingID) {
      return false;
    }
    return !item.path
      .split(',')
      .filter(Boolean)
      .map(Number)
      .includes(editingID);
  });
};

const flattenItems = (
  items: DictionaryItem[],
  depth = 0,
): ItemOption[] =>
  items.flatMap((item) => [
    {
      id: item.id,
      label: `${'  '.repeat(depth)}${item.label}`,
      path: item.path,
    },
    ...flattenItems(item.children, depth + 1),
  ]);

const countItems = (items: DictionaryItem[]): number =>
  items.reduce((total, item) => total + 1 + countItems(item.children), 0);

export default Dictionaries;
