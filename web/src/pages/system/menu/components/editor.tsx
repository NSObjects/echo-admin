import React, { useRef, useEffect, useState } from 'react';
import { message, Space } from 'antd';
import {
  ProForm,
  ModalForm,
  ProFormText,
  ProFormTreeSelect,
  ProFormRadio,
  ProFormDigit,
  ProFormSelect,
  ProFormList,
} from '@ant-design/pro-components';
import { Modal } from 'antd';
import { ProFormInstance } from '@ant-design/pro-form/lib';
import { postMenus, putMenusId } from '@/services/echo-admin/caidan';
import IconSelector from './iconSelector/index';
import * as AntdIcons from '@ant-design/icons';
import { type EnhancedMenuItem } from '@/pages/system/menu';
import { getApi } from '@/services/echo-admin/zhanghao';
import { LabeledValue } from 'antd/es/select';

const options = [
  {
    value: 'GET',
    label: 'GET',
  },
  {
    value: 'PUT',
    label: 'PUT',
  },
  {
    value: 'POST',
    label: 'POST',
  },
  {
    value: 'DELETE',
    label: 'DELETE',
  },
];

type Props = {
  modalVisit: boolean;
  setModalVisit: (modalVisit: boolean) => void;
  menu: EnhancedMenuItem[];
  menuValue: Partial<API.menu>;
  reload(): void;
};

const allIcons: Record<string, any> = AntdIcons;

function createIcon(icon: string | any): React.ReactNode | string {
  if (typeof icon === 'object') {
    return icon;
  }
  const ele = allIcons[icon];
  if (ele) {
    return React.createElement(allIcons[icon]);
  }
  return '';
}

const typeOption = [
  {
    label: '目录',
    value: 1,
  },
  {
    label: '菜单',
    value: 2,
  },
  {
    label: '按钮',
    value: 3,
  },
];

