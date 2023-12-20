import React, {useRef,useEffect,useState} from 'react'

import {  message,Form } from 'antd';
import {
  ProForm,
  ModalForm,
  ProFormText,
  ProFormSwitch,
  ProFormDigit,
  ProFormTextArea

} from '@ant-design/pro-components';
import type { DataNode } from 'antd/es/tree';

import {ProFormInstance} from "@ant-design/pro-form/lib";
import {getMenus} from "@/services/echo-admin/caidan";
import MenuTree from "@/pages/system/role/components/menuTree";
import {postRoles, putRolesId} from "@/services/echo-admin/jiaose";


type Props = {
  modalVisit: boolean;
  setModalVisit: (modalVisit: boolean) => void;
  values: Partial<API.role>;
}

const MenuItemTree = (menus: API.menu[]): DataNode[] => {
  return menus.map((item) => {
    const { name, id, children } = item;
    const newItem: DataNode = {
      title: name,
      key: id ?? 0,
    };
    if (children && children.length > 0) {
      newItem.children = MenuItemTree(children);
    }
    return newItem;
  });
};

const RoleEditor: React.FC<Props> = props => {
  const { modalVisit, setModalVisit } = props;
  const restFormRef = useRef<ProFormInstance>();
  const [treeData, setTreeData] = useState<DataNode[]>([]);
  useEffect(() => {
    restFormRef.current?.resetFields();
    restFormRef.current?.setFieldsValue({
      id: props.values.id,
      name:props.values.name,
      sort:props.values.sort,
      mark:props.values.mark,
      menus:props.values.menus,
    });
  }, [restFormRef, props]);
  return (
    <div>
      <ModalForm
        title="新建角色"
        formRef={restFormRef}
        open={modalVisit}
        request={async () => {
          let res = await getMenus()
          setTreeData(MenuItemTree(res.data.list??[]))
          return {}
        }}
        onFinish={async (fieldsValue: API.role) => {
          console.log(fieldsValue)
          let res: API.success
          if (props.values?.id) {
            res = await putRolesId({id: props.values.id}, fieldsValue)
          } else {
            res = await postRoles(fieldsValue)
          }

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
        <ProForm.Group>
          <ProFormText
            name="name"
            width="md"
            label="角色名称"
            placeholder="请填写角色名称"
          />
          <ProFormDigit
            name="sort"
            label="排序"
            width="md"
            min={1}
            max={10}
            fieldProps={{ precision: 0 }}
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormSwitch
            name="status"
            label="角色状态"
            width="md"
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormTextArea
            name="mark"
            label="角色描述"
            placeholder="请输入角色描述"
            width="xl"
          />
        </ProForm.Group>
        <Form.Item name="menus">
          <MenuTree treeData={treeData}/>
        </Form.Item>
      </ModalForm>
    </div>
  )
}

export default RoleEditor
