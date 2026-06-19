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
  it('allows admin delete when the permission is granted', () => {
    const result = access({
      currentUser: {
        ...baseUser,
        permissions: ['admin:delete'],
      },
    });

    expect(result.canAdminDelete).toBe(true);
  });

  it('denies admin management without the permission', () => {
    const result = access({
      currentUser: {
        ...baseUser,
        permissions: ['role:delete', 'dict:delete'],
      },
    });

    expect(result.canAdminRead).toBe(false);
    expect(result.canRoleDelete).toBe(true);
    expect(result.canMenuDelete).toBe(false);
    expect(result.canDictDelete).toBe(true);
  });

  it('denies admin management without a current user', () => {
    expect(access(undefined).canAdminRead).toBe(false);
  });
});
