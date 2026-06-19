import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Form,
  Input,
  Modal,
  message,
  Select,
  Space,
  Switch,
  Table,
  Tag,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  type AdminUser,
  createAdmin,
  type ListParams,
  listAdmins,
  listRoles,
  type PageMeta,
  type Role,
  updateAdmin,
} from '@/services/admin';

type AdminFormValues = {
  username: string;
  display_name: string;
  email?: string;
  password?: string;
  role_ids: number[];
  active_role_id?: number;
  active: boolean;
};

const formatDate = (value: string) => new Date(value).toLocaleString();

const Admins: React.FC = () => {
  const access = useAccess();
  const [admins, setAdmins] = useState<AdminUser[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<AdminUser>();
  const [form] = Form.useForm<AdminFormValues>();
  const selectedRoleIDs = Form.useWatch('role_ids', form) ?? [];

  const loadData = async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const [adminResponse, roleResponse] = await Promise.all([
        listAdmins(params),
        listRoles({ page_size: 100 }),
      ]);
      setAdmins(adminResponse.data);
      setPage(adminResponse.page);
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
    form.setFieldsValue({
      active: true,
      role_ids: [],
      active_role_id: undefined,
    });
    setModalOpen(true);
  };

  const openEdit = (record: AdminUser) => {
    setEditing(record);
    form.setFieldsValue({
      username: record.username,
      display_name: record.display_name,
      email: record.email,
      role_ids: record.role_ids,
      active_role_id: record.active_role_id,
      active: record.active,
    });
    setModalOpen(true);
  };

  useEffect(() => {
    const activeRoleID = form.getFieldValue('active_role_id');
    if (selectedRoleIDs.length === 0) {
      form.setFieldValue('active_role_id', undefined);
      return;
    }
    if (!selectedRoleIDs.includes(activeRoleID)) {
      form.setFieldValue('active_role_id', selectedRoleIDs[0]);
    }
  }, [form, selectedRoleIDs]);

  const submit = async () => {
    const values = await form.validateFields();
    if (editing) {
      const password = values.password?.trim();
      await updateAdmin(editing.id, {
        display_name: values.display_name,
        email: values.email,
        role_ids: values.role_ids,
        active_role_id: values.active_role_id,
        active: values.active,
        ...(password ? { password } : {}),
      });
      message.success('管理员已更新');
    } else {
      await createAdmin({
        username: values.username,
        display_name: values.display_name,
        email: values.email,
        password: values.password ?? '',
        role_ids: values.role_ids,
        active_role_id: values.active_role_id,
        active: values.active,
      });
      message.success('管理员已创建');
    }
    setModalOpen(false);
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const roleName = (roleID: number) =>
    roles.find((role) => role.id === roleID)?.name ?? `#${roleID}`;

  const columns: ColumnsType<AdminUser> = [
    { title: '用户名', dataIndex: 'username' },
    { title: '显示名', dataIndex: 'display_name' },
    {
      title: '邮箱',
      dataIndex: 'email',
      render: (value?: string) => value || '-',
    },
    {
      title: '角色',
      dataIndex: 'role_ids',
      render: (roleIDs: number[]) => (
        <Space wrap>
          {roleIDs.map((roleID) => (
            <Tag key={roleID}>{roleName(roleID)}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '当前角色',
      dataIndex: 'active_role_id',
      render: (roleID: number) => <Tag color="blue">{roleName(roleID)}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'active',
      render: (active: boolean) => (
        <Tag color={active ? 'green' : 'default'}>
          {active ? '启用' : '停用'}
        </Tag>
      ),
    },
    { title: '更新时间', dataIndex: 'updated_at', render: formatDate },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) =>
        access.canAdminUpdate ? (
          <Button type="link" onClick={() => openEdit(record)}>
            编辑
          </Button>
        ) : null,
    },
  ];

  const roleOptions = roles.map((role) => ({
    label: role.name,
    value: role.id,
  }));
  const activeRoleOptions = roleOptions.filter((option) =>
    selectedRoleIDs.includes(option.value),
  );

  return (
    <PageContainer title="管理员管理">
      <Table<AdminUser>
        rowKey="id"
        columns={columns}
        dataSource={admins}
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
            {access.canAdminCreate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={openCreate}
              >
                新增管理员
              </Button>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title={editing ? '编辑管理员' : '新增管理员'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form<AdminFormValues> form={form} layout="vertical">
          <Form.Item
            label="用户名"
            name="username"
            rules={[{ required: !editing, message: '请输入用户名' }]}
          >
            <Input disabled={Boolean(editing)} maxLength={64} />
          </Form.Item>
          <Form.Item
            label="显示名"
            name="display_name"
            rules={[{ required: true, message: '请输入显示名' }]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item label="邮箱" name="email" rules={[{ type: 'email' }]}>
            <Input maxLength={160} />
          </Form.Item>
          <Form.Item
            label={editing ? '新密码' : '密码'}
            name="password"
            rules={[{ required: !editing, min: 8, message: '密码至少 8 位' }]}
          >
            <Input.Password maxLength={72} />
          </Form.Item>
          <Form.Item
            label="角色"
            name="role_ids"
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select mode="multiple" options={roleOptions} />
          </Form.Item>
          <Form.Item
            label="当前角色"
            name="active_role_id"
            rules={[{ required: true, message: '请选择当前角色' }]}
          >
            <Select
              disabled={selectedRoleIDs.length === 0}
              options={activeRoleOptions}
            />
          </Form.Item>
          <Form.Item label="启用" name="active" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default Admins;
