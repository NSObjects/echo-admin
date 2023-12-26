import React from 'react';
import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import { Button, message } from 'antd';
import { useRef, useEffect } from 'react';
import { deleteUsersId, getUsers } from '@/services/echo-admin/yonghu';
import UserEditor from '@/pages/system/user/components/editor';
import { useState } from 'react';
import Department from '@/pages/system/user/components/deparment';
import { Card, Flex } from 'antd';
import { DataNode } from 'antd/es/tree';
import { getDepartments } from '@/services/echo-admin/bumen';

const departmentItemTree = (menus: API.department[]): DataNode[] =>
  menus.map(({ name, id, children }) => ({
    title: name,
    key: id ?? 0,
    children: children && children.length > 0 ? departmentItemTree(children) : [],
  }));

const User: React.FC = () => {
  // let deparment: DataNode[] = [];

  const [showModal, setShowModal] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.user>();
  const actionRef = useRef<ActionType>();
  const [selectDept, setSelectDept] = useState<number>(0);
  const [department, setDepartment] = useState<DataNode[]>([]);
  useEffect(() => {
    getDepartments({ page: 1, count: 1000 }).then((res) => {
      setDepartment(departmentItemTree(res.data.list ?? []));
    });
  }, []);
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
      onFilter: true,
      valueType: 'select',
      valueEnum: {
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
            setShowModal(true);
            setCurrentRow(record);
          }}
        >
          编辑
        </a>,
        <a
          key="delete"
          onClick={() => {
            deleteUsersId({ id: record.id ?? 0 }).then((res) => {
              if (res.code === 0) {
                actionRef.current?.reload();
                message.success('删除成功').then(() => {});
              } else {
                message.error('删除失败').then((r) => console.log(r));
              }
            });
          }}
        >
          删除
        </a>,
      ],
    },
  ];

  return (
    <>
      <Flex gap={'middle'}>
        <Card style={{ width: '300px' }}>
          <Department
            onSelect={async (value: any) => {
              console.log('value: ', value);
              setSelectDept(value);
              actionRef.current?.reload();
            }}
            department={department}
          ></Department>
        </Card>
        <ProTable<API.user, API.getUsersParams>
          columns={columns}
          actionRef={actionRef}
          onReset={() => {
            setSelectDept(0);
            actionRef.current?.reload();
          }}
          cardBordered
          request={async (p, sort) => {
            console.log(sort, p);
            const msg = await getUsers({
              count: p.pageSize,
              page: p.current,
              create_end: p.create_end,
              create_start: p.create_start,
              key: p.key,
              phone: p.phone,
              status: p.status,
              department_id: selectDept,
            });
            return {
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
                setShowModal(true);
              }}
              type="primary"
            >
              新建
            </Button>,
          ]}
        />
      </Flex>
      <UserEditor
        depts={department}
        modalVisit={showModal}
        setModalVisit={(modalVisit: boolean) => {
          setShowModal(modalVisit);
          if (!modalVisit) {
            setCurrentRow(undefined);
          }
        }}
        values={currentRow || {}}
        reload={() => {
          actionRef.current?.reload();
        }}
      ></UserEditor>
    </>
  );
};

export default User;
