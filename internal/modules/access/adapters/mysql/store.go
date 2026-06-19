// Package mysql persists access roles, menus, permissions, and API metadata in MySQL.
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

// Store persists access management data in MySQL.
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
	if err := db.WithContext(ctx).AutoMigrate(&roleModel{}, &menuModel{}, &permissionModel{}, &apiModel{}); err != nil {
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

// DeleteRole removes a role row by id.
func (s *Store) DeleteRole(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	result := s.db.WithContext(ctx).Delete(&roleModel{}, "id = ?", id)
	if result.Error != nil {
		return apperr.WrapDatabase(result.Error, "delete role")
	}
	if result.RowsAffected == 0 {
		return apperr.NewNotFound("role")
	}
	return nil
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

// DeleteMenu removes a menu row by id.
func (s *Store) DeleteMenu(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	result := s.db.WithContext(ctx).Delete(&menuModel{}, "id = ?", id)
	if result.Error != nil {
		return apperr.WrapDatabase(result.Error, "delete menu")
	}
	if result.RowsAffected == 0 {
		return apperr.NewNotFound("menu")
	}
	return nil
}

func (s *Store) seed(ctx context.Context) error {
	// Access seed data is one startup invariant: the root role must not be
	// refreshed without the permission, API, and menu catalogs it depends on.
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seedStore := &Store{db: tx}
		if err := seedStore.seedPermissions(ctx); err != nil {
			return err
		}
		if err := seedStore.seedAPIs(ctx); err != nil {
			return err
		}
		menuIDs, err := seedStore.seedMenus(ctx)
		if err != nil {
			return err
		}
		return seedStore.seedSuperAdminRole(ctx, menuIDs)
	})
}

