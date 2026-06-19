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
    canAdminRead: can('admin:read'),
    canAdminCreate: can('admin:create'),
    canAdminUpdate: can('admin:update'),
    canAdminDelete: can('admin:delete'),
    canRoleRead: can('role:read'),
    canRoleCreate: can('role:create'),
    canRoleUpdate: can('role:update'),
    canRoleDelete: can('role:delete'),
    canMenuRead: can('menu:read'),
    canMenuCreate: can('menu:create'),
    canMenuUpdate: can('menu:update'),
    canMenuDelete: can('menu:delete'),
    canConfigRead: can('config:read'),
    canConfigUpdate: can('config:update'),
    canDictRead: can('dict:read'),
    canDictCreate: can('dict:create'),
    canDictUpdate: can('dict:update'),
    canDictDelete: can('dict:delete'),
    canFileRead: can('file:read'),
    canFileUpload: can('file:upload'),
    canLogRead: can('log:read'),
  };
}
