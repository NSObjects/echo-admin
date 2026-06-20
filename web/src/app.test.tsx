import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { mockHistory, mockQueryCurrentUser, mockReplace } = vi.hoisted(() => {
  const replace = vi.fn();
  return {
    mockHistory: {
      location: {
        pathname: '/dashboard',
        search: '',
        hash: '',
      },
      replace,
    },
    mockQueryCurrentUser: vi.fn(),
    mockReplace: replace,
  };
});

vi.mock('@umijs/max', () => ({
  history: mockHistory,
  Link: ({ children }: { children: ReactNode }) => children,
}));

vi.mock('@/services/admin', () => ({
  currentUser: mockQueryCurrentUser,
}));

vi.mock('@/components', () => ({
  AvatarDropdown: ({ children }: { children: ReactNode }) => children,
  ErrorBoundary: ({ children }: { children: ReactNode }) => children,
  Footer: () => null,
  LangDropdown: () => null,
  OfflineBanner: () => null,
}));

vi.mock('./requestErrorConfig', () => ({
  errorConfig: {},
}));

vi.mock('../config/defaultSettings', () => ({
  default: { title: 'Echo Admin', locale: false },
}));

describe('app getInitialState', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockHistory.location = {
      pathname: '/dashboard',
      search: '',
      hash: '',
    };
  });

  it('loads current administrator outside login page', async () => {
    const { getInitialState } = await import('./app');
    const user = {
      id: 1,
      username: 'admin',
      display_name: '系统管理员',
      email: '',
      active_role_id: 1,
      active_role: {},
      default_path: '/dashboard',
      roles: [],
      permissions: ['admin:read'],
      menus: [],
    };
    mockQueryCurrentUser.mockResolvedValue(user);

    const state = await getInitialState();

    expect(mockQueryCurrentUser).toHaveBeenCalled();
    expect(state.currentUser).toEqual(user);
    expect(state.settings).toEqual({ title: 'Echo Admin', locale: false });
  });

  it('redirects to login when current administrator cannot be loaded', async () => {
    const { getInitialState } = await import('./app');
    mockHistory.location = {
      pathname: '/admins',
      search: '?page=2',
      hash: '#top',
    };
    mockQueryCurrentUser.mockRejectedValue(new Error('unauthorized'));

    const state = await getInitialState();

    expect(state.currentUser).toBeUndefined();
    expect(mockReplace).toHaveBeenCalledWith(
      `/user/login?redirect=${encodeURIComponent('/admins?page=2#top')}`,
    );
  });

  it('does not load current administrator on login page', async () => {
    const { getInitialState } = await import('./app');
    mockHistory.location = {
      pathname: '/user/login',
      search: '',
      hash: '',
    };

    const state = await getInitialState();

    expect(mockQueryCurrentUser).not.toHaveBeenCalled();
    expect(state.currentUser).toBeUndefined();
    expect(state.fetchUserInfo).toBeDefined();
  });
});

describe('app runtime exports', () => {
  it('only exports Umi runtime configuration keys', async () => {
    const app = await import('./app');

    expect(Object.keys(app).sort()).toEqual([
      'getInitialState',
      'layout',
      'request',
      'rootContainer',
    ]);
  });
});

describe('app menu filtering', () => {
  it('keeps only routes granted by backend menus', async () => {
    const { filterMenuDataByGrantedMenus } = await import('./runtime/menu');

    const filtered = filterMenuDataByGrantedMenus(
      [
        { path: '/dashboard', name: 'dashboard' },
        { path: '/admins', name: 'admins' },
        { path: '/roles', name: 'roles' },
      ],
      [
        {
          id: 1,
          parent_id: 0,
          name: '工作台',
          path: '/dashboard',
          icon: 'dashboard',
          hidden: false,
          component: './Dashboard',
          meta: {
            active_name: '',
            keep_alive: false,
            default_menu: false,
            close_tab: false,
            transition_type: '',
          },
          permission: '',
          sort: 10,
          active: true,
          buttons: [],
        },
        {
          id: 2,
          parent_id: 0,
          name: '角色权限',
          path: '/roles',
          icon: 'safety',
          hidden: false,
          component: './Roles',
          meta: {
            active_name: '',
            keep_alive: false,
            default_menu: false,
            close_tab: false,
            transition_type: '',
          },
          permission: 'role:read',
          sort: 20,
          active: true,
          buttons: [],
        },
      ],
    );

    expect(filtered.map((item) => item.path)).toEqual(['/dashboard', '/roles']);
  });

  it('keeps parent menus when a child route is granted', async () => {
    const { filterMenuDataByGrantedMenus } = await import('./runtime/menu');

    const filtered = filterMenuDataByGrantedMenus(
      [
        {
          path: '/system',
          name: 'system',
          routes: [
            { path: '/configs', name: 'configs' },
            { path: '/dictionaries', name: 'dictionaries' },
          ],
        },
      ],
      [
        {
          id: 1,
          parent_id: 0,
          name: '系统配置',
          path: '/configs',
          icon: 'setting',
          hidden: false,
          component: './Configs',
          meta: {
            active_name: '',
            keep_alive: false,
            default_menu: false,
            close_tab: false,
            transition_type: '',
          },
          permission: 'config:read',
          sort: 10,
          active: true,
          buttons: [],
        },
      ],
    );

    expect(filtered).toEqual([
      {
        path: '/system',
        name: 'system',
        routes: [{ path: '/configs', name: 'configs' }],
      },
    ]);
  });

  it('excludes backend hidden menus from layout navigation', async () => {
    const { filterMenuDataByGrantedMenus } = await import('./runtime/menu');

    const filtered = filterMenuDataByGrantedMenus(
      [
        { path: '/dashboard', name: 'dashboard' },
        { path: '/menus', name: 'menus' },
      ],
      [
        {
          id: 1,
          parent_id: 0,
          name: '工作台',
          path: '/dashboard',
          icon: 'dashboard',
          hidden: false,
          component: './Dashboard',
          meta: {
            active_name: '',
            keep_alive: false,
            default_menu: false,
            close_tab: false,
            transition_type: '',
          },
          permission: '',
          sort: 10,
          active: true,
          buttons: [],
        },
        {
          id: 2,
          parent_id: 0,
          name: '菜单管理',
          path: '/menus',
          icon: 'menu',
          hidden: true,
          component: './Menus',
          meta: {
            active_name: '',
            keep_alive: false,
            default_menu: false,
            close_tab: false,
            transition_type: '',
          },
          permission: 'menu:read',
          sort: 20,
          active: true,
          buttons: [],
        },
      ],
    );

    expect(filtered.map((item) => item.path)).toEqual(['/dashboard']);
  });
});
