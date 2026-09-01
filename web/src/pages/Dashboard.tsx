import {
  AppstoreOutlined,
  CheckCircleFilled,
  CloseCircleFilled,
  CloudServerOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons';
import {
  PageContainer,
  ProCard,
  ProDescriptions,
  StatisticCard,
} from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { Avatar, Spin, Tag, Typography } from 'antd';
import { createStyles } from 'antd-style';
import React, { useEffect, useMemo, useState } from 'react';

import { menuIconElement } from '@/runtime/menu';
import {
  type AppInfo,
  appInfo,
  type CapabilityStatus,
  capabilities,
} from '@/services/admin';
import { buildMenuTree } from '@/utils/menu-tree';

// 统计卡图标使用 antd 预设浅色底色，与默认 light 主题配套。
const statIcons = {
  roles: { bg: '#e6f4ff', color: '#1677ff', icon: <TeamOutlined /> },
  permissions: {
    bg: '#f9f0ff',
    color: '#722ed1',
    icon: <SafetyCertificateOutlined />,
  },
  menus: { bg: '#e6fffb', color: '#13c2c2', icon: <AppstoreOutlined /> },
  capabilities: {
    bg: '#f6ffed',
    color: '#52c41a',
    icon: <CloudServerOutlined />,
  },
} as const;

/** 统计卡前缀的彩色图标徽章。 */
const StatIcon: React.FC<{
  bg: string;
  color: string;
  children: React.ReactNode;
}> = ({ bg, color, children }) => (
  <span
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      width: 40,
      height: 40,
      borderRadius: 10,
      background: bg,
      color,
      fontSize: 18,
      verticalAlign: 'middle',
      marginInlineEnd: 12,
    }}
  >
    {children}
  </span>
);

const greeting = (): string => {
  const hour = new Date().getHours();
  if (hour < 6) return '夜深了';
  if (hour < 12) return '早上好';
  if (hour < 18) return '下午好';
  return '晚上好';
};

const formatDate = (date: Date): string =>
  new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  }).format(date);

const formatDateTime = (date: Date): string =>
  new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date);

/** 把 resource:action 权限 token 按资源前缀聚合，工作台按资源分组展示比平铺更易读。 */
const groupPermissions = (permissions: string[]): [string, string[]][] => {
  const groups = new Map<string, string[]>();
  for (const permission of permissions) {
    const separator = permission.indexOf(':');
    const resource =
      separator < 0 ? permission : permission.slice(0, separator);
    if (separator >= 0) {
      const action = permission.slice(separator + 1);
      const actions = groups.get(resource) ?? [];
      actions.push(action);
      groups.set(resource, actions);
    } else {
      groups.set(resource, groups.get(resource) ?? []);
    }
  }
  return [...groups.entries()];
};

const useStyles = createStyles(({ token }) => ({
  banner: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '24px 28px',
    borderRadius: token.borderRadiusLG,
    background: `linear-gradient(120deg, ${token.colorPrimary} 0%, ${token.colorPrimaryTextHover} 100%)`,
    marginBottom: 16,
  },
  bannerText: {
    color: 'rgba(255, 255, 255, 0.85)',
  },
  groupList: {
    display: 'flex',
    flexDirection: 'column',
    gap: 14,
  },
  tagWrap: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: 4,
  },
  statusMeta: {
    display: 'flex',
    justifyContent: 'space-between',
    flexWrap: 'wrap',
    gap: 8,
    paddingBottom: 12,
    borderBottom: `1px solid ${token.colorSplit}`,
  },
  statusRow: {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
  },
  statusIconOk: {
    color: token.colorSuccess,
    fontSize: 14,
  },
  statusIconError: {
    color: token.colorError,
    fontSize: 14,
  },
  statusMessage: {
    flex: 1,
    minWidth: 120,
    textAlign: 'right',
  },
  menuRow: {
    display: 'flex',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: 8,
  },
  menuIcon: {
    display: 'inline-flex',
    color: token.colorPrimary,
    fontSize: 16,
  },
}));

