import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProDescriptions,
  ProFormSelect,
  ProTable,
} from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import { Drawer, message, Tag } from 'antd';
import React, { useEffect, useRef, useState } from 'react';

import {
  type APIResource,
  listAPIRoles,
  listAPIs,
  listRoles,
  pageParams,
  type Role,
  readAPI,
  setAPIRoles,
  toTableResult,
} from '@/services/admin';

const methodColor: Record<string, string> = {
  GET: 'blue',
  POST: 'green',
  PUT: 'gold',
  PATCH: 'purple',
  DELETE: 'red',
};

const formatDate = (value?: string) =>
  value ? new Date(value).toLocaleString() : '-';

type RoleGrantTarget = {
  api: APIResource;
  role_ids: number[];
};

const APIs: React.FC = () => {
  const access = useAccess();
  const actionRef = useRef<ActionType | undefined>(undefined);
  const canGrant = access.canApiGrant && access.canRoleRead;
  const [roles, setRoles] = useState<Role[]>([]);
  const [detail, setDetail] = useState<APIResource>();
  const [roleTarget, setRoleTarget] = useState<RoleGrantTarget>();

  useEffect(() => {
    if (!canGrant) {
      return;
    }
    void listRoles({ page_size: 100 }).then((response) => {
      setRoles(response.data);
    });
  }, [canGrant]);

  const columns: ProColumns<APIResource>[] = [
    {
      title: '方法',
      dataIndex: 'method',
      width: 96,
      render: (_, record) => (
        <Tag color={methodColor[record.method] ?? 'default'}>
          {record.method}
        </Tag>
      ),
    },
    { title: '注册路由模式', dataIndex: 'path' },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '分组', dataIndex: 'group', width: 120 },
    {
      title: '权限',
      dataIndex: 'permission',
      render: (_, record) => record.permission || '-',
    },
    {
      title: '操作',
      valueType: 'option',
      width: 160,
      render: (_, record) => {
        const actions: React.ReactNode[] = [
          <a
            key="detail"
            onClick={() => {
              void readAPI(record.id).then(setDetail);
            }}
          >
            详情
          </a>,
        ];
        if (canGrant) {
          actions.push(
            <a
              key="grant"
              onClick={() => {
                void listAPIRoles(record.id).then((roleIDs) => {
                  setRoleTarget({ api: record, role_ids: roleIDs });
                });
              }}
            >
              授权角色
            </a>,
          );
        }
        return actions;
      },
    },
  ];

  return (
    <PageContainer
      title="受管 API 路由目录"
      subTitle="路由身份和元数据由部署代码维护，后台只提供查看与角色授权。"
    >
      <ProTable<APIResource>
        headerTitle="路由列表"
        rowKey="id"
        actionRef={actionRef}
        search={false}
        columns={columns}
        request={async (params) =>
          toTableResult(await listAPIs(pageParams(params)))
        }
      />
      <Drawer
        title="API详情"
        width={520}
        open={Boolean(detail)}
        onClose={() => setDetail(undefined)}
      >
        {detail && (
          <ProDescriptions<APIResource>
            column={1}
            size="small"
            dataSource={detail}
            columns={[
              {
                title: '方法',
                dataIndex: 'method',
                render: (_, entity) => (
                  <Tag color={methodColor[entity.method] ?? 'default'}>
                    {entity.method}
                  </Tag>
                ),
              },
              { title: '注册路由模式', dataIndex: 'path' },
              { title: '描述', dataIndex: 'description' },
              { title: '分组', dataIndex: 'group' },
              {
                title: '权限',
                dataIndex: 'permission',
                render: (_, entity) => entity.permission || '-',
              },
              {
                title: '创建时间',
                dataIndex: 'created_at',
                render: (_, entity) => formatDate(entity.created_at),
              },
              {
                title: '更新时间',
                dataIndex: 'updated_at',
                render: (_, entity) => formatDate(entity.updated_at),
              },
            ]}
          />
        )}
      </Drawer>
      <ModalForm<{ role_ids: number[] }>
        key={roleTarget?.api.id ?? 'idle'}
        title={
          roleTarget
            ? `授权角色 - ${roleTarget.api.method} ${roleTarget.api.path}`
            : '授权角色'
        }
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
          await setAPIRoles(roleTarget.api.id, values.role_ids ?? []);
          message.success('API授权角色已更新');
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

export default APIs;
