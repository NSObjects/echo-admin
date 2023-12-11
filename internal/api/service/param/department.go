/*
 *
 * department.go
 * param
 *
 * Created by lintao on 2023/11/21 10:50
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

type DepartmentQuery struct {
	APIQuery
	Name   string `json:"name"`
	Status int    `json:"status"`
}

type Department struct {
	Name      string `json:"name"`
	ParentID  *uint  `json:"parent_id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    int    `json:"status"`
	Sort      int    `json:"sort"`
	Principal string `json:"principal" `
}
