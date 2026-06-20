import {
  CopyOutlined,
  DeleteOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess, useModel } from '@umijs/max';
import {
  Button,
  DatePicker,
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
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs, { type Dayjs } from 'dayjs';
import React, { useEffect, useMemo, useState } from 'react';

import {
  type AdminUser,
  type APIToken,
  type APITokenInput,
  createAPIToken,
  deleteAPIToken,
  type ListParams,
  listAdmins,
  listAPITokens,
  listRoles,
  type PageMeta,
  type Role,
  updateAPIToken,
} from '@/services/admin';

type TokenFormValues = {
  admin_id?: number;
  role_id?: number;
  name: string;
  description?: string;
  active: boolean;
  days?: number;
  expires_at?: Dayjs | null;
};

const formatDate = (value?: string | null) =>
  value ? new Date(value).toLocaleString() : '-';

const formValuesToInput = (
  values: TokenFormValues,
  editing: boolean,
): APITokenInput => ({
  admin_id: editing ? undefined : values.admin_id,
  role_id: editing ? undefined : values.role_id,
  name: values.name,
  description: values.description,
  active: values.active,
  days: editing ? undefined : values.days,
  expires_at: editing
    ? values.expires_at
      ? values.expires_at.toISOString()
      : null
    : undefined,
});

const APITokens: React.FC = () => {
  const access = useAccess();
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;
  const [tokens, setTokens] = useState<APIToken[]>([]);
  const [admins, setAdmins] = useState<AdminUser[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<APIToken>();
  const [form] = Form.useForm<TokenFormValues>();
  const selectedAdminID = Form.useWatch('admin_id', form);

  const loadData = async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const fallbackAdmins: AdminUser[] = currentUser
        ? [
            {
              id: currentUser.id,
              username: currentUser.username,
              display_name: currentUser.display_name,
              email: currentUser.email,
              role_ids: currentUser.roles.map((role) => role.id),
              active_role_id:
                currentUser.active_role?.id ?? currentUser.roles[0]?.id ?? 0,
              active: true,
              created_at: '',
              updated_at: '',
            },
          ]
        : [];
      const [tokenResponse, adminResponse, roleResponse] = await Promise.all([
        listAPITokens(params),
        access.canAdminRead
          ? listAdmins({ page_size: 100 })
          : Promise.resolve({ data: fallbackAdmins }),
        access.canRoleRead
          ? listRoles({ page_size: 100 })
          : Promise.resolve({ data: currentUser?.roles ?? [] }),
      ]);
      setTokens(tokenResponse.data);
      setPage(tokenResponse.page);
      setAdmins(adminResponse.data);
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
    const adminID = currentUser?.id ?? admins[0]?.id;
    const roleID =
      currentUser?.active_role?.id ?? currentUser?.roles[0]?.id ?? roles[0]?.id;
    form.setFieldsValue({
      active: true,
      admin_id: adminID,
      role_id: roleID,
      days: 30,
    });
    setModalOpen(true);
  };

  const openEdit = (record: APIToken) => {
    setEditing(record);
    form.setFieldsValue({
      admin_id: record.admin_id,
      role_id: record.role_id,
      name: record.name,
      description: record.description,
      active: record.active,
      days: 30,
      expires_at: record.expires_at ? dayjs(record.expires_at) : null,
    });
    setModalOpen(true);
  };

  const selectedAdmin = useMemo(
    () => admins.find((admin) => admin.id === selectedAdminID),
    [admins, selectedAdminID],
  );
  const roleOptions = useMemo(
    () =>
      selectedAdmin && selectedAdmin.role_ids.length > 0
        ? roles.filter((role) => selectedAdmin.role_ids.includes(role.id))
        : roles,
    [roles, selectedAdmin],
  );

  useEffect(() => {
    if (editing || !selectedAdmin || roleOptions.length === 0) {
      return;
    }
    const roleID = form.getFieldValue('role_id');
    if (!selectedAdmin.role_ids.includes(roleID)) {
      form.setFieldValue('role_id', roleOptions[0].id);
    }
  }, [editing, form, roleOptions, selectedAdmin]);

  const showCreatedSecret = (secret: string) => {
    Modal.info({
      title: 'API Token已创建',
      width: 640,
      content: (
        <Space direction="vertical" style={{ width: '100%' }}>
          <Input.TextArea value={secret} readOnly autoSize />
          <Button
            icon={<CopyOutlined />}
            onClick={() => void navigator.clipboard.writeText(secret)}
          >
            复制
          </Button>
        </Space>
      ),
    });
  };

  const submit = async () => {
    const values = await form.validateFields();
    const input = formValuesToInput(values, Boolean(editing));
    if (editing) {
      await updateAPIToken(editing.id, input);
      message.success('API Token已更新');
    } else {
      const created = await createAPIToken(input);
      message.success('API Token已创建');
      showCreatedSecret(created.secret);
    }
    setModalOpen(false);
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const removeToken = async (record: APIToken) => {
    await deleteAPIToken(record.id);
    message.success('API Token已作废');
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const adminName = (adminID: number) => {
    const admin = admins.find((item) => item.id === adminID);
    return admin ? `${admin.display_name}(${admin.username})` : `#${adminID}`;
  };

  const roleName = (roleID: number) =>
    roles.find((role) => role.id === roleID)?.name ?? `#${roleID}`;

  const columns: ColumnsType<APIToken> = [
    { title: '名称', dataIndex: 'name' },
    { title: '前缀', dataIndex: 'prefix', width: 140 },
    {
      title: '管理员',
      dataIndex: 'admin_id',
      width: 180,
      render: (adminID: number) => adminName(adminID),
    },
    {
      title: '角色',
      dataIndex: 'role_id',
      width: 140,
      render: (roleID: number) => roleName(roleID),
    },
    {
      title: '状态',
      dataIndex: 'active',
      width: 96,
      render: (active: boolean) => (
        <Tag color={active ? 'green' : 'default'}>
          {active ? '启用' : '停用'}
        </Tag>
      ),
    },
    { title: '过期时间', dataIndex: 'expires_at', render: formatDate },
    { title: '最近使用', dataIndex: 'last_used_at', render: formatDate },
    {
      title: '操作',
      key: 'actions',
      width: 160,
      render: (_, record) => (
        <Space>
          {access.canApiTokenUpdate ? (
            <Button type="link" onClick={() => openEdit(record)}>
              编辑
            </Button>
          ) : null}
          {access.canApiTokenDelete ? (
            <Popconfirm
              title="作废API Token"
              description={`确认作废 ${record.name}？`}
              okText="作废"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeToken(record)}
            >
              <Button danger type="link" icon={<DeleteOutlined />}>
                作废
              </Button>
            </Popconfirm>
          ) : null}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title="API Token">
      <Table<APIToken>
        rowKey="id"
        columns={columns}
        dataSource={tokens}
        loading={loading}
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
            {access.canApiTokenCreate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={openCreate}
              >
                新增Token
              </Button>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title={editing ? '编辑API Token' : '新增API Token'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form<TokenFormValues> form={form} layout="vertical">
          <Form.Item
            label="管理员"
            name="admin_id"
            rules={[{ required: !editing, message: '请选择管理员' }]}
          >
            <Select
              disabled={Boolean(editing)}
              options={admins.map((admin) => ({
                value: admin.id,
                label: `${admin.display_name}(${admin.username})`,
              }))}
            />
          </Form.Item>
          <Form.Item
            label="角色"
            name="role_id"
            rules={[{ required: !editing, message: '请选择角色' }]}
          >
            <Select
              disabled={Boolean(editing)}
              options={roleOptions.map((role) => ({
                value: role.id,
                label: role.name,
              }))}
            />
          </Form.Item>
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item label="描述" name="description">
            <Input.TextArea maxLength={240} rows={3} />
          </Form.Item>
          <Form.Item label="启用" name="active" valuePropName="checked">
            <Switch />
          </Form.Item>
          {editing ? (
            <Form.Item label="过期时间" name="expires_at">
              <DatePicker showTime style={{ width: '100%' }} />
            </Form.Item>
          ) : (
            <Form.Item
              label="有效天数"
              name="days"
              rules={[{ required: true, message: '请输入有效天数' }]}
            >
              <InputNumber min={1} max={365} style={{ width: '100%' }} />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default APITokens;
