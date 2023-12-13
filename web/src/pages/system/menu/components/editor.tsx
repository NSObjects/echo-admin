import React, {useRef} from 'react'

import {  message } from 'antd';
import {
  ProForm,
  ModalForm,
  ProFormText,
  ProFormCascader,
  ProFormRadio,
  ProFormDigit,
  ProFormSelect

} from '@ant-design/pro-components';

import {ProFormInstance} from "@ant-design/pro-form/lib";
import {getApiMenus, postApiMenus} from "@/services/echo-admin/caidan";
import ms from "@umijs/utils/compiled/debug/ms";
import {getApiRoles} from "@/services/echo-admin/jiaose";

type Props = {
  modalVisit: boolean;
  setModalVisit: (modalVisit: boolean) => void;
}

const fixMenuItemIcon = (menus: API.menu[]): API.menu[] => {
  menus.forEach((item) => {
    const {name,id, children} = item
    item.label = name
    item.value = id
    // eslint-disable-next-line no-param-reassign,@typescript-eslint/no-unused-expressions
    children && children.length > 0 ? item.children = fixMenuItemIcon(children) : null
  });
  return menus;
};

const MenuEditor: React.FC<Props> = props => {
  const { modalVisit, setModalVisit } = props;
  const restFormRef = useRef<ProFormInstance>();
  return (
    <div>
      <ModalForm
        title="新建菜单"
        formRef={restFormRef}
        open={modalVisit}
        onFinish={async (fieldsValue: any) => {
          console.log(fieldsValue)
          const res = await postApiMenus({
            api: fieldsValue["api"],
            cache: fieldsValue["cache"],
            component: fieldsValue["component"],
            fixed: fieldsValue["fixed"],
            hidden: fieldsValue["hidden"],
            icon: fieldsValue["icon"],
            identify:  fieldsValue["identify"],
            layout: fieldsValue["layout"],
            link: fieldsValue["link"],
            name: fieldsValue["name"],
            path: fieldsValue["path"],
            pid:  fieldsValue["pid"],
            redirect: fieldsValue["redirect"],
            remark: fieldsValue["remark"],
            role: fieldsValue["role"],
            sort: fieldsValue["sort"],
            status: fieldsValue["status"],
            type: fieldsValue["type"]
          })
          if (res.code === 0) {
            message.success('提交成功');
            return true;
          } else {
            message.error('提交失败');
            return false;
          }
        }}
        onOpenChange={(visible:boolean)=>{
          if (!visible) {
            restFormRef.current?.resetFields();
          }
          setModalVisit(visible)
        }}
      >
        <ProFormCascader
          name="pid"
          request={async () => {
            const msg =  await getApiMenus()
            return fixMenuItemIcon(msg.data.list)
          }}
          label="上级菜单"
        />
        <ProFormRadio.Group
          label="菜单类型"
          name="type"
          initialValue="目录"
          options={[
                {
                  label: '目录',
                  value:1,
                },
                {
                  label: '菜单',
                  value:2,
                },
                {
                  label: '按钮',
                  value:3,
                }
          ]}
        />
        <ProForm.Group>
          <ProFormText
            name="name"
            width="md"
            label="菜单名称"
            placeholder="请填写菜单名称"
          />
          <ProFormText
            name="rule"
            width="md"
            label="接口规则"
            placeholder="后端api地址"
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormText
            name="path"
            label="路由路径"
            width="md"
            placeholder="路由中的path值"
          />
          <ProFormText
            name="redirect"
            label="重定向"
            width="md"
            placeholder="请输入路由重定向"
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormText
            name="icon"
            label="菜单图标"
            width="md"
            placeholder="请输入组件路径"
          />
          <ProFormText
            name="component"
            label="组件路径"
            width="md"
            placeholder="请输入组件路径"
          />
          </ProForm.Group>
        <ProForm.Group>
          <ProFormText
            name="url"
            label="链接地址"
            width="md"
            placeholder="请输入组件路径"
          />
          <ProFormSelect
            name="identify"
            label="权限标识"
            width="md"
            request={async () => {
              const msg = await getApiRoles({
                page: 1,
                count: 1000
              })
             return msg.data.list.map((item: any) => {
                return {
                  label: item.name,
                  value: item.id
                }
              })
            }}
          />
          </ProForm.Group>
        <ProForm.Group>
          <ProFormDigit
            name="sort"
            label="菜单排序"
            width="md"
            min={1}
            max={10}
            fieldProps={{ precision: 0 }}
          />
          <ProFormRadio.Group
            radioType="button"
            name="hidden"
            label="是否隐藏"
            width="md"
            options={[
              {
                label: '是',
                value:1,
              },
              {
                label: '否',
                value:2,
              },
            ]
            }
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormRadio.Group
            radioType="button"
            width="sm"
            label="页面缓存"
            name="cache"
            options={[
              {
                label: '缓存',
                value:1,
              },
              {
                label: '不缓存',
                value:2,
              },
            ]
            }
          />
          <ProFormRadio.Group
            width="sm"
            label="是否固定"
            name="fixed"
            // initialValue="固定"
            options={[
              {
                label: '固定',
                value:1,
              },
              {
                label: '不固定',
                value:2,
              },
            ]
            }
          />
        </ProForm.Group>
      </ModalForm>
    </div>
  )
}

export default MenuEditor
