import { request } from '@umijs/max';

import { clearAuthToken, setAuthToken } from './auth-token';

export type Envelope<T> = {
  code: number;
  message: string;
  data: T;
  page?: PageMeta;
};

export type PageMeta = {
  page: number;
  page_size: number;
  total: number;
  has_next: boolean;
};

export type ListParams = {
  page?: number;
  page_size?: number;
};

export type LoginResult = {
  token: string;
  user: CurrentUser;
};

export type AdminUser = {
  id: number;
  username: string;
  display_name: string;
  email: string;
  role_ids: number[];
  active_role_id: number;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type CurrentUser = {
  id: number;
  username: string;
  display_name: string;
  email: string;
  active_role_id: number;
  active_role: Role;
  default_path: string;
  roles: Role[];
  permissions: string[];
  menus: Menu[];
};

export type Role = {
  id: number;
  parent_id: number;
  code: string;
  name: string;
  permissions: string[];
  menu_ids: number[];
  default_path: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type PermissionDefinition = {
  token: string;
  resource: string;
  action: string;
  name: string;
};

export type Menu = {
  id: number;
  parent_id: number;
  name: string;
  path: string;
  icon: string;
  permission: string;
  sort: number;
  active: boolean;
};

export type SystemConfig = {
  key: string;
  name: string;
  value: string;
  public: boolean;
  updated_at: string;
};

export type Dictionary = {
  id: number;
  code: string;
  name: string;
  items: DictionaryItem[];
};

export type DictionaryItem = {
  id: number;
  label: string;
  value: string;
  sort: number;
  active: boolean;
};

export type FileObject = {
  id: number;
  name: string;
  url: string;
  size: number;
  content_type: string;
  created_at: string;
};

export type OperationLog = {
  id: number;
  actor_id: number;
  action: string;
  resource: string;
  resource_id: string;
  method: string;
  path: string;
  ip: string;
  success: boolean;
  message: string;
  created_at: string;
};

export type LoginLog = {
  id: number;
  admin_id: number;
  username: string;
  ip: string;
  success: boolean;
  reason: string;
  created_at: string;
};

export type AdminCreateInput = {
  username: string;
  display_name: string;
  email?: string;
  password: string;
  role_ids: number[];
  active_role_id?: number;
  active: boolean;
};

export type AdminUpdateInput = {
  display_name?: string;
  email?: string;
  password?: string;
  role_ids?: number[];
  active_role_id?: number;
  active?: boolean;
};

export type RoleCreateInput = {
  parent_id: number;
  code: string;
  name: string;
  permissions: string[];
  menu_ids?: number[];
  default_path?: string;
  active: boolean;
};

export type RoleUpdateInput = {
  parent_id?: number;
  name?: string;
  permissions?: string[];
  menu_ids?: number[];
  default_path?: string;
  active?: boolean;
};

export type MenuInput = {
  parent_id: number;
  name: string;
  path: string;
  icon?: string;
  permission?: string;
  sort: number;
  active: boolean;
};

export type ConfigInput = {
  name: string;
  value: string;
  public: boolean;
};

export type DictionaryCreateInput = {
  code: string;
  name: string;
};

export type DictionaryItemInput = {
  label: string;
  value: string;
  sort: number;
  active: boolean;
};

export async function login(body: {
  username: string;
  password: string;
}): Promise<LoginResult> {
  const response = await request<Envelope<LoginResult>>('/api/auth/login', {
    method: 'POST',
    data: body,
  });
  setAuthToken(response.data.token);
  return response.data;
}

export async function logout(): Promise<void> {
  clearAuthToken();
}

export async function currentUser(): Promise<CurrentUser> {
  const response = await request<Envelope<CurrentUser>>('/api/auth/me', {
    method: 'GET',
  });
  return response.data;
}

export async function switchRole(roleID: number): Promise<LoginResult> {
  const response = await request<Envelope<LoginResult>>('/api/auth/role', {
    method: 'POST',
    data: { role_id: roleID },
  });
  setAuthToken(response.data.token);
  return response.data;
}

export async function listAdmins(
  params?: ListParams,
): Promise<Envelope<AdminUser[]>> {
  return request<Envelope<AdminUser[]>>('/api/admins', { params });
}

export async function createAdmin(body: AdminCreateInput): Promise<void> {
  await request('/api/admins', { method: 'POST', data: body });
}

export async function updateAdmin(
  id: number,
  body: AdminUpdateInput,
): Promise<void> {
  await request(`/api/admins/${id}`, { method: 'PATCH', data: body });
}

export async function listRoles(
  params?: ListParams,
): Promise<Envelope<Role[]>> {
  return request<Envelope<Role[]>>('/api/roles', { params });
}

export async function listPermissions(): Promise<PermissionDefinition[]> {
  const response =
    await request<Envelope<PermissionDefinition[]>>('/api/permissions');
  return response.data;
}

export async function createRole(body: RoleCreateInput): Promise<void> {
  await request('/api/roles', { method: 'POST', data: body });
}

export async function updateRole(
  id: number,
  body: RoleUpdateInput,
): Promise<void> {
  await request(`/api/roles/${id}`, { method: 'PATCH', data: body });
}

export async function listMenus(): Promise<Menu[]> {
  const response = await request<Envelope<Menu[]>>('/api/menus');
  return response.data;
}

export async function createMenu(body: MenuInput): Promise<void> {
  await request('/api/menus', { method: 'POST', data: body });
}

export async function updateMenu(id: number, body: MenuInput): Promise<void> {
  await request(`/api/menus/${id}`, { method: 'PATCH', data: body });
}

export async function listConfigs(): Promise<SystemConfig[]> {
  const response = await request<Envelope<SystemConfig[]>>('/api/system/configs');
  return response.data;
}

export async function upsertConfig(key: string, body: ConfigInput): Promise<void> {
  await request(`/api/system/configs/${encodeURIComponent(key)}`, {
    method: 'PUT',
    data: body,
  });
}

export async function listDictionaries(): Promise<Dictionary[]> {
  const response = await request<Envelope<Dictionary[]>>('/api/dictionaries');
  return response.data;
}

export async function createDictionary(
  body: DictionaryCreateInput,
): Promise<void> {
  await request('/api/dictionaries', { method: 'POST', data: body });
}

export async function addDictionaryItem(
  code: string,
  body: DictionaryItemInput,
): Promise<void> {
  await request(`/api/dictionaries/${encodeURIComponent(code)}/items`, {
    method: 'POST',
    data: body,
  });
}

export async function updateDictionaryItem(
  code: string,
  itemID: number,
  body: DictionaryItemInput,
): Promise<void> {
  await request(
    `/api/dictionaries/${encodeURIComponent(code)}/items/${itemID}`,
    { method: 'PATCH', data: body },
  );
}

export async function listFiles(
  params?: ListParams,
): Promise<Envelope<FileObject[]>> {
  return request<Envelope<FileObject[]>>('/api/files', { params });
}

export async function uploadFile(file: File): Promise<FileObject> {
  const body = new FormData();
  body.append('file', file);
  const response = await request<Envelope<FileObject>>('/api/files', {
    method: 'POST',
    data: body,
  });
  return response.data;
}

export async function listOperationLogs(
  params?: ListParams,
): Promise<Envelope<OperationLog[]>> {
  return request<Envelope<OperationLog[]>>('/api/logs/operations', { params });
}

export async function listLoginLogs(
  params?: ListParams,
): Promise<Envelope<LoginLog[]>> {
  return request<Envelope<LoginLog[]>>('/api/logs/logins', { params });
}
