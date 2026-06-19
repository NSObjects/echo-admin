import { ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Table, Tabs, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  type ListParams,
  type LoginLog,
  type OperationLog,
  type PageMeta,
  listLoginLogs,
  listOperationLogs,
} from '@/services/admin';

const formatDate = (value: string) => new Date(value).toLocaleString();

const Logs: React.FC = () => {
  const [operations, setOperations] = useState<OperationLog[]>([]);
  const [logins, setLogins] = useState<LoginLog[]>([]);
  const [operationPage, setOperationPage] = useState<PageMeta>();
  const [loginPage, setLoginPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);

  const loadData = async (
    operationParams: ListParams = {},
    loginParams: ListParams = {},
  ) => {
    setLoading(true);
    try {
      const [operationResponse, loginResponse] = await Promise.all([
        listOperationLogs(operationParams),
        listLoginLogs(loginParams),
      ]);
      setOperations(operationResponse.data);
      setOperationPage(operationResponse.page);
      setLogins(loginResponse.data);
      setLoginPage(loginResponse.page);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

  const operationColumns: ColumnsType<OperationLog> = [
    { title: '操作者', dataIndex: 'actor_id', width: 96 },
    { title: '动作', dataIndex: 'action', width: 96 },
    { title: '资源', dataIndex: 'resource', width: 120 },
    { title: '资源 ID', dataIndex: 'resource_id' },
    { title: '方法', dataIndex: 'method', width: 88 },
    { title: '路径', dataIndex: 'path' },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    {
      title: '结果',
      dataIndex: 'success',
      render: (success: boolean) => (
        <Tag color={success ? 'green' : 'red'}>{success ? '成功' : '失败'}</Tag>
      ),
    },
    { title: '时间', dataIndex: 'created_at', render: formatDate },
  ];

  const loginColumns: ColumnsType<LoginLog> = [
    { title: '管理员', dataIndex: 'admin_id', width: 96 },
    { title: '用户名', dataIndex: 'username' },
    { title: 'IP', dataIndex: 'ip' },
    {
      title: '结果',
      dataIndex: 'success',
      render: (success: boolean) => (
        <Tag color={success ? 'green' : 'red'}>{success ? '成功' : '失败'}</Tag>
      ),
    },
    { title: '原因', dataIndex: 'reason', render: (value?: string) => value || '-' },
    { title: '时间', dataIndex: 'created_at', render: formatDate },
  ];

  return (
    <PageContainer
      title="系统日志"
      extra={
        <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
          刷新
        </Button>
      }
    >
      <Tabs
        items={[
          {
            key: 'operations',
            label: '操作日志',
            children: (
              <Table<OperationLog>
                rowKey="id"
                columns={operationColumns}
                dataSource={operations}
                loading={loading}
                pagination={{
                  current: operationPage?.page,
                  pageSize: operationPage?.page_size,
                  total: operationPage?.total,
                  showSizeChanger: true,
                }}
                onChange={(pagination) =>
                  void loadData(
                    {
                      page: pagination.current,
                      page_size: pagination.pageSize,
                    },
                    {
                      page: loginPage?.page,
                      page_size: loginPage?.page_size,
                    },
                  )
                }
              />
            ),
          },
          {
            key: 'logins',
            label: '登录日志',
            children: (
              <Table<LoginLog>
                rowKey="id"
                columns={loginColumns}
                dataSource={logins}
                loading={loading}
                pagination={{
                  current: loginPage?.page,
                  pageSize: loginPage?.page_size,
                  total: loginPage?.total,
                  showSizeChanger: true,
                }}
                onChange={(pagination) =>
                  void loadData(
                    {
                      page: operationPage?.page,
                      page_size: operationPage?.page_size,
                    },
                    {
                      page: pagination.current,
                      page_size: pagination.pageSize,
                    },
                  )
                }
              />
            ),
          },
        ]}
      />
    </PageContainer>
  );
};

export default Logs;
