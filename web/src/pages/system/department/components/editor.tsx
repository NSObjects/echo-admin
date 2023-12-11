import React, {useRef,useEffect} from 'react'

import {  message } from 'antd';
import {
  ProForm,
  ProFormText,
  ModalForm, ProFormTreeSelect,
  ProFormDigit, ProFormInstance,
  ProFormSwitch
} from '@ant-design/pro-components';

import { Form, Modal} from 'antd';

import {getApiDepartments, postApiDepartments} from "@/services/echo-admin/bumen";

type Props = {
  modalVisit: boolean;
  setModalVisit: (modalVisit: boolean) => void;
  reload: () => void;
  // values: Partial<API.department>;
}


interface TransformedNode {
  value: number;
  title: string;
  children?: TransformedNode[];
}


const DepartmentEditor: React.FC<Props> = props => {

  // const [form] = Form.useForm();
  // useEffect(() => {
  //   form.resetFields();
  //   form.setFieldsValue({
  //     id: props.values.id,
  //     // parentId: props.values.parentId,
  //     // ancestors: props.values.ancestors,
  //     // deptName: props.values.deptName,
  //     // orderNum: props.values.orderNum,
  //     // leader: props.values.leader,
  //     phone: props.values.phone,
  //     email: props.values.email,
  //     status: props.values.status,
  //     // delFlag: props.values.delFlag,
  //     // createBy: props.values.createBy,
  //     // createTime: props.values.createTime,
  //     // updateBy: props.values.updateBy,
  //     // updateTime: props.values.updateTime,
  //   });
  // }, [form, props]);
  const { modalVisit, setModalVisit } = props;
  const restFormRef = useRef<ProFormInstance>();
  function transformNode(node:API.department): TransformedNode {
    const item: TransformedNode = {
      value: node.id,
      title: node.name,
    };
    if (node.children) {
      item.children = node.children.map(child => transformNode(child));
    }
    return item;
  }
  return (
    <div>
      <ModalForm
        // @ts-ignore
        initialValues={props?.obj}
        title="添加部门"
        modalProps={{
          destroyOnClose: true
        }}
        formRef={restFormRef}
        open={modalVisit}
        onFinish={async (fieldsValue: any) => {
          console.log(fieldsValue)
          const res = await getApiDepartments({
            email: fieldsValue["email"],
            name: fieldsValue["name"],
            parent_id: fieldsValue["parent_id"],
            phone: fieldsValue["phone"],
            principal: fieldsValue["principal"],
            sort: fieldsValue["sort"],
            status: fieldsValue["status"] ? 1 : 2,
          })
          if (res.code === 0) {
            message.success('提交成功');
            props.reload();
            return true;
          } else {
            message.error('提交失败');
            return false;
          }
        }}
        onOpenChange={(visible:boolean)=>{
          setModalVisit(visible)
        }}
      >
        <ProFormTreeSelect
          name="parent_id"
          label="上级部门"
          request={async () => {
            const res = await getApiDepartments({page: 0, count: 1000})
            return res.data.list.map((item: API.department) => {
              return transformNode(item)
            })
          }}

        />
        <ProForm.Group>
          <ProFormText
            width="md"
            name="name"
            label="部门名称"
            tooltip="最长为 24 位"
            placeholder="请输入部门名称名称"
          />
          <ProFormText
            width="md"
            name="principal"
            label="负责人"
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormText
            name="phone"
            width="md"
            label="手机号"
            placeholder="请输入手机号"
          />
          <ProFormText
            name="email"
            width="md"
            label="邮箱"
            placeholder="请输入邮箱"
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormDigit
            label="排序"
            name="sort"
            width="md"
            fieldProps={{ precision: 0 }}
          />
          <ProFormSwitch
            colProps={{
              span: 4,
            }}
            // fieldProps={{
            //   onChange: setGrid,
            // }}
            width="md"
            initialValue={true}
            label="是否启用"
            name="status"
          />
        </ProForm.Group>

      </ModalForm>
    </div>
  )
}

export default DepartmentEditor
