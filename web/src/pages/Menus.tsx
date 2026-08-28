import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProDescriptions,
  ProForm,
  ProFormDigit,
  ProFormList,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTreeSelect,
  ProTable,
} from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Drawer,
  Form,
  InputNumber,
  message,
  Popconfirm,
  Tag,
} from 'antd';
import React, { useEffect, useRef, useState } from 'react';

import {
  createMenu,
  deleteMenu,
  listMenuRoles,
  listMenus,
  listPermissions,
  listRoles,
  type Menu,
  type PermissionDefinition,
  type Role,
  readMenu,
  setMenuRoles,
  updateMenu,
} from '@/services/admin';

type MenuFormValues = {
  parent_id: number;
  name: string;
  path: string;
  icon?: string;
  hidden: boolean;
  component: string;
  meta: {
    active_name?: string;
    keep_alive: boolean;
    default_menu: boolean;
    close_tab: boolean;
    transition_type?: string;
  };
  permission?: string;
  sort: number;
  active: boolean;
  buttons?: {
    id?: number;
    name: string;
    description?: string;
  }[];
};

type MenuTreeNode = {
  title: string;
  value: number;
  children: MenuTreeNode[];
};

type RoleGrantTarget = {
  menu: Menu;
  role_ids: number[];
};

// 菜单接口返回扁平列表，这里按 parent_id 组装成树供父级选择。
const buildMenuTree = (
  menus: Menu[],
  parentID: number,
  excludeID?: number,
): MenuTreeNode[] =>
  menus
    .filter((menu) => menu.parent_id === parentID && menu.id !== excludeID)
    .map((menu) => ({
      title: menu.name,
      value: menu.id,
      children: buildMenuTree(menus, menu.id, excludeID),
    }));

