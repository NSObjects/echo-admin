import React from 'react'
import {  PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import {Button, message} from 'antd';
import { useRef } from 'react';
import {deleteUsersId, getUsers} from "@/services/echo-admin/yonghu";
import UserEditor from "@/pages/system/user/components/editor";
import { useState } from 'react';
import MenuEditor from "@/pages/system/menu/components/editor";

const User: React.FC = () => {
  const [showModal, setShowModal] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.user>();
  const actionRef = useRef<ActionType>();
  const columns: ProColumns<API.user>[] = [
    {
      dataIndex: 'index',
      valueType: 'indexBorder',
      width: 48,
    },
    {
      title: '昵称',
      dataIndex: 'name',
      hideInSearch: true,
    },
    {
      title: '手机号码',
      dataIndex: 'phone',
    },
    {
      title: '关键词搜索',
      dataIndex: 'key',
      onFilter: true,
      hideInTable: true,
    },
    {
      disable: true,
      title: '状态',
      dataIndex: 'status',
      // filters: true,
       onFilter: true,
      // ellipsis: true,
      valueType: 'select',
      valueEnum: {
        all: { text: '未知' ,  disabled: true,},
        1: {
          text: '启用',
          status: 'Error',
        },
        2: {
          text: '停用',
          status: 'Success',

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
      valueType: 'dateTimeRange',
      hideInTable: true,
      search: {
        transform: (value) => {
          return {
            create_start: value[0],
            create_end: value[1],
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
            setShowModal(true)
            setCurrentRow(record)
          }}
        >
          编辑
        </a>,
        <a
          key="delete"
          onClick={() => {
            deleteUsersId({id:record.id ?? 0}).then((res)=>{
              if (res.code === 0) {
                message.success('删除成功').then(() => {
                  actionRef.current?.reload()})
              } else {
                message.error('删除失败').then(r => console.log(r));
              }
            })
          }}
        >
          删除
        </a>,
      ],
    },
  ];

  return<>
    <ProTable<API.user,API.getUsersParams>
      columns={columns}

      actionRef={actionRef}
      cardBordered
      request={async (p, sort, filter) => {
        console.log(sort, p);
        const msg = await getUsers({
          count: p.pageSize,
          page: p.current,
          create_end: p.create_end,
          create_start: p.create_start,
          key: p.key,
          phone: p.phone,
          status: p.status,
        })
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
        collapsed: false,
      }}
      options={{
        setting: {
          listsHeight: 400,
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
            setShowModal(true)
          }}
          type="primary"
        >
          新建
        </Button>,
      ]}
    />
    <UserEditor modalVisit={showModal} setModalVisit={(modalVisit: boolean)=>{
      setShowModal(modalVisit)
      if (!modalVisit) {
        setCurrentRow(undefined)
      }
    }
    }  values={currentRow || {}} reload={()=>{actionRef.current?.reload()}}></UserEditor>
  </>
}

export default User
