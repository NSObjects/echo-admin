import {
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  ExportOutlined,
  EyeOutlined,
  ImportOutlined,
  PlusOutlined,
  ReloadOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  DatePicker,
  Descriptions,
  Form,
  Input,
  Modal,
  message,
  Popconfirm,
  Select,
  Space,
  Table,
  Upload,
} from 'antd';
import type { DescriptionsProps, UploadProps } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs, { type Dayjs } from 'dayjs';
import React, { useEffect, useState } from 'react';

import {
  type APIResource,
  batchDeleteVersions,
  createVersion,
  deleteVersion,
  type Dictionary,
  downloadVersionJSON,
  exportVersion,
  type ExportVersionInput,
  readVersion,
  importVersion,
  listAPIs,
  listDictionaries,
  listMenus,
  listVersions,
  type Menu,
  type SystemVersion,
  updateVersion,
  type VersionBundle,
  type VersionInput,
} from '@/services/admin';

type VersionFormValues = {
  version: string;
  name: string;
  description?: string;
  published_at?: Dayjs;
};

type ExportFormValues = {
  version: string;
  name: string;
  description?: string;
  menu_ids?: number[];
  api_ids?: number[];
  dictionary_ids?: number[];
};

type ImportFormValues = {
  data: string;
};

type ExportResources = {
  menus: Menu[];
  apis: APIResource[];
  dictionaries: Dictionary[];
};

const emptyResources: ExportResources = {
  menus: [],
  apis: [],
  dictionaries: [],
};

const formatDate = (value?: string) =>
  value ? new Date(value).toLocaleString() : '-';

const formValuesToInput = (values: VersionFormValues): VersionInput => ({
  version: values.version,
  name: values.name,
  description: values.description,
  published_at: values.published_at?.toISOString(),
});

const exportValuesToInput = (
  values: ExportFormValues,
): ExportVersionInput => ({
  version: values.version,
  name: values.name,
  description: values.description,
  menu_ids: values.menu_ids ?? [],
  api_ids: values.api_ids ?? [],
  dictionary_ids: values.dictionary_ids ?? [],
});

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value);

const isVersionBundle = (value: unknown): value is VersionBundle => {
  if (!isRecord(value) || !isRecord(value.version)) {
    return false;
  }
  return (
    typeof value.version.code === 'string' &&
    typeof value.version.name === 'string' &&
    (value.menus === undefined || Array.isArray(value.menus)) &&
    (value.apis === undefined || Array.isArray(value.apis)) &&
    (value.dictionaries === undefined || Array.isArray(value.dictionaries))
  );
};

const loadAllAPIs = async (): Promise<APIResource[]> => {
  const out: APIResource[] = [];
  for (let page = 1; ; page += 1) {
    const response = await listAPIs({ page, page_size: 100 });
    out.push(...response.data);
    if (!response.page?.has_next) {
      return out;
    }
  }
};

const saveBlob = (blob: Blob, filename: string) => {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
};

