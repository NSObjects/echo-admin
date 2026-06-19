import type { CurrentUser } from '@/services/admin';

/**
 * @see https://umijs.org/docs/max/access#access
 * */
export default function access(
  initialState: { currentUser?: CurrentUser } | undefined,
) {
  const { currentUser } = initialState ?? {};
  const can = (permission: string) =>
    currentUser?.permissions.includes(permission) ?? false;

  return {
    canAdminManage: can('admin:manage'),
    canRoleManage: can('role:manage'),
    canMenuManage: can('menu:manage'),
    canConfigManage: can('config:manage'),
    canDictManage: can('dict:manage'),
    canFileUpload: can('file:upload'),
    canLogRead: can('log:read'),
  };
}
