import { describe, expect, it } from 'vitest';
import access from './access';

const baseUser = {
  id: 1,
  username: 'admin',
  display_name: '系统管理员',
  email: '',
  active_role_id: 1,
  active_role: {} as never,
  default_path: '/dashboard',
  roles: [],
  menus: [],
};

describe('access', () => {
  it('allows admin management when the permission is granted', () => {
    const result = access({
      currentUser: {
        ...baseUser,
        permissions: ['admin:read'],
      },
    });

    expect(result.canAdminRead).toBe(true);
  });

  it('denies admin management without the permission', () => {
    const result = access({
      currentUser: {
        ...baseUser,
        permissions: ['role:read'],
      },
    });

    expect(result.canAdminRead).toBe(false);
    expect(result.canRoleRead).toBe(true);
  });

  it('denies admin management without a current user', () => {
    expect(access(undefined).canAdminRead).toBe(false);
  });
});
