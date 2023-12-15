import React, {useRef, useState} from 'react'
import {ActionType, ProColumns, ProTable} from "@ant-design/pro-components";

import {Button} from "antd";
import {PlusOutlined} from "@ant-design/icons";
import UserEditor from "@/pages/system/user/components/editor";
import {getRoles} from "@/services/echo-admin/jiaose";


const Role: React.FC = () => {
  const [showModal, setShowModal] = useState<boolean>(false);
  const columns: ProColumns<API.role>[] = [
    {
      dataIndex: 'index',
      valueType: 'indexBorder',
      width: 48,
    },
    {
      title: '角色名称',
      dataIndex: 'name',
    },
    {
      title: '排序',
      dataIndex: 'sort',
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
      title: '角色描述',
      dataIndex: 'name',
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
            console.log(record.name)
          }}
        >
          编辑
        </a>,
        <a
          key="delete"
          onClick={() => {
            console.log(record.name)
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
    <ProTable<API.role>
      columns={columns}
      actionRef={actionRef}
      cardBordered
      // request={getUsers}
      request={async (p, sort, filter) => {
        console.log(sort, filter);
        const msg = await getRoles({
          name: getStringValue(filter,"name"),
          page: p.current , count: p.pageSize
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
            // setShowModal(true)
          }}
          type="primary"
        >
          新建
        </Button>,
      ]}
    />
    {/*<UserEditor modalVisit={showModal} setModalVisit={(modalVisit: boolean)=>*/}
    {/*  setShowModal(modalVisit)}></UserEditor>*/}
  </>
}

export default Role
