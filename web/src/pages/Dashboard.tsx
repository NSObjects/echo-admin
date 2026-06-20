import { PageContainer } from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { Card, Col, Descriptions, Row, Spin, Tag } from 'antd';
import React, { useEffect, useState } from 'react';

import {
  type AppInfo,
  type CapabilityStatus,
  appInfo,
  capabilities,
} from '@/services/admin';

const Dashboard: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const user = initialState?.currentUser;
  const [info, setInfo] = useState<AppInfo>();
  const [capabilityRows, setCapabilityRows] = useState<CapabilityStatus[]>([]);
  const [loadingStatus, setLoadingStatus] = useState(false);

  useEffect(() => {
    let active = true;
    const loadStatus = async () => {
      setLoadingStatus(true);
      try {
        const [infoResponse, capabilitiesResponse] = await Promise.all([
          appInfo(),
          capabilities(),
        ]);
        if (!active) {
          return;
        }
        setInfo(infoResponse);
        setCapabilityRows(capabilitiesResponse.capabilities);
      } finally {
        if (active) {
          setLoadingStatus(false);
        }
      }
    };

    void loadStatus();
    return () => {
      active = false;
    };
  }, []);

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
          <Card title="系统状态">
            <Spin spinning={loadingStatus}>
              <Descriptions column={1} size="small">
                <Descriptions.Item label="应用">{info?.name ?? '-'}</Descriptions.Item>
                <Descriptions.Item label="版本">
                  {info?.version ?? '-'}
                </Descriptions.Item>
                <Descriptions.Item label="服务时间">
                  {info ? new Date(info.time).toLocaleString() : '-'}
                </Descriptions.Item>
                <Descriptions.Item label="Capability">
                  {capabilityRows.map((item) => (
                    <Tag
                      key={item.name}
                      color={item.available ? 'green' : 'red'}
                    >
                      {item.name}:{item.state}
                    </Tag>
                  ))}
                </Descriptions.Item>
              </Descriptions>
            </Spin>
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
