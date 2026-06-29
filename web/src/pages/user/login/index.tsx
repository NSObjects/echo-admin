import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { LoginForm, ProFormText } from '@ant-design/pro-components';
import { Helmet, history, useModel } from '@umijs/max';
import { App } from 'antd';
import { createStyles } from 'antd-style';
import React from 'react';
import { flushSync } from 'react-dom';

import { Footer } from '@/components';
import { login } from '@/services/admin';
import Settings from '../../../../config/defaultSettings';

const useStyles = createStyles(({ token }) => ({
  container: {
    minHeight: '100vh',
    display: 'flex',
    flexDirection: 'column',
    background: token.colorBgLayout,
  },
  main: {
    flex: 1,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
  },
  panel: {
    width: '100%',
    maxWidth: 420,
  },
}));

const loginPath = '/user/login';

const safeRedirect = (redirect: string | null, fallback: string): string => {
  if (!redirect?.startsWith('/') || redirect.startsWith('//')) return fallback;
  try {
    const parsed = new URL(redirect, window.location.origin);
    if (parsed.origin !== window.location.origin) return fallback;
    return `${parsed.pathname}${parsed.search}${parsed.hash}`;
  } catch {
    return fallback;
  }
};

const Login: React.FC = () => {
  const { styles } = useStyles();
  const { message } = App.useApp();
  const { setInitialState } = useModel('@@initialState');

  return (
    <div className={styles.container}>
      <Helmet>
        <title>{`登录 - ${Settings.title}`}</title>
      </Helmet>
      <main className={styles.main}>
        <div className={styles.panel}>
          <LoginForm
            title="Echo Admin"
            subTitle="统一后台管理基础框架"
            initialValues={{ username: 'admin' }}
            onFinish={async (values) => {
              const result = await login({
                username: String(values.username ?? ''),
                password: String(values.password ?? ''),
              });
              const currentUser = result.user;
              flushSync(() => {
                setInitialState((state) => ({ ...state, currentUser }));
              });
              message.success('登录成功');
              const redirect = new URL(window.location.href).searchParams.get(
                'redirect',
              );
              const fallbackPath = currentUser?.default_path || '/dashboard';
              history.replace(
                safeRedirect(
                  redirect === loginPath ? null : redirect,
                  fallbackPath,
                ),
              );
            }}
          >
            <ProFormText
              name="username"
              fieldProps={{
                size: 'large',
                prefix: <UserOutlined />,
                autoComplete: 'username',
              }}
              placeholder="用户名：admin"
              rules={[{ required: true, message: '请输入用户名' }]}
            />
            <ProFormText.Password
              name="password"
              fieldProps={{
                size: 'large',
                prefix: <LockOutlined />,
                autoComplete: 'current-password',
              }}
              placeholder="请输入密码"
              rules={[{ required: true, message: '请输入密码' }]}
            />
          </LoginForm>
        </div>
      </main>
      <Footer />
    </div>
  );
};

export default Login;
