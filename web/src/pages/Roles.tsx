import { PlusOutlined } from '@ant-design/icons';
import {
  DrawerForm,
  PageContainer,
  ProCard,
  ProDescriptions,
  ProForm,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProList,
  StatisticCard,
} from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Avatar,
  Button,
  Checkbox,
  Empty,
  Form,
  Input,
  message,
  Popconfirm,
  Space,
  Spin,
  Tag,
  Tree,
  Typography,
} from 'antd';
import React, { useEffect, useMemo, useState } from 'react';

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
  type Role,
  setRoleAdmins,
  updateRole,
} from '@/services/admin';
import {
  deriveAPIIDs,
  deriveButtonIDs,
  deriveMenuIDs,
} from '@/utils/grant-derivation';
import { toAntdTreeData } from '@/utils/menu-tree';

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

// 角色头像的循环取色，让左侧列表一眼能区分不同角色。
const avatarColors = [
  '#1677ff',
  '#13c2c2',
  '#722ed1',
  '#fa8c16',
  '#52c41a',
  '#eb2f96',
  '#f5222d',
  '#2f54eb',
];

const avatarColor = (roleID: number) =>
  avatarColors[roleID % avatarColors.length];

const methodColor: Record<string, string> = {
  GET: 'blue',
  POST: 'green',
  PUT: 'gold',
  PATCH: 'purple',
  DELETE: 'red',
};

// 详情页签里分组 Tag 列表。
const groupTags = (
  groups: Map<string, { label: string; sub?: string }[]>,
): React.ReactNode[] =>
  [...groups.entries()].map(([group, items]) => (
    <Space
      key={group}
      direction="vertical"
      size={4}
      style={{ display: 'flex' }}
    >
      <Typography.Text type="secondary">{group}</Typography.Text>
      <Space wrap size={[8, 8]}>
        {items.map((item) => (
          <Tag key={item.label} color={item.sub}>
            {item.label}
          </Tag>
        ))}
      </Space>
    </Space>
  ));

// 功能权限字段：按资源分组的卡片式勾选，是角色授权的单一控制面。
// 菜单可见性、按钮显示、API 放行均由此派生，卡片底部实时预览派生结果。
const PermissionGroupsField = ({
  value = [],
  onChange,
  permissions,
  apis,
  menus,
}: {
  value?: string[];
  onChange?: (value: string[]) => void;
  permissions: PermissionDefinition[];
  apis: APIResource[];
  menus: Menu[];
}) => {
  const groups = new Map<string, PermissionDefinition[]>();
  for (const permission of permissions) {
    const list = groups.get(permission.resource) ?? [];
    list.push(permission);
    groups.set(permission.resource, list);
  }
  const menuCount = deriveMenuIDs(value, menus).length;
  const apiCount = deriveAPIIDs(value, apis).length;
  return (
    <Space direction="vertical" size={8} style={{ display: 'flex' }}>
      {[...groups.entries()].map(([resource, items]) => {
        const tokens = items.map((item) => item.token);
        const checkedCount = tokens.filter((token) =>
          value.includes(token),
        ).length;
        return (
          <div
            key={resource}
            style={{
              border: '1px solid #f0f0f0',
              borderRadius: 8,
              padding: '8px 12px 10px',
              background: '#fafafa',
            }}
          >
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <Typography.Text strong style={{ fontSize: 13 }}>
                {resource}
                <Typography.Text
                  type="secondary"
                  style={{ fontSize: 12, fontWeight: 400, marginLeft: 8 }}
                >
                  {items.length} 项
                </Typography.Text>
              </Typography.Text>
              <Checkbox
                checked={checkedCount === tokens.length}
                indeterminate={checkedCount > 0 && checkedCount < tokens.length}
                onChange={(event) => {
                  const next = event.target.checked
                    ? [...new Set([...value, ...tokens])]
                    : value.filter((token) => !tokens.includes(token));
                  onChange?.(next);
                }}
              >
                全选
              </Checkbox>
            </div>
            <Checkbox.Group
              style={{ display: 'flex', flexWrap: 'wrap', columnGap: 16 }}
              value={value.filter((token) => tokens.includes(token))}
              options={items.map((item) => ({
                label: item.name,
                value: item.token,
              }))}
              onChange={(next) => {
                const others = value.filter((token) => !tokens.includes(token));
                onChange?.([...others, ...next]);
              }}
            />
          </div>
        );
      })}
      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
        当前选择将自动授权 {menuCount} 个菜单、{apiCount} 条受管 API。
      </Typography.Text>
    </Space>
  );
};

