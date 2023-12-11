import React, {useRef, useState} from 'react'
import {ActionType, ProColumns, ProTable} from "@ant-design/pro-components";
import {Button} from "antd";
import {PlusOutlined} from "@ant-design/icons";
import {getApiMenus} from "@/services/echo-admin/caidan";
import MenuEditor from "@/pages/system/menu/components/editor";

const Menu: React.FC = () => {
  const [showModal, setShowModal] = useState<boolean>(false);
  const columns: ProColumns<API.menuData>[] = [
    {
      title: '菜单名称',
      dataIndex: 'name',
      ellipsis: true,
    },
    {
      title: '路由路径',
      dataIndex: 'path',
      ellipsis: true,
    },
    {
      title: '组件路径',
      dataIndex: 'component',
      ellipsis: true,
    },
    {
      title: 'API接口',
      dataIndex: 'api',
      ellipsis: true,
    },
    {
      title: '排序',
      dataIndex: 'sort',
      ellipsis: true,
    },
    {
      disable: true,
      title: '类型',
      dataIndex: 'type',
      filters: true,
      onFilter: true,
      ellipsis: true,
      valueType: 'select',
      valueEnum: {
        all: { text: '未知' },
        1: {
          text: '目录',
          status: 'Error',
        },
        2: {
          text: '菜单',
          status: 'Default',
          disabled: true,
        },
        3: {
          text: '按钮',
          status: 'Processing',
          disabled: true,
        },
      },
    },

    {
      title: '显示状态',
      dataIndex: 'status',
      ellipsis: true,
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

  // const actionRef = useRef<ActionType>();
  return<>
    <ProTable<API.menuData>
      columns={columns}
      // actionRef={actionRef}
      cardBordered
      request={async (p, sort, filter) => {
        const msg = await getApiMenus()
        return  {
          data: msg.data.list ?? [],
          total: msg.data.list?.length ?? 0,
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
      // form={{
      //   // 由于配置了 transform，提交的参与与定义的不同这里需要转化一下
      //   syncToUrl: (values, type) => {
      //     if (type === 'get') {
      //       return {
      //         ...values,
      //         created_at: [values.startTime, values.endTime],
      //       };
      //     }
      //     return values;
      //   },
      // }}
      dateFormatter="string"
      headerTitle="菜单列表"
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
    <MenuEditor modalVisit={showModal} setModalVisit={(modalVisit: boolean)=>
      setShowModal(modalVisit)}></MenuEditor>
  </>
}

export default Menu
