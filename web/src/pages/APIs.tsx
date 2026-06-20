import {
  DeleteOutlined,
  EyeOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  AutoComplete,
  Button,
  Descriptions,
  Form,
  Input,
  Modal,
  message,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
} from 'antd';
import type { DescriptionsProps } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  type APIInput,
  type APIResource,
  batchDeleteAPIs,
  createAPI,
  deleteAPI,
  readAPI,
  listAPIRoles,
  type ListParams,
  listAPIs,
  listAPIGroups,
  listPermissions,
  listRoles,
  type PageMeta,
  type PermissionDefinition,
  type Role,
  setAPIRoles,
  updateAPI,
} from '@/services/admin';

const methodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map(
  (method) => ({ label: method, value: method }),
);

const methodColor: Record<string, string> = {
  GET: 'blue',
  POST: 'green',
  PUT: 'gold',
  PATCH: 'purple',
  DELETE: 'red',
};

const APIs: React.FC = () => {
  const access = useAccess();
  const [apis, setAPIs] = useState<APIResource[]>([]);
  const [groups, setGroups] = useState<string[]>([]);
  const [permissions, setPermissions] = useState<PermissionDefinition[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [roleModalOpen, setRoleModalOpen] = useState(false);
  const [roleLoading, setRoleLoading] = useState(false);
  const [editing, setEditing] = useState<APIResource>();
  const [roleTarget, setRoleTarget] = useState<APIResource>();
  const [selectedAPIIDs, setSelectedAPIIDs] = useState<React.Key[]>([]);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailItems, setDetailItems] = useState<DescriptionsProps['items']>(
    [],
  );
  const [form] = Form.useForm<APIInput>();
  const [roleForm] = Form.useForm<{ role_ids: number[] }>();

  const loadData = async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const [apiResponse, groupResponse, permissionResponse, roleResponse] =
        await Promise.all([
          listAPIs(params),
          listAPIGroups(),
          listPermissions(),
          access.canRoleRead
            ? listRoles({ page_size: 100 })
            : Promise.resolve({ data: [] }),
        ]);
      setAPIs(apiResponse.data);
      setGroups(groupResponse);
      setPage(apiResponse.page);
      setPermissions(permissionResponse);
      setRoles(roleResponse.data);
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
    form.setFieldsValue({ method: 'GET', group: 'api', public: false });
    setModalOpen(true);
  };

  const openRoleModal = async (record: APIResource) => {
    setRoleTarget(record);
    setRoleModalOpen(true);
    setRoleLoading(true);
    try {
      const roleIDs = await listAPIRoles(record.id);
      roleForm.setFieldsValue({ role_ids: roleIDs });
    } finally {
      setRoleLoading(false);
    }
  };

  const openEdit = (record: APIResource) => {
    setEditing(record);
    form.setFieldsValue({
      method: record.method,
      path: record.path,
      description: record.description,
      group: record.group,
      permission: record.permission,
      public: record.public,
    });
    setModalOpen(true);
  };

  const openDetail = async (record: APIResource) => {
    const detail = await readAPI(record.id);
    setDetailItems([
      { key: 'method', label: '方法', children: detail.method },
      { key: 'path', label: '路径', children: detail.path },
      { key: 'description', label: '描述', children: detail.description },
      { key: 'group', label: '分组', children: detail.group },
      { key: 'permission', label: '权限', children: detail.permission || '-' },
      { key: 'public', label: '公开', children: detail.public ? '是' : '否' },
      {
        key: 'created_at',
        label: '创建时间',
        children: new Date(detail.created_at).toLocaleString(),
      },
      {
        key: 'updated_at',
        label: '更新时间',
        children: new Date(detail.updated_at).toLocaleString(),
      },
    ]);
    setDetailOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    if (editing) {
      await updateAPI(editing.id, values);
      message.success('API已更新');
    } else {
      await createAPI(values);
      message.success('API已创建');
    }
    setModalOpen(false);
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const removeSelectedAPIs = async () => {
    const ids = selectedAPIIDs.map(Number);
    await batchDeleteAPIs(ids);
    message.success('API已批量删除');
    setSelectedAPIIDs([]);
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const removeAPI = async (record: APIResource) => {
    await deleteAPI(record.id);
    message.success('API已删除');
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const submitRoles = async () => {
    if (!roleTarget) {
      return;
    }
    const values = await roleForm.validateFields();
    await setAPIRoles(roleTarget.id, values.role_ids ?? []);
    message.success('API授权角色已更新');
    setRoleModalOpen(false);
  };

  const permissionOptions = permissions.map((permission) => ({
    label: `${permission.name} (${permission.token})`,
    value: permission.token,
  }));
  const groupOptions = groups.map((group) => ({ label: group, value: group }));

  const columns: ColumnsType<APIResource> = [
    {
      title: '方法',
      dataIndex: 'method',
      width: 96,
      render: (method: string) => (
        <Tag color={methodColor[method] ?? 'default'}>{method}</Tag>
      ),
    },
    { title: '路径', dataIndex: 'path' },
    { title: '描述', dataIndex: 'description' },
    { title: '分组', dataIndex: 'group', width: 120 },
    {
      title: '权限',
      dataIndex: 'permission',
      render: (permission?: string) => permission || '-',
    },
    {
      title: '公开',
      dataIndex: 'public',
      width: 88,
      render: (publicAPI: boolean) => (
        <Tag color={publicAPI ? 'green' : 'default'}>
          {publicAPI ? '是' : '否'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => void openDetail(record)}
          >
            详情
          </Button>
          {access.canApiUpdate ? (
            <Button type="link" onClick={() => openEdit(record)}>
              编辑
            </Button>
          ) : null}
          {access.canApiUpdate && access.canRoleRead ? (
            <Button type="link" onClick={() => void openRoleModal(record)}>
              授权角色
            </Button>
          ) : null}
          {access.canApiDelete ? (
            <Popconfirm
              title="删除API"
              description={`确认删除 ${record.method} ${record.path}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeAPI(record)}
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
    <PageContainer title="API管理">
      <Table<APIResource>
        rowKey="id"
        columns={columns}
        dataSource={apis}
        loading={loading}
        rowSelection={
          access.canApiDelete
            ? {
                selectedRowKeys: selectedAPIIDs,
                onChange: setSelectedAPIIDs,
              }
            : undefined
        }
        pagination={{
          current: page?.page,
          pageSize: page?.page_size,
          total: page?.total,
          showSizeChanger: true,
        }}
        onChange={(pagination) =>
          void loadData({
            page: pagination.current,
            page_size: pagination.pageSize,
          })
        }
        title={() => (
          <Space>
            {access.canApiCreate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={openCreate}
              >
                新增API
              </Button>
            ) : null}
            {access.canApiDelete ? (
              <Popconfirm
                title="批量删除API"
                description={`确认删除已选 ${selectedAPIIDs.length} 个 API？`}
                okText="删除"
                okButtonProps={{ danger: true }}
                disabled={selectedAPIIDs.length === 0}
                onConfirm={() => void removeSelectedAPIs()}
              >
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  disabled={selectedAPIIDs.length === 0}
                >
                  批量删除
                </Button>
              </Popconfirm>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title={editing ? '编辑API' : '新增API'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form<APIInput> form={form} layout="vertical">
          <Form.Item
            label="方法"
            name="method"
            rules={[{ required: true, message: '请选择请求方法' }]}
          >
            <Select options={methodOptions} />
          </Form.Item>
          <Form.Item
            label="路径"
            name="path"
            rules={[{ required: true, message: '请输入API路径' }]}
          >
            <Input maxLength={180} />
          </Form.Item>
          <Form.Item
            label="描述"
            name="description"
            rules={[{ required: true, message: '请输入API描述' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
          <Form.Item
            label="分组"
            name="group"
            rules={[{ required: true, message: '请输入API分组' }]}
          >
            <AutoComplete options={groupOptions}>
              <Input maxLength={80} />
            </AutoComplete>
          </Form.Item>
          <Form.Item label="权限" name="permission">
            <Select allowClear options={permissionOptions} />
          </Form.Item>
          <Form.Item label="公开" name="public" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title="API详情"
        open={detailOpen}
        footer={null}
        onCancel={() => setDetailOpen(false)}
        destroyOnHidden
      >
        <Descriptions column={1} size="small" items={detailItems} />
      </Modal>
      <Modal
        title={
          roleTarget
            ? `授权角色 - ${roleTarget.method} ${roleTarget.path}`
            : '授权角色'
        }
        open={roleModalOpen}
        onOk={() => void submitRoles()}
        onCancel={() => setRoleModalOpen(false)}
        confirmLoading={roleLoading}
        destroyOnHidden
      >
        <Form<{ role_ids: number[] }> form={roleForm} layout="vertical">
          <Form.Item label="角色" name="role_ids">
            <Select
              mode="multiple"
              options={roles.map((role) => ({
                value: role.id,
                label: role.name,
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default APIs;
