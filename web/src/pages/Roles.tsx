import { PlusOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProTable,
} from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import { Button, message, Popconfirm, Space, Tag } from 'antd';
import React, { useEffect, useRef, useState } from 'react';

import {
  type AdminUser,
  type APIResource,
  copyRole,
  createRole,
  deleteRole,
  listAdmins,
  listAPIs,
  listMenus,
  listPermissions,
  listRoleAdmins,
  listRoles,
  type Menu,
  type PermissionDefinition,
  pageParams,
  type Role,
  setRoleAdmins,
  toTableResult,
  updateRole,
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

type MemberTarget = {
  role: Role;
  admin_ids: number[];
};

const Roles: React.FC = () => {
  const access = useAccess();
  const actionRef = useRef<ActionType | undefined>(undefined);
  const [admins, setAdmins] = useState<AdminUser[]>([]);
  const [menus, setMenus] = useState<Menu[]>([]);
  const [apis, setAPIs] = useState<APIResource[]>([]);
  const [permissions, setPermissions] = useState<PermissionDefinition[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Role>();
  const [copying, setCopying] = useState<Role>();
  const [memberTarget, setMemberTarget] = useState<MemberTarget>();
  // 列表数据缓存在组件内，供上级角色展示名和表单选项复用。
  const [rolesCache, setRolesCache] = useState<Role[]>([]);

  useEffect(() => {
    void Promise.all([
      access.canAdminRead
        ? listAdmins({ page_size: 100 })
        : Promise.resolve({ data: [] }),
      access.canMenuRead ? listMenus() : Promise.resolve([]),
      access.canApiRead
        ? listAPIs({ page_size: 100 })
        : Promise.resolve({ data: [] }),
      listPermissions(),
    ]).then(
      ([adminResponse, menuResponse, apiResponse, permissionResponse]) => {
        setAdmins(adminResponse.data);
        setMenus(menuResponse);
        setAPIs(apiResponse.data);
        setPermissions(permissionResponse);
      },
    );
  }, [access]);

  const openCreate = () => {
    setEditing(undefined);
    setCopying(undefined);
    setModalOpen(true);
  };

  const openEdit = (record: Role) => {
    setCopying(undefined);
    setEditing(record);
    setModalOpen(true);
  };

  const openCopy = (record: Role) => {
    setEditing(undefined);
    setCopying(record);
    setModalOpen(true);
  };

  // 上级角色的展示名需要引用当前列表数据。
  const parentRoleName = (roleID: number) =>
    rolesCache.find((role) => role.id === roleID)?.name;

  const columns: ProColumns<Role>[] = [
    { title: '编码', dataIndex: 'code' },
    { title: '名称', dataIndex: 'name' },
    {
      title: '上级',
      dataIndex: 'parent_id',
      render: (_, record) =>
        record.parent_id === 0
          ? '顶级角色'
          : (parentRoleName(record.parent_id) ?? `#${record.parent_id}`),
    },
    {
      title: '权限',
      dataIndex: 'permissions',
      render: (_, record) => (
        <Space wrap>
          {record.permissions.map((permission) => (
            <Tag key={permission}>{permission}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '菜单',
      dataIndex: 'menu_ids',
      render: (_, record) => `${record.menu_ids.length} 个`,
    },
    {
      title: 'API',
      dataIndex: 'api_ids',
      render: (_, record) => `${record.api_ids.length} 个`,
    },
    {
      title: '按钮',
      dataIndex: 'button_ids',
      render: (_, record) => `${record.button_ids.length} 个`,
    },
    {
      title: '数据权限',
      dataIndex: 'data_role_ids',
      render: (_, record) => `${record.data_role_ids.length} 个`,
    },
    { title: '入口', dataIndex: 'default_path' },
    {
      title: '状态',
      dataIndex: 'active',
      render: (_, record) => (
        <Tag color={record.active ? 'green' : 'default'}>
          {record.active ? '启用' : '停用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      valueType: 'option',
      width: 200,
      render: (_, record) => {
        const actions: React.ReactNode[] = [];
        if (access.canRoleUpdate) {
          actions.push(
            <a key="edit" onClick={() => openEdit(record)}>
              编辑
            </a>,
          );
        }
        if (access.canRoleCreate) {
          actions.push(
            <a key="copy" onClick={() => openCopy(record)}>
              复制
            </a>,
          );
        }
        if (access.canRoleUpdate && access.canAdminRead) {
          actions.push(
            <a
              key="members"
              onClick={() => {
                void listRoleAdmins(record.id).then((adminIDs) => {
                  setMemberTarget({ role: record, admin_ids: adminIDs });
                });
              }}
            >
              成员
            </a>,
          );
        }
        if (access.canRoleDelete) {
          actions.push(
            <Popconfirm
              key="delete"
              title="删除角色"
              description={`确认删除 ${record.name}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={async () => {
                await deleteRole(record.id);
                message.success('角色已删除');
                actionRef.current?.reload();
              }}
            >
              <Button type="link" danger size="small">
                删除
              </Button>
            </Popconfirm>,
          );
        }
        return actions;
      },
    },
  ];

  const rolesRequest = async (params: {
    current?: number;
    pageSize?: number;
  }) => {
    const response = await listRoles(pageParams(params));
    setRolesCache(response.data);
    return toTableResult(response);
  };

  const permissionOptions = permissions.map((permission) => ({
    label: `${permission.name} (${permission.token})`,
    value: permission.token,
  }));
  const parentRoleOptions = [
    { label: '顶级角色', value: 0 },
    ...rolesCache
      .filter((role) => role.id !== editing?.id)
      .map((role) => ({ label: role.name, value: role.id })),
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
  const dataRoleOptions = rolesCache.map((role) => ({
    label: role.name,
    value: role.id,
  }));
  const adminOptions = admins.map((admin) => ({
    label: `${admin.display_name} (${admin.username})`,
    value: admin.id,
  }));

  const formInitialValues = editing
    ? {
        parent_id: editing.parent_id,
        code: editing.code,
        name: editing.name,
        permissions: editing.permissions,
        menu_ids: editing.menu_ids,
        api_ids: editing.api_ids,
        button_ids: editing.button_ids,
        data_role_ids: editing.data_role_ids,
        default_path: editing.default_path,
        active: editing.active,
      }
    : copying
      ? {
          parent_id: copying.parent_id,
          code: `${copying.code}_copy`,
          name: `${copying.name}副本`,
          permissions: copying.permissions,
          menu_ids: copying.menu_ids,
          api_ids: copying.api_ids,
          button_ids: copying.button_ids,
          data_role_ids: copying.data_role_ids,
          default_path: copying.default_path,
          active: copying.active,
        }
      : {
          active: true,
          parent_id: 0,
          permissions: [],
          menu_ids: [],
          api_ids: [],
          button_ids: [],
          data_role_ids: [],
          default_path: '/dashboard',
        };

  return (
    <PageContainer title="角色权限">
      <ProTable<Role>
        headerTitle="角色列表"
        rowKey="id"
        actionRef={actionRef}
        search={false}
        columns={columns}
        request={rolesRequest}
        toolBarRender={() =>
          access.canRoleCreate
            ? [
                <Button
                  key="create"
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={openCreate}
                >
                  新增角色
                </Button>,
              ]
            : []
        }
      />
      <ModalForm<RoleFormValues>
        key={
          editing
            ? `edit-${editing.id}`
            : copying
              ? `copy-${copying.id}`
              : 'create'
        }
        title={editing ? '编辑角色' : copying ? '复制角色' : '新增角色'}
        open={modalOpen}
        onOpenChange={(open) => {
          setModalOpen(open);
          if (!open) {
            setCopying(undefined);
          }
        }}
        width={520}
        modalProps={{ destroyOnHidden: true }}
        initialValues={formInitialValues}
        onFinish={async (values) => {
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
          actionRef.current?.reload();
          return true;
        }}
      >
        <ProFormText
          name="code"
          label="编码"
          disabled={Boolean(editing)}
          fieldProps={{ maxLength: 64 }}
          rules={editing ? [] : [{ required: true, message: '请输入角色编码' }]}
        />
        <ProFormSelect
          name="parent_id"
          label="上级角色"
          options={parentRoleOptions}
        />
        <ProFormText
          name="name"
          label="名称"
          fieldProps={{ maxLength: 80 }}
          rules={[{ required: true, message: '请输入角色名称' }]}
        />
        <ProFormSelect
          name="permissions"
          label="权限"
          mode="multiple"
          disabled={Boolean(copying)}
          options={permissionOptions}
          fieldProps={{ maxTagCount: 'responsive' }}
          rules={copying ? [] : [{ required: true, message: '请选择权限' }]}
        />
        <ProFormSelect
          name="menu_ids"
          label="菜单"
          mode="multiple"
          disabled={Boolean(copying)}
          options={menuOptions}
          fieldProps={{ maxTagCount: 'responsive' }}
        />
        <ProFormSelect
          name="api_ids"
          label="API"
          mode="multiple"
          disabled={Boolean(copying)}
          options={apiOptions}
          fieldProps={{ maxTagCount: 'responsive' }}
        />
        <ProFormSelect
          name="button_ids"
          label="按钮"
          mode="multiple"
          disabled={Boolean(copying)}
          options={buttonOptions}
          fieldProps={{ maxTagCount: 'responsive' }}
        />
        <ProFormSelect
          name="data_role_ids"
          label="数据权限"
          mode="multiple"
          disabled={Boolean(copying)}
          options={dataRoleOptions}
          fieldProps={{ maxTagCount: 'responsive' }}
        />
        <ProFormText
          name="default_path"
          label="默认入口"
          fieldProps={{ maxLength: 160 }}
          rules={[{ required: true, message: '请输入默认入口' }]}
        />
        <ProFormSwitch name="active" label="启用" />
      </ModalForm>
      <ModalForm<{ admin_ids: number[] }>
        key={memberTarget?.role.id ?? 'idle'}
        title={memberTarget ? `${memberTarget.role.name}成员` : '角色成员'}
        open={Boolean(memberTarget)}
        onOpenChange={(open) => {
          if (!open) {
            setMemberTarget(undefined);
          }
        }}
        modalProps={{ destroyOnHidden: true }}
        initialValues={{ admin_ids: memberTarget?.admin_ids ?? [] }}
        onFinish={async (values) => {
          if (!memberTarget) {
            return true;
          }
          const assignedIDs = await setRoleAdmins(
            memberTarget.role.id,
            values.admin_ids ?? [],
          );
          setMemberTarget((previous) =>
            previous ? { ...previous, admin_ids: assignedIDs } : previous,
          );
          message.success('角色成员已更新');
          return true;
        }}
      >
        <ProFormSelect
          name="admin_ids"
          label="成员"
          mode="multiple"
          options={adminOptions}
          fieldProps={{ maxTagCount: 'responsive' }}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default Roles;
