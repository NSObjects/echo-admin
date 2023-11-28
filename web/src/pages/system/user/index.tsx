import React from 'react'
import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import { Button  } from 'antd';
import { useRef } from 'react';
import {getUsers} from "@/services/echo-admin/yonghu";

const User: React.FC = () => {

  const columns: ProColumns<API.user>[] = [
    {
      dataIndex: 'index',
      valueType: 'indexBorder',
      width: 48,
    },
    {
      title: '昵称',
      dataIndex: 'name',
      ellipsis: true,
      tip: '标题过长会自动收缩',
    },
    {
      disable: true,
      title: '状态',
      dataIndex: 'status',
      filters: true,
      onFilter: true,
      ellipsis: true,
      valueType: 'select',
      valueEnum: {
        all: { text: '未知' },
        1: {
          text: '启用',
          status: 'Error',
        },
        2: {
          text: '停用',
          status: 'Success',
          disabled: true,
        },
      },
    },

    {
      title: '创建时间',
      key: 'showTime',
      dataIndex: 'created_at',
      valueType: 'date',
      sorter: true,
      hideInSearch: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      valueType: 'dateRange',
      hideInTable: true,
      search: {
        transform: (value) => {
          return {
            startTime: value[0],
            endTime: value[1],
          };
        },
      },
    },
    {
      title: '操作',
      valueType: 'option',
      key: 'option',
      render: (text, record) => [
        <a
          key="editable"
          onClick={() => {
            console.log(record.id)
          }}
        >
          编辑
        </a>,
        <a
          key="delete"
          onClick={() => {
            console.log(record.id)
          }}
        >
          删除
        </a>,
      ],
    },
  ];

  function getStringValue(filter: { [key: string]: any }, propertyName: string): string {
    return typeof filter[propertyName] === 'string' ? filter[propertyName] : '';
  }

  const actionRef = useRef<ActionType>();
  return<>
    <ProTable<API.user>
      columns={columns}
      actionRef={actionRef}
      cardBordered
      // request={getUsers}
      request={async (p, sort, filter) => {
        console.log(sort, filter);
        const msg = await getUsers({name:getStringValue(filter,"name") ,
          phone: getStringValue(filter,"phone"),
          page: p.current , count: p.pageSize})
        return  {
            data: msg.data.list,
            total: msg.data.total,
            success: true,
        };
      }}
      editable={{
        type: 'multiple',
      }}
      columnsState={{
        persistenceKey: 'pro-table-singe-demos',
        persistenceType: 'localStorage',
        onChange(value) {
          console.log('value: ', value);
        },
      }}
      rowKey="id"
      search={{
        labelWidth: 'auto',
      }}
      options={{
        setting: {
          listsHeight: 400,
        },
      }}
      form={{
        // 由于配置了 transform，提交的参与与定义的不同这里需要转化一下
        syncToUrl: (values, type) => {
          if (type === 'get') {
            return {
              ...values,
              created_at: [values.startTime, values.endTime],
            };
          }
          return values;
        },
      }}
      pagination={{
        pageSize: 5,
        onChange: (page) => console.log(page),
      }}
      dateFormatter="string"
      headerTitle="用户列表"
      toolBarRender={() => [
        <Button
          key="button"
          icon={<PlusOutlined />}
          onClick={() => {
            actionRef.current?.reload();
          }}
          type="primary"
        >
          新建
        </Button>,
      ]}
    />
  </>
}

export default User