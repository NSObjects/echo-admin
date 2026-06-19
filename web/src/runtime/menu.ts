import type { Menu } from '@/services/admin';

type NavigationItem = {
  path?: string;
  children?: NavigationItem[];
  routes?: NavigationItem[];
  [key: string]: unknown;
};

const normalizeMenuPath = (path?: string): string => {
  if (!path) return '';
  return path.length > 1 && path.endsWith('/') ? path.slice(0, -1) : path;
};

// The backend owns menu visibility; the static route table only registers pages.
export function filterMenuDataByGrantedMenus(
  menuData: NavigationItem[],
  menus: Menu[],
): NavigationItem[] {
  const allowedPaths = new Set(
    menus
      .filter((menu) => menu.active)
      .map((menu) => normalizeMenuPath(menu.path))
      .filter(Boolean),
  );
  if (allowedPaths.size === 0) return [];
  return filterNavigationItems(menuData, allowedPaths);
}

function filterNavigationItems(
  items: NavigationItem[] | undefined,
  allowedPaths: Set<string>,
): NavigationItem[] {
  if (!items) return [];
  const out: NavigationItem[] = [];
  for (const item of items) {
    const children = filterNavigationItems(item.children, allowedPaths);
    const routes = filterNavigationItems(item.routes, allowedPaths);
    const path = normalizeMenuPath(item.path);
    if (!allowedPaths.has(path) && children.length === 0 && routes.length === 0) {
      continue;
    }
    out.push({
      ...item,
      ...(item.children ? { children } : {}),
      ...(item.routes ? { routes } : {}),
    });
  }
  return out;
}
