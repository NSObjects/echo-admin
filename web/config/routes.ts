export default [
  {
    path: '/user',
    layout: false,
    routes: [
      {
        path: '/user/login',
        name: 'login',
        component: './user/login',
      },
      {
        path: '/user',
        redirect: '/user/login',
      },
    ],
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    icon: 'dashboard',
    component: './Dashboard',
  },
  {
    path: '/admins',
    name: 'admins',
    icon: 'user',
    access: 'canAdminManage',
    component: './Admins',
  },
  {
    path: '/roles',
    name: 'roles',
    icon: 'safety',
    access: 'canRoleManage',
    component: './Roles',
  },
  {
    path: '/menus',
    name: 'menus',
    icon: 'menu',
    access: 'canMenuManage',
    component: './Menus',
  },
  {
    path: '/configs',
    name: 'configs',
    icon: 'setting',
    access: 'canConfigManage',
    component: './Configs',
  },
  {
    path: '/dictionaries',
    name: 'dictionaries',
    icon: 'profile',
    access: 'canDictManage',
    component: './Dictionaries',
  },
  {
    path: '/files',
    name: 'files',
    icon: 'upload',
    access: 'canFileUpload',
    component: './Files',
  },
  {
    path: '/logs',
    name: 'logs',
    icon: 'fileSearch',
    access: 'canLogRead',
    component: './Logs',
  },
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    component: './exception/404',
    path: '/*',
  },
];
