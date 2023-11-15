/*
 *
 * role.go
 * param
 *
 * Created by lintao on 2023/11/15 16:20
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

type RoleQuery struct {
	Name      string `json:"name" `
	Identify  string `json:"identify" `
	State     int    `json:"state" `
	StartDate int64  `json:"start_date" `
	EndDate   int64  `json:"end_date" `
	APIQuery
}
