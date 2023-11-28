// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';
const basePath = '/api';
/** api路径 GET /api */
export async function getApi(options?: { [key: string]: any }) {
  return request<{ method?: string; path?: string; name?: string }[]>(`${basePath}/api`, {
    method: 'GET',
    ...(options || {}),
  });
}

/** 登录 POST /login/account */
export async function postLoginAccount(body: API.account, options?: { [key: string]: any }) {
  return request<API.login>(`${basePath}/login/account`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 登出 POST /login/out */
export async function postLoginOut(options?: { [key: string]: any }) {
  return request<API.success>(`${basePath}/login/out`, {
    method: 'POST',
    ...(options || {}),
  });
}
