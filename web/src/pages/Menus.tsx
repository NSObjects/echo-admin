import {
  DeleteOutlined,
  EyeOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Descriptions,
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
import type { DescriptionsProps } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  createMenu,
  deleteMenu,
  readMenu,
  listMenuRoles,
  type ListParams,
  listMenus,
  listPermissions,
  listRoles,
  type Menu,
  type PermissionDefinition,
  type Role,
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

const Menus: React.FC = () => {
  const access = useAccess();
  const [menus, setMenus] = useState<Menu[]>([]);
  const [permissions, setPermissions] = useState<PermissionDefinition[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [roleModalOpen, setRoleModalOpen] = useState(false);
  const [roleLoading, setRoleLoading] = useState(false);
  const [editing, setEditing] = useState<Menu>();
  const [roleTarget, setRoleTarget] = useState<Menu>();
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailItems, setDetailItems] = useState<DescriptionsProps['items']>(
    [],
  );
  const [form] = Form.useForm<MenuFormValues>();
  const [roleForm] = Form.useForm<{ role_ids: number[] }>();

  const loadData = async () => {
    setLoading(true);
    try {
      const [menuResponse, permissionResponse, roleResponse] =
        await Promise.all([
          listMenus(),
          listPermissions(),
          access.canRoleRead
            ? listRoles({ page_size: 100 } as ListParams)
            : Promise.resolve({ data: [] }),
        ]);
      setMenus(menuResponse);
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
    form.setFieldsValue({
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
    });
    setModalOpen(true);
  };

  const openRoleModal = async (record: Menu) => {
    setRoleTarget(record);
    setRoleModalOpen(true);
    setRoleLoading(true);
    try {
      const roleIDs = await listMenuRoles(record.id);
      roleForm.setFieldsValue({ role_ids: roleIDs });
    } finally {
      setRoleLoading(false);
    }
  };

  const openEdit = (record: Menu) => {
    setEditing(record);
    form.setFieldsValue({
      parent_id: record.parent_id,
      name: record.name,
      path: record.path,
      icon: record.icon,
      hidden: record.hidden,
      component: record.component,
      meta: record.meta,
      permission: record.permission,
      sort: record.sort,
      active: record.active,
      buttons: record.buttons.map((button) => ({
        id: button.id,
        name: button.name,
        description: button.description,
      })),
    });
    setModalOpen(true);
  };

  const openDetail = async (record: Menu) => {
    const detail = await readMenu(record.id);
    setDetailItems([
      { key: 'name', label: '名称', children: detail.name },
      { key: 'path', label: '路径', children: detail.path },
      { key: 'component', label: '组件', children: detail.component },
      { key: 'parent', label: '上级', children: detail.parent_id || '顶级菜单' },
      { key: 'icon', label: '图标', children: detail.icon || '-' },
      { key: 'permission', label: '权限', children: detail.permission || '-' },
      { key: 'hidden', label: '隐藏', children: detail.hidden ? '是' : '否' },
      { key: 'active', label: '启用', children: detail.active ? '是' : '否' },
      { key: 'sort', label: '排序', children: detail.sort },
      {
        key: 'keep_alive',
        label: 'KeepAlive',
        children: detail.meta.keep_alive ? '是' : '否',
      },
      {
        key: 'default_menu',
        label: '默认菜单',
        children: detail.meta.default_menu ? '是' : '否',
      },
      {
        key: 'close_tab',
        label: '关闭标签',
        children: detail.meta.close_tab ? '是' : '否',
      },
      {
        key: 'transition',
        label: '过渡',
        children: detail.meta.transition_type || '-',
      },
      {
        key: 'buttons',
        label: '按钮',
        children:
          detail.buttons
            .map((button) => button.description || button.name)
            .join(', ') || '-',
      },
    ]);
    setDetailOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
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
    setModalOpen(false);
    await loadData();
  };

  const removeMenu = async (record: Menu) => {
    await deleteMenu(record.id);
    message.success('菜单已删除');
    await loadData();
  };

  const submitRoles = async () => {
    if (!roleTarget) {
      return;
    }
    const values = await roleForm.validateFields();
    await setMenuRoles(roleTarget.id, values.role_ids ?? []);
    message.success('菜单授权角色已更新');
    setRoleModalOpen(false);
  };

  const columns: ColumnsType<Menu> = [
    { title: '名称', dataIndex: 'name' },
    { title: '路径', dataIndex: 'path' },
    { title: '组件', dataIndex: 'component' },
    {
      title: '上级',
      dataIndex: 'parent_id',
      render: (parentID: number) =>
        parentID === 0
          ? '顶级菜单'
          : (menus.find((menu) => menu.id === parentID)?.name ??
            `#${parentID}`),
    },
    {
      title: '权限',
      dataIndex: 'permission',
      render: (value?: string) => value || '-',
    },
    {
      title: '隐藏',
      dataIndex: 'hidden',
      render: (hidden: boolean) => (
        <Tag color={hidden ? 'default' : 'blue'}>
          {hidden ? '隐藏' : '显示'}
        </Tag>
      ),
    },
    {
      title: '按钮',
      dataIndex: 'buttons',
      render: (buttons: Menu['buttons']) => `${buttons.length} 个`,
    },
    { title: '排序', dataIndex: 'sort', width: 88 },
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
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => void openDetail(record)}
          >
            详情
          </Button>
          {access.canMenuUpdate ? (
            <Button type="link" onClick={() => openEdit(record)}>
              编辑
            </Button>
          ) : null}
          {access.canMenuUpdate && access.canRoleRead ? (
            <Button type="link" onClick={() => void openRoleModal(record)}>
              授权角色
            </Button>
          ) : null}
          {access.canMenuDelete ? (
            <Popconfirm
              title="删除菜单"
              description={`确认删除 ${record.name}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeMenu(record)}
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

  return (
    <PageContainer title="菜单管理">
      <Table<Menu>
        rowKey="id"
        columns={columns}
        dataSource={menus}
        loading={loading}
        pagination={false}
        title={() => (
          <Space>
            {access.canMenuCreate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={openCreate}
              >
                新增菜单
              </Button>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title={editing ? '编辑菜单' : '新增菜单'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form<MenuFormValues> form={form} layout="vertical">
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入菜单名称' }]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item
            label="路径"
            name="path"
            rules={[{ required: true, message: '请输入菜单路径' }]}
          >
            <Input maxLength={160} />
          </Form.Item>
          <Form.Item
            label="组件"
            name="component"
            rules={[{ required: true, message: '请输入组件路径' }]}
          >
            <Input maxLength={160} />
          </Form.Item>
          <Form.Item label="上级菜单" name="parent_id">
            <Select
              options={[
                { label: '顶级菜单', value: 0 },
                ...menus
                  .filter((menu) => menu.id !== editing?.id)
                  .map((menu) => ({ label: menu.name, value: menu.id })),
              ]}
            />
          </Form.Item>
          <Form.Item label="图标" name="icon">
            <Input maxLength={80} />
          </Form.Item>
          <Space size="large" wrap>
            <Form.Item label="隐藏" name="hidden" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item
              label="缓存"
              name={['meta', 'keep_alive']}
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
            <Form.Item
              label="默认菜单"
              name={['meta', 'default_menu']}
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
            <Form.Item
              label="允许关闭"
              name={['meta', 'close_tab']}
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
          </Space>
          <Form.Item label="激活菜单名" name={['meta', 'active_name']}>
            <Input maxLength={160} />
          </Form.Item>
          <Form.Item label="切换动画" name={['meta', 'transition_type']}>
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item label="权限" name="permission">
            <Select allowClear options={permissionOptions} />
          </Form.Item>
          <Form.Item label="排序" name="sort">
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="启用" name="active" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.List name="buttons">
            {(fields, { add, remove }) => (
              <Space direction="vertical" style={{ width: '100%' }}>
                {fields.map((field) => (
                  <Space key={field.key} align="baseline" wrap>
                    <Form.Item name={[field.name, 'id']} hidden>
                      <InputNumber />
                    </Form.Item>
                    <Form.Item
                      label="按钮 key"
                      name={[field.name, 'name']}
                      rules={[{ required: true, message: '请输入按钮 key' }]}
                    >
                      <Input maxLength={80} />
                    </Form.Item>
                    <Form.Item
                      label="按钮说明"
                      name={[field.name, 'description']}
                    >
                      <Input maxLength={120} />
                    </Form.Item>
                    <Button
                      danger
                      icon={<DeleteOutlined />}
                      onClick={() => remove(field.name)}
                    />
                  </Space>
                ))}
                <Button icon={<PlusOutlined />} onClick={() => add()}>
                  添加按钮
                </Button>
              </Space>
            )}
          </Form.List>
        </Form>
      </Modal>
      <Modal
        title="菜单详情"
        open={detailOpen}
        footer={null}
        onCancel={() => setDetailOpen(false)}
        destroyOnHidden
      >
        <Descriptions column={1} size="small" items={detailItems} />
      </Modal>
      <Modal
        title={roleTarget ? `授权角色 - ${roleTarget.name}` : '授权角色'}
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

export default Menus;
