// Package mysql persists access roles and menus in MySQL.
package mysql

import (
	"context"
	"errors"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/mysqljson"
)

// Store persists roles and menus in MySQL.
type Store struct {
	db *gorm.DB
}

// NewStore migrates and seeds the MySQL access tables.
func NewStore(ctx context.Context, db *gorm.DB) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("create access store: nil context")
	}
	if db == nil {
		return nil, errors.New("create access store: nil db")
	}
	store := &Store{db: db}
	if err := db.WithContext(ctx).AutoMigrate(&roleModel{}, &menuModel{}); err != nil {
		return nil, apperr.WrapDatabase(err, "migrate access tables")
	}
	if err := store.seed(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

// FindRoleByID returns a role by id.
func (s *Store) FindRoleByID(ctx context.Context, id int64) (domain.Role, error) {
	if err := ctx.Err(); err != nil {
		return domain.Role{}, err
	}
	var model roleModel
	err := s.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		return domain.Role{}, mapReadError(err, "role", "find role")
	}
	return model.toDomain()
}

// FindRoleByCode returns a role by stable role code for bootstrapping.
func (s *Store) FindRoleByCode(ctx context.Context, code string) (domain.Role, error) {
	if err := ctx.Err(); err != nil {
		return domain.Role{}, err
	}
	var model roleModel
	err := s.db.WithContext(ctx).First(&model, "code = ?", code).Error
	if err != nil {
		return domain.Role{}, mapReadError(err, "role", "find role by code")
	}
	return model.toDomain()
}

// ListAllRoles returns all roles ordered for scoped delegation checks.
func (s *Store) ListAllRoles(ctx context.Context) ([]domain.Role, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var models []roleModel
	err := s.db.WithContext(ctx).Order("id DESC").Find(&models).Error
	if err != nil {
		return nil, apperr.WrapDatabase(err, "list all roles")
	}
	roles := make([]domain.Role, 0, len(models))
	for _, model := range models {
		role, err := model.toDomain()
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// CreateRole inserts a role.
func (s *Store) CreateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	if err := ctx.Err(); err != nil {
		return domain.Role{}, err
	}
	model := roleModelFromDomain(role)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Role{}, mapWriteError(err, "role code already exists", "create role")
	}
	return model.toDomain()
}

// UpdateRole replaces mutable role fields.
func (s *Store) UpdateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	if err := ctx.Err(); err != nil {
		return domain.Role{}, err
	}
	model := roleModelFromDomain(role)
	result := s.db.WithContext(ctx).Save(&model)
	if result.Error != nil {
		return domain.Role{}, mapWriteError(result.Error, "role code already exists", "update role")
	}
	if result.RowsAffected == 0 {
		return domain.Role{}, apperr.NewNotFound("role")
	}
	return model.toDomain()
}

// ListMenus returns menus ordered for display.
func (s *Store) ListMenus(ctx context.Context) ([]domain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var models []menuModel
	err := s.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&models).Error
	if err != nil {
		return nil, apperr.WrapDatabase(err, "list menus")
	}
	menus := make([]domain.Menu, 0, len(models))
	for _, model := range models {
		menu, err := model.toDomain()
		if err != nil {
			return nil, err
		}
		menus = append(menus, menu)
	}
	return menus, nil
}

// FindMenuByID returns a menu by id.
func (s *Store) FindMenuByID(ctx context.Context, id int64) (domain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return domain.Menu{}, err
	}
	var model menuModel
	err := s.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		return domain.Menu{}, mapReadError(err, "menu", "find menu")
	}
	return model.toDomain()
}

// CreateMenu inserts a menu.
func (s *Store) CreateMenu(ctx context.Context, menu domain.Menu) (domain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return domain.Menu{}, err
	}
	model := menuModelFromDomain(menu)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Menu{}, mapWriteError(err, "menu path already exists", "create menu")
	}
	return model.toDomain()
}

// UpdateMenu replaces mutable menu fields.
func (s *Store) UpdateMenu(ctx context.Context, menu domain.Menu) (domain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return domain.Menu{}, err
	}
	model := menuModelFromDomain(menu)
	result := s.db.WithContext(ctx).Save(&model)
	if result.Error != nil {
		return domain.Menu{}, mapWriteError(result.Error, "menu path already exists", "update menu")
	}
	if result.RowsAffected == 0 {
		return domain.Menu{}, apperr.NewNotFound("menu")
	}
	return model.toDomain()
}

func (s *Store) seed(ctx context.Context) error {
	menuIDs, err := s.seedMenus(ctx)
	if err != nil {
		return err
	}
	return s.seedSuperAdminRole(ctx, menuIDs)
}

func (s *Store) seedMenus(ctx context.Context) ([]int64, error) {
	seeds := []menuSeed{
		{name: "工作台", path: "/dashboard", icon: "dashboard", sort: 10},
		{name: "管理员管理", path: "/admins", icon: "user", permission: domain.PermissionAdminRead, sort: 20},
		{name: "角色权限", path: "/roles", icon: "safety", permission: domain.PermissionRoleRead, sort: 30},
		{name: "菜单管理", path: "/menus", icon: "menu", permission: domain.PermissionMenuRead, sort: 40},
		{name: "系统配置", path: "/configs", icon: "setting", permission: domain.PermissionConfigRead, sort: 50},
		{name: "数据字典", path: "/dictionaries", icon: "profile", permission: domain.PermissionDictRead, sort: 60},
		{name: "文件上传", path: "/files", icon: "upload", permission: domain.PermissionFileRead, sort: 70},
		{name: "审计日志", path: "/logs", icon: "fileSearch", permission: domain.PermissionLogRead, sort: 80},
	}
	menuIDs := make([]int64, 0, len(seeds))
	for _, seed := range seeds {
		id, err := s.ensureMenu(ctx, seed)
		if err != nil {
			return nil, err
		}
		menuIDs = append(menuIDs, id)
	}
	return menuIDs, nil
}

