/*
 *
 * menu.go
 * db
 *
 * Created by lintao on 2023/11/10 16:35
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package model

type Menu struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Redirect  string `json:"redirect"`
	Routes    []Menu `json:"routes"`
}
