package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/NSObjects/echo-admin/internal/modules/settings/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

// ListConfigs returns system configs.
func (u *Usecase) ListConfigs(ctx context.Context) ([]SystemConfig, error) {
	if err := u.ready(); err != nil {
		return nil, err
	}
	configs, err := u.store.ListConfigs(ctx)
	if err != nil {
		return nil, err
	}
	return mapConfigs(configs), nil
}

// UpsertConfig creates or updates a system config.
func (u *Usecase) UpsertConfig(ctx context.Context, input ConfigInput) (SystemConfig, error) {
	if err := u.ready(); err != nil {
		return SystemConfig{}, err
	}
	config, err := domain.RestoreSystemConfig(input.Key, input.Name, input.Value, input.Public, time.Time{})
	if err != nil {
		return SystemConfig{}, mapDomainError(err)
	}
	saved, err := u.store.UpsertConfig(ctx, config)
	if err != nil {
		return SystemConfig{}, err
	}
	return fromConfig(saved), nil
}

// ListDictionaries returns dictionaries with their items.
func (u *Usecase) ListDictionaries(ctx context.Context) ([]Dictionary, error) {
	if err := u.ready(); err != nil {
		return nil, err
	}
	dictionaries, err := u.store.ListDictionaries(ctx)
	if err != nil {
		return nil, err
	}
	return mapDictionaries(dictionaries), nil
}

// CreateDictionary validates and stores a dictionary.
func (u *Usecase) CreateDictionary(ctx context.Context, input DictionaryInput) (Dictionary, error) {
	if err := u.ready(); err != nil {
		return Dictionary{}, err
	}
	dictionary, err := domain.RestoreDictionary(0, input.Code, input.Name, nil, time.Time{}, time.Time{})
	if err != nil {
		return Dictionary{}, mapDomainError(err)
	}
	created, err := u.store.CreateDictionary(ctx, dictionary)
	if err != nil {
		return Dictionary{}, err
	}
	return fromDictionary(created), nil
}

// AddDictionaryItem appends one dictionary item.
func (u *Usecase) AddDictionaryItem(ctx context.Context, code string, input DictionaryItemInput) (Dictionary, error) {
	if err := u.ready(); err != nil {
		return Dictionary{}, err
	}
	item, err := domain.RestoreDictionaryItem(0, input.Label, input.Value, input.Sort, input.Active)
	if err != nil {
		return Dictionary{}, mapDomainError(err)
	}
	dictionary, err := u.store.AddDictionaryItem(ctx, code, item)
	if err != nil {
		return Dictionary{}, err
	}
	return fromDictionary(dictionary), nil
}

// UpdateDictionaryItem updates one dictionary item.
func (u *Usecase) UpdateDictionaryItem(ctx context.Context, code string, input DictionaryItemInput) (Dictionary, error) {
	if err := u.ready(); err != nil {
		return Dictionary{}, err
	}
	item, err := domain.RestoreDictionaryItem(input.ID, input.Label, input.Value, input.Sort, input.Active)
	if err != nil {
		return Dictionary{}, mapDomainError(err)
	}
	dictionary, err := u.store.UpdateDictionaryItem(ctx, code, item)
	if err != nil {
		return Dictionary{}, err
	}
	return fromDictionary(dictionary), nil
}

func (u *Usecase) ready() error {
	if u == nil || u.store == nil {
		return apperr.New(apperr.ErrInternalServer, "settings store is not configured")
	}
	return nil
}

func mapConfigs(configs []domain.SystemConfig) []SystemConfig {
	out := make([]SystemConfig, 0, len(configs))
	for _, config := range configs {
		out = append(out, fromConfig(config))
	}
	return out
}

func mapDictionaries(dictionaries []domain.Dictionary) []Dictionary {
	out := make([]Dictionary, 0, len(dictionaries))
	for _, dictionary := range dictionaries {
		out = append(out, fromDictionary(dictionary))
	}
	return out
}

func mapDictionaryItems(items []domain.DictionaryItem) []DictionaryItem {
	out := make([]DictionaryItem, 0, len(items))
	for _, item := range items {
		out = append(out, fromDictionaryItem(item))
	}
	return out
}

func fromConfig(config domain.SystemConfig) SystemConfig {
	return SystemConfig{
		Key:       config.Key,
		Name:      config.Name,
		Value:     config.Value,
		Public:    config.Public,
		UpdatedAt: config.UpdatedAt,
	}
}

func fromDictionary(dictionary domain.Dictionary) Dictionary {
	return Dictionary{
		ID:        dictionary.ID,
		Code:      dictionary.Code,
		Name:      dictionary.Name,
		Items:     mapDictionaryItems(dictionary.Items),
		CreatedAt: dictionary.CreatedAt,
		UpdatedAt: dictionary.UpdatedAt,
	}
}

func fromDictionaryItem(item domain.DictionaryItem) DictionaryItem {
	return DictionaryItem{
		ID:     item.ID,
		Label:  item.Label,
		Value:  item.Value,
		Sort:   item.Sort,
		Active: item.Active,
	}
}

func mapDomainError(err error) error {
	for _, entry := range domainErrorMessages {
		if errors.Is(err, entry.err) {
			return apperr.NewBadRequest(entry.message)
		}
	}
	return err
}

var domainErrorMessages = []struct {
	err     error
	message string
}{
	{domain.ErrInvalidConfigKey, "invalid config key"},
	{domain.ErrInvalidConfigName, "invalid config name"},
	{domain.ErrInvalidDictID, "invalid dictionary id"},
	{domain.ErrInvalidDictCode, "invalid dictionary code"},
	{domain.ErrInvalidDictName, "invalid dictionary name"},
	{domain.ErrInvalidDictItemID, "invalid dictionary item id"},
	{domain.ErrInvalidDictLabel, "invalid dictionary item label"},
	{domain.ErrInvalidDictValue, "invalid dictionary item value"},
}
