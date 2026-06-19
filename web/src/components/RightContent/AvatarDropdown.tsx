import {
  CheckOutlined,
  LogoutOutlined,
  SkinOutlined,
  SwapOutlined,
} from '@ant-design/icons';
import { history, useModel } from '@umijs/max';
import type { MenuProps } from 'antd';
import { message, Spin } from 'antd';
import React, { startTransition } from 'react';
import { logout, switchRole } from '@/services/admin';
import HeaderDropdown from '../HeaderDropdown';

type GlobalHeaderRightProps = {
  children?: React.ReactNode;
};

export const AvatarDropdown: React.FC<GlobalHeaderRightProps> = ({
  children,
}) => {
  const loginOut = async () => {
    await logout();
    const { search, pathname } = window.location;
    const urlParams = new URL(window.location.href).searchParams;
    const searchParams = new URLSearchParams({
      redirect: pathname + search,
    });
    const redirect = urlParams.get('redirect');
    if (window.location.pathname !== '/user/login' && !redirect) {
      history.replace({
        pathname: '/user/login',
        search: searchParams.toString(),
      });
    }
  };
  const { initialState, setInitialState } = useModel('@@initialState');

  const onMenuClick: MenuProps['onClick'] = async (event) => {
    const { key } = event;
    if (key === 'logout') {
      startTransition(() => {
        setInitialState((s) => ({ ...s, currentUser: undefined }));
      });
      await loginOut();
      return;
    }
    if (key === 'theme') {
      setInitialState((s) => ({ ...s, settingDrawerOpen: true }));
      return;
    }
    if (typeof key === 'string' && key.startsWith('role:')) {
      const roleID = Number(key.slice('role:'.length));
      if (!Number.isFinite(roleID)) return;
      const result = await switchRole(roleID);
      startTransition(() => {
        setInitialState((s) => ({ ...s, currentUser: result.user }));
      });
      message.success('角色已切换');
      history.replace(result.user.default_path || '/dashboard');
    }
  };

  if (!initialState) {
    return <Spin size="small" />;
  }

  const { currentUser } = initialState;

  if (!currentUser) {
    return <Spin size="small" />;
  }

  const roleItems: MenuProps['items'] = currentUser.roles.map((role) => ({
    key: `role:${role.id}`,
    icon:
      role.id === currentUser.active_role_id ? <CheckOutlined /> : undefined,
    label: role.name,
    disabled: role.id === currentUser.active_role_id,
  }));

  const menuItems: MenuProps['items'] = [
    {
      key: 'roles',
      icon: <SwapOutlined />,
      label: currentUser.active_role?.name ?? '切换角色',
      children: roleItems,
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'theme',
      icon: <SkinOutlined />,
      label: '主题设置',
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
    },
  ];

  return (
    <HeaderDropdown
      placement="bottomRight"
      menu={{
        selectedKeys: [],
        onClick: onMenuClick,
        items: menuItems,
      }}
      arrow
    >
      {children}
    </HeaderDropdown>
  );
};
