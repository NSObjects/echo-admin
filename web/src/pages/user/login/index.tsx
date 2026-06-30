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

const useStyles = createStyles(() => ({
  container: {
    minHeight: '100dvh',
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 1fr) minmax(420px, 520px)',
    background: '#f5f7f6',
    color: '#18211f',
    overflow: 'hidden',
    '@media (max-width: 920px)': {
      display: 'block',
      minHeight: '100dvh',
      overflow: 'auto',
    },
  },
  visual: {
    position: 'relative',
    minHeight: '100dvh',
    padding: 48,
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
    backgroundImage:
      'linear-gradient(90deg, rgba(13, 29, 28, 0.86), rgba(13, 29, 28, 0.18) 58%, rgba(245, 247, 246, 0.15)), url("/images/login-operations.png")',
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    '@media (max-width: 920px)': {
      display: 'none',
    },
  },
  visualBrand: {
    display: 'flex',
    alignItems: 'center',
    gap: 14,
    color: '#f7fbf8',
  },
  mark: {
    width: 44,
    height: 44,
    borderRadius: 8,
    display: 'grid',
    placeItems: 'center',
    background: 'rgba(255, 255, 255, 0.12)',
    border: '1px solid rgba(255, 255, 255, 0.28)',
    boxShadow: '0 18px 50px rgba(0, 0, 0, 0.24)',
    fontSize: 17,
    fontWeight: 700,
    letterSpacing: 0,
  },
  brandText: {
    display: 'flex',
    flexDirection: 'column',
    gap: 3,
  },
  brandName: {
    fontSize: 18,
    lineHeight: 1.2,
    fontWeight: 700,
  },
  brandMeta: {
    fontSize: 13,
    lineHeight: 1.35,
    color: 'rgba(247, 251, 248, 0.68)',
    '@media (max-width: 920px)': {
      color: '#66706c',
    },
  },
  visualCopy: {
    maxWidth: 560,
    color: '#f7fbf8',
  },
  visualTitle: {
    margin: 0,
    fontSize: 40,
    lineHeight: 1.12,
    fontWeight: 760,
    letterSpacing: 0,
  },
  visualLine: {
    width: 72,
    height: 3,
    marginTop: 22,
    borderRadius: 999,
    background: 'linear-gradient(90deg, #77e2d4 0%, #f3b56f 58%, #e87c68 100%)',
  },
  formSide: {
    position: 'relative',
    minHeight: '100dvh',
    padding: '40px 48px',
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'center',
    background:
      'linear-gradient(180deg, rgba(255,255,255,0.96) 0%, rgba(246,248,247,0.98) 100%)',
    '@media (max-width: 920px)': {
      minHeight: '100dvh',
      padding: '40px 24px 72px',
      justifyContent: 'center',
      backgroundImage:
        'linear-gradient(180deg, rgba(246,248,247,0.9), rgba(246,248,247,0.98)), url("/images/login-operations.png")',
      backgroundSize: 'cover',
      backgroundPosition: 'center left',
    },
    '@media (max-width: 520px)': {
      padding: '72px 18px 70px',
      justifyContent: 'flex-start',
    },
  },
  panel: {
    width: '100%',
    maxWidth: 400,
    margin: '0 auto',
    '@media (max-width: 920px)': {
      maxWidth: 420,
    },
  },
  formHeader: {
    marginBottom: 28,
  },
  formTitle: {
    margin: 0,
    fontSize: 30,
    lineHeight: 1.18,
    fontWeight: 760,
    letterSpacing: 0,
    color: '#18211f',
    '@media (max-width: 920px)': {
      fontSize: 28,
    },
    '@media (max-width: 520px)': {
      fontSize: 26,
    },
  },
  formSubtitle: {
    margin: '10px 0 0',
    fontSize: 14,
    lineHeight: 1.7,
    color: '#66706c',
  },
  loginForm: {
    '.ant-pro-form-login-container': {
      width: '100%',
      padding: 0,
    },
    '.ant-pro-form-login-main': {
      width: '100%',
      minWidth: 0,
    },
    '.ant-pro-form-login-header': {
      display: 'none',
    },
    '.ant-input-affix-wrapper': {
      minHeight: 46,
      borderRadius: 8,
      borderColor: '#d9e1de',
      background: '#fff',
      boxShadow: 'none',
    },
    '.ant-input-affix-wrapper:hover': {
      borderColor: '#7ecfc4',
    },
    '.ant-input-affix-wrapper-focused': {
      borderColor: '#34b7aa',
      boxShadow: '0 0 0 3px rgba(52, 183, 170, 0.14)',
    },
    '.ant-input-prefix': {
      marginInlineEnd: 10,
      color: '#7a8581',
    },
    '.ant-btn-primary': {
      minHeight: 46,
      borderRadius: 8,
      background: '#1d5c58',
      boxShadow: '0 14px 30px rgba(29, 92, 88, 0.22)',
      fontWeight: 650,
    },
    '.ant-btn-primary:not(:disabled):not(.ant-btn-disabled):hover': {
      background: '#174d49',
    },
  },
  mobileBrand: {
    display: 'none',
    alignItems: 'center',
    gap: 12,
    marginBottom: 32,
    color: '#18211f',
    '@media (max-width: 920px)': {
      display: 'flex',
    },
  },
  mobileMark: {
    width: 40,
    height: 40,
    borderRadius: 8,
    display: 'grid',
    placeItems: 'center',
    background: '#1d5c58',
    color: '#f7fbf8',
    fontSize: 16,
    fontWeight: 700,
    letterSpacing: 0,
  },
  footer: {
    position: 'absolute',
    right: 48,
    bottom: 22,
    left: 48,
    color: '#7a8581',
    '@media (max-width: 920px)': {
      right: 24,
      left: 24,
      bottom: 18,
    },
  },
}));

const loginPath = '/user/login';

const safeRedirect = (redirect: string | null, fallback: string): string => {
  if (!redirect?.startsWith('/') || redirect.startsWith('//')) return fallback;
  try {
    const parsed = new URL(redirect, window.location.origin);
    if (parsed.origin !== window.location.origin) return fallback;
    if (parsed.pathname === loginPath) return fallback;
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
      <section className={styles.visual} aria-hidden="true">
        <div className={styles.visualBrand}>
          <div className={styles.mark}>EA</div>
          <div className={styles.brandText}>
            <span className={styles.brandName}>Echo Admin</span>
            <span className={styles.brandMeta}>后台管理模板</span>
          </div>
        </div>
        <div className={styles.visualCopy}>
          <h1 className={styles.visualTitle}>统一后台管理入口</h1>
          <div className={styles.visualLine} />
        </div>
      </section>
      <main className={styles.formSide}>
        <div className={styles.panel}>
          <div className={styles.mobileBrand}>
            <div className={styles.mobileMark}>EA</div>
            <div className={styles.brandText}>
              <span className={styles.brandName}>Echo Admin</span>
              <span className={styles.brandMeta}>后台管理模板</span>
            </div>
          </div>
          <div className={styles.formHeader}>
            <h2 className={styles.formTitle}>管理员登录</h2>
            <p className={styles.formSubtitle}>请输入账号和密码。</p>
          </div>
          <LoginForm
            className={styles.loginForm}
            submitter={{
              searchConfig: {
                submitText: '登录',
              },
            }}
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
              placeholder="请输入用户名"
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
        <div className={styles.footer}>
          <Footer />
        </div>
      </main>
    </div>
  );
};

export default Login;
