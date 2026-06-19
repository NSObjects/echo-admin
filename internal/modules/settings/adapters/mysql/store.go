// Package mysql persists system settings and dictionaries in MySQL.
package mysql

import (
	"context"
	"errors"
	"sort"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/modules/settings/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
)

// Store persists settings and dictionaries in MySQL.
type Store struct {
	db *gorm.DB
}

// NewStore migrates and seeds the MySQL settings tables.
func NewStore(ctx context.Context, db *gorm.DB) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("create settings store: nil context")
	}
	if db == nil {
		return nil, errors.New("create settings store: nil db")
	}
	store := &Store{db: db}
	if err := db.WithContext(ctx).AutoMigrate(&configModel{}, &dictionaryModel{}, &dictionaryItemModel{}); err != nil {
		return nil, apperr.WrapDatabase(err, "migrate settings tables")
	}
	if err := store.seed(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

// ListConfigs returns configs ordered by key.
func (s *Store) ListConfigs(ctx context.Context) ([]domain.SystemConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var models []configModel
	if err := s.db.WithContext(ctx).Order("`key` ASC").Find(&models).Error; err != nil {
		return nil, apperr.WrapDatabase(err, "list configs")
	}
	configs := make([]domain.SystemConfig, 0, len(models))
	for _, model := range models {
		config, err := model.toDomain()
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// UpsertConfig creates or updates one config by key.
func (s *Store) UpsertConfig(ctx context.Context, config domain.SystemConfig) (domain.SystemConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.SystemConfig{}, err
	}
	var model configModel
	err := s.db.WithContext(ctx).First(&model, "`key` = ?", config.Key).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.SystemConfig{}, apperr.WrapDatabase(err, "find config")
	}
	now := time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model = configModelFromDomain(config, now)
		if createErr := s.db.WithContext(ctx).Create(&model).Error; createErr != nil {
			return domain.SystemConfig{}, mapWriteError(createErr, "config key already exists", "create config")
		}
		return model.toDomain()
	}
	model.Name = config.Name
	model.Value = config.Value
	model.Public = config.Public
	model.UpdatedAt = now
	if saveErr := s.db.WithContext(ctx).Save(&model).Error; saveErr != nil {
		return domain.SystemConfig{}, mapWriteError(saveErr, "config key already exists", "update config")
	}
	return model.toDomain()
}

// ListDictionaries returns dictionaries with items ordered for display.
func (s *Store) ListDictionaries(ctx context.Context) ([]domain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var models []dictionaryModel
	err := s.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("sort ASC, id ASC") }).
		Order("id DESC").
		Find(&models).Error
	if err != nil {
		return nil, apperr.WrapDatabase(err, "list dictionaries")
	}
	dictionaries := make([]domain.Dictionary, 0, len(models))
	for _, model := range models {
		dictionary, err := model.toDomain()
		if err != nil {
			return nil, err
		}
		dictionaries = append(dictionaries, dictionary)
	}
	return dictionaries, nil
}

// CreateDictionary inserts a dictionary.
func (s *Store) CreateDictionary(ctx context.Context, dictionary domain.Dictionary) (domain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return domain.Dictionary{}, err
	}
	now := time.Now().UTC()
	model := dictionaryModelFromDomain(dictionary, now)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Dictionary{}, mapWriteError(err, "dictionary code already exists", "create dictionary")
	}
	return s.findDictionaryByCode(ctx, model.Code)
}

// AddDictionaryItem inserts one dictionary item under code.
func (s *Store) AddDictionaryItem(ctx context.Context, code string, item domain.DictionaryItem) (domain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return domain.Dictionary{}, err
	}
	var dictionary dictionaryModel
	if err := s.db.WithContext(ctx).First(&dictionary, "code = ?", code).Error; err != nil {
		return domain.Dictionary{}, mapReadError(err, "dictionary", "find dictionary")
	}
	model := dictionaryItemModelFromDomain(dictionary.ID, item)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Dictionary{}, apperr.WrapDatabase(err, "create dictionary item")
	}
	return s.findDictionaryByCode(ctx, code)
}

// UpdateDictionaryItem updates one dictionary item under code.
func (s *Store) UpdateDictionaryItem(ctx context.Context, code string, item domain.DictionaryItem) (domain.Dictionary, error) {
	if err := ctx.Err(); err != nil {
		return domain.Dictionary{}, err
	}
	var dictionary dictionaryModel
	if err := s.db.WithContext(ctx).First(&dictionary, "code = ?", code).Error; err != nil {
		return domain.Dictionary{}, mapReadError(err, "dictionary", "find dictionary")
	}
	model := dictionaryItemModelFromDomain(dictionary.ID, item)
	result := s.db.WithContext(ctx).Model(&dictionaryItemModel{}).
		Where("id = ? AND dictionary_id = ?", item.ID, dictionary.ID).
		Updates(map[string]interface{}{
			"label":  model.Label,
			"value":  model.Value,
			"sort":   model.Sort,
			"active": model.Active,
		})
	if result.Error != nil {
		return domain.Dictionary{}, apperr.WrapDatabase(result.Error, "update dictionary item")
	}
	if result.RowsAffected == 0 {
		return domain.Dictionary{}, apperr.NewNotFound("dictionary item")
	}
	return s.findDictionaryByCode(ctx, code)
}

