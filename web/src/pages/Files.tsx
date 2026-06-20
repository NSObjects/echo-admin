import {
  DeleteOutlined,
  EditOutlined,
  FolderAddOutlined,
  LinkOutlined,
  ReloadOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import {
  Button,
  Form,
  Input,
  message,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tree,
  Upload,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { DataNode } from 'antd/es/tree';
import React, { useEffect, useState } from 'react';

import {
  createFileCategory,
  deleteFileCategory,
  deleteFile,
  type FileCategory,
  type FileObject,
  type FileListParams,
  importFileURL,
  listFileCategories,
  listFiles,
  type PageMeta,
  renameFile,
  updateFileCategory,
  uploadFile,
} from '@/services/admin';

const formatDate = (value: string) => new Date(value).toLocaleString();

type ImportURLForm = {
  name?: string;
  url: string;
  category_id?: number;
};

type RenameFileForm = {
  name: string;
};

type CategoryForm = {
  name: string;
  parent_id?: number;
};

type CategoryOption = {
  id: number;
  parent_id: number;
  name: string;
  label: string;
};

const Files: React.FC = () => {
  const access = useAccess();
  const [files, setFiles] = useState<FileObject[]>([]);
  const [categories, setCategories] = useState<FileCategory[]>([]);
  const [selectedCategoryID, setSelectedCategoryID] = useState(0);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);
  const [categoryLoading, setCategoryLoading] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [importing, setImporting] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [renaming, setRenaming] = useState(false);
  const [categoryOpen, setCategoryOpen] = useState(false);
  const [savingCategory, setSavingCategory] = useState(false);
  const [deletingID, setDeletingID] = useState<number>();
  const [deletingCategoryID, setDeletingCategoryID] = useState<number>();
  const [selectedFile, setSelectedFile] = useState<FileObject>();
  const [editingCategory, setEditingCategory] = useState<CategoryOption>();
  const [importForm] = Form.useForm<ImportURLForm>();
  const [renameForm] = Form.useForm<RenameFileForm>();
  const [categoryForm] = Form.useForm<CategoryForm>();

  const loadCategories = async () => {
    setCategoryLoading(true);
    try {
      setCategories(await listFileCategories());
    } finally {
      setCategoryLoading(false);
    }
  };

  const loadData = async (params: FileListParams = {}) => {
    setLoading(true);
    try {
      const categoryID = params.category_id ?? selectedCategoryID;
      const response = await listFiles({
        ...params,
        category_id: categoryID > 0 ? categoryID : undefined,
      });
      setFiles(response.data);
      setPage(response.page);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadCategories();
    void loadData();
  }, []);

  const categoryOptions = flattenCategories(categories);
  const selectedCategory = categoryOptions.find(
    (category) => category.id === selectedCategoryID,
  );
  const categoryNameByID = new Map(
    categoryOptions.map((category) => [category.id, category.name]),
  );

  const currentPageParams = (): FileListParams => ({
    page: page?.page,
    page_size: page?.page_size,
    category_id: selectedCategoryID,
  });

  const selectCategory = async (keys: React.Key[]) => {
    const nextID = Number(keys[0] ?? 0);
    setSelectedCategoryID(nextID);
    await loadData({
      page: 1,
      page_size: page?.page_size,
      category_id: nextID,
    });
  };

  const submitImport = async () => {
    const values = await importForm.validateFields();
    setImporting(true);
    try {
      await importFileURL({
        name: values.name?.trim() || undefined,
        url: values.url.trim(),
        category_id: values.category_id,
      });
      message.success('URL已导入');
      setImportOpen(false);
      importForm.resetFields();
      await loadData(currentPageParams());
    } finally {
      setImporting(false);
    }
  };

  const openImport = () => {
    importForm.setFieldsValue({
      category_id: selectedCategoryID > 0 ? selectedCategoryID : undefined,
    });
    setImportOpen(true);
  };

  const closeRename = () => {
    setRenameOpen(false);
    setSelectedFile(undefined);
    renameForm.resetFields();
  };

  const openRename = (file: FileObject) => {
    setSelectedFile(file);
    renameForm.setFieldsValue({ name: file.name });
    setRenameOpen(true);
  };

  const submitRename = async () => {
    if (!selectedFile) {
      return;
    }
    const values = await renameForm.validateFields();
    setRenaming(true);
    try {
      await renameFile(selectedFile.id, values.name.trim());
      message.success('文件已重命名');
      closeRename();
      await loadData(currentPageParams());
    } finally {
      setRenaming(false);
    }
  };

  const removeFile = async (file: FileObject) => {
    setDeletingID(file.id);
    try {
      await deleteFile(file.id);
      message.success('文件已删除');
      await loadData(currentPageParams());
    } finally {
      setDeletingID(undefined);
    }
  };

  const openCreateCategory = () => {
    setEditingCategory(undefined);
    categoryForm.setFieldsValue({
      name: '',
      parent_id: selectedCategoryID > 0 ? selectedCategoryID : undefined,
    });
    setCategoryOpen(true);
  };

  const openEditCategory = () => {
    if (!selectedCategory) {
      return;
    }
    setEditingCategory(selectedCategory);
    categoryForm.setFieldsValue({
      name: selectedCategory.name,
      parent_id:
        selectedCategory.parent_id > 0 ? selectedCategory.parent_id : undefined,
    });
    setCategoryOpen(true);
  };

  const closeCategory = () => {
    setCategoryOpen(false);
    setEditingCategory(undefined);
    categoryForm.resetFields();
  };

  const submitCategory = async () => {
    const values = await categoryForm.validateFields();
    const body = {
      name: values.name.trim(),
      parent_id: values.parent_id ?? 0,
    };
    setSavingCategory(true);
    try {
      if (editingCategory) {
        await updateFileCategory(editingCategory.id, body);
        message.success('分类已更新');
      } else {
        await createFileCategory(body);
        message.success('分类已创建');
      }
      closeCategory();
      await loadCategories();
    } finally {
      setSavingCategory(false);
    }
  };

  const removeCategory = async (category: CategoryOption) => {
    setDeletingCategoryID(category.id);
    try {
      await deleteFileCategory(category.id);
      message.success('分类已删除');
      setSelectedCategoryID(0);
      await loadCategories();
      await loadData({ page: 1, page_size: page?.page_size, category_id: 0 });
    } finally {
      setDeletingCategoryID(undefined);
    }
  };

  const columns: ColumnsType<FileObject> = [
    {
      title: '文件名',
      dataIndex: 'name',
      render: (name: string, record) => (
        <a href={record.url} target="_blank" rel="noreferrer">
          {name}
        </a>
      ),
    },
    {
      title: '分类',
      dataIndex: 'category_id',
      width: 140,
      render: (categoryID: number) =>
        categoryID > 0 ? categoryNameByID.get(categoryID) ?? `#${categoryID}` : '未分类',
    },
    { title: '类型', dataIndex: 'content_type' },
    {
      title: '大小',
      dataIndex: 'size',
      render: (size: number, record) =>
        record.content_type === 'external/url' ? '外部链接' : `${size} B`,
    },
    { title: '上传时间', dataIndex: 'created_at', render: formatDate },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      render: (_, record) => {
        const actions: React.ReactNode[] = [];
        if (access.canFileUpdate) {
          actions.push(
            <Button
              key="rename"
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => openRename(record)}
            >
              重命名
            </Button>,
          );
        }
        if (access.canFileDelete) {
          actions.push(
            <Popconfirm
              key="delete"
              title="删除文件"
              description="确认删除这条文件记录？"
              okText="删除"
              okButtonProps={{ danger: true }}
              cancelText="取消"
              onConfirm={() => removeFile(record)}
            >
              <Button
                danger
                type="link"
                size="small"
                icon={<DeleteOutlined />}
                loading={deletingID === record.id}
              >
                删除
              </Button>
            </Popconfirm>,
          );
        }
        return actions.length > 0 ? <Space size="small">{actions}</Space> : '-';
      },
    },
  ];

  return (
    <PageContainer title="文件上传">
      <div
        style={{
          display: 'grid',
          gap: 16,
          gridTemplateColumns: '260px minmax(0, 1fr)',
          alignItems: 'start',
        }}
      >
        <div>
          <Space style={{ marginBottom: 12 }} wrap>
            {access.canFileCategoryCreate ? (
              <Button icon={<FolderAddOutlined />} onClick={openCreateCategory}>
                新增分类
              </Button>
            ) : null}
            {selectedCategory && access.canFileCategoryUpdate ? (
              <Button icon={<EditOutlined />} onClick={openEditCategory}>
                编辑
              </Button>
            ) : null}
            {selectedCategory && access.canFileCategoryDelete ? (
              <Popconfirm
                title="删除分类"
                description="确认删除当前分类？"
                okText="删除"
                okButtonProps={{ danger: true }}
                cancelText="取消"
                onConfirm={() => removeCategory(selectedCategory)}
              >
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  loading={deletingCategoryID === selectedCategory.id}
                >
                  删除
                </Button>
              </Popconfirm>
            ) : null}
          </Space>
          <Tree
            blockNode
            defaultExpandAll
            disabled={categoryLoading}
            selectedKeys={[String(selectedCategoryID)]}
            treeData={categoryTreeData(categories)}
            onSelect={(keys) => void selectCategory(keys)}
          />
        </div>
        <Table<FileObject>
          rowKey="id"
          columns={columns}
          dataSource={files}
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
              category_id: selectedCategoryID,
            })
          }
          title={() => (
            <Space>
              {access.canFileUpload ? (
                <Upload
                  maxCount={1}
                  showUploadList={false}
                  beforeUpload={async (file) => {
                    await uploadFile(
                      file,
                      selectedCategoryID > 0 ? selectedCategoryID : undefined,
                    );
                    message.success('文件已上传');
                    await loadData(currentPageParams());
                    return Upload.LIST_IGNORE;
                  }}
                >
                  <Button type="primary" icon={<UploadOutlined />}>
                    上传文件
                  </Button>
                </Upload>
              ) : null}
              {access.canFileUpload ? (
                <Button icon={<LinkOutlined />} onClick={openImport}>
                  导入URL
                </Button>
              ) : null}
              <Button
                icon={<ReloadOutlined />}
                onClick={() => void loadData(currentPageParams())}
              >
                刷新
              </Button>
            </Space>
          )}
        />
      </div>
      <Modal
        title="导入URL"
        open={importOpen}
        confirmLoading={importing}
        onCancel={() => {
          setImportOpen(false);
          importForm.resetFields();
        }}
        onOk={() => void submitImport()}
      >
        <Form<ImportURLForm> form={importForm} layout="vertical">
          <Form.Item label="名称" name="name" rules={[{ max: 180 }]}>
            <Input maxLength={180} />
          </Form.Item>
          <Form.Item label="分类" name="category_id">
            <Select
              allowClear
              options={categoryOptions.map((category) => ({
                label: category.label,
                value: category.id,
              }))}
            />
          </Form.Item>
          <Form.Item
            label="URL"
            name="url"
            rules={[
              { required: true, message: '请输入URL' },
              { type: 'url', message: '请输入有效URL' },
              { max: 2048, message: 'URL不能超过2048个字符' },
            ]}
          >
            <Input maxLength={2048} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title={editingCategory ? '编辑分类' : '新增分类'}
        open={categoryOpen}
        confirmLoading={savingCategory}
        onCancel={closeCategory}
        onOk={() => void submitCategory()}
      >
        <Form<CategoryForm> form={categoryForm} layout="vertical">
          <Form.Item
            label="名称"
            name="name"
            rules={[
              { required: true, message: '请输入分类名称' },
              { whitespace: true, message: '分类名称不能只包含空白' },
              { max: 80, message: '分类名称不能超过80个字符' },
            ]}
          >
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item label="父级分类" name="parent_id">
            <Select
              allowClear
              options={categoryOptions
                .filter((category) => category.id !== editingCategory?.id)
                .map((category) => ({
                  label: category.label,
                  value: category.id,
                }))}
            />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title="重命名文件"
        open={renameOpen}
        confirmLoading={renaming}
        onCancel={closeRename}
        onOk={() => void submitRename()}
      >
        <Form<RenameFileForm> form={renameForm} layout="vertical">
          <Form.Item
            label="文件名"
            name="name"
            rules={[
              { required: true, message: '请输入文件名' },
              { whitespace: true, message: '文件名不能只包含空白' },
              { max: 180, message: '文件名不能超过180个字符' },
            ]}
          >
            <Input maxLength={180} />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

const categoryTreeData = (categories: FileCategory[]): DataNode[] => [
  {
    key: '0',
    title: '全部文件',
    children: categories.map(categoryNode),
  },
];

const categoryNode = (category: FileCategory): DataNode => ({
  key: String(category.id),
  title: category.name,
  children: category.children.map(categoryNode),
});

const flattenCategories = (
  categories: FileCategory[],
  depth = 0,
): CategoryOption[] =>
  categories.flatMap((category) => [
    {
      id: category.id,
      parent_id: category.parent_id,
      name: category.name,
      label: `${'  '.repeat(depth)}${category.name}`,
    },
    ...flattenCategories(category.children, depth + 1),
  ]);

export default Files;
