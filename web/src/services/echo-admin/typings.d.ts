declare namespace API {
  type account = {
    account: string;
    password: string;
    type: string;
  };

  type deleteDepartmentsIdParams = {
    id: number;
  };

  type deleteMenusIdParams = {
    id: number;
  };

  type deleteRolesIdParams = {
    id: number;
  };

  type deleteUsersIdParams = {
    id: number;
  };

  type department = {
    name: string;
    parent_id?: number;
    email?: string;
    phone?: string;
    status?: number;
    sort?: number;
    principal?: string;
    created_at?: string;
    updated_at?: string;
    id?: number;
    children?: department[];
  };

  type departmentResp = {
    code: number;
    msg: string;
    data: department;
  };

  type getDepartmentsIdParams = {
    id: number;
  };

  type getDepartmentsParams = {
    page?: number;
    count?: number;
    name?: string;
    status?: number;
  };

  type getRolesParams = {
    name?: string;
    identify?: string;
    state?: number;
    start_date?: number;
    end_date?: number;
    page?: number;
    count?: number;
  };

  type getUsersIdParams = {
    id: number;
  };

  type getUsersParams = {
    name: string;
    phone: string;
    page?: number;
    count?: number;
  };

  type listDepartmentsResp = {
    code: number;
    msg: string;
    data: { total?: number; list?: department[] };
  };

  type listMenuResp = {
    code: number;
    msg: string;
    data: { total?: number; list?: menu[] };
  };

  type listResp = {
    code: number;
    msg: string;
    data: { total?: number; list?: Record<string, any>[] };
  };

  type listRoleResp = {
    code: number;
    msg: string;
    data: { total?: number; list?: role[] };
  };

  type listUserResp = {
    code: number;
    msg: string;
    data: { total?: number; list?: user[] };
  };

  type login = {
    code: number;
    msg: string;
    data: { token?: string; type?: string };
  };

  type menu = {
    /** 父菜单id */
    pid: number;
    /** 菜单id */
    id?: number;
    /** 类型 1=目录 2=菜单 3=按钮 */
    type: 1 | 2 | 3;
    /** 菜单名称 */
    name: string;
    /** api接口 */
    api: string;
    /** 路由路径 */
    path: string;
    /** 组件路径 */
    component: string;
    layout?: number;
    /** 重定向 */
    redirect?: string;
    /** 排序 */
    sort?: number;
    /** 状态 1=启用 2=禁用 */
    status?: 1 | 2;
    /** 图标 */
    icon?: string;
    /** 外链地址s */
    link?: string;
    /** 备注 */
    remark?: string;
    /** 是否隐藏 1=是 2=否 */
    hidden?: number;
    /** 是否缓存 1=是 2=否 */
    cache?: number;
    /** 是否固定 1=是 2=否 */
    fixed?: number;
    label?: string;
    value?: number;
    /** 菜单标识符 */
    identify?: string;
    /** 角色id列表 */
    role?: number[];
    children?: menu[];
  };

  type putDepartmentsIdParams = {
    id: number;
  };

  type putMenu = {
    /** 父菜单id */
    pid?: number;
    /** 类型 1=目录 2=菜单 3=按钮 */
    type?: 1 | 2 | 3;
    /** 菜单名称 */
    name?: string;
    /** api接口 */
    api?: string;
    /** 路由路径 */
    path?: string;
    /** 组件路径 */
    component?: string;
    layout?: number;
    /** 重定向 */
    redirect?: string;
    /** 排序 */
    sort?: number;
    /** 状态 1=启用 2=禁用 */
    status?: 1 | 2;
    /** 图标 */
    icon?: string;
    /** 外链地址s */
    link?: string;
    /** 备注 */
    remark?: string;
    /** 是否隐藏 1=是 2=否 */
    hidden?: number;
    /** 是否缓存 1=是 2=否 */
    cache?: number;
    /** 是否固定 1=是 2=否 */
    fixed?: number;
    /** 菜单标识符 */
    identify?: string;
    /** 角色id列表 */
    role?: number[];
  };

  type putMenusIdParams = {
    id: number;
  };

  type putRolesIdMenusParams = {
    id: number;
  };

  type putRolesIdParams = {
    id: number;
  };

  type putUsersIdParams = {
    id: number;
  };

  type resp = {
    code: number;
    msg: string;
    data?: Record<string, any>;
  };

  type role = {
    name: string;
    order: number;
    identify: string;
    state: number;
  };

  type roleResp = {
    code: number;
    msg: string;
    data?: role;
  };

  type success = {
    code: number;
    msg: string;
  };

  type user = {
    /** 昵称 */
    name?: string;
    /** 手机号码 */
    phone?: string;
    /** 状态 */
    status?: number;
    /** 密码 */
    password?: string;
    /** 账号 */
    account?: string;
    /** 头像 */
    avatar?: string;
    /** 角色id */
    role_id?: number;
    /** 部门Id */
    department_id?: number;
    id?: number;
    /** 邮箱 */
    email?: string;
    /** 性别 */
    sex?: 1 | 2;
    /** 岗位 */
    posts?: string;
  };

  type userResp = {
    code: number;
    msg: string;
    data?: user;
  };
}
