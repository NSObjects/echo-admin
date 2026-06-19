import { PageContainer } from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { Card, Col, Descriptions, Row, Tag } from 'antd';
import React from 'react';

const Dashboard: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const user = initialState?.currentUser;

  return (
    <PageContainer title="工作台">
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card title="当前管理员">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="用户名">{user?.username}</Descriptions.Item>
              <Descriptions.Item label="显示名">{user?.display_name}</Descriptions.Item>
              <Descriptions.Item label="邮箱">{user?.email || '-'}</Descriptions.Item>
              <Descriptions.Item label="角色">
                {user?.roles.map((role) => (
                  <Tag key={role.id} color="blue">
                    {role.name}
                  </Tag>
                ))}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
        <Col xs={24} lg={14}>
          <Card title="已授权能力">
            {user?.permissions.map((permission) => (
              <Tag key={permission} color="geekblue">
                {permission}
              </Tag>
            ))}
          </Card>
        </Col>
        <Col span={24}>
          <Card title="后台菜单">
            {user?.menus.map((menu) => (
              <Tag key={menu.id}>{menu.name}</Tag>
            ))}
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Dashboard;
