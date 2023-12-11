// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';
const basePath = "/api"
/** 查询菜单 GET /menus */
export async function getMenus(options?: { [key: string]: any }) {
  return request<API.listMenuResp>(`${basePath}/menus`, {
    method: 'GET',
    ...(options || {}),
  });
}

/** 创建菜单 POST /menus */
export async function postMenus(body: {}, options?: { [key: string]: any }) {
  return request<API.success>(`${basePath}/menus`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 更新菜单 PUT /menus/${param0} */
export async function putMenusId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.putMenusIdParams,
  body: {},
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`${basePath}/menus/${param0}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** 删除菜单 DELETE /menus/${param0} */
export async function deleteMenusId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteMenusIdParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`${basePath}/menus/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}
