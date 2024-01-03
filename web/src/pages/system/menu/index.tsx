import React, { useState, useRef, useContext } from 'react';
import { ProColumns, ProTable, ActionType, ProProvider } from '@ant-design/pro-components';
import { Button, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { deleteMenusId, getMenus } from '@/services/echo-admin/caidan';
import MenuEditor from '@/pages/system/menu/components/editor';

export interface EnhancedMenuItem {
  title: string;
  value: number;
  children?: EnhancedMenuItem[];
}

const fixMenuItemIcon = (menus: API.menu[]): EnhancedMenuItem[] => {
  return menus.map((item) => {
    const { name, id, children } = item;
    const newItem: EnhancedMenuItem = {
      title: name ?? '',
      value: id ?? 0,
    };
    if (children && children.length > 0) {
      newItem.children = fixMenuItemIcon(children);
    }
    return newItem;
  });
};

const Menu: React.FC = () => {
  const [showModal, setShowModal] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.menu>();
  const [menu, setMenu] = useState<EnhancedMenuItem[]>([]);
  const actionRef = useRef<ActionType>();

  const columns: ProColumns<API.menu, 'api'>[] = [
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
      title: '类型',
      dataIndex: 'type',
      filters: true,
      onFilter: true,
      ellipsis: true,
      valueType: 'select',
      valueEnum: {
        1: {
          text: '目录',
          status: 'Error',
        },
        2: {
          text: '菜单',
          status: 'Default',
        },
        3: {
          text: '按钮',
          status: 'Processing',
        },
      },
    },

    {
      title: '操作',
      valueType: 'option',
      key: 'option',
      render: (text, record) => [
        <a
          hidden={record.type === 3}
          key="add"
          onClick={() => {
            if (record.type === 3) {
              message.error('按钮不允许添加子菜单');
              return;
            }
            setCurrentRow({
              pid: record.id,
              type: record.type,
            });
            setShowModal(true);
          }}
        >
          新增
        </a>,
        <a
          key="editable"
          onClick={() => {
            setShowModal(true);
            setCurrentRow(record);
          }}
        >
          修改
        </a>,
        <a
          key="delete"
          onClick={() => {
            deleteMenusId({ id: record.id ?? 0 }).then((res) => {
              if (res.code === 0) {
                message.success('删除成功');
                actionRef.current?.reload();
              } else {
                message.error('删除失败');
              }
            });
          }}
        >
          删除
        </a>,
      ],
    },
  ];
  const values = useContext(ProProvider);
  return (
    <>
      <ProProvider.Provider
        value={{
          ...values,
          valueTypeMap: {
            api: {
              render: (text) => {
                return <>{text.url}</>;
              },
            },
          },
        }}
      >
        <ProTable<API.menu, Record<string, any>, 'api'>
          columns={columns}
          actionRef={actionRef}
          cardBordered
          request={async () => {
            const msg = await getMenus();
            setMenu(fixMenuItemIcon(msg.data.list ?? []));
            return {
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
            collapsed: false,
          }}
          options={{
            setting: {
              listsHeight: 400,
            },
          }}
          dateFormatter="string"
          headerTitle="菜单列表"
          toolBarRender={() => [
            <Button
              key="button"
              icon={<PlusOutlined />}
              onClick={() => {
                setCurrentRow({ type: 1 });
                setShowModal(true);
              }}
              type="primary"
            >
              新建
            </Button>,
          ]}
        />
        <MenuEditor
          modalVisit={showModal}
          setModalVisit={(modalVisit: boolean) => {
            setShowModal(modalVisit);
            if (!modalVisit) {
              setCurrentRow(undefined);
            }
          }}
          menuValue={currentRow || {}}
          menu={menu}
          reload={() => {
            actionRef.current?.reload();
          }}
        ></MenuEditor>
      </ProProvider.Provider>
    </>
  );
};

export default Menu;
