/*
 *
 * menu.go
 * code
 *
 * Created by lintao on 2023/11/15 15:17
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

//go:generate codegen -type=int
//go:generate codegen -type=int -doc -output ./error_code_generated.md

package code

// 菜单类业务错误：
const (
	// ErrParentMenuExisted - 201: 父菜单id不存在，请确认后再选择.
	ErrParentMenuExisted int = iota + 100501
	// ErrNotAllowCreate - 201: 当上级菜单不是目录时，不允许创建子菜单.菜单类型只能创建按钮子集
	ErrNotAllowCreate
)
