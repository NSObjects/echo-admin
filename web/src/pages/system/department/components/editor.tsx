import React, { useRef, useEffect } from 'react';

import { message } from 'antd';
import {
  ProForm,
  ProFormText,
  ModalForm,
  ProFormTreeSelect,
  ProFormDigit,
  ProFormInstance,
  ProFormSwitch,
} from '@ant-design/pro-components';

import { getDepartments, postDepartments, putDepartmentsId } from '@/services/echo-admin/bumen';

type Props = {
  modalVisit: boolean;
  setModalVisit: (modalVisit: boolean) => void;
  values: Partial<API.department>;
  reload: () => void;
};

interface TransformedNode {
  value: number;
  title: string;
  children?: TransformedNode[];
}

const DepartmentEditor: React.FC<Props> = (props) => {
  const { modalVisit, setModalVisit, values } = props;
  const restFormRef = useRef<ProFormInstance>();
  useEffect(() => {
    restFormRef.current?.resetFields();
    restFormRef.current?.setFieldsValue({
      id: values.id,
      phone: values.phone,
      email: values.email,
      status: values.id !== undefined ? values.status === 1 : true,
      parent_id: values.parent_id,
      sort: values.sort,
      principal: values.principal,
      name: values.name,
    });
  }, [restFormRef, props]);

  function transformNode(node: API.department): TransformedNode {
    const item: TransformedNode = {
      value: node.id ?? 0,
      title: node.name,
    };
    if (node.children) {
      item.children = node.children.map((child) => transformNode(child));
    }
    return item;
  }

  return (
    <div>
      <ModalForm
        title="添加部门"
        formRef={restFormRef}
        open={modalVisit}
        onFinish={async (fieldsValue: API.department) => {
          fieldsValue.status = fieldsValue.status ? 1 : 2;
          let res: API.success;
          if (values?.id) {
            res = await putDepartmentsId({ id: props.values.id ?? 0 }, fieldsValue);
          } else {
            res = await postDepartments(fieldsValue);
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
        onOpenChange={(visible: boolean) => {
          if (!visible) {
            restFormRef.current?.resetFields();
          }
          setModalVisit(visible);
        }}
      >
        <ProFormTreeSelect
          name="parent_id"
          label="上级部门"
          request={async () => {
            const res = await getDepartments({ page: 0, count: 1000 });
            return (res.data.list ?? []).map((item: API.department) => {
              return transformNode(item);
            });
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
          <ProFormText width="md" name="principal" label="负责人" />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormText name="phone" width="md" label="手机号" placeholder="请输入手机号" />
          <ProFormText name="email" width="md" label="邮箱" placeholder="请输入邮箱" />
        </ProForm.Group>
        <ProForm.Group>
          <ProFormDigit label="排序" name="sort" width="md" fieldProps={{ precision: 0 }} />
          <ProFormSwitch
            colProps={{
              span: 4,
            }}
            width="md"
            initialValue={true}
            label="是否启用"
            name="status"
          />
        </ProForm.Group>
      </ModalForm>
    </div>
  );
};

export default DepartmentEditor;
