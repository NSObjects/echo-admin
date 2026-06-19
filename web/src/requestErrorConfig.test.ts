import { message } from 'antd';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { errorConfig } from './requestErrorConfig';

const { mockClearAuthToken, mockGetAuthToken, mockReplace } = vi.hoisted(() => ({
  mockClearAuthToken: vi.fn(),
  mockGetAuthToken: vi.fn(),
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

vi.mock('@/services/auth-token', () => ({
  clearAuthToken: mockClearAuthToken,
  getAuthToken: mockGetAuthToken,
}));

type TestError = Error & {
  response?: { status?: number };
  info?: { message?: string };
};

describe('requestErrorConfig', () => {
  const errorThrower = errorConfig.errorConfig?.errorThrower;
  const errorHandler = errorConfig.errorConfig?.errorHandler;
  const interceptor = errorConfig.requestInterceptors?.[0] as (config: {
    headers?: Record<string, string>;
  }) => { headers?: Record<string, string> };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAuthToken.mockReturnValue('');
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

  it('redirects and clears token on unauthorized responses', () => {
    const error: TestError = new Error('unauthorized');
    error.response = { status: 401 };

    errorHandler?.(error, {});

    expect(mockClearAuthToken).toHaveBeenCalled();
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

  it('adds bearer token when a token exists', () => {
    mockGetAuthToken.mockReturnValue('token-1');

    const result = interceptor({ headers: { Accept: 'application/json' } });

    expect(result.headers).toEqual({
      Accept: 'application/json',
      Authorization: 'Bearer token-1',
    });
  });

  it('keeps request config unchanged without a token', () => {
    const config = { headers: { Accept: 'application/json' } };

    expect(interceptor(config)).toBe(config);
  });
});
