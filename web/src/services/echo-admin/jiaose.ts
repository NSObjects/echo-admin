// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';
const basePath = "/api"
/** 查询角色信息 GET /roles */
export async function getRoles(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getRolesParams,
  options?: { [key: string]: any },
) {
  return request<API.listRoleResp>(`${basePath}/roles`, {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 创建角色 POST /roles */
export async function postRoles(body: API.role, options?: { [key: string]: any }) {
  return request<API.success>(`${basePath}/roles`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 更新角色信息 PUT /roles/${param0} */
export async function putRolesId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.putRolesIdParams,
  body: API.role,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`${basePath}/roles/${param0}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** 删除角色 DELETE /roles/${param0} */
export async function deleteRolesId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteRolesIdParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`${basePath}/roles/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** 更新角色菜单 PUT /roles/${param0}/menus */
export async function putRolesIdMenus(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.putRolesIdMenusParams,
  body: {
    menu_id?: number[];
    creator?: string;
  },
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`${basePath}/roles/${param0}/menus`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}