func (s *Store) ensureMenu(ctx context.Context, seed menuSeed) (int64, error) {
	var existing menuModel
	err := s.db.WithContext(ctx).Where("path = ?", seed.path).First(&existing).Error
	if err == nil {
		existing.Name = seed.name
		existing.Icon = seed.icon
		existing.Permission = seed.permission
		existing.Sort = seed.sort
		existing.Active = true
		if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			return 0, apperr.WrapDatabase(saveErr, "update seed menu")
		}
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, apperr.WrapDatabase(err, "find seed menu")
	}
	now := time.Now().UTC()
	menu, err := domain.RestoreMenu(0, 0, seed.name, seed.path, seed.icon, seed.permission, seed.sort, true, now, now)
	if err != nil {
		return 0, err
	}
	model := menuModelFromDomain(menu)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return 0, apperr.WrapDatabase(err, "create seed menu")
	}
	return model.ID, nil
}

func (s *Store) seedSuperAdminRole(ctx context.Context, menuIDs []int64) error {
	var existing roleModel
	err := s.db.WithContext(ctx).Where("code = ?", domain.RoleCodeSuperAdmin).First(&existing).Error
	if err == nil {
		existing.ParentID = 0
		existing.Name = "超级管理员"
		existing.Permissions = mysqljson.Strings(usecase.AllPermissions())
		existing.MenuIDs = mysqljson.Int64s(menuIDs)
		existing.DefaultPath = domain.DefaultRolePath
		existing.Active = true
		if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			return apperr.WrapDatabase(saveErr, "update seed role")
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.WrapDatabase(err, "find seed role")
	}
	now := time.Now().UTC()
	role, err := domain.RestoreRole(0, 0, domain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), menuIDs, domain.DefaultRolePath, true, now, now)
	if err != nil {
		return err
	}
	model := roleModelFromDomain(role)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return apperr.WrapDatabase(err, "create seed role")
	}
	return nil
}

type menuSeed struct {
	name       string
	path       string
	icon       string
	permission string
	sort       int
}

type roleModel struct {
	ID          int64             `gorm:"primaryKey"`
	ParentID    int64             `gorm:"not null;index"`
	Code        string            `gorm:"type:varchar(64);not null;uniqueIndex"`
	Name        string            `gorm:"type:varchar(80);not null"`
	Permissions mysqljson.Strings `gorm:"type:json;not null"`
	MenuIDs     mysqljson.Int64s  `gorm:"type:json;not null"`
	DefaultPath string            `gorm:"type:varchar(160);not null"`
	Active      bool              `gorm:"not null"`
	CreatedAt   time.Time         `gorm:"not null"`
	UpdatedAt   time.Time         `gorm:"not null"`
}

func (roleModel) TableName() string {
	return "access_roles"
}

func roleModelFromDomain(role domain.Role) roleModel {
	return roleModel{
		ID:          role.ID,
		ParentID:    role.ParentID,
		Code:        role.Code,
		Name:        role.Name,
		Permissions: mysqljson.Strings(role.Permissions),
		MenuIDs:     mysqljson.Int64s(role.MenuIDs),
		DefaultPath: role.DefaultPath,
		Active:      role.Active,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func (m roleModel) toDomain() (domain.Role, error) {
	return domain.RestoreRole(m.ID, m.ParentID, m.Code, m.Name, []string(m.Permissions), []int64(m.MenuIDs), m.DefaultPath, m.Active, m.CreatedAt, m.UpdatedAt)
}

type menuModel struct {
	ID         int64     `gorm:"primaryKey"`
	ParentID   int64     `gorm:"not null"`
	Name       string    `gorm:"type:varchar(80);not null"`
	Path       string    `gorm:"type:varchar(160);not null;uniqueIndex"`
	Icon       string    `gorm:"type:varchar(80);not null"`
	Permission string    `gorm:"type:varchar(80);not null"`
	Sort       int       `gorm:"not null"`
	Active     bool      `gorm:"not null"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (menuModel) TableName() string {
	return "access_menus"
}

func menuModelFromDomain(menu domain.Menu) menuModel {
	return menuModel{
		ID:         menu.ID,
		ParentID:   menu.ParentID,
		Name:       menu.Name,
		Path:       menu.Path,
		Icon:       menu.Icon,
		Permission: menu.Permission,
		Sort:       menu.Sort,
		Active:     menu.Active,
		CreatedAt:  menu.CreatedAt,
		UpdatedAt:  menu.UpdatedAt,
	}
}

func (m menuModel) toDomain() (domain.Menu, error) {
	return domain.RestoreMenu(m.ID, m.ParentID, m.Name, m.Path, m.Icon, m.Permission, m.Sort, m.Active, m.CreatedAt, m.UpdatedAt)
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
