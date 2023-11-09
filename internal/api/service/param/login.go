/*
 *
 * login.go
 * param
 *
 * Created by lintao on 2023/11/9 16:20
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package param

type Login struct {
	Account  string `json:"account" `
	Password string `json:"password" `
}

type LoginResponse struct {
	Token string `json:"token"`
}
