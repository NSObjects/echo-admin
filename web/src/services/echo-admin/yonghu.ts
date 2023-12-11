// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 查询用户 GET /api/users */
export async function getApiUsers(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getApiUsersParams,
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
export async function postApiUsers(body: API.user, options?: { [key: string]: any }) {
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
export async function getApiUsersId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getApiUsersIdParams,
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
export async function putApiUsersId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.putApiUsersIdParams,
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
export async function deleteApiUsersId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteApiUsersIdParams,
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
export async function getApiUsersCurrent(options?: { [key: string]: any }) {
  return request<API.userResp>('/api/users/current', {
    method: 'GET',
    ...(options || {}),
  });
}
