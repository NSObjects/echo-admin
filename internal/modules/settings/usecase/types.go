// Package usecase coordinates system setting and dictionary workflows.
package usecase

import (
	"context"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/settings/domain"
)

// Store persists system settings and dictionaries.
type Store interface {
	ListConfigs(context.Context) ([]domain.SystemConfig, error)
	UpsertConfig(context.Context, domain.SystemConfig) (domain.SystemConfig, error)
	ListDictionaries(context.Context) ([]domain.Dictionary, error)
	CreateDictionary(context.Context, domain.Dictionary) (domain.Dictionary, error)
	UpdateDictionary(context.Context, domain.Dictionary) (domain.Dictionary, error)
	DeleteDictionary(context.Context, string) error
	AddDictionaryItem(context.Context, string, domain.DictionaryItem) (domain.Dictionary, error)
	UpdateDictionaryItem(context.Context, string, domain.DictionaryItem) (domain.Dictionary, error)
	DeleteDictionaryItem(context.Context, string, int64) (domain.Dictionary, error)
}

// Usecase coordinates system setting and dictionary rules.
type Usecase struct {
	store Store
}

// New creates a settings usecase.
func New(store Store) *Usecase {
	return &Usecase{store: store}
}

// ConfigInput carries a system config value update.
type ConfigInput struct {
	Key    string
	Name   string
	Value  string
	Public bool
}

// DictionaryInput carries dictionary creation fields.
type DictionaryInput struct {
	Code string
	Name string
}

// UpdateDictionaryInput carries mutable dictionary fields.
type UpdateDictionaryInput struct {
	Code string
	Name string
}

// DictionaryItemInput carries dictionary item fields.
type DictionaryItemInput struct {
	ID     int64
	Label  string
	Value  string
	Sort   int
	Active bool
}

// SystemConfig is the adapter-facing config DTO.
type SystemConfig struct {
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	Public    bool      `json:"public"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Dictionary is the adapter-facing dictionary DTO.
type Dictionary struct {
	ID        int64            `json:"id"`
	Code      string           `json:"code"`
	Name      string           `json:"name"`
	Items     []DictionaryItem `json:"items"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// DictionaryItem is the adapter-facing dictionary item DTO.
type DictionaryItem struct {
	ID     int64  `json:"id"`
	Label  string `json:"label"`
	Value  string `json:"value"`
	Sort   int    `json:"sort"`
	Active bool   `json:"active"`
}
