// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 查询菜单 GET /api/menus */
export async function getMenus(options?: { [key: string]: any }) {
  return request<API.listMenuResp>('/api/menus', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 创建菜单 POST /api/menus */
export async function postMenus(
  body: API.putMenu,
  options?: {
    [key: string]: any;
  },
) {
  return request<API.success>('/api/menus', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 更新菜单 PUT /api/menus/${param0} */
export async function putMenusId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.putMenusIdParams,
  body: API.putMenu,
  options?: {
    [key: string]: any;
  },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`/api/menus/${param0}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** 删除菜单 DELETE /api/menus/${param0} */
export async function deleteMenusId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteMenusIdParams,
  options?: {
    [key: string]: any;
  },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`/api/menus/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}
