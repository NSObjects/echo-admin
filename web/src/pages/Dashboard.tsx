import {
  PageContainer,
  ProCard,
  ProDescriptions,
  StatisticCard,
} from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { Spin, Tag } from 'antd';
import React, { useEffect, useState } from 'react';

import {
  type AppInfo,
  appInfo,
  type CapabilityStatus,
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
      <ProCard gutter={[16, 16]} wrap>
        <StatisticCard
          colSpan={{ xs: 12, md: 6 }}
          statistic={{
            title: '角色',
            value: user?.roles.length ?? 0,
            description: '当前管理员拥有的角色',
          }}
        />
        <StatisticCard
          colSpan={{ xs: 12, md: 6 }}
          statistic={{
            title: '已授权能力',
            value: user?.permissions.length ?? 0,
            description: 'resource:action 权限 token',
          }}
        />
        <StatisticCard
          colSpan={{ xs: 12, md: 6 }}
          statistic={{
            title: '后台菜单',
            value: user?.menus.length ?? 0,
            description: '可见的后台菜单',
          }}
        />
        <StatisticCard
          colSpan={{ xs: 12, md: 6 }}
          statistic={{
            title: '可用 Capability',
            value: capabilityRows.filter((item) => item.available).length,
            description: `共 ${capabilityRows.length} 项基础设施能力`,
          }}
        />
        <ProCard colSpan={{ xs: 24, lg: 10 }} title="当前管理员">
          <ProDescriptions
            column={1}
            size="small"
            dataSource={user}
            columns={[
              { title: '用户名', dataIndex: 'username' },
              { title: '显示名', dataIndex: 'display_name' },
              {
                title: '邮箱',
                dataIndex: 'email',
                render: (_, entity) => entity?.email || '-',
              },
              {
                title: '角色',
                dataIndex: 'roles',
                render: (_, entity) =>
                  entity?.roles.map((role) => (
                    <Tag key={role.id} color="blue">
                      {role.name}
                    </Tag>
                  )),
              },
            ]}
          />
        </ProCard>
        <ProCard colSpan={{ xs: 24, lg: 14 }} title="系统状态">
          <Spin spinning={loadingStatus}>
            <ProDescriptions
              column={1}
              size="small"
              dataSource={info}
              columns={[
                { title: '应用', dataIndex: 'name' },
                { title: '版本', dataIndex: 'version' },
                {
                  title: '服务时间',
                  dataIndex: 'time',
                  render: (_, entity) =>
                    entity?.time ? new Date(entity.time).toLocaleString() : '-',
                },
                {
                  title: 'Capability',
                  dataIndex: 'capabilities',
                  render: () =>
                    capabilityRows.map((item) => (
                      <Tag
                        key={item.name}
                        color={item.available ? 'green' : 'red'}
                      >
                        {item.name}:{item.state}
                      </Tag>
                    )),
                },
              ]}
            />
          </Spin>
        </ProCard>
        <ProCard colSpan={{ xs: 24, lg: 14 }} title="已授权能力">
          {user?.permissions.map((permission) => (
            <Tag key={permission} color="geekblue">
              {permission}
            </Tag>
          ))}
        </ProCard>
        <ProCard colSpan={24} title="后台菜单">
          {user?.menus.map((menu) => (
            <Tag key={menu.id}>{menu.name}</Tag>
          ))}
        </ProCard>
      </ProCard>
    </PageContainer>
  );
};

export default Dashboard;