func (s *Store) seedPermissions(ctx context.Context) error {
	now := time.Now().UTC()
	for _, permission := range domain.PermissionCatalog() {
		if err := s.ensurePermission(ctx, permission, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensurePermission(ctx context.Context, permission domain.PermissionDefinition, now time.Time) error {
	var existing permissionModel
	err := s.db.WithContext(ctx).Where("token = ?", permission.Token).First(&existing).Error
	if err == nil {
		existing.Resource = permission.Resource
		existing.Action = permission.Action
		existing.Name = permission.Name
		if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			return apperr.WrapDatabase(saveErr, "update seed permission")
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.WrapDatabase(err, "find seed permission")
	}
	model := permissionModel{
		Token:     permission.Token,
		Resource:  permission.Resource,
		Action:    permission.Action,
		Name:      permission.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if createErr := s.db.WithContext(ctx).Create(&model).Error; createErr != nil {
		return mapWriteError(createErr, "permission token already exists", "create seed permission")
	}
	return nil
}

func (s *Store) seedAPIs(ctx context.Context) error {
	now := time.Now().UTC()
	for _, api := range apiSeeds {
		if err := s.ensureAPI(ctx, api, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensureAPI(ctx context.Context, seed apiSeed, now time.Time) error {
	var existing apiModel
	err := s.db.WithContext(ctx).Where("method = ? AND path = ?", seed.method, seed.path).First(&existing).Error
	if err == nil {
		existing.Name = seed.name
		existing.Permission = seed.permission
		existing.Public = seed.public
		if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			return apperr.WrapDatabase(saveErr, "update seed api")
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.WrapDatabase(err, "find seed api")
	}
	model := apiModel{
		Method:     seed.method,
		Path:       seed.path,
		Name:       seed.name,
		Permission: seed.permission,
		Public:     seed.public,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if createErr := s.db.WithContext(ctx).Create(&model).Error; createErr != nil {
		return mapWriteError(createErr, "api route already exists", "create seed api")
	}
	return nil
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

type apiSeed struct {
	method     string
	path       string
	name       string
	permission string
	public     bool
}

// apiSeeds is the audited route catalog seeded into MySQL. Authorization still
// uses role permission tokens as the source of truth; this table documents which
// API routes are public and which permission token protects each private route.
var apiSeeds = []apiSeed{
	{method: "GET", path: "/api/health", name: "进程存活检查", public: true},
	{method: "GET", path: "/api/info", name: "应用信息", public: true},
	{method: "GET", path: "/api/ready", name: "整体 readiness", public: true},
	{method: "GET", path: "/api/capabilities", name: "capability 状态", public: true},
	{method: "POST", path: "/api/auth/login", name: "管理员登录", public: true},
	{method: "POST", path: "/api/auth/logout", name: "客户端退出登录"},
	{method: "POST", path: "/api/auth/role", name: "切换当前角色"},
	{method: "GET", path: "/api/auth/me", name: "当前管理员"},
	{method: "GET", path: "/api/admins", name: "管理员列表", permission: domain.PermissionAdminRead},
	{method: "POST", path: "/api/admins", name: "创建管理员", permission: domain.PermissionAdminCreate},
	{method: "PATCH", path: "/api/admins/:id", name: "更新管理员", permission: domain.PermissionAdminUpdate},
	{method: "DELETE", path: "/api/admins/:id", name: "删除管理员", permission: domain.PermissionAdminDelete},
	{method: "GET", path: "/api/roles", name: "角色列表", permission: domain.PermissionRoleRead},
	{method: "POST", path: "/api/roles", name: "创建角色", permission: domain.PermissionRoleCreate},
	{method: "PATCH", path: "/api/roles/:id", name: "更新角色", permission: domain.PermissionRoleUpdate},
	{method: "DELETE", path: "/api/roles/:id", name: "删除角色", permission: domain.PermissionRoleDelete},
	{method: "POST", path: "/api/roles/:id/copy", name: "复制角色", permission: domain.PermissionRoleCreate},
	{method: "GET", path: "/api/permissions", name: "权限目录元数据"},
	{method: "GET", path: "/api/menus", name: "菜单列表", permission: domain.PermissionMenuRead},
	{method: "POST", path: "/api/menus", name: "创建菜单", permission: domain.PermissionMenuCreate},
	{method: "PATCH", path: "/api/menus/:id", name: "更新菜单", permission: domain.PermissionMenuUpdate},
	{method: "DELETE", path: "/api/menus/:id", name: "删除菜单", permission: domain.PermissionMenuDelete},
	{method: "GET", path: "/api/system/configs", name: "系统配置列表", permission: domain.PermissionConfigRead},
	{method: "PUT", path: "/api/system/configs/:key", name: "创建或更新系统配置", permission: domain.PermissionConfigUpdate},
	{method: "GET", path: "/api/dictionaries", name: "字典列表", permission: domain.PermissionDictRead},
	{method: "POST", path: "/api/dictionaries", name: "创建字典", permission: domain.PermissionDictCreate},
	{method: "PATCH", path: "/api/dictionaries/:code", name: "更新字典", permission: domain.PermissionDictUpdate},
	{method: "DELETE", path: "/api/dictionaries/:code", name: "删除字典", permission: domain.PermissionDictDelete},
	{method: "POST", path: "/api/dictionaries/:code/items", name: "新增字典项", permission: domain.PermissionDictCreate},
	{method: "PATCH", path: "/api/dictionaries/:code/items/:item_id", name: "更新字典项", permission: domain.PermissionDictUpdate},
	{method: "DELETE", path: "/api/dictionaries/:code/items/:item_id", name: "删除字典项", permission: domain.PermissionDictDelete},
	{method: "GET", path: "/api/files", name: "文件列表", permission: domain.PermissionFileRead},
	{method: "POST", path: "/api/files", name: "上传文件", permission: domain.PermissionFileUpload},
	{method: "GET", path: "/api/uploads/*", name: "上传文件静态访问"},
	{method: "GET", path: "/api/logs/operations", name: "操作日志", permission: domain.PermissionLogRead},
	{method: "GET", path: "/api/logs/logins", name: "登录日志", permission: domain.PermissionLogRead},
}

type permissionModel struct {
	ID        int64     `gorm:"primaryKey"`
	Token     string    `gorm:"type:varchar(80);not null;uniqueIndex"`
	Resource  string    `gorm:"type:varchar(80);not null"`
	Action    string    `gorm:"type:varchar(40);not null"`
	Name      string    `gorm:"type:varchar(120);not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (permissionModel) TableName() string {
	return "access_permissions"
}

type apiModel struct {
	ID         int64     `gorm:"primaryKey"`
	Method     string    `gorm:"type:varchar(12);not null;uniqueIndex:idx_access_api_method_path"`
	Path       string    `gorm:"type:varchar(180);not null;uniqueIndex:idx_access_api_method_path"`
	Name       string    `gorm:"type:varchar(120);not null"`
	Permission string    `gorm:"type:varchar(80);not null;index"`
	Public     bool      `gorm:"not null"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (apiModel) TableName() string {
	return "access_apis"
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
