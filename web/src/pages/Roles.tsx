import {
  CopyOutlined,
  DeleteOutlined,
  PlusOutlined,
  ReloadOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
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
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  copyRole,
  createRole,
  deleteRole,
  listRoleAdmins,
  listAdmins,
  listAPIs,
  type AdminUser,
  type ListParams,
  listMenus,
  listPermissions,
  listRoles,
  type Menu,
  type PageMeta,
  type PermissionDefinition,
  type Role,
  setRoleAdmins,
  updateRole,
  type APIResource,
} from '@/services/admin';

type RoleFormValues = {
  parent_id: number;
  code: string;
  name: string;
  permissions: string[];
  menu_ids?: number[];
  api_ids?: number[];
  button_ids?: number[];
  data_role_ids?: number[];
  default_path?: string;
  active: boolean;
};

const Roles: React.FC = () => {
  const access = useAccess();
  const [roles, setRoles] = useState<Role[]>([]);
  const [admins, setAdmins] = useState<AdminUser[]>([]);
  const [menus, setMenus] = useState<Menu[]>([]);
  const [apis, setAPIs] = useState<APIResource[]>([]);
  const [permissions, setPermissions] = useState<PermissionDefinition[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [membersOpen, setMembersOpen] = useState(false);
  const [editing, setEditing] = useState<Role>();
  const [copying, setCopying] = useState<Role>();
  const [memberRole, setMemberRole] = useState<Role>();
  const [memberAdminIDs, setMemberAdminIDs] = useState<number[]>([]);
  const [form] = Form.useForm<RoleFormValues>();

  const loadData = async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const [
        roleResponse,
        adminResponse,
        menuResponse,
        apiResponse,
        permissionResponse,
      ] = await Promise.all([
        listRoles(params),
        access.canAdminRead
          ? listAdmins({ page_size: 100 })
          : Promise.resolve({ data: [] }),
        access.canMenuRead ? listMenus() : Promise.resolve([]),
        access.canApiRead
          ? listAPIs({ page_size: 100 })
          : Promise.resolve({ data: [] }),
        listPermissions(),
      ]);
      setRoles(roleResponse.data);
      setAdmins(adminResponse.data);
      setPage(roleResponse.page);
      setMenus(menuResponse);
      setAPIs(apiResponse.data);
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
    setCopying(undefined);
    form.resetFields();
    form.setFieldsValue({
      active: true,
      parent_id: 0,
      permissions: [],
      menu_ids: [],
      api_ids: [],
      button_ids: [],
      data_role_ids: [],
      default_path: '/dashboard',
    });
    setModalOpen(true);
  };

  const openEdit = (record: Role) => {
    setEditing(record);
    setCopying(undefined);
    form.setFieldsValue({
      parent_id: record.parent_id,
      code: record.code,
      name: record.name,
      permissions: record.permissions,
      menu_ids: record.menu_ids,
      api_ids: record.api_ids,
      button_ids: record.button_ids,
      data_role_ids: record.data_role_ids,
      default_path: record.default_path,
      active: record.active,
    });
    setModalOpen(true);
  };

  const openCopy = (record: Role) => {
    setEditing(undefined);
    setCopying(record);
    form.setFieldsValue({
      parent_id: record.parent_id,
      code: `${record.code}_copy`,
      name: `${record.name}副本`,
      permissions: record.permissions,
      menu_ids: record.menu_ids,
      api_ids: record.api_ids,
      button_ids: record.button_ids,
      data_role_ids: record.data_role_ids,
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
        api_ids: values.api_ids,
        button_ids: values.button_ids,
        data_role_ids: values.data_role_ids,
        default_path: values.default_path,
        active: values.active,
      });
      message.success('角色已更新');
    } else if (copying) {
      await copyRole(copying.id, {
        parent_id: values.parent_id,
        code: values.code,
        name: values.name,
        default_path: values.default_path,
        active: values.active,
      });
      message.success('角色已复制');
    } else {
      await createRole({
        parent_id: values.parent_id,
        code: values.code,
        name: values.name,
        permissions: values.permissions,
        menu_ids: values.menu_ids,
        api_ids: values.api_ids,
        button_ids: values.button_ids,
        data_role_ids: values.data_role_ids,
        default_path: values.default_path,
        active: values.active,
      });
      message.success('角色已创建');
    }
    setModalOpen(false);
    setCopying(undefined);
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const openMembers = async (record: Role) => {
    setMemberRole(record);
    setMembersOpen(true);
    setMemberAdminIDs(await listRoleAdmins(record.id));
  };

  const submitMembers = async () => {
    if (!memberRole) {
      return;
    }
    const assignedIDs = await setRoleAdmins(memberRole.id, memberAdminIDs);
    setMemberAdminIDs(assignedIDs);
    message.success('角色成员已更新');
    setMembersOpen(false);
    setMemberRole(undefined);
    await loadData({ page: page?.page, page_size: page?.page_size });
  };

  const removeRole = async (record: Role) => {
    await deleteRole(record.id);
    message.success('角色已删除');
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
    {
      title: 'API',
      dataIndex: 'api_ids',
      render: (apiIDs: number[]) => `${apiIDs.length} 个`,
    },
    {
      title: '按钮',
      dataIndex: 'button_ids',
      render: (buttonIDs: number[]) => `${buttonIDs.length} 个`,
    },
    {
      title: '数据权限',
      dataIndex: 'data_role_ids',
      render: (roleIDs: number[]) => `${roleIDs.length} 个`,
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
      render: (_, record) => (
        <Space>
          {access.canRoleUpdate ? (
            <Button type="link" onClick={() => openEdit(record)}>
              编辑
            </Button>
          ) : null}
          {access.canRoleCreate ? (
            <Button
              type="link"
              icon={<CopyOutlined />}
              onClick={() => openCopy(record)}
            >
              复制
            </Button>
          ) : null}
          {access.canRoleUpdate && access.canAdminRead ? (
            <Button
              type="link"
              icon={<TeamOutlined />}
              onClick={() => void openMembers(record)}
            >
              成员
            </Button>
          ) : null}
          {access.canRoleDelete ? (
            <Popconfirm
              title="删除角色"
              description={`确认删除 ${record.name}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeRole(record)}
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
  const apiOptions = apis.map((api) => ({
    label: `${api.group} / ${api.description} (${api.method} ${api.path})`,
    value: api.id,
  }));
  const buttonOptions = menus.flatMap((menu) =>
    menu.buttons.map((button) => ({
      label: `${menu.name} / ${button.description || button.name} (${button.name})`,
      value: button.id,
    })),
  );
  const dataRoleOptions = roles.map((role) => ({
    label: role.name,
    value: role.id,
  }));
  const adminOptions = admins.map((admin) => ({
    label: `${admin.display_name} (${admin.username})`,
    value: admin.id,
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
        title={editing ? '编辑角色' : copying ? '复制角色' : '新增角色'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => {
          setModalOpen(false);
          setCopying(undefined);
        }}
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
            <Select
              disabled={Boolean(copying)}
              mode="multiple"
              options={permissionOptions}
            />
          </Form.Item>
          <Form.Item label="菜单" name="menu_ids">
            <Select
              disabled={Boolean(copying)}
              mode="multiple"
              options={menuOptions}
            />
          </Form.Item>
          <Form.Item label="API" name="api_ids">
            <Select
              disabled={Boolean(copying)}
              mode="multiple"
              options={apiOptions}
            />
          </Form.Item>
          <Form.Item label="按钮" name="button_ids">
            <Select
              disabled={Boolean(copying)}
              mode="multiple"
              options={buttonOptions}
            />
          </Form.Item>
          <Form.Item label="数据权限" name="data_role_ids">
            <Select
              disabled={Boolean(copying)}
              mode="multiple"
              options={dataRoleOptions}
            />
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
      <Modal
        title={memberRole ? `${memberRole.name}成员` : '角色成员'}
        open={membersOpen}
        onOk={() => void submitMembers()}
        onCancel={() => {
          setMembersOpen(false);
          setMemberRole(undefined);
        }}
        destroyOnHidden
      >
        <Select
          mode="multiple"
          style={{ width: '100%' }}
          value={memberAdminIDs}
          options={adminOptions}
          onChange={setMemberAdminIDs}
        />
      </Modal>
    </PageContainer>
  );
};

export default Roles;
