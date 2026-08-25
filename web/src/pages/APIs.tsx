import { EyeOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Descriptions,
  Form,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  message,
} from 'antd';
import type { DescriptionsProps } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  type APIResource,
  type ListParams,
  type PageMeta,
  type Role,
  listAPIRoles,
  listAPIs,
  listRoles,
  readAPI,
  setAPIRoles,
} from '@/services/admin';

const methodColor: Record<string, string> = {
  GET: 'blue',
  POST: 'green',
  PUT: 'gold',
  PATCH: 'purple',
  DELETE: 'red',
};

const APIs: React.FC = () => {
  const access = useAccess();
  const [apis, setAPIs] = useState<APIResource[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [roleModalOpen, setRoleModalOpen] = useState(false);
  const [roleLoading, setRoleLoading] = useState(false);
  const [roleTarget, setRoleTarget] = useState<APIResource>();
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailItems, setDetailItems] = useState<DescriptionsProps['items']>(
    [],
  );
  const [roleForm] = Form.useForm<{ role_ids: number[] }>();

  const loadData = async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const [apiResponse, roleResponse] = await Promise.all([
        listAPIs(params),
        access.canApiGrant && access.canRoleRead
          ? listRoles({ page_size: 100 })
          : Promise.resolve({ data: [] }),
      ]);
      setAPIs(apiResponse.data);
      setPage(apiResponse.page);
      setRoles(roleResponse.data);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

  const openRoleModal = async (record: APIResource) => {
    setRoleTarget(record);
    setRoleModalOpen(true);
    setRoleLoading(true);
    try {
      const roleIDs = await listAPIRoles(record.id);
      roleForm.setFieldsValue({ role_ids: roleIDs });
    } finally {
      setRoleLoading(false);
    }
  };

  const openDetail = async (record: APIResource) => {
    const detail = await readAPI(record.id);
    setDetailItems([
      { key: 'method', label: '方法', children: detail.method },
      { key: 'path', label: '注册路由模式', children: detail.path },
      { key: 'description', label: '描述', children: detail.description },
      { key: 'group', label: '分组', children: detail.group },
      { key: 'permission', label: '权限', children: detail.permission || '-' },
      {
        key: 'created_at',
        label: '创建时间',
        children: new Date(detail.created_at).toLocaleString(),
      },
      {
        key: 'updated_at',
        label: '更新时间',
        children: new Date(detail.updated_at).toLocaleString(),
      },
    ]);
    setDetailOpen(true);
  };

  const submitRoles = async () => {
    if (!roleTarget) {
      return;
    }
    const values = await roleForm.validateFields();
    await setAPIRoles(roleTarget.id, values.role_ids ?? []);
    message.success('API授权角色已更新');
    setRoleModalOpen(false);
  };

  const columns: ColumnsType<APIResource> = [
    {
      title: '方法',
      dataIndex: 'method',
      width: 96,
      render: (method: string) => (
        <Tag color={methodColor[method] ?? 'default'}>{method}</Tag>
      ),
    },
    { title: '注册路由模式', dataIndex: 'path' },
    { title: '描述', dataIndex: 'description' },
    { title: '分组', dataIndex: 'group', width: 120 },
    {
      title: '权限',
      dataIndex: 'permission',
      render: (permission?: string) => permission || '-',
    },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => void openDetail(record)}
          >
            详情
          </Button>
          {access.canApiGrant && access.canRoleRead ? (
            <Button type="link" onClick={() => void openRoleModal(record)}>
              授权角色
            </Button>
          ) : null}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer
      title="受管 API 路由目录"
      subTitle="路由身份和元数据由部署代码维护，后台只提供查看与角色授权。"
    >
      <Table<APIResource>
        rowKey="id"
        columns={columns}
        dataSource={apis}
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
          <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
            刷新
          </Button>
        )}
      />
      <Modal
        title="API详情"
        open={detailOpen}
        footer={null}
        onCancel={() => setDetailOpen(false)}
        destroyOnHidden
      >
        <Descriptions column={1} size="small" items={detailItems} />
      </Modal>
      <Modal
        title={
          roleTarget
            ? `授权角色 - ${roleTarget.method} ${roleTarget.path}`
            : '授权角色'
        }
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

export default APIs;
