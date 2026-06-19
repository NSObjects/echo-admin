import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Form, Input, Modal, Select, Space, Switch, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  type ListParams,
  type Menu,
  type PageMeta,
  type Role,
  createRole,
  listMenus,
  listRoles,
  updateRole,
} from '@/services/admin';

type RoleFormValues = {
  code: string;
  name: string;
  permissions: string[];
  menu_ids?: number[];
  active: boolean;
};

const permissionOptions = [
  { label: '管理员管理', value: 'admin:manage' },
  { label: '角色权限', value: 'role:manage' },
  { label: '菜单管理', value: 'menu:manage' },
  { label: '系统配置', value: 'config:manage' },
  { label: '数据字典', value: 'dict:manage' },
  { label: '文件上传', value: 'file:upload' },
  { label: '日志查看', value: 'log:read' },
];

const Roles: React.FC = () => {
  const [roles, setRoles] = useState<Role[]>([]);
  const [menus, setMenus] = useState<Menu[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Role>();
  const [form] = Form.useForm<RoleFormValues>();

  const loadData = async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const [roleResponse, menuResponse] = await Promise.all([
        listRoles(params),
        listMenus(),
      ]);
      setRoles(roleResponse.data);
      setPage(roleResponse.page);
      setMenus(menuResponse);
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
    form.setFieldsValue({ active: true, permissions: [], menu_ids: [] });
    setModalOpen(true);
  };

  const openEdit = (record: Role) => {
    setEditing(record);
    form.setFieldsValue({
      code: record.code,
      name: record.name,
      permissions: record.permissions,
      menu_ids: record.menu_ids,
      active: record.active,
    });
    setModalOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    if (editing) {
      await updateRole(editing.id, {
        name: values.name,
        permissions: values.permissions,
        menu_ids: values.menu_ids,
        active: values.active,
      });
      message.success('角色已更新');
    } else {
      await createRole({
        code: values.code,
        name: values.name,
        permissions: values.permissions,
        menu_ids: values.menu_ids,
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
        <Button type="link" onClick={() => openEdit(record)}>
          编辑
        </Button>
      ),
    },
  ];

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
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新增角色
            </Button>
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
            <Select
              mode="multiple"
              options={menus.map((menu) => ({
                label: menu.name,
                value: menu.id,
              }))}
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

export default Roles;
