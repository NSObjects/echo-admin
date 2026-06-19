package accesshttp

type createRoleRequest struct {
	Code        string   `json:"code" validate:"required,max=64"`
	Name        string   `json:"name" validate:"required,max=80"`
	Permissions []string `json:"permissions" validate:"required,min=1,dive,required,max=80"`
	MenuIDs     []int64  `json:"menu_ids" validate:"omitempty,dive,gt=0"`
	Active      bool     `json:"active"`
}

type updateRoleRequest struct {
	Name        *string  `json:"name" validate:"omitempty,max=80"`
	Permissions []string `json:"permissions" validate:"omitempty,dive,required,max=80"`
	MenuIDs     []int64  `json:"menu_ids" validate:"omitempty,dive,gt=0"`
	Active      *bool    `json:"active"`
}

type menuRequest struct {
	ParentID   int64  `json:"parent_id" validate:"gte=0"`
	Name       string `json:"name" validate:"required,max=80"`
	Path       string `json:"path" validate:"required,max=160"`
	Icon       string `json:"icon" validate:"omitempty,max=80"`
	Permission string `json:"permission" validate:"omitempty,max=80"`
	Sort       int    `json:"sort"`
	Active     bool   `json:"active"`
}
