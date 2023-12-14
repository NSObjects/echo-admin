// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 查询部门 GET /api/departments */
export async function getDepartments(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getDepartmentsParams,
  options?: { [key: string]: any },
) {
  return request<API.listDepartmentsResp>('/api/departments', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 创建部门 POST /api/departments */
export async function postDepartments(body: API.department, options?: { [key: string]: any }) {
  return request<{ code: number; msg: string }>('/api/departments', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 查询部门详情 GET /api/departments/${param0} */
export async function getDepartmentsId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getDepartmentsIdParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.departmentResp>(`/api/departments/${param0}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** 更新部门信息 PUT /api/departments/${param0} */
export async function putDepartmentsId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.putDepartmentsIdParams,
  body: API.department,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<{ code: number; msg: string }>(`/api/departments/${param0}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** 删除部门 DELETE /api/departments/${param0} */
export async function deleteDepartmentsId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteDepartmentsIdParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.success>(`/api/departments/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}
