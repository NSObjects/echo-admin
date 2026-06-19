import { ReloadOutlined, UploadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Space, Table, Upload, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';

import {
  type FileObject,
  type ListParams,
  type PageMeta,
  listFiles,
  uploadFile,
} from '@/services/admin';

const formatDate = (value: string) => new Date(value).toLocaleString();

const Files: React.FC = () => {
  const [files, setFiles] = useState<FileObject[]>([]);
  const [page, setPage] = useState<PageMeta>();
  const [loading, setLoading] = useState(false);

  const loadData = async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const response = await listFiles(params);
      setFiles(response.data);
      setPage(response.page);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

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
    { title: '类型', dataIndex: 'content_type' },
    { title: '大小', dataIndex: 'size', render: (size: number) => `${size} B` },
    { title: '上传时间', dataIndex: 'created_at', render: formatDate },
  ];

  return (
    <PageContainer title="文件上传">
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
          })
        }
        title={() => (
          <Space>
            <Upload
              maxCount={1}
              showUploadList={false}
              beforeUpload={async (file) => {
                await uploadFile(file);
                message.success('文件已上传');
                await loadData({ page: page?.page, page_size: page?.page_size });
                return Upload.LIST_IGNORE;
              }}
            >
              <Button type="primary" icon={<UploadOutlined />}>
                上传文件
              </Button>
            </Upload>
            <Button icon={<ReloadOutlined />} onClick={() => void loadData()}>
              刷新
            </Button>
          </Space>
        )}
      />
    </PageContainer>
  );
};

export default Files;
