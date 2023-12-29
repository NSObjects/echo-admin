// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 查询当前用户菜单 GET /api/user/menus */
export async function getUserMenus(options?: { [key: string]: any }) {
  return request<API.listMenuResp>('/api/user/menus', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 查询用户 GET /api/users */
export async function getUsers(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getUsersParams,
  options?: { [key: string]: any },
) {
  return request<API.listUserResp>('/api/users', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 创建用户 POST /api/users */
export async function postUsers(body: API.user, options?: { [key: string]: any }) {
  return request<API.success>('/api/users', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 查询用户详情 GET /api/users/${param0} */
export async function getUsersId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getUsersIdParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.userResp>(`/api/users/${param0}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** 更新用户信息 PUT /api/users/${param0} */
export async function putUsersId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.putUsersIdParams,
  body: API.user,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`/api/users/${param0}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** 删除用户信息 DELETE /api/users/${param0} */
export async function deleteUsersId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteUsersIdParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`/api/users/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** 查询当前用户 GET /api/users/current */
export async function getUsersCurrent(options?: { [key: string]: any }) {
  return request<API.userResp>('/api/users/current', {
    method: 'GET',
    ...(options || {}),
  });
}
