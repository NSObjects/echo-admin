declare namespace API {
  type account = {
    account: string;
    password: string;
    type: string;
  };

  type deleteApiDepartmentsIdParams = {
    id: number;
  };

  type deleteApiMenusIdParams = {
    id: string;
  };

  type deleteApiRolesIdParams = {
    id: string;
  };

  type deleteApiUsersIdParams = {
    id: string;
  };

  // type department = {
  //   id : number;
  //   name: string;
  //   parent_id?: number;
  //   email?: string;
  //   phone?: string;
  //   status?: number;
  //   sort?: number;
  //   principal?: string;
  //   created_at?: string;
  //   updated_at?: string;
  // };

  type departmentResp = {
    code: number;
    msg: string;
    data: department;
  };

  type getApiDepartmentsIdParams = {
    id?: number;

  };

  type getApiDepartmentsParams = {
    page?: number;
    count?: number;
    name?: string;
    email?: string,
    parent_id?: number,
    phone?:string,
    principal?:string,
    sort?: string,
    status?: number,
  };

  type getApiRolesParams = {
    name?: string;
    identify?: string;
    state?: number;
    start_date?: number;
    end_date?: number;
    page?: number;
    count?: number;
  };

  type getApiUsersIdParams = {
    id: number;
  };

  type getApiUsersParams = {
    name: string;
    phone: string;
    page?: number;
    count?: number;
  };

  type listDepartmentsResp = {
    code: number;
    msg: string;
    data: {
      total?: number;
      list: department[];
    };
  };

  type department = {
    id: number;
    name: string;
    status?: number;
    sort?: number;
    principal?: string;
    phone?: string;
    email?: string;
    created_at?: string;
    updated_at?: string;
    children?: department[];
  }

  type listMenuResp = {
    code: number;
    msg: string;
    data: {
      list:menuData[],
      total: number
    };
  };

  type menuData = {
    component?: string;
    parent_id?: number;
    layout?: boolean;
    path?: string;
    name?: string;
    redirect?: string;
    api?: string;
    sort?: number;
    type?: 1 | 2 | 3;
    status?: 1 | 2;
    children: menu[];
  }

  type listResp = {
    code: number;
    msg: string;
    data: { total?: number; list?: Record<string, any>[] };
  };

  type listRoleResp = {
    code: number;
    msg: string;
    data: { total?: number; list: role[] };
  };

  type listUserResp = {
    code: number;
    msg: string;
    data: { total?: number; list?: user[] };
  };

  type login = {
    code: number;
    msg: string;
    data: { token: string; type?: string };
  };

  type menu = {
    /** 组件路径 */
    component: string;
    /** 父菜单id */
    parent_id: number;
    layout: boolean;
    /** 路由路径 */
    path: string;
    /** 菜单名称 */
    name: string;
    redirect: string;
    /** api接口 */
    api: string;
    /** 排序 */
    sort: number;
    /** 类型 1=目录 2=菜单 3=按钮 */
    type: 1 | 2 | 3;
    /** 状态 1=启用 2=禁用 */
    status: 1 | 2;
  };

  type putApiDepartmentsIdParams = {
    id: number;
  };

  type putApiRolesIdMenusParams = {
    id: string;
  };

  type putApiRolesIdParams = {
    id: string;
  };

  type putApiUsersIdParams = {
    id: string;
  };

  type putMenusIdParams = {
    id: string;
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
    name: string;
    phone: string;
    status?: number;
    password: string;
    account: string;
    avatar?: string;
    role_id?: number;
    department_id?: number;
  };

  type userResp = {
    code: number;
    msg: string;
    data?: user;
  };
}
