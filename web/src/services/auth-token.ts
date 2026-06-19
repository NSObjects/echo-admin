const authTokenKey = 'echo_admin_token';

export function getAuthToken(): string {
  if (typeof window === 'undefined') return '';
  return window.localStorage.getItem(authTokenKey) ?? '';
}

export function setAuthToken(token: string): void {
  window.localStorage.setItem(authTokenKey, token);
}

export function clearAuthToken(): void {
  window.localStorage.removeItem(authTokenKey);
}
