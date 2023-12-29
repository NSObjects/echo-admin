// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** api路径 GET /api/api */
export async function getApi(options?: { [key: string]: any }) {
  return request<API.apiResp>('/api/api', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 登录 POST /api/login/account */
export async function postLoginAccount(body: API.account, options?: { [key: string]: any }) {
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
export async function postLoginOut(options?: { [key: string]: any }) {
  return request<API.success>('/api/login/out', {
    method: 'POST',
    ...(options || {}),
  });
}
