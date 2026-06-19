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
    access: 'canAdminRead',
    component: './Admins',
  },
  {
    path: '/roles',
    name: 'roles',
    icon: 'safety',
    access: 'canRoleRead',
    component: './Roles',
  },
  {
    path: '/menus',
    name: 'menus',
    icon: 'menu',
    access: 'canMenuRead',
    component: './Menus',
  },
  {
    path: '/configs',
    name: 'configs',
    icon: 'setting',
    access: 'canConfigRead',
    component: './Configs',
  },
  {
    path: '/dictionaries',
    name: 'dictionaries',
    icon: 'profile',
    access: 'canDictRead',
    component: './Dictionaries',
  },
  {
    path: '/files',
    name: 'files',
    icon: 'upload',
    access: 'canFileRead',
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