const Dashboard: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const user = initialState?.currentUser;
  const [info, setInfo] = useState<AppInfo>();
  const [capabilityRows, setCapabilityRows] = useState<CapabilityStatus[]>([]);
  const [loadingStatus, setLoadingStatus] = useState(false);
  const { styles } = useStyles();

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

  const visibleMenus = useMemo(
    () => (user?.menus ?? []).filter((menu) => menu.active && !menu.hidden),
    [user?.menus],
  );
  const menuTree = useMemo(() => buildMenuTree(visibleMenus), [visibleMenus]);
  const permissionGroups = useMemo(
    () => groupPermissions(user?.permissions ?? []),
    [user?.permissions],
  );
  const availableCapabilities = capabilityRows.filter(
    (item) => item.available,
  ).length;
  const activeRole = user?.roles.find(
    (role) => role.id === user?.active_role_id,
  );

  return (
    <PageContainer title="工作台">
      <div className={styles.banner}>
        <div>
          <Typography.Title level={3} style={{ color: '#fff', margin: 0 }}>
            {greeting()}，{user?.display_name || user?.username || '管理员'}
          </Typography.Title>
          <Typography.Text className={styles.bannerText}>
            {formatDate(new Date())}
            {activeRole ? ` · 当前角色：${activeRole.name}` : ''}
          </Typography.Text>
        </div>
        <Avatar
          size={56}
          icon={<UserOutlined />}
          style={{ background: 'rgba(255, 255, 255, 0.25)', color: '#fff' }}
        />
      </div>
      <ProCard gutter={[16, 16]} wrap>
        <StatisticCard
          colSpan={{ xs: 12, md: 6 }}
          statistic={{
            title: '角色',
            value: user?.roles.length ?? 0,
            description: '当前管理员拥有的角色',
            prefix: (
              <StatIcon bg={statIcons.roles.bg} color={statIcons.roles.color}>
                {statIcons.roles.icon}
              </StatIcon>
            ),
          }}
        />
        <StatisticCard
          colSpan={{ xs: 12, md: 6 }}
          statistic={{
            title: '已授权能力',
            value: user?.permissions.length ?? 0,
            description: 'resource:action 权限 token',
            prefix: (
              <StatIcon
                bg={statIcons.permissions.bg}
                color={statIcons.permissions.color}
              >
                {statIcons.permissions.icon}
              </StatIcon>
            ),
          }}
        />
        <StatisticCard
          colSpan={{ xs: 12, md: 6 }}
          statistic={{
            title: '后台菜单',
            value: visibleMenus.length,
            description: '当前角色可见的后台菜单',
            prefix: (
              <StatIcon bg={statIcons.menus.bg} color={statIcons.menus.color}>
                {statIcons.menus.icon}
              </StatIcon>
            ),
          }}
        />
        <StatisticCard
          colSpan={{ xs: 12, md: 6 }}
          statistic={{
            title: '可用 Capability',
            value: loadingStatus ? '-' : availableCapabilities,
            description: `共 ${capabilityRows.length} 项基础设施能力`,
            prefix: (
              <StatIcon
                bg={statIcons.capabilities.bg}
                color={statIcons.capabilities.color}
              >
                {statIcons.capabilities.icon}
              </StatIcon>
            ),
          }}
        />
        <ProCard
          colSpan={{ xs: 24, lg: 14 }}
          title="已授权能力"
          extra={
            <Typography.Text type="secondary">
              {user?.permissions.length ?? 0} 个 token
            </Typography.Text>
          }
        >
          {permissionGroups.length === 0 ? (
            <Typography.Text type="secondary">暂无授权能力</Typography.Text>
          ) : (
            <div className={styles.groupList}>
              {permissionGroups.map(([resource, actions]) => (
                <div key={resource}>
                  <span
                    style={{
                      display: 'inline-flex',
                      gap: 8,
                      alignItems: 'center',
                    }}
                  >
                    <Typography.Text code strong>
                      {resource}
                    </Typography.Text>
                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                      × {actions.length}
                    </Typography.Text>
                  </span>
                  <div className={styles.tagWrap}>
                    {actions.map((action) => (
                      <Tag key={action} style={{ marginInlineEnd: 0 }}>
                        {action}
                      </Tag>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </ProCard>
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
                title: '默认路径',
                dataIndex: 'default_path',
                render: (_, entity) => entity?.default_path || '-',
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
        <ProCard
          colSpan={{ xs: 24, lg: 14 }}
          title="系统状态"
          extra={
            info ? (
              <Typography.Text type="secondary">
                v{info.version}
              </Typography.Text>
            ) : null
          }
        >
          <Spin spinning={loadingStatus}>
            <div className={styles.groupList}>
              {info && (
                <div className={styles.statusMeta}>
                  <Typography.Text strong>{info.name}</Typography.Text>
                  <Typography.Text type="secondary">
                    服务时间：{formatDateTime(new Date(info.time))}
                  </Typography.Text>
                </div>
              )}
              {capabilityRows.map((item) => (
                <div key={item.name} className={styles.statusRow}>
                  {item.available ? (
                    <CheckCircleFilled className={styles.statusIconOk} />
                  ) : (
                    <CloseCircleFilled className={styles.statusIconError} />
                  )}
                  <Typography.Text strong>{item.name}</Typography.Text>
                  <Tag
                    color={item.available ? 'success' : 'error'}
                    style={{ marginInlineEnd: 0 }}
                  >
                    {item.state}
                  </Tag>
                  {item.message && (
                    <Typography.Text
                      type="secondary"
                      ellipsis
                      className={styles.statusMessage}
                    >
                      {item.message}
                    </Typography.Text>
                  )}
                </div>
              ))}
            </div>
          </Spin>
        </ProCard>
        <ProCard colSpan={{ xs: 24, lg: 10 }} title="后台菜单">
          {menuTree.length === 0 ? (
            <Typography.Text type="secondary">暂无可见菜单</Typography.Text>
          ) : (
            <div className={styles.groupList}>
              {menuTree.map((menu) => (
                <div key={menu.id} className={styles.menuRow}>
                  <span className={styles.menuIcon}>
                    {menuIconElement(menu.icon) ?? <AppstoreOutlined />}
                  </span>
                  <Typography.Text strong>{menu.name}</Typography.Text>
                  <div className={styles.tagWrap}>
                    {menu.children.map((child) => (
                      <Tag key={child.id} style={{ marginInlineEnd: 0 }}>
                        {child.name}
                      </Tag>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </ProCard>
      </ProCard>
    </PageContainer>
  );
};

export default Dashboard;
