/*
 * Created by lintao on 2023/7/18 下午3:56
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package resp

type StatusCode int

const (
	StatusOK                 StatusCode = 0
	StatusDBErr              StatusCode = 100001
	StatusParamErr           StatusCode = 100002
	StatusAuth               StatusCode = 100003
	StatusServiceUnavailable StatusCode = 100004
	StatusAlterMsg           StatusCode = 100005
)
