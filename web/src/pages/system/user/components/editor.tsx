import React, {useRef} from 'react'

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
import {postUsers} from "@/services/echo-admin/yonghu";
import {ProFormInstance} from "@ant-design/pro-form/lib";
type Props = {
  modalVisit: boolean;
  setModalVisit: (modalVisit: boolean) => void;
}
const UserEditor: React.FC<Props> = props => {
  const { modalVisit, setModalVisit } = props;
  const restFormRef = useRef<ProFormInstance>();
  return (
    <div>
      <ModalForm
        // @ts-ignore
        title="新建用户"
        formRef={restFormRef}
        open={modalVisit}
        onFinish={async (fieldsValue: any) => {
          console.log(fieldsValue)
          const res = await postUsers({
            account: fieldsValue["phone"],
            avatar: "",
            name: fieldsValue["name"],
            password:  fieldsValue["password"],
            phone: fieldsValue["phone"],
            status: fieldsValue["status"] ? 1 : 0,
          })
          if (res.code === 200) {
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
            width="md"
            name="name"
            label="用户昵称"
            tooltip="最长为 24 位"
            placeholder="请输入名称"
          />
          <ProFormText
            width="md"
            name="phone"
            label="电话号码"
            placeholder="请输入名称"
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormText.Password
            name="password"
            width="md"
            label="密码"
            placeholder="请输入密码"
          />
          <ProFormSwitch
            colProps={{
              span: 4,
            }}
            // fieldProps={{
            //   onChange: setGrid,
            // }}
            initialValue={true}
            label="是否启用"
            name="status"
          />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormSelect
            name="role"
            label="角色"
            width="md"
            request={async () => {
              const res = await getRoles({page: 0, count: 1000})
              return (res.data.list ?? []).map((item: any) => {
                return {label: item.name, value: item.id}
              })
            }}

          />

          <ProFormSelect
            name="depratment"
            label="部门"
            width="md"
            request={async () => {
              const res = await getRoles({page: 0, count: 1000})
              return (res.data.list ?? []).map((item: any) => {
                return {label: item.name, value: item.id}
              })
            }}
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
