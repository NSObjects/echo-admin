package authhttp

type loginRequest struct {
	Username string `json:"username" validate:"required,max=64"`
	Password string `json:"password" validate:"required,max=72"`
}