const Versions: React.FC = () => {
  const access = useAccess();
  const [versions, setVersions] = useState<SystemVersion[]>([]);
  const [selectedVersionIDs, setSelectedVersionIDs] = useState<React.Key[]>([]);
  const [resources, setResources] = useState<ExportResources>(emptyResources);
  const [loading, setLoading] = useState(false);
  const [resourceLoading, setResourceLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [exportModalOpen, setExportModalOpen] = useState(false);
  const [importModalOpen, setImportModalOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailItems, setDetailItems] = useState<DescriptionsProps['items']>(
    [],
  );
  const [editing, setEditing] = useState<SystemVersion>();
  const [form] = Form.useForm<VersionFormValues>();
  const [exportForm] = Form.useForm<ExportFormValues>();
  const [importForm] = Form.useForm<ImportFormValues>();

  const loadData = async () => {
    setLoading(true);
    try {
      setVersions(await listVersions());
    } finally {
      setLoading(false);
    }
  };

  const loadResources = async () => {
    setResourceLoading(true);
    try {
      const [menus, apis, dictionaries] = await Promise.all([
        access.canMenuRead ? listMenus() : Promise.resolve([]),
        access.canApiRead ? loadAllAPIs() : Promise.resolve([]),
        access.canDictRead ? listDictionaries() : Promise.resolve([]),
      ]);
      setResources({ menus, apis, dictionaries });
    } finally {
      setResourceLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

  const openCreate = () => {
    setEditing(undefined);
    form.resetFields();
    form.setFieldsValue({ published_at: dayjs() });
    setModalOpen(true);
  };

  const openExport = () => {
    exportForm.resetFields();
    exportForm.setFieldsValue({
      version: `v${dayjs().format('YYYY.MM.DD.HHmm')}`,
      name: '后台配置包',
    });
    setExportModalOpen(true);
    void loadResources();
  };

  const openImport = () => {
    importForm.resetFields();
    setImportModalOpen(true);
  };

  const openEdit = (record: SystemVersion) => {
    setEditing(record);
    form.setFieldsValue({
      version: record.version,
      name: record.name,
      description: record.description,
      published_at: dayjs(record.published_at),
    });
    setModalOpen(true);
  };

  const openDetail = async (record: SystemVersion) => {
    const detail = await readVersion(record.id);
    setDetailItems([
      { key: 'version', label: '版本号', children: detail.version },
      { key: 'name', label: '名称', children: detail.name },
      { key: 'description', label: '说明', children: detail.description || '-' },
      {
        key: 'published_at',
        label: '发布时间',
        children: formatDate(detail.published_at),
      },
      {
        key: 'created_at',
        label: '创建时间',
        children: formatDate(detail.created_at),
      },
      {
        key: 'updated_at',
        label: '更新时间',
        children: formatDate(detail.updated_at),
      },
    ]);
    setDetailOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    const input = formValuesToInput(values);
    if (editing) {
      await updateVersion(editing.id, input);
      message.success('版本记录已更新');
    } else {
      await createVersion(input);
      message.success('版本记录已创建');
    }
    setModalOpen(false);
    await loadData();
  };

  const submitExport = async () => {
    const values = await exportForm.validateFields();
    const created = await exportVersion(exportValuesToInput(values));
    message.success('版本包已导出');
    setExportModalOpen(false);
    await downloadVersion(created);
    await loadData();
  };

  const submitImport = async () => {
    const values = await importForm.validateFields();
    try {
      const parsed: unknown = JSON.parse(values.data);
      if (!isVersionBundle(parsed)) {
        message.error('版本包结构不正确');
        return;
      }
      await importVersion(parsed);
      message.success('版本包已导入');
      setImportModalOpen(false);
      await loadData();
    } catch (error) {
      message.error(
        error instanceof SyntaxError
          ? 'JSON格式不正确'
          : error instanceof Error
            ? error.message
            : '版本包导入失败',
      );
    }
  };

  const removeVersion = async (record: SystemVersion) => {
    await deleteVersion(record.id);
    message.success('版本记录已删除');
    await loadData();
  };

  const removeSelectedVersions = async () => {
    await batchDeleteVersions(selectedVersionIDs.map(Number));
    message.success('版本记录已批量删除');
    setSelectedVersionIDs([]);
    await loadData();
  };

  const downloadVersion = async (record: SystemVersion) => {
    try {
      const blob = await downloadVersionJSON(record.id);
      saveBlob(blob, `version_${record.version}.json`);
      message.success('版本JSON已下载');
    } catch (error) {
      message.error(error instanceof Error ? error.message : '版本JSON下载失败');
    }
  };

  const uploadProps: UploadProps = {
    accept: 'application/json,.json',
    maxCount: 1,
    beforeUpload: (file) => {
      const reader = new FileReader();
      reader.onload = () => {
        importForm.setFieldsValue({
          data: typeof reader.result === 'string' ? reader.result : '',
        });
      };
      reader.onerror = () => {
        message.error('文件读取失败');
      };
      reader.readAsText(file);
      return false;
    },
  };

  const menuOptions = resources.menus.map((menu) => ({
    label: `${menu.name} (${menu.path})`,
    value: menu.id,
  }));
  const apiOptions = resources.apis.map((api) => ({
    label: `${api.method} ${api.path}`,
    value: api.id,
  }));
  const dictionaryOptions = resources.dictionaries.map((dictionary) => ({
    label: `${dictionary.name} (${dictionary.code})`,
    value: dictionary.id,
  }));

  const columns: ColumnsType<SystemVersion> = [
    { title: '版本号', dataIndex: 'version', width: 160 },
    { title: '名称', dataIndex: 'name', width: 200 },
    { title: '说明', dataIndex: 'description', ellipsis: true },
    {
      title: '发布时间',
      dataIndex: 'published_at',
      width: 200,
      render: formatDate,
    },
    {
      title: '操作',
      key: 'actions',
      width: 260,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => void openDetail(record)}
          >
            详情
          </Button>
          {access.canVersionUpdate ? (
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => openEdit(record)}
            >
              编辑
            </Button>
          ) : null}
          {access.canVersionRead ? (
            <Button
              type="link"
              icon={<DownloadOutlined />}
              onClick={() => void downloadVersion(record)}
            >
              下载
            </Button>
          ) : null}
          {access.canVersionDelete ? (
            <Popconfirm
              title="删除版本记录"
              description={`确认删除 ${record.version}？`}
              okText="删除"
              okButtonProps={{ danger: true }}
              onConfirm={() => void removeVersion(record)}
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
    <PageContainer title="版本管理">
      <Table<SystemVersion>
        rowKey="id"
        columns={columns}
        dataSource={versions}
        loading={loading}
        pagination={false}
        rowSelection={
          access.canVersionDelete
            ? {
                selectedRowKeys: selectedVersionIDs,
                onChange: setSelectedVersionIDs,
              }
            : undefined
        }
        title={() => (
          <Space wrap>
            {access.canVersionCreate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={openCreate}
              >
                新增版本
              </Button>
            ) : null}
            {access.canVersionCreate ? (
              <Button icon={<ExportOutlined />} onClick={openExport}>
                导出
              </Button>
            ) : null}
            {access.canVersionCreate ? (
              <Button icon={<ImportOutlined />} onClick={openImport}>
                导入
              </Button>
            ) : null}
            {access.canVersionDelete && selectedVersionIDs.length > 0 ? (
              <Popconfirm
                title="批量删除版本记录"
                description={`确认删除选中的 ${selectedVersionIDs.length} 条记录？`}
                okText="删除"
                okButtonProps={{ danger: true }}
                onConfirm={() => void removeSelectedVersions()}
              >
                <Button danger icon={<DeleteOutlined />}>
                  批量删除
                </Button>
              </Popconfirm>
            ) : null}
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
      <Modal
        title={editing ? '编辑版本' : '新增版本'}
        open={modalOpen}
        onOk={() => void submit()}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form<VersionFormValues> form={form} layout="vertical">
          <Form.Item
            label="版本号"
            name="version"
            rules={[
              { required: true, message: '请输入版本号' },
              {
                pattern: /^[A-Za-z0-9._+-]+$/,
                message: '版本号只能包含字母、数字、点、横线、下划线和加号',
              },
            ]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入版本名称' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
          <Form.Item label="说明" name="description">
            <Input.TextArea maxLength={4000} rows={5} />
          </Form.Item>
          <Form.Item
            label="发布时间"
            name="published_at"
            rules={[{ required: true, message: '请选择发布时间' }]}
          >
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title="导出版本包"
        open={exportModalOpen}
        onOk={() => void submitExport()}
        onCancel={() => setExportModalOpen(false)}
        destroyOnHidden
      >
        <Form<ExportFormValues> form={exportForm} layout="vertical">
          <Form.Item
            label="版本号"
            name="version"
            rules={[
              { required: true, message: '请输入版本号' },
              {
                pattern: /^[A-Za-z0-9._+-]+$/,
                message: '版本号只能包含字母、数字、点、横线、下划线和加号',
              },
            ]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入版本名称' }]}
          >
            <Input maxLength={120} />
          </Form.Item>
          <Form.Item label="说明" name="description">
            <Input.TextArea maxLength={4000} rows={3} />
          </Form.Item>
          <Form.Item label="菜单" name="menu_ids">
            <Select
              mode="multiple"
              allowClear
              loading={resourceLoading}
              maxTagCount="responsive"
              options={menuOptions}
            />
          </Form.Item>
          <Form.Item label="API" name="api_ids">
            <Select
              mode="multiple"
              allowClear
              loading={resourceLoading}
              maxTagCount="responsive"
              options={apiOptions}
            />
          </Form.Item>
          <Form.Item label="字典" name="dictionary_ids">
            <Select
              mode="multiple"
              allowClear
              loading={resourceLoading}
              maxTagCount="responsive"
              options={dictionaryOptions}
            />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title="导入版本包"
        open={importModalOpen}
        onOk={() => void submitImport()}
        onCancel={() => setImportModalOpen(false)}
        destroyOnHidden
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Upload {...uploadProps}>
            <Button icon={<UploadOutlined />}>选择JSON文件</Button>
          </Upload>
          <Form<ImportFormValues> form={importForm} layout="vertical">
            <Form.Item
              label="JSON内容"
              name="data"
              rules={[{ required: true, message: '请粘贴或选择版本JSON' }]}
            >
              <Input.TextArea rows={12} />
            </Form.Item>
          </Form>
        </Space>
      </Modal>
      <Modal
        title="版本详情"
        open={detailOpen}
        footer={null}
        onCancel={() => setDetailOpen(false)}
        destroyOnHidden
      >
        <Descriptions column={1} items={detailItems} />
      </Modal>
    </PageContainer>
  );
};

export default Versions;
