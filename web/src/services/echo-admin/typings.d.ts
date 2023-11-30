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
    id: string;
  };

  type deleteRolesIdParams = {
    id: string;
  };

  type deleteUsersIdParams = {
    id: string;
  };

  type department = {
    name: string;
    parent_id: number;
    email: string;
    phone: string;
    status: number;
    sort: number;
    principal_id: number;
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
    data: menu[];
  };

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
    data: { token: string; type: string };
  };

  type menu = {
    component: string;
    parent_id: number;
    layout: boolean;
    path: string;
    name: string;
    redirect: string;
  };

  type putDepartmentsIdParams = {
    id: number;
  };

  type putMenusIdParams = {
    id: string;
  };

  type putRolesIdMenusParams = {
    id: string;
  };

  type putRolesIdParams = {
    id: string;
  };

  type putUsersIdParams = {
    id: string;
  };

  type resp = {
    code: number;
    msg: string;
    data?: Record<string, any>;
  };

  type role = {
    name: string;
    order: string;
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
    status: number;
    password: string;
    account: string;
    avatar: string;
  };

  type userResp = {
    code: number;
    msg: string;
    data?: user;
  };
}
