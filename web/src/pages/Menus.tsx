import {
  DeleteOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
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
import React, { useEffect, useState } from 'react';

import {
  createMenu,
  deleteMenu,
  listMenus,
  listPermissions,
  type Menu,
  type PermissionDefinition,
  updateMenu,
} from '@/services/admin';

type MenuFormValues = {
  parent_id: number;
  name: string;
  path: string;
  icon?: string;
  permission?: string;
  sort: number;
  active: boolean;
};

const Menus: React.FC = () => {
  const access = useAccess();
  const [menus, setMenus] = useState<Menu[]>([]);
  const [permissions, setPermissions] = useState<PermissionDefinition[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Menu>();
  const [form] = Form.useForm<MenuFormValues>();

  const loadData = async () => {
    setLoading(true);
    try {
      const [menuResponse, permissionResponse] = await Promise.all([
        listMenus(),
        listPermissions(),
      ]);
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
    form.setFieldsValue({ active: true, parent_id: 0, sort: 100 });
    setModalOpen(true);
  };

  const openEdit = (record: Menu) => {
    setEditing(record);
    form.setFieldsValue(record);
    setModalOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    const body = {
      parent_id: values.parent_id,
      name: values.name,
      path: values.path,
      icon: values.icon,
      permission: values.permission,
      sort: values.sort,
      active: values.active,
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

  const columns: ColumnsType<Menu> = [
    { title: '名称', dataIndex: 'name' },
    { title: '路径', dataIndex: 'path' },
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
          {access.canMenuUpdate ? (
            <Button type="link" onClick={() => openEdit(record)}>
              编辑
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
          <Form.Item label="权限" name="permission">
            <Select allowClear options={permissionOptions} />
          </Form.Item>
          <Form.Item label="排序" name="sort">
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="启用" name="active" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default Menus;
