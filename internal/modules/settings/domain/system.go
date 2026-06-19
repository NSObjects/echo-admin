// Package domain contains system configuration and dictionary rules.
package domain

import (
	"errors"
	"strings"
	"time"
)

// System data validation errors.
var (
	ErrInvalidConfigKey  = errors.New("invalid config key")
	ErrInvalidConfigName = errors.New("invalid config name")
	ErrInvalidDictID     = errors.New("invalid dictionary id")
	ErrInvalidDictCode   = errors.New("invalid dictionary code")
	ErrInvalidDictName   = errors.New("invalid dictionary name")
	ErrInvalidDictItemID = errors.New("invalid dictionary item id")
	ErrInvalidDictLabel  = errors.New("invalid dictionary item label")
	ErrInvalidDictValue  = errors.New("invalid dictionary item value")
)

// SystemConfig is a key-value runtime setting managed from the back office.
type SystemConfig struct {
	key       string
	name      string
	value     string
	public    bool
	updatedAt time.Time
}

// RestoreSystemConfig rebuilds a config item from a trusted store representation.
func RestoreSystemConfig(key, name, value string, public bool, updatedAt time.Time) (SystemConfig, error) {
	key = normalizeToken(key)
	name = strings.TrimSpace(name)
	if key == "" || len(key) > 80 {
		return SystemConfig{}, ErrInvalidConfigKey
	}
	if name == "" || len(name) > 120 {
		return SystemConfig{}, ErrInvalidConfigName
	}
	return SystemConfig{key: key, name: name, value: value, public: public, updatedAt: updatedAt}, nil
}

// Key returns the unique config key.
func (c SystemConfig) Key() string { return c.key }

// Name returns the operator-facing config name.
func (c SystemConfig) Name() string { return c.name }

// Value returns the stored config value.
func (c SystemConfig) Value() string { return c.value }

// Public reports whether non-sensitive clients may read the config value.
func (c SystemConfig) Public() bool { return c.public }

// UpdatedAt returns the last update timestamp.
func (c SystemConfig) UpdatedAt() time.Time { return c.updatedAt }

// Dictionary groups controlled values by code.
type Dictionary struct {
	id        int64
	code      string
	name      string
	items     []DictionaryItem
	createdAt time.Time
	updatedAt time.Time
}

// RestoreDictionary rebuilds a dictionary from a trusted store representation.
func RestoreDictionary(id int64, code, name string, items []DictionaryItem, createdAt, updatedAt time.Time) (Dictionary, error) {
	code = normalizeToken(code)
	name = strings.TrimSpace(name)
	if id < 0 {
		return Dictionary{}, ErrInvalidDictID
	}
	if code == "" || len(code) > 80 {
		return Dictionary{}, ErrInvalidDictCode
	}
	if name == "" || len(name) > 120 {
		return Dictionary{}, ErrInvalidDictName
	}
	copiedItems := append([]DictionaryItem(nil), items...)
	return Dictionary{id: id, code: code, name: name, items: copiedItems, createdAt: createdAt, updatedAt: updatedAt}, nil
}

// ID returns the persisted dictionary id.
func (d Dictionary) ID() int64 { return d.id }

// Code returns the unique dictionary code.
func (d Dictionary) Code() string { return d.code }

// Name returns the operator-facing dictionary name.
func (d Dictionary) Name() string { return d.name }

// Items returns a copy of dictionary items.
func (d Dictionary) Items() []DictionaryItem { return append([]DictionaryItem(nil), d.items...) }

// CreatedAt returns the creation timestamp.
func (d Dictionary) CreatedAt() time.Time { return d.createdAt }

// UpdatedAt returns the last update timestamp.
func (d Dictionary) UpdatedAt() time.Time { return d.updatedAt }

// DictionaryItem is one selectable value under a dictionary.
type DictionaryItem struct {
	id     int64
	label  string
	value  string
	sort   int
	active bool
}

// RestoreDictionaryItem rebuilds a dictionary item from a trusted store representation.
func RestoreDictionaryItem(id int64, label, value string, sort int, active bool) (DictionaryItem, error) {
	label = strings.TrimSpace(label)
	value = strings.TrimSpace(value)
	if id < 0 {
		return DictionaryItem{}, ErrInvalidDictItemID
	}
	if label == "" || len(label) > 120 {
		return DictionaryItem{}, ErrInvalidDictLabel
	}
	if value == "" || len(value) > 120 {
		return DictionaryItem{}, ErrInvalidDictValue
	}
	return DictionaryItem{id: id, label: label, value: value, sort: sort, active: active}, nil
}

// ID returns the persisted dictionary item id.
func (i DictionaryItem) ID() int64 { return i.id }

// Label returns the visible dictionary item label.
func (i DictionaryItem) Label() string { return i.label }

// Value returns the stable dictionary item value.
func (i DictionaryItem) Value() string { return i.value }

// Sort returns the item ordering key.
func (i DictionaryItem) Sort() int { return i.sort }

// Active reports whether the item is selectable.
func (i DictionaryItem) Active() bool { return i.active }

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
