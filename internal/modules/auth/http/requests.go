package authhttp

type loginRequest struct {
	Username string `json:"username" validate:"required,max=64"`
	Password string `json:"password" validate:"required,max=72"`
}

type switchRoleRequest struct {
	RoleID int64 `json:"role_id" validate:"required,gt=0"`
}
