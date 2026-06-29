import type { RequestOptions } from '@@/plugin-request/request';
import type { RequestConfig } from '@umijs/max';
import { history } from '@umijs/max';
import { message } from 'antd';

import { getCSRFToken } from '@/services/csrf-token';

type ApiEnvelope = {
  code?: number;
  message?: string;
  data?: unknown;
};

type RequestError = Error & {
  response?: { status?: number };
  info?: ApiEnvelope;
};

const successCode = 100001;
const loginPath = '/user/login';
const csrfHeader = 'X-CSRF-Token';
const unsafeMethods = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

export const errorConfig: RequestConfig = {
  errorConfig: {
    errorThrower: (response) => {
      const envelope = response as ApiEnvelope;
      if (typeof envelope.code === 'number' && envelope.code !== successCode) {
        const error = new Error(envelope.message ?? 'Request failed') as RequestError;
        error.name = 'BizError';
        error.info = envelope;
        throw error;
      }
    },
    errorHandler: (error: RequestError, opts) => {
      if (opts?.skipErrorHandler) throw error;

      if (error.response?.status === 401) {
        if (history.location.pathname !== loginPath) {
          history.replace(
            `${loginPath}?redirect=${encodeURIComponent(
              history.location.pathname + history.location.search + history.location.hash,
            )}`,
          );
        }
        return;
      }

      if (error.name === 'BizError') {
        message.error(error.info?.message ?? 'Request failed');
        return;
      }

      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        message.error('网络不可用，请检查连接后重试');
        return;
      }
      message.error(error.message || '请求失败，请稍后重试');
    },
  },
  requestInterceptors: [
    (config: RequestOptions) => {
      const method = String(config.method ?? 'GET').toUpperCase();
      const csrfToken = unsafeMethods.has(method) ? getCSRFToken() : '';
      return {
        ...config,
        withCredentials: true,
        headers: csrfToken
          ? {
              ...config.headers,
              [csrfHeader]: csrfToken,
            }
          : config.headers,
      };
    },
  ],
  responseInterceptors: [],
};
