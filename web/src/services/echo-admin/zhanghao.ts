// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** api路径 GET /api/api */
export async function getApiApi(options?: { [key: string]: any }) {
  return request<{ method?: string; path?: string; name?: string }[]>('/api/api', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 登录 POST /api/login/account */
export async function postApiLoginAccount(body: API.account, options?: { [key: string]: any }) {
  return request<API.login>('/api/login/account', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 登出 POST /api/login/out */
export async function postApiLoginOut(options?: { [key: string]: any }) {
  return request<API.success>('/api/login/out', {
    method: 'POST',
    ...(options || {}),
  });
}
