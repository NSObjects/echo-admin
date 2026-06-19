import { describe, expect, it } from 'vitest';
import access from './access';

const baseUser = {
  id: 1,
  username: 'admin',
  display_name: '系统管理员',
  email: '',
  roles: [],
  menus: [],
};

describe('access', () => {
  it('allows admin management when the permission is granted', () => {
    const result = access({
      currentUser: {
        ...baseUser,
        permissions: ['admin:manage'],
      },
    });

    expect(result.canAdminManage).toBe(true);
  });

  it('denies admin management without the permission', () => {
    const result = access({
      currentUser: {
        ...baseUser,
        permissions: ['role:manage'],
      },
    });

    expect(result.canAdminManage).toBe(false);
    expect(result.canRoleManage).toBe(true);
  });

  it('denies admin management without a current user', () => {
    expect(access(undefined).canAdminManage).toBe(false);
  });
});
