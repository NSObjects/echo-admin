
import { request } from '@umijs/max';

/** 查询菜单 GET /api/menus */
export async function getMenus(

  options?: { [key: string]: any },
) {
  return request<API.FakeCaptcha>('/api/login/captcha', {
    method: 'GET',
    ...(options || {}),
  });
}
