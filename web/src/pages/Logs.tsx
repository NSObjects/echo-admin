import {
  CheckCircleOutlined,
  DeleteOutlined,
  EyeOutlined,
  ReloadOutlined,
  UndoOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Descriptions,
  Form,
  Input,
  message,
  Modal,
  Popconfirm,
  Space,
  Table,
  Tabs,
  Tag,
} from 'antd';
import type { DescriptionsProps } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  batchDeleteLoginLogs,
  batchDeleteOperationLogs,
  batchDeleteSystemErrorLogs,
  type ListParams,
  type LoginLog,
  type OperationLog,
  type PageMeta,
  type SystemErrorLog,
  deleteLoginLog,
  deleteOperationLog,
  deleteSystemErrorLog,
  readLoginLog,
  readOperationLog,
  readSystemErrorLog,
  listLoginLogs,
  listOperationLogs,
  listSystemErrorLogs,
  reopenSystemErrorLog,
  resolveSystemErrorLog,
} from '@/services/admin';

const formatDate = (value: string) => new Date(value).toLocaleString();

type ResolveFormValues = {
  note?: string;
};

const Logs: React.FC = () => {
  const access = useAccess();
  const [operations, setOperations] = useState<OperationLog[]>([]);
  const [logins, setLogins] = useState<LoginLog[]>([]);
  const [errors, setErrors] = useState<SystemErrorLog[]>([]);
  const [selectedOperationIDs, setSelectedOperationIDs] = useState<React.Key[]>(
    [],
  );
  const [selectedLoginIDs, setSelectedLoginIDs] = useState<React.Key[]>([]);
  const [selectedErrorIDs, setSelectedErrorIDs] = useState<React.Key[]>([]);
  const [operationPage, setOperationPage] = useState<PageMeta>();
  const [loginPage, setLoginPage] = useState<PageMeta>();
  const [errorPage, setErrorPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailTitle, setDetailTitle] = useState('');
  const [detailItems, setDetailItems] = useState<DescriptionsProps['items']>(
    [],
  );
  const [resolveOpen, setResolveOpen] = useState(false);
  const [resolving, setResolving] = useState(false);
  const [selectedError, setSelectedError] = useState<SystemErrorLog>();
  const [resolveForm] = Form.useForm<ResolveFormValues>();

  const loadData = async (
    operationParams: ListParams = {},
    loginParams: ListParams = {},
    errorParams: ListParams = {},
  ) => {
    setLoading(true);
    try {
      const [operationResponse, loginResponse, errorResponse] = await Promise.all([
        listOperationLogs(operationParams),
        listLoginLogs(loginParams),
        listSystemErrorLogs(errorParams),
      ]);
      setOperations(operationResponse.data);
      setOperationPage(operationResponse.page);
      setLogins(loginResponse.data);
      setLoginPage(loginResponse.page);
      setErrors(errorResponse.data);
      setErrorPage(errorResponse.page);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

  const currentParams = {
    operations: {
      page: operationPage?.page,
      page_size: operationPage?.page_size,
    },
    logins: {
      page: loginPage?.page,
      page_size: loginPage?.page_size,
    },
    errors: {
      page: errorPage?.page,
      page_size: errorPage?.page_size,
    },
  };

  const removeOperation = async (record: OperationLog) => {
    await deleteOperationLog(record.id);
    message.success('操作日志已删除');
    await loadData(currentParams.operations, currentParams.logins, currentParams.errors);
  };

  const removeSelectedOperations = async () => {
    await batchDeleteOperationLogs(selectedOperationIDs.map(Number));
    message.success('操作日志已批量删除');
    setSelectedOperationIDs([]);
    await loadData(currentParams.operations, currentParams.logins, currentParams.errors);
  };

  const removeLogin = async (record: LoginLog) => {
    await deleteLoginLog(record.id);
    message.success('登录日志已删除');
    await loadData(currentParams.operations, currentParams.logins, currentParams.errors);
  };

  const removeSelectedLogins = async () => {
    await batchDeleteLoginLogs(selectedLoginIDs.map(Number));
    message.success('登录日志已批量删除');
    setSelectedLoginIDs([]);
    await loadData(currentParams.operations, currentParams.logins, currentParams.errors);
  };

  const removeSystemError = async (record: SystemErrorLog) => {
    await deleteSystemErrorLog(record.id);
    message.success('系统错误日志已删除');
    await loadData(currentParams.operations, currentParams.logins, currentParams.errors);
  };

  const removeSelectedSystemErrors = async () => {
    await batchDeleteSystemErrorLogs(selectedErrorIDs.map(Number));
    message.success('系统错误日志已批量删除');
    setSelectedErrorIDs([]);
    await loadData(currentParams.operations, currentParams.logins, currentParams.errors);
  };

  const openResolve = (record: SystemErrorLog) => {
    setSelectedError(record);
    resolveForm.setFieldsValue({ note: record.resolve_note });
    setResolveOpen(true);
  };

  const submitResolve = async () => {
    if (!selectedError) {
      return;
    }
    const values = await resolveForm.validateFields();
    setResolving(true);
    try {
      await resolveSystemErrorLog(selectedError.id, values.note?.trim());
      message.success('系统错误已标记为已处理');
      setResolveOpen(false);
      resolveForm.resetFields();
      setSelectedError(undefined);
      await loadData(currentParams.operations, currentParams.logins, currentParams.errors);
    } finally {
      setResolving(false);
    }
  };

  const reopenSystemError = async (record: SystemErrorLog) => {
    await reopenSystemErrorLog(record.id);
    message.success('系统错误已取消处理');
    await loadData(currentParams.operations, currentParams.logins, currentParams.errors);
  };

  const showDetail = (
    title: string,
    items: NonNullable<DescriptionsProps['items']>,
  ) => {
    setDetailTitle(title);
    setDetailItems(items);
    setDetailOpen(true);
  };

  const openOperationDetail = async (record: OperationLog) => {
    const detail = await readOperationLog(record.id);
    showDetail(`操作日志 #${detail.id}`, [
      { key: 'actor', label: '操作者', children: detail.actor_id },
      { key: 'action', label: '动作', children: detail.action },
      { key: 'resource', label: '资源', children: detail.resource },
      { key: 'resource_id', label: '资源 ID', children: detail.resource_id },
      { key: 'method', label: '方法', children: detail.method },
      { key: 'path', label: '路径', children: detail.path },
      { key: 'ip', label: 'IP', children: detail.ip },
      { key: 'user_agent', label: 'User-Agent', children: detail.user_agent },
      { key: 'message', label: '消息', children: detail.message },
      { key: 'time', label: '时间', children: formatDate(detail.created_at) },
    ]);
  };

  const openLoginDetail = async (record: LoginLog) => {
    const detail = await readLoginLog(record.id);
    showDetail(`登录日志 #${detail.id}`, [
      { key: 'admin', label: '管理员', children: detail.admin_id },
      { key: 'username', label: '用户名', children: detail.username },
      { key: 'ip', label: 'IP', children: detail.ip },
      { key: 'user_agent', label: 'User-Agent', children: detail.user_agent },
      { key: 'reason', label: '原因', children: detail.reason || '-' },
      { key: 'time', label: '时间', children: formatDate(detail.created_at) },
    ]);
  };

  const openSystemErrorDetail = async (record: SystemErrorLog) => {
    const detail = await readSystemErrorLog(record.id);
    showDetail(`系统错误 #${detail.id}`, [
      { key: 'code', label: '代码', children: detail.code },
      { key: 'message', label: '消息', children: detail.message },
      { key: 'method', label: '方法', children: detail.method },
      { key: 'path', label: '路径', children: detail.path },
      { key: 'ip', label: 'IP', children: detail.ip },
      { key: 'user_agent', label: 'User-Agent', children: detail.user_agent },
      { key: 'request_id', label: '请求ID', children: detail.request_id },
      { key: 'user_id', label: '用户', children: detail.user_id || '-' },
      { key: 'detail', label: '详情', children: detail.detail || '-' },
      {
        key: 'resolved',
        label: '处理状态',
        children: detail.resolved ? '已处理' : '未处理',
      },
      {
        key: 'resolved_by',
        label: '处理人',
        children: detail.resolved_by || '-',
      },
      {
        key: 'resolve_note',
        label: '处理备注',
        children: detail.resolve_note || '-',
      },
      {
        key: 'resolved_at',
        label: '处理时间',
        children: detail.resolved_at ? formatDate(detail.resolved_at) : '-',
      },
      { key: 'time', label: '时间', children: formatDate(detail.created_at) },
    ]);
  };

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
    {
      title: '操作',
      key: 'actions',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => void openOperationDetail(record)}
          >
            详情
          </Button>
          {access.canLogDelete ? (
            <Popconfirm
              title="删除操作日志"
              description={`确认删除 #${record.id}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeOperation(record)}
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
    {
      title: '操作',
      key: 'actions',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => void openLoginDetail(record)}
          >
            详情
          </Button>
          {access.canLogDelete ? (
            <Popconfirm
              title="删除登录日志"
              description={`确认删除 #${record.id}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeLogin(record)}
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

  const errorColumns: ColumnsType<SystemErrorLog> = [
    { title: '代码', dataIndex: 'code', width: 96 },
    { title: '消息', dataIndex: 'message' },
    { title: '方法', dataIndex: 'method', width: 88 },
    { title: '路径', dataIndex: 'path' },
    { title: '用户', dataIndex: 'user_id', width: 96, render: (value?: string) => value || '-' },
    {
      title: '状态',
      dataIndex: 'resolved',
      width: 96,
      render: (resolved: boolean) => (
        <Tag color={resolved ? 'green' : 'red'}>
          {resolved ? '已处理' : '未处理'}
        </Tag>
      ),
    },
    { title: '请求ID', dataIndex: 'request_id', ellipsis: true },
    { title: '时间', dataIndex: 'created_at', render: formatDate },
    {
      title: '操作',
      key: 'actions',
      width: 260,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => void openSystemErrorDetail(record)}
          >
            详情
          </Button>
          {access.canLogResolve ? (
            record.resolved ? (
              <Popconfirm
                title="取消处理状态"
                description={`确认把 #${record.id} 标记为未处理？`}
                okText="确认"
                onConfirm={() => void reopenSystemError(record)}
              >
                <Button type="link" icon={<UndoOutlined />}>
                  取消处理
                </Button>
              </Popconfirm>
            ) : (
              <Button
                type="link"
                icon={<CheckCircleOutlined />}
                onClick={() => openResolve(record)}
              >
                处理
              </Button>
            )
          ) : null}
          {access.canLogDelete ? (
            <Popconfirm
              title="删除系统错误日志"
              description={`确认删除 #${record.id}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeSystemError(record)}
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
                rowSelection={
                  access.canLogDelete
                    ? {
                        selectedRowKeys: selectedOperationIDs,
                        onChange: setSelectedOperationIDs,
                      }
                    : undefined
                }
                title={() =>
                  access.canLogDelete && selectedOperationIDs.length > 0 ? (
                    <Popconfirm
                      title="批量删除操作日志"
                      description={`确认删除选中的 ${selectedOperationIDs.length} 条操作日志？`}
                      okText="删除"
                      okButtonProps={{ danger: true }}
                      onConfirm={() => void removeSelectedOperations()}
                    >
                      <Button danger icon={<DeleteOutlined />}>
                        批量删除
                      </Button>
                    </Popconfirm>
                  ) : null
                }
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
                    {
                      page: errorPage?.page,
                      page_size: errorPage?.page_size,
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
                rowSelection={
                  access.canLogDelete
                    ? {
                        selectedRowKeys: selectedLoginIDs,
                        onChange: setSelectedLoginIDs,
                      }
                    : undefined
                }
                title={() =>
                  access.canLogDelete && selectedLoginIDs.length > 0 ? (
                    <Popconfirm
                      title="批量删除登录日志"
                      description={`确认删除选中的 ${selectedLoginIDs.length} 条登录日志？`}
                      okText="删除"
                      okButtonProps={{ danger: true }}
                      onConfirm={() => void removeSelectedLogins()}
                    >
                      <Button danger icon={<DeleteOutlined />}>
                        批量删除
                      </Button>
                    </Popconfirm>
                  ) : null
                }
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
                    {
                      page: errorPage?.page,
                      page_size: errorPage?.page_size,
                    },
                  )
                }
              />
            ),
          },
          {
            key: 'errors',
            label: '系统错误',
            children: (
              <Table<SystemErrorLog>
                rowKey="id"
                columns={errorColumns}
                dataSource={errors}
                loading={loading}
                rowSelection={
                  access.canLogDelete
                    ? {
                        selectedRowKeys: selectedErrorIDs,
                        onChange: setSelectedErrorIDs,
                      }
                    : undefined
                }
                title={() =>
                  access.canLogDelete && selectedErrorIDs.length > 0 ? (
                    <Popconfirm
                      title="批量删除系统错误日志"
                      description={`确认删除选中的 ${selectedErrorIDs.length} 条系统错误日志？`}
                      okText="删除"
                      okButtonProps={{ danger: true }}
                      onConfirm={() => void removeSelectedSystemErrors()}
                    >
                      <Button danger icon={<DeleteOutlined />}>
                        批量删除
                      </Button>
                    </Popconfirm>
                  ) : null
                }
                expandable={{ expandedRowRender: (record) => record.detail || '-' }}
                pagination={{
                  current: errorPage?.page,
                  pageSize: errorPage?.page_size,
                  total: errorPage?.total,
                  showSizeChanger: true,
                }}
                onChange={(pagination) =>
                  void loadData(
                    {
                      page: operationPage?.page,
                      page_size: operationPage?.page_size,
                    },
                    {
                      page: loginPage?.page,
                      page_size: loginPage?.page_size,
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
      <Modal
        title={detailTitle}
        open={detailOpen}
        footer={null}
        onCancel={() => setDetailOpen(false)}
        destroyOnHidden
      >
        <Descriptions column={1} size="small" items={detailItems} />
      </Modal>
      <Modal
        title={selectedError ? `处理系统错误 #${selectedError.id}` : '处理系统错误'}
        open={resolveOpen}
        confirmLoading={resolving}
        onOk={() => void submitResolve()}
        onCancel={() => {
          setResolveOpen(false);
          resolveForm.resetFields();
          setSelectedError(undefined);
        }}
        destroyOnHidden
      >
        <Form<ResolveFormValues> form={resolveForm} layout="vertical">
          <Form.Item name="note" label="处理备注" rules={[{ max: 1000 }]}>
            <Input.TextArea rows={4} maxLength={1000} showCount />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default Logs;