func (s *Store) seed(ctx context.Context) error {
	config, err := domain.RestoreSystemConfig("site_name", "站点名称", "Echo Admin", true, time.Now().UTC())
	if err != nil {
		return err
	}
	if _, err := s.UpsertConfig(ctx, config); err != nil {
		return err
	}
	return s.seedStatusDictionary(ctx)
}

func (s *Store) seedStatusDictionary(ctx context.Context) error {
	var existing dictionaryModel
	err := s.db.WithContext(ctx).Where("code = ?", "status").First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.WrapDatabase(err, "find seed dictionary")
	}
	now := time.Now().UTC()
	enabled, err := domain.RestoreDictionaryItem(0, "启用", "enabled", 10, true)
	if err != nil {
		return err
	}
	disabled, err := domain.RestoreDictionaryItem(0, "禁用", "disabled", 20, true)
	if err != nil {
		return err
	}
	dictionary, err := domain.RestoreDictionary(0, "status", "状态", []domain.DictionaryItem{enabled, disabled}, now, now)
	if err != nil {
		return err
	}
	model := dictionaryModelFromDomain(dictionary, now)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return apperr.WrapDatabase(err, "create seed dictionary")
	}
	return nil
}

func (s *Store) findDictionaryByCode(ctx context.Context, code string) (domain.Dictionary, error) {
	var model dictionaryModel
	err := s.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("sort ASC, id ASC") }).
		First(&model, "code = ?", code).Error
	if err != nil {
		return domain.Dictionary{}, mapReadError(err, "dictionary", "find dictionary")
	}
	return model.toDomain()
}

type configModel struct {
	Key       string    `gorm:"primaryKey;type:varchar(80)"`
	Name      string    `gorm:"type:varchar(120);not null"`
	Value     string    `gorm:"type:text;not null"`
	Public    bool      `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (configModel) TableName() string {
	return "settings_configs"
}

func configModelFromDomain(config domain.SystemConfig, updatedAt time.Time) configModel {
	return configModel{
		Key:       config.Key,
		Name:      config.Name,
		Value:     config.Value,
		Public:    config.Public,
		UpdatedAt: updatedAt,
	}
}

func (m configModel) toDomain() (domain.SystemConfig, error) {
	return domain.RestoreSystemConfig(m.Key, m.Name, m.Value, m.Public, m.UpdatedAt)
}

type dictionaryModel struct {
	ID        int64                 `gorm:"primaryKey"`
	Code      string                `gorm:"type:varchar(80);not null;uniqueIndex"`
	Name      string                `gorm:"type:varchar(120);not null"`
	Items     []dictionaryItemModel `gorm:"foreignKey:DictionaryID"`
	CreatedAt time.Time             `gorm:"not null"`
	UpdatedAt time.Time             `gorm:"not null"`
}

func (dictionaryModel) TableName() string {
	return "settings_dictionaries"
}

func dictionaryModelFromDomain(dictionary domain.Dictionary, now time.Time) dictionaryModel {
	items := dictionary.Items
	modelItems := make([]dictionaryItemModel, 0, len(items))
	for _, item := range items {
		modelItems = append(modelItems, dictionaryItemModelFromDomain(0, item))
	}
	return dictionaryModel{
		ID:        dictionary.ID,
		Code:      dictionary.Code,
		Name:      dictionary.Name,
		Items:     modelItems,
		CreatedAt: coalesceTime(dictionary.CreatedAt, now),
		UpdatedAt: coalesceTime(dictionary.UpdatedAt, now),
	}
}

func (m dictionaryModel) toDomain() (domain.Dictionary, error) {
	items := make([]domain.DictionaryItem, 0, len(m.Items))
	for _, itemModel := range m.Items {
		item, err := itemModel.toDomain()
		if err != nil {
			return domain.Dictionary{}, err
		}
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].ID < items[j].ID
		}
		return items[i].Sort < items[j].Sort
	})
	return domain.RestoreDictionary(m.ID, m.Code, m.Name, items, m.CreatedAt, m.UpdatedAt)
}

type dictionaryItemModel struct {
	ID           int64  `gorm:"primaryKey"`
	DictionaryID int64  `gorm:"not null;index"`
	Label        string `gorm:"type:varchar(120);not null"`
	Value        string `gorm:"type:varchar(120);not null"`
	Sort         int    `gorm:"not null"`
	Active       bool   `gorm:"not null"`
}

func (dictionaryItemModel) TableName() string {
	return "settings_dictionary_items"
}

func dictionaryItemModelFromDomain(dictionaryID int64, item domain.DictionaryItem) dictionaryItemModel {
	return dictionaryItemModel{
		ID:           item.ID,
		DictionaryID: dictionaryID,
		Label:        item.Label,
		Value:        item.Value,
		Sort:         item.Sort,
		Active:       item.Active,
	}
}

func (m dictionaryItemModel) toDomain() (domain.DictionaryItem, error) {
	return domain.RestoreDictionaryItem(m.ID, m.Label, m.Value, m.Sort, m.Active)
}

func mapReadError(err error, resource, operation string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.NewNotFound(resource)
	}
	return apperr.WrapDatabase(err, operation)
}

func mapWriteError(err error, conflictMessage, operation string) error {
	var mysqlErr *drivermysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return apperr.NewConflict(conflictMessage)
	}
	return apperr.WrapDatabase(err, operation)
}

func coalesceTime(value, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}
