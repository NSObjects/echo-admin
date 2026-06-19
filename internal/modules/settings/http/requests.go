package settingshttp

type configRequest struct {
	Name   string `json:"name" validate:"required,max=120"`
	Value  string `json:"value" validate:"max=4000"`
	Public bool   `json:"public"`
}

type dictionaryRequest struct {
	Code string `json:"code" validate:"required,max=80"`
	Name string `json:"name" validate:"required,max=120"`
}

type updateDictionaryRequest struct {
	Name string `json:"name" validate:"required,max=120"`
}

type dictionaryItemRequest struct {
	Label  string `json:"label" validate:"required,max=120"`
	Value  string `json:"value" validate:"required,max=120"`
	Sort   int    `json:"sort"`
	Active bool   `json:"active"`
}
