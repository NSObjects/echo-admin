import React, {useRef,useEffect} from 'react'

import {  message } from 'antd';
import {
  ProForm,
  ProFormTextArea,
  ProFormSelect,
  ProFormText,
  ModalForm

} from '@ant-design/pro-components';
import {getRoles} from "@/services/echo-admin/jiaose";
import ProFormSwitch from "@ant-design/pro-form/es/components/Switch";
import {postUsers, putUsersId} from "@/services/echo-admin/yonghu";
import {ProFormInstance} from "@ant-design/pro-form/lib";
type Props = {
  modalVisit: boolean;
  setModalVisit: (modalVisit: boolean) => void;
  values: Partial<API.user>;
  reload:()=>void;
}

const UserEditor: React.FC<Props> = props => {
  const { modalVisit, setModalVisit } = props;
  const restFormRef = useRef<ProFormInstance>();
  useEffect(() => {
    restFormRef.current?.resetFields();
    restFormRef.current?.setFieldsValue({
      id: props.values.id,
      name: props.values.name,
      phone: props.values.phone,
      status: props.values.id !== undefined ? props.values.status === 1 : true,
      // password: props.values.password,
      account: props.values.account,
      avatar: props.values.avatar,
      role_id: props.values.role_id,
      department_id: props.values.department_id,
      email: props.values.email,
      sex: props.values.sex,
      posts:props.values.posts,
    });
  }, [restFormRef, props]);
  return (
    <div>
      <ModalForm
        // @ts-ignore
        title="新建用户"
        formRef={restFormRef}
        open={modalVisit}
        onFinish={async (fieldsValue: any) => {
          console.log(fieldsValue)
          let body = {
            account: fieldsValue["phone"],
            avatar: fieldsValue["avatar"],
            name: fieldsValue["name"],
            password:  fieldsValue["password"],
            phone: fieldsValue["phone"],
            department_id: fieldsValue["department_id"],
            posts: fieldsValue["posts"],
            role_id: fieldsValue["role_id"],
            sex: fieldsValue["sex"],
            email: fieldsValue["email"],
            status: fieldsValue["status"] ? 1 : 0,
          }
          let res: API.success
          if (props.values?.id) {
             res = await putUsersId({id: props.values.id}, body)
          } else {
             res = await postUsers(body)
          }

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
          if (!visible) {
            restFormRef.current?.resetFields();
          }
          setModalVisit(visible)
        }}
      >
        <ProForm.Group>
          <ProFormText
            width="md"
            name="account"
            label="用户名"
            placeholder="请输入账户名称"
          />
          <ProFormText.Password
            width="md"
            name="password"
            label="账户密码"
            placeholder="请输入"
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormText
            name="name"
            width="md"
            label="用户昵称"
            placeholder="请输入用户昵称"
          />
          <ProFormSelect
            name="role_id"
            label="关联角色"
            width="md"
            request={async () => {
              const res = await getRoles({page: 0, count: 1000})
              return (res.data.list ?? []).map((item: any) => {
                return {label: item.name, value: item.id}
              })
            }}

          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormSelect
            name="department_id"
            label="部门"
            width="md"
            request={async () => {
              const res = await getRoles({page: 0, count: 1000})
              return (res.data.list ?? []).map((item: any) => {
                return {label: item.name, value: item.id}
              })
            }}
          />
          <ProFormText
            name="phone"
            width="md"
            label="手机号"
            placeholder="请输入手机号"
          />

        </ProForm.Group>
        <ProForm.Group>
          <ProFormText
            name="email"
            width="md"
            label="邮箱"
            placeholder="请输入邮箱"
          />
          <ProFormSelect
            options={[
              {
                value: 1,
                label: "男",
              },
              {
                value: 2,
                label: "女",
              },
            ]}
            width="md"
            name="sex"
            label="性别"
          />

        </ProForm.Group>
        <ProForm.Group>
          <ProFormSelect
            name="posts"
            label="岗位"
            width="md"
            request={async () => {
              const res = await getRoles({page: 0, count: 1000})
              return (res.data.list ?? []).map((item: any) => {
                return {label: item.name, value: item.id}
              })
            }}
          />
          <ProFormSwitch
            colProps={{
              span: 4,
            }}
            width="md"
            // initialValue={true}
            label="是否启用"
            name="status"
          />
        </ProForm.Group>
        <ProFormTextArea
          name="mark"
          label="备注"
          placeholder="请输入名称"
        />
      </ModalForm>
    </div>
  )
}

export default UserEditor
