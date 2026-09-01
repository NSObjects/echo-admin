import { PlusOutlined } from '@ant-design/icons';
import {
  DrawerForm,
  PageContainer,
  ProCard,
  ProDescriptions,
  ProForm,
  ProFormDependency,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTreeSelect,
  ProList,
  StatisticCard,
} from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Avatar,
  Button,
  Checkbox,
  Collapse,
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
  toAntdTreeData,
  toMenuTreeNodes,
  withMenuAncestors,
} from '@/utils/menu-tree';

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

// 把权限定义按 resource 归组，避免几十个 token 平铺在一个下拉里。
const permissionOptionGroups = (permissions: PermissionDefinition[]) => {
  const groups = new Map<string, { label: string; value: string }[]>();
  for (const permission of permissions) {
    const options = groups.get(permission.resource) ?? [];
    options.push({
      label: `${permission.name} (${permission.token})`,
      value: permission.token,
    });
    groups.set(permission.resource, options);
  }
  return [...groups.entries()].map(([resource, options]) => ({
    label: resource,
    options,
  }));
};

// 受管 API 按业务分组归组，下拉里能按模块快速定位。
const apiOptionGroups = (apis: APIResource[]) => {
  const groups = new Map<string, { label: string; value: number }[]>();
  for (const api of apis) {
    const options = groups.get(api.group) ?? [];
    options.push({
      label: `${api.description ?? api.path} (${api.method} ${api.path})`,
      value: api.id,
    });
    groups.set(api.group, options);
  }
  return [...groups.entries()].map(([group, options]) => ({
    label: group,
    options,
  }));
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

// 授权分区折叠面板标题：名称 + 已选数量，收起时也能看清授权规模。
const grantPanelLabel = (
  title: string,
  count: string,
  required = false,
): React.ReactNode => (
  <Space size={8}>
    {required ? <span style={{ color: '#ff4d4f' }}>*</span> : null}
    <span>{title}</span>
    <Typography.Text type="secondary" style={{ fontSize: 12, fontWeight: 400 }}>
      {count}
    </Typography.Text>
  </Space>
);

// 按钮权限字段：按菜单分组的卡片式勾选，替代几十个复选框平铺的墙。
// 作为受控组件接入 antd Form（由 Form.Item 注入 value/onChange）。
const ButtonGroupsField = ({
  value = [],
  onChange,
  menus,
  menuIds,
}: {
  value?: number[];
  onChange?: (value: number[]) => void;
  menus: Menu[];
  menuIds: number[];
}) => {
  const selectedMenus = menus.filter(
    (menu) => menu.buttons.length > 0 && menuIds.includes(menu.id),
  );
  if (selectedMenus.length === 0) {
    return (
      <Typography.Text type="secondary">
        先在上方选择菜单，再勾选对应按钮。
      </Typography.Text>
    );
  }
  return (
    <Space direction="vertical" size={8} style={{ display: 'flex' }}>
      {selectedMenus.map((menu) => {
        const menuButtonIds = menu.buttons.map((button) => button.id);
        const checkedCount = menuButtonIds.filter((id) =>
          value.includes(id),
        ).length;
        return (
          <div
            key={menu.id}
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
                {menu.name}
              </Typography.Text>
              <Checkbox
                checked={checkedCount === menuButtonIds.length}
                indeterminate={
                  checkedCount > 0 && checkedCount < menuButtonIds.length
                }
                onChange={(event) => {
                  const next = event.target.checked
                    ? [...new Set([...value, ...menuButtonIds])]
                    : value.filter((id) => !menuButtonIds.includes(id));
                  onChange?.(next);
                }}
              >
                全选
              </Checkbox>
            </div>
            <Checkbox.Group
              style={{ display: 'flex', flexWrap: 'wrap', columnGap: 16 }}
              value={value.filter((id) => menuButtonIds.includes(id))}
              options={menu.buttons.map((button) => ({
                label: button.description || button.name,
                value: button.id,
              }))}
              onChange={(next) => {
                const others = value.filter(
                  (id) => !menuButtonIds.includes(id),
                );
                onChange?.([...others, ...next]);
              }}
            />
          </div>
        );
      })}
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
  // 授权分区折叠面板的展开状态；默认全部收起，保证抽屉在矮窗口内不用滚动。
  const [openPanels, setOpenPanels] = useState<string[]>([]);
  // 表单值的镜像，用于在折叠面板标题上实时显示已选数量。
  const [liveValues, setLiveValues] = useState<Partial<RoleFormValues>>({});

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
  const menuSelectNodes = toMenuTreeNodes(menus);
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

  // 抽屉每次打开时同步计数镜像并收起全部分区。
  useEffect(() => {
    if (drawerOpen) {
      setLiveValues(formInitialValues);
      setOpenPanels([]);
    }
    // biome 忽略：formInitialValues 随 editing/copying 变化，即为目标依赖。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [drawerOpen, editing, copying]);

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
                      勾选状态为该角色可见的菜单，点击「编辑」调整授权。
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
        onValuesChange={(_, values) => setLiveValues(values)}
        onFinishFailed={(errorInfo) => {
          // 报错字段可能位于收起的折叠面板里，展开对应分区并滚动到错误处。
          const fields = errorInfo.errorFields.map((field) => field.name?.[0]);
          if (fields.includes('permissions')) {
            setOpenPanels((previous) =>
              previous.includes('func') ? previous : [...previous, 'func'],
            );
          }
          if (fields.includes('menu_ids') || fields.includes('button_ids')) {
            setOpenPanels((previous) =>
              previous.includes('menus') ? previous : [...previous, 'menus'],
            );
          }
          setTimeout(() => {
            document
              .querySelector('.ant-drawer-body .ant-form-item-explain-error')
              ?.scrollIntoView({ block: 'center', behavior: 'smooth' });
          }, 300);
        }}
        onFinish={async (values) => {
          // 菜单授权做祖先闭包：只勾子菜单时自动带上父级，保证侧边栏完整。
          const menu_ids = withMenuAncestors(values.menu_ids ?? [], menus);
          if (editing) {
            await updateRole(editing.id, {
              parent_id: values.parent_id,
              name: values.name,
              permissions: values.permissions,
              menu_ids,
              api_ids: values.api_ids,
              button_ids: values.button_ids,
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
              api_ids: values.api_ids,
              button_ids: values.button_ids,
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
            权限、菜单、API、按钮和数据角色授权将从源角色「{copying.name}
            」复制。
          </Typography.Text>
        ) : (
          <Collapse
            ghost
            activeKey={openPanels}
            onChange={(keys) =>
              setOpenPanels(Array.isArray(keys) ? keys : [keys])
            }
            items={[
              {
                key: 'menus',
                label: grantPanelLabel(
                  '菜单与按钮权限',
                  `${(liveValues.menu_ids ?? []).length} 菜单 · ${(liveValues.button_ids ?? []).length} 按钮`,
                ),
                children: (
                  <>
                    <ProFormTreeSelect
                      name="menu_ids"
                      label="可见菜单"
                      colProps={{ span: 24 }}
                      fieldProps={{
                        treeData: menuSelectNodes,
                        treeCheckable: true,
                        treeDefaultExpandAll: true,
                        showSearch: true,
                        treeNodeFilterProp: 'title',
                        maxTagCount: 'responsive',
                      }}
                    />
                    <ProFormDependency name={['menu_ids']}>
                      {({ menu_ids }) => {
                        // 按钮权限只展示已选菜单下的按钮，避免全量按钮无处下手。
                        const hasButtons = menus.some(
                          (menu) =>
                            menu.buttons.length > 0 &&
                            menu_ids?.includes(menu.id),
                        );
                        if (!hasButtons) {
                          return null;
                        }
                        return (
                          <Form.Item name="button_ids" label="按钮权限">
                            <ButtonGroupsField
                              menus={menus}
                              menuIds={menu_ids ?? []}
                            />
                          </Form.Item>
                        );
                      }}
                    </ProFormDependency>
                  </>
                ),
              },
              {
                key: 'func',
                label: grantPanelLabel(
                  '功能权限',
                  `${(liveValues.permissions ?? []).length} 项`,
                  true,
                ),
                children: (
                  <ProFormSelect
                    name="permissions"
                    label="权限"
                    mode="multiple"
                    colProps={{ span: 24 }}
                    options={permissionOptionGroups(permissions)}
                    fieldProps={{
                      showSearch: true,
                      optionFilterProp: 'label',
                      maxTagCount: 'responsive',
                    }}
                    rules={[{ required: true, message: '请选择权限' }]}
                  />
                ),
              },
              {
                key: 'api',
                label: grantPanelLabel(
                  'API 权限',
                  `${(liveValues.api_ids ?? []).length} 条`,
                ),
                children: (
                  <ProFormSelect
                    name="api_ids"
                    label="受管路由"
                    mode="multiple"
                    colProps={{ span: 24 }}
                    options={apiOptionGroups(apis)}
                    fieldProps={{
                      showSearch: true,
                      optionFilterProp: 'label',
                      maxTagCount: 'responsive',
                    }}
                  />
                ),
              },
              {
                key: 'data',
                label: grantPanelLabel(
                  '数据权限',
                  `${(liveValues.data_role_ids ?? []).length} 个角色`,
                ),
                children: (
                  <ProFormSelect
                    name="data_role_ids"
                    label="数据角色"
                    mode="multiple"
                    colProps={{ span: 24 }}
                    options={roles.map((role) => ({
                      label: role.name,
                      value: role.id,
                    }))}
                    fieldProps={{ maxTagCount: 'responsive' }}
                  />
                ),
              },
            ]}
          />
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
