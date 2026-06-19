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

// SystemConfig is a validated key-value runtime setting managed from the back office.
type SystemConfig struct {
	Key       string
	Name      string
	Value     string
	Public    bool
	UpdatedAt time.Time
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
	return SystemConfig{Key: key, Name: name, Value: value, Public: public, UpdatedAt: updatedAt}, nil
}

// Dictionary is a validated group of controlled values.
type Dictionary struct {
	ID        int64
	Code      string
	Name      string
	Items     []DictionaryItem
	CreatedAt time.Time
	UpdatedAt time.Time
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
	return Dictionary{ID: id, Code: code, Name: name, Items: copiedItems, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

// DictionaryItem is a validated selectable value under a dictionary.
type DictionaryItem struct {
	ID     int64
	Label  string
	Value  string
	Sort   int
	Active bool
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
	return DictionaryItem{ID: id, Label: label, Value: value, Sort: sort, Active: active}, nil
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