const Menus: React.FC = () => {
  const access = useAccess();
  const actionRef = useRef<ActionType | undefined>(undefined);
  const [menus, setMenus] = useState<Menu[]>([]);
  const [permissions, setPermissions] = useState<PermissionDefinition[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Menu>();
  const [detail, setDetail] = useState<Menu>();
  const [roleTarget, setRoleTarget] = useState<RoleGrantTarget>();

  useEffect(() => {
    void Promise.all([
      listPermissions(),
      access.canRoleRead
        ? listRoles({ page_size: 100 })
        : Promise.resolve({ data: [] }),
    ]).then(([permissionResponse, roleResponse]) => {
      setPermissions(permissionResponse);
      setRoles(roleResponse.data);
    });
  }, [access]);

  const menuName = (menuID: number) =>
    menus.find((menu) => menu.id === menuID)?.name ?? `#${menuID}`;

  const columns: ProColumns<Menu>[] = [
    { title: '名称', dataIndex: 'name' },
    { title: '路径', dataIndex: 'path' },
    { title: '组件', dataIndex: 'component' },
    {
      title: '上级',
      dataIndex: 'parent_id',
      render: (_, record) =>
        record.parent_id === 0 ? '顶级菜单' : menuName(record.parent_id),
    },
    {
      title: '权限',
      dataIndex: 'permission',
      render: (_, record) => record.permission || '-',
    },
    {
      title: '隐藏',
      dataIndex: 'hidden',
      render: (_, record) => (
        <Tag color={record.hidden ? 'default' : 'blue'}>
          {record.hidden ? '隐藏' : '显示'}
        </Tag>
      ),
    },
    {
      title: '按钮',
      dataIndex: 'buttons',
      render: (_, record) => `${record.buttons.length} 个`,
    },
    { title: '排序', dataIndex: 'sort', width: 88 },
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
      width: 220,
      render: (_, record) => {
        const actions: React.ReactNode[] = [
          <a
            key="detail"
            onClick={() => {
              void readMenu(record.id).then(setDetail);
            }}
          >
            详情
          </a>,
        ];
        if (access.canMenuUpdate) {
          actions.push(
            <a
              key="edit"
              onClick={() => {
                setEditing(record);
                setModalOpen(true);
              }}
            >
              编辑
            </a>,
          );
        }
        if (access.canMenuUpdate && access.canRoleRead) {
          actions.push(
            <a
              key="grant"
              onClick={() => {
                void listMenuRoles(record.id).then((roleIDs) => {
                  setRoleTarget({ menu: record, role_ids: roleIDs });
                });
              }}
            >
              授权角色
            </a>,
          );
        }
        if (access.canMenuDelete) {
          actions.push(
            <Popconfirm
              key="delete"
              title="删除菜单"
              description={`确认删除 ${record.name}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={async () => {
                await deleteMenu(record.id);
                message.success('菜单已删除');
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

  const menuTreeData: MenuTreeNode[] = [
    {
      title: '顶级菜单',
      value: 0,
      children: buildMenuTree(menus, 0, editing?.id),
    },
  ];

  const permissionOptions = permissions.map((permission) => ({
    label: `${permission.name} (${permission.token})`,
    value: permission.token,
  }));

  return (
    <PageContainer title="菜单管理">
      <ProTable<Menu>
        headerTitle="菜单列表"
        rowKey="id"
        actionRef={actionRef}
        search={false}
        pagination={false}
        columns={columns}
        request={async () => {
          const data = await listMenus();
          setMenus(data);
          return { data, success: true, total: data.length };
        }}
        toolBarRender={() =>
          access.canMenuCreate
            ? [
                <Button
                  key="create"
                  type="primary"
                  onClick={() => {
                    setEditing(undefined);
                    setModalOpen(true);
                  }}
                >
                  新增菜单
                </Button>,
              ]
            : []
        }
      />
      <ModalForm<MenuFormValues>
        key={editing?.id ?? 'create'}
        title={editing ? '编辑菜单' : '新增菜单'}
        open={modalOpen}
        onOpenChange={setModalOpen}
        width={560}
        modalProps={{ destroyOnHidden: true }}
        initialValues={
          editing
            ? {
                parent_id: editing.parent_id,
                name: editing.name,
                path: editing.path,
                icon: editing.icon,
                hidden: editing.hidden,
                component: editing.component,
                meta: editing.meta,
                permission: editing.permission,
                sort: editing.sort,
                active: editing.active,
                buttons: editing.buttons.map((button) => ({
                  id: button.id,
                  name: button.name,
                  description: button.description,
                })),
              }
            : {
                active: true,
                hidden: false,
                parent_id: 0,
                sort: 100,
                meta: {
                  keep_alive: false,
                  default_menu: false,
                  close_tab: false,
                },
                buttons: [],
              }
        }
        onFinish={async (values) => {
          const body = {
            parent_id: values.parent_id,
            name: values.name,
            path: values.path,
            icon: values.icon,
            hidden: values.hidden,
            component: values.component,
            meta: {
              active_name: values.meta?.active_name,
              keep_alive: values.meta?.keep_alive ?? false,
              default_menu: values.meta?.default_menu ?? false,
              close_tab: values.meta?.close_tab ?? false,
              transition_type: values.meta?.transition_type,
            },
            permission: values.permission,
            sort: values.sort,
            active: values.active,
            buttons: values.buttons ?? [],
          };
          if (editing) {
            await updateMenu(editing.id, body);
            message.success('菜单已更新');
          } else {
            await createMenu(body);
            message.success('菜单已创建');
          }
          actionRef.current?.reload();
          return true;
        }}
      >
        <ProFormText
          name="name"
          label="名称"
          fieldProps={{ maxLength: 80 }}
          rules={[{ required: true, message: '请输入菜单名称' }]}
        />
        <ProFormText
          name="path"
          label="路径"
          fieldProps={{ maxLength: 160 }}
          rules={[{ required: true, message: '请输入菜单路径' }]}
        />
        <ProFormText
          name="component"
          label="组件"
          fieldProps={{ maxLength: 160 }}
          rules={[{ required: true, message: '请输入组件路径' }]}
        />
        <ProFormTreeSelect
          name="parent_id"
          label="上级菜单"
          fieldProps={{ treeData: menuTreeData, treeDefaultExpandAll: true }}
        />
        <ProFormText name="icon" label="图标" fieldProps={{ maxLength: 80 }} />
        <ProForm.Group>
          <ProFormSwitch name="hidden" label="隐藏" />
          <ProFormSwitch name={['meta', 'keep_alive']} label="缓存" />
          <ProFormSwitch name={['meta', 'default_menu']} label="默认菜单" />
          <ProFormSwitch name={['meta', 'close_tab']} label="允许关闭" />
        </ProForm.Group>
        <ProFormText
          name={['meta', 'active_name']}
          label="激活菜单名"
          fieldProps={{ maxLength: 160 }}
        />
        <ProFormText
          name={['meta', 'transition_type']}
          label="切换动画"
          fieldProps={{ maxLength: 80 }}
        />
        <ProFormSelect
          name="permission"
          label="权限"
          options={permissionOptions}
          fieldProps={{ allowClear: true, showSearch: true }}
        />
        <ProFormDigit name="sort" label="排序" min={0} />
        <ProFormSwitch name="active" label="启用" />
        <ProFormList
          name="buttons"
          label="菜单按钮"
          creatorButtonProps={{ creatorButtonText: '添加按钮' }}
        >
          <ProForm.Group key="button-row">
            {/* 保留已有按钮的 id，后端按 id 识别是更新还是新增。 */}
            <Form.Item name="id" hidden key="button-id">
              <InputNumber />
            </Form.Item>
            <ProFormText
              name="name"
              label="按钮 key"
              fieldProps={{ maxLength: 80 }}
              rules={[{ required: true, message: '请输入按钮 key' }]}
            />
            <ProFormText
              name="description"
              label="按钮说明"
              fieldProps={{ maxLength: 120 }}
            />
          </ProForm.Group>
        </ProFormList>
      </ModalForm>
      <Drawer
        title="菜单详情"
        width={520}
        open={Boolean(detail)}
        onClose={() => setDetail(undefined)}
      >
        {detail && (
          <ProDescriptions<Menu>
            column={1}
            size="small"
            dataSource={detail}
            columns={[
              { title: '名称', dataIndex: 'name' },
              { title: '路径', dataIndex: 'path' },
              { title: '组件', dataIndex: 'component' },
              {
                title: '上级',
                dataIndex: 'parent_id',
                render: (_, entity) =>
                  entity.parent_id === 0
                    ? '顶级菜单'
                    : menuName(entity.parent_id),
              },
              {
                title: '图标',
                dataIndex: 'icon',
                render: (_, entity) => entity.icon || '-',
              },
              {
                title: '权限',
                dataIndex: 'permission',
                render: (_, entity) => entity.permission || '-',
              },
              {
                title: '隐藏',
                dataIndex: 'hidden',
                render: (_, entity) => (entity.hidden ? '是' : '否'),
              },
              {
                title: '启用',
                dataIndex: 'active',
                render: (_, entity) => (entity.active ? '是' : '否'),
              },
              { title: '排序', dataIndex: 'sort' },
              {
                title: 'KeepAlive',
                dataIndex: ['meta', 'keep_alive'],
                render: (_, entity) => (entity.meta.keep_alive ? '是' : '否'),
              },
              {
                title: '默认菜单',
                dataIndex: ['meta', 'default_menu'],
                render: (_, entity) => (entity.meta.default_menu ? '是' : '否'),
              },
              {
                title: '关闭标签',
                dataIndex: ['meta', 'close_tab'],
                render: (_, entity) => (entity.meta.close_tab ? '是' : '否'),
              },
              {
                title: '过渡',
                dataIndex: ['meta', 'transition_type'],
                render: (_, entity) => entity.meta.transition_type || '-',
              },
              {
                title: '按钮',
                dataIndex: 'buttons',
                render: (_, entity) =>
                  entity.buttons
                    .map((button) => button.description || button.name)
                    .join(', ') || '-',
              },
            ]}
          />
        )}
      </Drawer>
      <ModalForm<{ role_ids: number[] }>
        key={roleTarget?.menu.id ?? 'idle'}
        title={roleTarget ? `授权角色 - ${roleTarget.menu.name}` : '授权角色'}
        open={Boolean(roleTarget)}
        onOpenChange={(open) => {
          if (!open) {
            setRoleTarget(undefined);
          }
        }}
        modalProps={{ destroyOnHidden: true }}
        initialValues={{ role_ids: roleTarget?.role_ids ?? [] }}
        onFinish={async (values) => {
          if (!roleTarget) {
            return true;
          }
          await setMenuRoles(roleTarget.menu.id, values.role_ids ?? []);
          message.success('菜单授权角色已更新');
          return true;
        }}
      >
        <ProFormSelect
          name="role_ids"
          label="角色"
          mode="multiple"
          options={roles.map((role) => ({
            value: role.id,
            label: role.name,
          }))}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default Menus;