const Roles: React.FC = () => {
  const access = useAccess();
  const [roles, setRoles] = useState<Role[]>([]);
  const [rolesLoading, setRolesLoading] = useState(false);
  const [selectedRoleID, setSelectedRoleID] = useState<number>();
  const [keyword, setKeyword] = useState('');
  const [admins, setAdmins] = useState<AdminUser[]>([]);
  const [menus, setMenus] = useState<Menu[]>([]);
  const [apis, setAPIs] = useState<APIResource[]>([]);
  const [permissions, setPermissions] = useState<PermissionDefinition[]>([]);
  const [memberIDs, setMemberIDs] = useState<number[]>([]);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editing, setEditing] = useState<Role>();
  const [copying, setCopying] = useState<Role>();
  const [memberTarget, setMemberTarget] = useState<MemberTarget>();

  const loadRoles = async (preferID?: number) => {
    setRolesLoading(true);
    try {
      const response = await listRoles({ page_size: 100 });
      setRoles(response.data);
      setSelectedRoleID((previous) => {
        const wanted = preferID ?? previous;
        if (
          wanted !== undefined &&
          response.data.some((role) => role.id === wanted)
        ) {
          return wanted;
        }
        return response.data[0]?.id;
      });
    } finally {
      setRolesLoading(false);
    }
  };

  useEffect(() => {
    void loadRoles();
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

  // 成员页签的数据随选中角色切换重新拉取。
  const canManageMembers = access.canRoleUpdate && access.canAdminRead;
  useEffect(() => {
    if (!selectedRoleID || !canManageMembers) {
      setMemberIDs([]);
      return;
    }
    void listRoleAdmins(selectedRoleID).then(setMemberIDs);
  }, [selectedRoleID, canManageMembers, roles]);

  const selectedRole = roles.find((role) => role.id === selectedRoleID);
  const filteredRoles = useMemo(() => {
    const query = keyword.trim().toLowerCase();
    if (!query) {
      return roles;
    }
    return roles.filter(
      (role) =>
        role.name.toLowerCase().includes(query) ||
        role.code.toLowerCase().includes(query),
    );
  }, [roles, keyword]);

  const roleName = (roleID: number) =>
    roles.find((role) => role.id === roleID)?.name ?? `#${roleID}`;
  const apiByID = useMemo(
    () => new Map(apis.map((api) => [api.id, api])),
    [apis],
  );
  const permissionByToken = useMemo(
    () =>
      new Map(permissions.map((permission) => [permission.token, permission])),
    [permissions],
  );

  // 功能权限页签分组：只展示该角色已持有的 token。
  const grantedPermissionGroups = useMemo(() => {
    const groups = new Map<string, { label: string }[]>();
    if (!selectedRole) {
      return groups;
    }
    for (const token of selectedRole.permissions) {
      const permission = permissionByToken.get(token);
      const key = permission?.resource ?? 'other';
      const options = groups.get(key) ?? [];
      options.push({
        label: permission ? `${permission.name} (${token})` : token,
      });
      groups.set(key, options);
    }
    return groups;
  }, [selectedRole, permissionByToken]);

  const grantedAPIGroups = useMemo(() => {
    const groups = new Map<string, { label: string; sub: string }[]>();
    if (!selectedRole) {
      return groups;
    }
    for (const apiID of selectedRole.api_ids) {
      const api = apiByID.get(apiID);
      const key = api?.group ?? 'other';
      const options = groups.get(key) ?? [];
      options.push({
        label: `${api?.description ?? api?.path ?? `#${apiID}`} (${api?.method ?? ''} ${api?.path ?? ''})`,
        sub: methodColor[api?.method ?? ''] ?? 'default',
      });
      groups.set(key, options);
    }
    return groups;
  }, [selectedRole, apiByID]);

  const menuTreeData = useMemo(() => toAntdTreeData(menus), [menus]);
  const parentRoleOptions = [
    { label: '顶级角色', value: 0 },
    ...roles
      .filter((role) => role.id !== editing?.id)
      .map((role) => ({ label: role.name, value: role.id })),
  ];
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

  const headerActions = selectedRole ? (
    <Space>
      {access.canRoleUpdate ? (
        <Button
          key="edit"
          type="primary"
          onClick={() => {
            setCopying(undefined);
            setEditing(selectedRole);
            setDrawerOpen(true);
          }}
        >
          编辑
        </Button>
      ) : null}
      {access.canRoleCreate ? (
        <Button
          key="copy"
          onClick={() => {
            setEditing(undefined);
            setCopying(selectedRole);
            setDrawerOpen(true);
          }}
        >
          复制
        </Button>
      ) : null}
      {access.canRoleDelete ? (
        <Popconfirm
          key="delete"
          title="删除角色"
          description={`确认删除 ${selectedRole.name}？`}
          okText="删除"
          okButtonProps={{ danger: true }}
          onConfirm={async () => {
            await deleteRole(selectedRole.id);
            message.success('角色已删除');
            await loadRoles();
          }}
        >
          <Button danger>删除</Button>
        </Popconfirm>
      ) : null}
    </Space>
  ) : undefined;

  return (
    <PageContainer title="角色权限">
      <ProCard gutter={[16, 16]} wrap>
        <ProCard
          colSpan={{ xs: 24, lg: 7, xxl: 6 }}
          title={`角色（${roles.length}）`}
          extra={
            access.canRoleCreate ? (
              <Button
                key="create"
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  setCopying(undefined);
                  setEditing(undefined);
                  setDrawerOpen(true);
                }}
              >
                新增
              </Button>
            ) : undefined
          }
        >
          <Input.Search
            allowClear
            placeholder="搜索角色名称或编码"
            style={{ marginBottom: 12 }}
            onChange={(event) => setKeyword(event.target.value)}
          />
          <Spin spinning={rolesLoading}>
            <ProList<Role>
              rowKey="id"
              dataSource={filteredRoles}
              split={false}
              size="small"
              rowSelection={{
                type: 'radio',
                selectedRowKeys:
                  selectedRoleID !== undefined ? [selectedRoleID] : [],
                onChange: (keys) => setSelectedRoleID(Number(keys[0])),
              }}
              onItem={(record) => ({
                onClick: () => setSelectedRoleID(record.id),
              })}
              metas={{
                avatar: {
                  render: (_, record) => (
                    <Avatar style={{ backgroundColor: avatarColor(record.id) }}>
                      {record.name.slice(0, 1)}
                    </Avatar>
                  ),
                },
                title: {
                  render: (_, record) => (
                    <Typography.Text strong>{record.name}</Typography.Text>
                  ),
                },
                subTitle: {
                  render: (_, record) => (
                    <Tag color={record.active ? 'green' : 'default'}>
                      {record.active ? '启用' : '停用'}
                    </Tag>
                  ),
                },
                description: {
                  render: (_, record) => (
                    <Typography.Text
                      type="secondary"
                      style={{ fontSize: 12 }}
                      ellipsis
                    >
                      {`${record.code} · ${record.permissions.length} 功能 · ${record.menu_ids.length} 菜单 · ${record.api_ids.length} API`}
                    </Typography.Text>
                  ),
                },
              }}
              pagination={{
                pageSize: 8,
                size: 'small',
                simple: true,
                hideOnSinglePage: true,
              }}
            />
          </Spin>
        </ProCard>
        <ProCard
          colSpan={{ xs: 24, lg: 17, xxl: 18 }}
          title={
            selectedRole ? (
              <Space>
                <Avatar
                  size="small"
                  style={{ backgroundColor: avatarColor(selectedRole.id) }}
                >
                  {selectedRole.name.slice(0, 1)}
                </Avatar>
                <span>{selectedRole.name}</span>
                <Typography.Text code>{selectedRole.code}</Typography.Text>
              </Space>
            ) : (
              '角色详情'
            )
          }
          extra={headerActions}
          tabs={{
            items: [
              {
                key: 'overview',
                label: '概览',
                children: selectedRole ? (
                  <Space
                    direction="vertical"
                    size={16}
                    style={{ display: 'flex' }}
                  >
                    <StatisticCard.Group>
                      <StatisticCard
                        statistic={{
                          title: '功能权限',
                          value: selectedRole.permissions.length,
                          description: 'resource:action token',
                        }}
                      />
                      <StatisticCard.Divider />
                      <StatisticCard
                        statistic={{
                          title: '可见菜单',
                          value: selectedRole.menu_ids.length,
                          description: `${selectedRole.button_ids.length} 个菜单按钮`,
                        }}
                      />
                      <StatisticCard.Divider />
                      <StatisticCard
                        statistic={{
                          title: 'API 授权',
                          value: selectedRole.api_ids.length,
                          description: '受管路由',
                        }}
                      />
                      <StatisticCard.Divider />
                      <StatisticCard
                        statistic={{
                          title: '数据角色',
                          value: selectedRole.data_role_ids.length,
                          description:
                            selectedRole.data_role_ids
                              .map((roleID) => roleName(roleID))
                              .join('、') || '无',
                        }}
                      />
                    </StatisticCard.Group>
                    <ProDescriptions<Role>
                      column={2}
                      dataSource={selectedRole}
                      columns={[
                        {
                          title: '状态',
                          dataIndex: 'active',
                          render: (_, role) => (
                            <Tag color={role.active ? 'green' : 'default'}>
                              {role.active ? '启用' : '停用'}
                            </Tag>
                          ),
                        },
                        {
                          title: '上级角色',
                          dataIndex: 'parent_id',
                          render: (_, role) =>
                            role.parent_id === 0
                              ? '顶级角色'
                              : roleName(role.parent_id),
                        },
                        { title: '默认入口', dataIndex: 'default_path' },
                        { title: '编码', dataIndex: 'code' },
                      ]}
                    />
                  </Space>
                ) : (
                  <Empty description="选择左侧角色查看授权详情" />
                ),
              },
              {
                key: 'menus',
                label: '菜单权限',
                children: selectedRole ? (
                  <Space
                    direction="vertical"
                    size={8}
                    style={{ display: 'flex' }}
                  >
                    <Tree
                      checkable
                      checkedKeys={selectedRole.menu_ids}
                      treeData={menuTreeData}
                      defaultExpandAll
                    />
                    <Typography.Text type="secondary">
                      菜单授权由功能权限自动派生，点击「编辑」调整功能权限。
                    </Typography.Text>
                  </Space>
                ) : (
                  <Empty description="选择左侧角色查看菜单授权" />
                ),
              },
              {
                key: 'permissions',
                label: '功能权限',
                children: selectedRole ? (
                  grantedPermissionGroups.size > 0 ? (
                    <Space
                      direction="vertical"
                      size={12}
                      style={{ display: 'flex' }}
                    >
                      {groupTags(grantedPermissionGroups)}
                    </Space>
                  ) : (
                    <Empty description="该角色没有功能权限" />
                  )
                ) : (
                  <Empty description="选择左侧角色查看功能权限" />
                ),
              },
              {
                key: 'apis',
                label: 'API 权限',
                children: selectedRole ? (
                  grantedAPIGroups.size > 0 ? (
                    <Space
                      direction="vertical"
                      size={12}
                      style={{ display: 'flex' }}
                    >
                      {groupTags(grantedAPIGroups)}
                    </Space>
                  ) : (
                    <Empty description="该角色没有 API 授权" />
                  )
                ) : (
                  <Empty description="选择左侧角色查看 API 授权" />
                ),
              },
              {
                key: 'members',
                label: '成员',
                children: selectedRole ? (
                  <Space
                    direction="vertical"
                    size={12}
                    style={{ display: 'flex' }}
                  >
                    {canManageMembers ? (
                      <Button
                        onClick={() => {
                          setMemberTarget({
                            role: selectedRole,
                            admin_ids: memberIDs,
                          });
                        }}
                      >
                        编辑成员
                      </Button>
                    ) : null}
                    {memberIDs.length > 0 ? (
                      <Space wrap size={[8, 8]}>
                        {memberIDs.map((adminID) => {
                          const admin = admins.find(
                            (candidate) => candidate.id === adminID,
                          );
                          return (
                            <Tag key={adminID} color="blue">
                              {admin
                                ? `${admin.display_name} (${admin.username})`
                                : `#${adminID}`}
                            </Tag>
                          );
                        })}
                      </Space>
                    ) : (
                      <Empty description="该角色暂无成员" />
                    )}
                  </Space>
                ) : (
                  <Empty description="选择左侧角色查看成员" />
                ),
              },
            ],
          }}
        />
      </ProCard>
      <DrawerForm<RoleFormValues>
        key={
          editing
            ? `edit-${editing.id}`
            : copying
              ? `copy-${copying.id}`
              : 'create'
        }
        title={editing ? '编辑角色' : copying ? '复制角色' : '新增角色'}
        open={drawerOpen}
        onOpenChange={(open) => {
          setDrawerOpen(open);
          if (!open) {
            setCopying(undefined);
          }
        }}
        width="min(720px, 100%)"
        grid
        drawerProps={{ destroyOnHidden: true }}
        initialValues={formInitialValues}
        onFinish={async (values) => {
          // 功能权限是单一控制面：菜单/按钮/API 授权全部由 token 派生，
          // 勾一处即同时决定菜单可见、按钮显示和后端路由放行。
          const menu_ids = deriveMenuIDs(values.permissions, menus);
          const api_ids = deriveAPIIDs(values.permissions, apis);
          const button_ids = deriveButtonIDs(menu_ids, menus);
          if (editing) {
            await updateRole(editing.id, {
              parent_id: values.parent_id,
              name: values.name,
              permissions: values.permissions,
              menu_ids,
              api_ids,
              button_ids,
              data_role_ids: values.data_role_ids,
              default_path: values.default_path,
              active: values.active,
            });
            message.success('角色已更新');
            await loadRoles(editing.id);
          } else if (copying) {
            await copyRole(copying.id, {
              parent_id: values.parent_id,
              code: values.code,
              name: values.name,
              default_path: values.default_path,
              active: values.active,
            });
            message.success('角色已复制');
            await loadRoles();
          } else {
            await createRole({
              parent_id: values.parent_id,
              code: values.code,
              name: values.name,
              permissions: values.permissions,
              menu_ids,
              api_ids,
              button_ids,
              data_role_ids: values.data_role_ids,
              default_path: values.default_path,
              active: values.active,
            });
            message.success('角色已创建');
            await loadRoles();
          }
          return true;
        }}
      >
        <ProForm.Group title="基本信息" grid>
          <ProFormText
            name="name"
            label="名称"
            colProps={{ xs: 24, md: 12 }}
            fieldProps={{ maxLength: 80 }}
            rules={[{ required: true, message: '请输入角色名称' }]}
          />
          <ProFormText
            name="code"
            label="编码"
            colProps={{ xs: 24, md: 12 }}
            disabled={Boolean(editing)}
            fieldProps={{ maxLength: 64 }}
            rules={
              editing ? [] : [{ required: true, message: '请输入角色编码' }]
            }
          />
          <ProFormSelect
            name="parent_id"
            label="上级角色"
            colProps={{ xs: 24, md: 12 }}
            options={parentRoleOptions}
          />
          <ProFormText
            name="default_path"
            label="默认入口"
            colProps={{ xs: 24, md: 12 }}
            fieldProps={{ maxLength: 160 }}
            rules={[{ required: true, message: '请输入默认入口' }]}
          />
          <ProFormSwitch
            name="active"
            label="启用"
            colProps={{ xs: 24, md: 12 }}
          />
        </ProForm.Group>
        {copying ? (
          <Typography.Text type="secondary">
            功能权限、菜单、API、按钮和数据角色授权将从源角色「{copying.name}
            」复制。
          </Typography.Text>
        ) : (
          <>
            <Form.Item
              name="permissions"
              label="功能权限"
              required
              tooltip="勾选后自动派生菜单可见性、按钮显示和 API 放行，无需分别配置"
            >
              <PermissionGroupsField
                permissions={permissions}
                apis={apis}
                menus={menus}
              />
            </Form.Item>
            <ProFormSelect
              name="data_role_ids"
              label="数据权限（可见的管理员数据范围）"
              mode="multiple"
              colProps={{ span: 24 }}
              options={roles.map((role) => ({
                label: role.name,
                value: role.id,
              }))}
              fieldProps={{ maxTagCount: 'responsive' }}
            />
          </>
        )}
      </DrawerForm>
      <DrawerForm<{ admin_ids: number[] }>
        key={memberTarget?.role.id ?? 'idle'}
        title={memberTarget ? `${memberTarget.role.name}成员` : '角色成员'}
        open={Boolean(memberTarget)}
        onOpenChange={(open) => {
          if (!open) {
            setMemberTarget(undefined);
          }
        }}
        drawerProps={{ destroyOnHidden: true }}
        initialValues={{ admin_ids: memberTarget?.admin_ids ?? [] }}
        onFinish={async (values) => {
          if (!memberTarget) {
            return true;
          }
          const assignedIDs = await setRoleAdmins(
            memberTarget.role.id,
            values.admin_ids ?? [],
          );
          setMemberIDs(assignedIDs);
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
          fieldProps={{
            showSearch: true,
            optionFilterProp: 'label',
            maxTagCount: 'responsive',
          }}
        />
      </DrawerForm>
    </PageContainer>
  );
};

export default Roles;