const MenuEditor: React.FC<Props> = (props) => {
  const { modalVisit, setModalVisit, menu, menuValue } = props;
  const restFormRef = useRef<ProFormInstance>();
  const [menuIconName, setMenuIconName] = useState<string>();
  const [menuTypeId, setMenuTypeId] = useState<number>(menuValue.type || 1);
  const [iconSelectorOpen, setIconSelectorOpen] = useState<boolean>(false);
  const [title, setTitle] = useState<string>('新建菜单'); // 初始值设置为 '新建菜单'

  useEffect(() => {
    restFormRef.current?.resetFields();
    setTitle(menuValue?.id ? '编辑菜单' : '新建菜单');

    if (menuValue.type === 2 && menuValue.id === undefined && menuValue.pid) {
      restFormRef.current?.setFieldsValue({
        pid: menuValue.pid,
        type: 3,
      });
      console.log('menuValue.type === 2 && menuValue.id !== undefined', menuValue.id);
      setMenuTypeId(3);
    } else {
      restFormRef.current?.setFieldsValue({
        ...menuValue,
        sort: menuValue.sort === 0 ? undefined : menuValue.sort,
        // 如果有其他需要设置的字段，请在这里添加
      });
      setMenuTypeId(menuValue.type === undefined ? 1 : menuValue.type);
    }
  }, [menuValue]);
  return (
    <div>
      <ModalForm
        title={title}
        formRef={restFormRef}
        open={modalVisit}
        onFinish={async (fieldsValue: any) => {
          console.log(fieldsValue);
          const body = {
            apis: fieldsValue['apis'],
            cache: fieldsValue['cache'],
            component: fieldsValue['component'],
            fixed: fieldsValue['fixed'],
            hidden: fieldsValue['hidden'],
            icon: fieldsValue['icon'],
            identify: fieldsValue['identify'],
            layout: fieldsValue['layout'],
            link: fieldsValue['link'],
            name: fieldsValue['name'],
            path: fieldsValue['path'],
            pid: fieldsValue['pid'],
            redirect: fieldsValue['redirect'],
            remark: fieldsValue['remark'],
            role: fieldsValue['role'],
            sort: fieldsValue['sort'],
            status: fieldsValue['status'],
            type: fieldsValue['type'] === 0 ? 1 : fieldsValue['type'],
          };

          let res: API.success;
          if (props.menuValue?.id) {
            res = await putMenusId({ id: props.menuValue.id }, body);
          } else {
            res = await postMenus(body);
          }
          console.log('Onfinish');
          if (res.code === 0) {
            props.reload();
            message.success('提交成功');
            return true;
          } else {
            message.error('提交失败:');
            return false;
          }
        }}
        onOpenChange={(visible: boolean) => {
          // restFormRef.current?.resetFields();
          //
          // if (!visible) {
          //   setMenuTypeId(1);
          // }
          setModalVisit(visible);
        }}
      >
        <ProFormTreeSelect
          name="pid"
          request={async () => {
            return menu;
          }}
          label="上级菜单"
        />
        <ProFormRadio.Group
          name="type"
          options={typeOption}
          label="菜单类型"
          placeholder="请输入菜单类型"
          rules={[
            {
              required: true,
              message: '请输入菜单类型',
            },
          ]}
          fieldProps={{
            // defaultValue: { menuTypeId },
            onChange: (e) => {
              setMenuTypeId(e.target.value);
            },
          }}
        />
        <ProForm.Group>
          <ProFormText
            name="name"
            width="md"
            label="菜单名称"
            placeholder="请填写菜单名称"
            rules={[{ required: true, message: '菜单名称不能为空' }]}
          />
          <ProFormSelect
            name="icon"
            label="菜单图标"
            valueEnum={{}}
            width="md"
            hidden={menuTypeId === 3}
            addonBefore={createIcon(menuIconName)}
            fieldProps={{
              onClick: () => {
                setIconSelectorOpen(true);
              },
            }}
            placeholder="请输入菜单图标"
            rules={[
              {
                required: false,
                message: '请输入菜单图标',
              },
            ]}
          />
          <ProFormText
            name="path"
            label="路由路径"
            width="md"
            hidden={menuTypeId === 3}
            placeholder="路由中的path值"
            rules={[{ required: true, message: '路由地址不能为空' }]}
          />

          <ProFormText
            name="component"
            hidden={menuTypeId === 3}
            label="组件路径"
            width="md"
            placeholder="请输入组件路径"
          />
          <ProFormText
            hidden={menuTypeId === 3}
            name="url"
            label="链接地址"
            width="md"
            placeholder="请输入组件路径"
          />
          <ProFormDigit
            name="sort"
            label="菜单排序"
            width="md"
            min={1}
            max={10}
            fieldProps={{ precision: 0 }}
          />
        </ProForm.Group>
        <ProFormList name="apis">
          {(meta, index, action) => {
            return (
              <ProForm.Group>
                <ProFormSelect
                  showSearch
                  // name="api"
                  fieldProps={{
                    labelInValue: true,
                    placeholder: 'API快捷输入',
                  }}
                  onChange={(value: LabeledValue) => {
                    console.log('onChange', value);
                    let values = value?.value.toString().split('@');
                    action.setCurrentRowData({
                      name: value?.label,
                      method: values[1],
                      url: values[0],
                    });
                  }}
                  request={async ({ keyWords }) => {
                    let resp = await getApi();
                    let apis = resp.data.list;
                    if (keyWords) {
                      apis = apis.filter((item: API.api) => {
                        return item.name.includes(keyWords) || item.path.includes(keyWords);
                      });
                    }
                    return apis.map((item: API.api) => {
                      return { label: item.name, value: item.path + '@' + item.method };
                    });
                  }}
                />
                <Space.Compact>
                  <ProFormSelect style={{ width: '120px' }} options={options} name="method" />
                  <ProFormText placeholder="接口名称" name="name" />
                  <ProFormText placeholder="后端api路径" name="url" />
                </Space.Compact>
              </ProForm.Group>
            );
          }}
        </ProFormList>
        <Modal
          width={600}
          open={iconSelectorOpen}
          onCancel={() => {
            setIconSelectorOpen(false);
          }}
          footer={null}
        >
          <IconSelector
            onSelect={(name: string) => {
              restFormRef.current?.setFieldsValue({ icon: name });
              setMenuIconName(name);
              setIconSelectorOpen(false);
            }}
          />
        </Modal>
      </ModalForm>
    </div>
  );
};

export default MenuEditor;
