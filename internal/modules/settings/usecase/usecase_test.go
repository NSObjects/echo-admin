package usecase_test

import (
	"context"
	"testing"
	"time"

	settingsdomain "github.com/NSObjects/echo-admin/internal/modules/settings/domain"
	"github.com/NSObjects/echo-admin/internal/modules/settings/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

func TestDeleteDictionaryRejectsSeedDictionary(t *testing.T) {
	store := &storeSpy{}
	uc := usecase.New(store)

	err := uc.DeleteDictionary(context.Background(), "status")
	if err == nil {
		t.Fatal("DeleteDictionary(status) error = nil, want bad request")
	}
	def, ok := apperr.ParseRegistered(err)
	if !ok || def.Kind != apperr.KindBadRequest {
		t.Fatalf("DeleteDictionary(status) kind = %v, want %s", def.Kind, apperr.KindBadRequest)
	}
	if store.deletedDictionaryCode != "" {
		t.Fatalf("deletedDictionaryCode = %q, want empty", store.deletedDictionaryCode)
	}
}

func TestDeleteDictionaryDeletesCustomDictionary(t *testing.T) {
	store := &storeSpy{}
	uc := usecase.New(store)

	if err := uc.DeleteDictionary(context.Background(), "Color"); err != nil {
		t.Fatalf("DeleteDictionary() error = %v", err)
	}
	if store.deletedDictionaryCode != "color" {
		t.Fatalf("deletedDictionaryCode = %q, want color", store.deletedDictionaryCode)
	}
}

type storeSpy struct {
	deletedDictionaryCode string
}

func (s *storeSpy) ListConfigs(context.Context) ([]settingsdomain.SystemConfig, error) {
	return nil, nil
}

func (s *storeSpy) UpsertConfig(_ context.Context, config settingsdomain.SystemConfig) (settingsdomain.SystemConfig, error) {
	return config, nil
}

func (s *storeSpy) ListDictionaries(context.Context) ([]settingsdomain.Dictionary, error) {
	return nil, nil
}

func (s *storeSpy) CreateDictionary(_ context.Context, dictionary settingsdomain.Dictionary) (settingsdomain.Dictionary, error) {
	return dictionary, nil
}

func (s *storeSpy) UpdateDictionary(_ context.Context, dictionary settingsdomain.Dictionary) (settingsdomain.Dictionary, error) {
	return dictionary, nil
}

func (s *storeSpy) DeleteDictionary(_ context.Context, code string) error {
	s.deletedDictionaryCode = code
	return nil
}

func (s *storeSpy) AddDictionaryItem(_ context.Context, code string, item settingsdomain.DictionaryItem) (settingsdomain.Dictionary, error) {
	return dictionaryWithItem(code, item)
}

func (s *storeSpy) UpdateDictionaryItem(_ context.Context, code string, item settingsdomain.DictionaryItem) (settingsdomain.Dictionary, error) {
	return dictionaryWithItem(code, item)
}

func (s *storeSpy) DeleteDictionaryItem(_ context.Context, code string, itemID int64) (settingsdomain.Dictionary, error) {
	item, err := settingsdomain.RestoreDictionaryItem(itemID, "启用", "enabled", 10, true)
	if err != nil {
		return settingsdomain.Dictionary{}, err
	}
	return dictionaryWithItem(code, item)
}

func dictionaryWithItem(code string, item settingsdomain.DictionaryItem) (settingsdomain.Dictionary, error) {
	return settingsdomain.RestoreDictionary(1, code, "字典", []settingsdomain.DictionaryItem{item}, fixedTime(), fixedTime())
}

func fixedTime() time.Time {
	return time.Unix(1_800_000_000, 0).UTC()
}
