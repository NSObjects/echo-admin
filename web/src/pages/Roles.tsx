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
  createRole,
  type ListParams,
  listMenus,
  listPermissions,
  listRoles,
  type Menu,
  type PageMeta,
  type PermissionDefinition,
  type Role,
  updateRole,
} from '@/services/admin';

type RoleFormValues = {
  parent_id: number;
  code: string;
  name: string;
  permissions: string[];
  menu_ids?: number[];
  default_path?: string;
  active: boolean;
};

const Roles: React.FC = () => {
  const access = useAccess();
  const [roles, setRoles] = useState<Role[]>([]);
  const [menus, setMenus] = useState<Menu[]>([]);
  const [permissions, setPermissions] = useState<PermissionDefinition[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Role>();
  const [form] = Form.useForm<RoleFormValues>();

  const loadData = async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const [roleResponse, menuResponse, permissionResponse] =
        await Promise.all([
          listRoles(params),
          access.canMenuRead ? listMenus() : Promise.resolve([]),
          listPermissions(),
        ]);
      setRoles(roleResponse.data);
      setPage(roleResponse.page);
      setMenus(menuResponse);
      setPermissions(permissionResponse);
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
      parent_id: 0,
      permissions: [],
      menu_ids: [],
      default_path: '/dashboard',
    });
    setModalOpen(true);
  };

  const openEdit = (record: Role) => {
    setEditing(record);
    form.setFieldsValue({
      parent_id: record.parent_id,
      code: record.code,
      name: record.name,
      permissions: record.permissions,
      menu_ids: record.menu_ids,
      default_path: record.default_path,
      active: record.active,
    });
    setModalOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    if (editing) {
      await updateRole(editing.id, {
        parent_id: values.parent_id,
        name: values.name,
        permissions: values.permissions,
        menu_ids: values.menu_ids,
        default_path: values.default_path,
        active: values.active,
      });
      message.success('角色已更新');
    } else {
      await createRole({
        parent_id: values.parent_id,
        code: values.code,
        name: values.name,
        permissions: values.permissions,
        menu_ids: values.menu_ids,
        default_path: values.default_path,
        active: values.active,
      });
      message.success('角色已创建');
    }
    setModalOpen(false);
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const columns: ColumnsType<Role> = [
    { title: '编码', dataIndex: 'code' },
    { title: '名称', dataIndex: 'name' },
    {
      title: '上级',
      dataIndex: 'parent_id',
      render: (parentID: number) =>
        parentID === 0
          ? '顶级角色'
          : (roles.find((role) => role.id === parentID)?.name ??
            `#${parentID}`),
    },
    {
      title: '权限',
      dataIndex: 'permissions',
      render: (permissions: string[]) => (
        <Space wrap>
          {permissions.map((permission) => (
            <Tag key={permission}>{permission}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '菜单',
      dataIndex: 'menu_ids',
      render: (menuIDs: number[]) => `${menuIDs.length} 个`,
    },
    { title: '入口', dataIndex: 'default_path' },
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
      render: (_, record) =>
        access.canRoleUpdate ? (
          <Button type="link" onClick={() => openEdit(record)}>
            编辑
          </Button>
        ) : null,
    },
  ];

  const permissionOptions = permissions.map((permission) => ({
    label: `${permission.name} (${permission.token})`,
    value: permission.token,
  }));
  const roleOptions = [
    { label: '顶级角色', value: 0 },
    ...roles
      .filter((role) => role.id !== editing?.id)
      .map((role) => ({
        label: role.name,
        value: role.id,
      })),
  ];
  const menuOptions = menus.map((menu) => ({
    label: menu.name,
    value: menu.id,
  }));

  return (
    <PageContainer title="角色权限">
      <Table<Role>
        rowKey="id"
        columns={columns}
        dataSource={roles}
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
            {access.canRoleCreate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={openCreate}
              >
                新增角色
              </Button>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title={editing ? '编辑角色' : '新增角色'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form<RoleFormValues> form={form} layout="vertical">
          <Form.Item
            label="编码"
            name="code"
            rules={[{ required: !editing, message: '请输入角色编码' }]}
          >
            <Input disabled={Boolean(editing)} maxLength={64} />
          </Form.Item>
          <Form.Item label="上级角色" name="parent_id">
            <Select options={roleOptions} />
          </Form.Item>
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入角色名称' }]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item
            label="权限"
            name="permissions"
            rules={[{ required: true, message: '请选择权限' }]}
          >
            <Select mode="multiple" options={permissionOptions} />
          </Form.Item>
          <Form.Item label="菜单" name="menu_ids">
            <Select mode="multiple" options={menuOptions} />
          </Form.Item>
          <Form.Item
            label="默认入口"
            name="default_path"
            rules={[{ required: true, message: '请输入默认入口' }]}
          >
            <Input maxLength={160} />
          </Form.Item>
          <Form.Item label="启用" name="active" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default Roles;
