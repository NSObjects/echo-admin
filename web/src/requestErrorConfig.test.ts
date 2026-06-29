import { message } from 'antd';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { errorConfig } from './requestErrorConfig';

const { mockGetCSRFToken, mockReplace } = vi.hoisted(() => ({
  mockGetCSRFToken: vi.fn(),
  mockReplace: vi.fn(),
}));

vi.mock('antd', () => ({
  message: {
    error: vi.fn(),
  },
}));

vi.mock('@umijs/max', () => ({
  history: {
    location: {
      pathname: '/admins',
      search: '?page=1',
      hash: '#top',
    },
    replace: mockReplace,
  },
}));

vi.mock('@/services/csrf-token', () => ({
  getCSRFToken: mockGetCSRFToken,
}));

type TestError = Error & {
  response?: { status?: number };
  info?: { message?: string };
};

describe('requestErrorConfig', () => {
  const errorThrower = errorConfig.errorConfig?.errorThrower;
  const errorHandler = errorConfig.errorConfig?.errorHandler;
  const interceptor = errorConfig.requestInterceptors?.[0] as (config: {
    method?: string;
    headers?: Record<string, string>;
  }) => { headers?: Record<string, string>; withCredentials?: boolean };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCSRFToken.mockReturnValue('');
  });

  it('throws BizError for non-success API envelopes', () => {
    expect(() =>
      errorThrower?.({ code: 400001, message: 'invalid input' }),
    ).toThrow('invalid input');
  });

  it('accepts success API envelopes', () => {
    expect(() =>
      errorThrower?.({ code: 100001, message: 'OK', data: {} }),
    ).not.toThrow();
  });

  it('redirects on unauthorized responses', () => {
    const error: TestError = new Error('unauthorized');
    error.response = { status: 401 };

    errorHandler?.(error, {});

    expect(mockReplace).toHaveBeenCalledWith(
      `/user/login?redirect=${encodeURIComponent('/admins?page=1#top')}`,
    );
  });

  it('shows business error messages', () => {
    const error: TestError = new Error('bad request');
    error.name = 'BizError';
    error.info = { message: 'bad request' };

    errorHandler?.(error, {});

    expect(message.error).toHaveBeenCalledWith('bad request');
  });

  it('rethrows when the request opts out of shared handling', () => {
    const error = new Error('boom');

    expect(() => errorHandler?.(error, { skipErrorHandler: true })).toThrow(
      'boom',
    );
  });

  it('includes credentials and csrf header for unsafe requests', () => {
    mockGetCSRFToken.mockReturnValue('csrf-1');

    const result = interceptor({
      method: 'POST',
      headers: { Accept: 'application/json' },
    });

    expect(result.withCredentials).toBe(true);
    expect(result.headers).toEqual({
      Accept: 'application/json',
      'X-CSRF-Token': 'csrf-1',
    });
  });

  it('includes credentials without csrf header for safe requests', () => {
    const config = { headers: { Accept: 'application/json' } };
    const result = interceptor(config);

    expect(result).not.toBe(config);
    expect(result.withCredentials).toBe(true);
    expect(result.headers).toEqual({ Accept: 'application/json' });
  });
});
