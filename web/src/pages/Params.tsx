import {
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Descriptions,
  Form,
  Input,
  Modal,
  message,
  Popconfirm,
  Space,
  Table,
} from 'antd';
import type { DescriptionsProps } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  batchDeleteParams,
  createParam,
  deleteParam,
  readParam,
  type PageMeta,
  type ParamInput,
  listParams,
  type SystemParam,
  updateParam,
} from '@/services/admin';

type ParamSearchValues = {
  name?: string;
  key?: string;
};

const formatDate = (value?: string) =>
  value ? new Date(value).toLocaleString() : '-';

const Params: React.FC = () => {
  const access = useAccess();
  const [params, setParams] = useState<SystemParam[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailItems, setDetailItems] = useState<DescriptionsProps['items']>(
    [],
  );
  const [editing, setEditing] = useState<SystemParam>();
  const [selectedIDs, setSelectedIDs] = useState<React.Key[]>([]);
  const [form] = Form.useForm<ParamInput>();
  const [searchForm] = Form.useForm<ParamSearchValues>();

  const loadData = async (
    nextPage = page?.page ?? 1,
    nextPageSize = page?.page_size ?? 20,
  ) => {
    setLoading(true);
    try {
      const filters = searchForm.getFieldsValue();
      const response = await listParams({
        page: nextPage,
        page_size: nextPageSize,
        name: filters.name,
        key: filters.key,
      });
      setParams(response.data);
      setPage(response.page);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData(1, 20);
  }, []);

  const openCreate = () => {
    setEditing(undefined);
    form.resetFields();
    setModalOpen(true);
  };

  const openEdit = (record: SystemParam) => {
    setEditing(record);
    form.setFieldsValue({
      name: record.name,
      key: record.key,
      value: record.value,
      desc: record.desc,
    });
    setModalOpen(true);
  };

  const openDetail = async (record: SystemParam) => {
    const detail = await readParam(record.id);
    setDetailItems([
      { key: 'name', label: '名称', children: detail.name },
      { key: 'key', label: '键', children: detail.key },
      { key: 'value', label: '值', children: detail.value },
      { key: 'desc', label: '说明', children: detail.desc || '-' },
      {
        key: 'created_at',
        label: '创建时间',
        children: formatDate(detail.created_at),
      },
      {
        key: 'updated_at',
        label: '更新时间',
        children: formatDate(detail.updated_at),
      },
    ]);
    setDetailOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    if (editing) {
      await updateParam(editing.id, values);
      message.success('参数已更新');
    } else {
      await createParam(values);
      message.success('参数已创建');
    }
    setModalOpen(false);
    await loadData();
  };

  const removeParam = async (record: SystemParam) => {
    await deleteParam(record.id);
    message.success('参数已删除');
    await loadData();
  };

  const removeSelected = async () => {
    await batchDeleteParams(selectedIDs.map(Number));
    message.success('参数已批量删除');
    setSelectedIDs([]);
    await loadData();
  };

  const columns: ColumnsType<SystemParam> = [
    { title: '名称', dataIndex: 'name', width: 180 },
    { title: '键', dataIndex: 'key', width: 180 },
    { title: '值', dataIndex: 'value', ellipsis: true },
    { title: '说明', dataIndex: 'desc', ellipsis: true },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 180,
      render: formatDate,
    },
    {
      title: '操作',
      key: 'actions',
      width: 220,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => void openDetail(record)}
          >
            详情
          </Button>
          {access.canParamUpdate ? (
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => openEdit(record)}
            >
              编辑
            </Button>
          ) : null}
          {access.canParamDelete ? (
            <Popconfirm
              title="删除参数"
              description={`确认删除 ${record.key}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeParam(record)}
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
    <PageContainer title="系统参数">
      <Table<SystemParam>
        rowKey="id"
        columns={columns}
        dataSource={params}
        loading={loading}
        rowSelection={
          access.canParamDelete
            ? {
                selectedRowKeys: selectedIDs,
                onChange: setSelectedIDs,
              }
            : undefined
        }
        pagination={{
          current: page?.page ?? 1,
          pageSize: page?.page_size ?? 20,
          total: page?.total ?? 0,
          showSizeChanger: true,
          onChange: (nextPage, nextPageSize) => {
            void loadData(nextPage, nextPageSize);
          },
        }}
        title={() => (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Form<ParamSearchValues> form={searchForm} layout="inline">
              <Form.Item name="name">
                <Input allowClear placeholder="名称" />
              </Form.Item>
              <Form.Item name="key">
                <Input allowClear placeholder="键" />
              </Form.Item>
              <Button
                icon={<SearchOutlined />}
                onClick={() => void loadData(1, page?.page_size ?? 20)}
              >
                查询
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => {
                  searchForm.resetFields();
                  void loadData(1, page?.page_size ?? 20);
                }}
              >
                重置
              </Button>
            </Form>
            <Space wrap>
              {access.canParamCreate ? (
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={openCreate}
                >
                  新增参数
                </Button>
              ) : null}
              {access.canParamDelete && selectedIDs.length > 0 ? (
                <Popconfirm
                  title="批量删除参数"
                  description={`确认删除选中的 ${selectedIDs.length} 条参数？`}
                  okText="删除"
                  okButtonProps={{ danger: true }}
                  onConfirm={() => void removeSelected()}
                >
                  <Button danger icon={<DeleteOutlined />}>
                    批量删除
                  </Button>
                </Popconfirm>
              ) : null}
              <Button
                icon={<ReloadOutlined />}
                onClick={() => void loadData()}
              >
                刷新
              </Button>
            </Space>
          </Space>
        )}
      />
      <Modal
        title={editing ? '编辑参数' : '新增参数'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form<ParamInput> form={form} layout="vertical">
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入参数名称' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
          <Form.Item
            label="键"
            name="key"
            rules={[{ required: true, message: '请输入参数键' }]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item
            label="值"
            name="value"
            rules={[{ required: true, message: '请输入参数值' }]}
          >
            <Input.TextArea maxLength={4000} rows={4} />
          </Form.Item>
          <Form.Item label="说明" name="desc">
            <Input.TextArea maxLength={4000} rows={4} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title="参数详情"
        open={detailOpen}
        footer={null}
        onCancel={() => setDetailOpen(false)}
        destroyOnHidden
      >
        <Descriptions column={1} items={detailItems} />
      </Modal>
    </PageContainer>
  );
};

export default Params;
