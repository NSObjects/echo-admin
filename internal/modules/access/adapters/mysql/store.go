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
	if err := db.WithContext(ctx).AutoMigrate(&roleModel{}, &menuModel{}, &menuButtonModel{}, &permissionModel{}, &apiModel{}); err != nil {
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

// ListAPIs returns API routes ordered by group and path.
func (s *Store) ListAPIs(ctx context.Context) ([]domain.API, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var models []apiModel
	err := s.db.WithContext(ctx).Order("api_group ASC, path ASC, method ASC").Find(&models).Error
	if err != nil {
		return nil, apperr.WrapDatabase(err, "list apis")
	}
	apis := make([]domain.API, 0, len(models))
	for _, model := range models {
		api, err := model.toDomain()
		if err != nil {
			return nil, err
		}
		apis = append(apis, api)
	}
	return apis, nil
}

// FindAPIByID returns an API route by id.
func (s *Store) FindAPIByID(ctx context.Context, id int64) (domain.API, error) {
	if err := ctx.Err(); err != nil {
		return domain.API{}, err
	}
	var model apiModel
	err := s.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		return domain.API{}, mapReadError(err, "api", "find api")
	}
	return model.toDomain()
}

// CreateAPI inserts an API route.
func (s *Store) CreateAPI(ctx context.Context, api domain.API) (domain.API, error) {
	if err := ctx.Err(); err != nil {
		return domain.API{}, err
	}
	model := apiModelFromDomain(api)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.API{}, mapWriteError(err, "api route already exists", "create api")
	}
	return model.toDomain()
}

// UpdateAPI replaces mutable API route fields.
func (s *Store) UpdateAPI(ctx context.Context, api domain.API) (domain.API, error) {
	if err := ctx.Err(); err != nil {
		return domain.API{}, err
	}
	model := apiModelFromDomain(api)
	result := s.db.WithContext(ctx).Save(&model)
	if result.Error != nil {
		return domain.API{}, mapWriteError(result.Error, "api route already exists", "update api")
	}
	if result.RowsAffected == 0 {
		return domain.API{}, apperr.NewNotFound("api")
	}
	return model.toDomain()
}

// DeleteAPI removes an API route row by id.
func (s *Store) DeleteAPI(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	result := s.db.WithContext(ctx).Delete(&apiModel{}, "id = ?", id)
	if result.Error != nil {
		return apperr.WrapDatabase(result.Error, "delete api")
	}
	if result.RowsAffected == 0 {
		return apperr.NewNotFound("api")
	}
	return nil
}

// ListMenus returns menus ordered for display.
func (s *Store) ListMenus(ctx context.Context) ([]domain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var models []menuModel
	err := s.db.WithContext(ctx).
		Preload("Buttons", func(tx *gorm.DB) *gorm.DB { return tx.Order("id ASC") }).
		Order("sort ASC, id ASC").
		Find(&models).Error
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
	err := s.db.WithContext(ctx).
		Preload("Buttons", func(tx *gorm.DB) *gorm.DB { return tx.Order("id ASC") }).
		First(&model, "id = ?", id).Error
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
	var created domain.Menu
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
			return mapWriteError(err, "menu path already exists", "create menu")
		}
		if err := replaceMenuButtons(ctx, tx, model.ID, menu.Buttons); err != nil {
			return err
		}
		loaded, err := loadMenu(ctx, tx, model.ID)
		if err != nil {
			return err
		}
		created = loaded
		return nil
	})
	if err != nil {
		return domain.Menu{}, err
	}
	return created, nil
}

// UpdateMenu replaces mutable menu fields.
func (s *Store) UpdateMenu(ctx context.Context, menu domain.Menu) (domain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return domain.Menu{}, err
	}
	model := menuModelFromDomain(menu)
	var updated domain.Menu
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Save(&model)
		if result.Error != nil {
			return mapWriteError(result.Error, "menu path already exists", "update menu")
		}
		if result.RowsAffected == 0 {
			return apperr.NewNotFound("menu")
		}
		if err := replaceMenuButtons(ctx, tx, model.ID, menu.Buttons); err != nil {
			return err
		}
		loaded, err := loadMenu(ctx, tx, model.ID)
		if err != nil {
			return err
		}
		updated = loaded
		return nil
	})
	if err != nil {
		return domain.Menu{}, err
	}
	return updated, nil
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
		apiIDs, err := seedStore.seedAPIs(ctx)
		if err != nil {
			return err
		}
		menuIDs, buttonIDs, err := seedStore.seedMenus(ctx)
		if err != nil {
			return err
		}
		return seedStore.seedSuperAdminRole(ctx, menuIDs, apiIDs, buttonIDs)
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

func (s *Store) seedAPIs(ctx context.Context) ([]int64, error) {
	now := time.Now().UTC()
	apiIDs := make([]int64, 0, len(apiSeeds))
	for _, api := range apiSeeds {
		id, err := s.ensureAPI(ctx, api, now)
		if err != nil {
			return nil, err
		}
		apiIDs = append(apiIDs, id)
	}
	return apiIDs, nil
}

func (s *Store) ensureAPI(ctx context.Context, seed apiSeed, now time.Time) (int64, error) {
	var existing apiModel
	err := s.db.WithContext(ctx).Where("method = ? AND path = ?", seed.method, seed.path).First(&existing).Error
	if err == nil {
		existing.Description = seed.description
		existing.Group = seed.group
		existing.Permission = seed.permission
		existing.Public = seed.public
		if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			return 0, apperr.WrapDatabase(saveErr, "update seed api")
		}
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, apperr.WrapDatabase(err, "find seed api")
	}
	model := apiModel{
		Method:      seed.method,
		Path:        seed.path,
		Description: seed.description,
		Group:       seed.group,
		Permission:  seed.permission,
		Public:      seed.public,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if createErr := s.db.WithContext(ctx).Create(&model).Error; createErr != nil {
		return 0, mapWriteError(createErr, "api route already exists", "create seed api")
	}
	return model.ID, nil
}

func (s *Store) seedMenus(ctx context.Context) ([]int64, []int64, error) {
	menuIDs := make([]int64, 0, len(defaultMenuSeeds))
	buttonIDs := make([]int64, 0, len(defaultMenuSeeds)*3)
	for _, seed := range defaultMenuSeeds {
		id, ids, err := s.ensureMenu(ctx, seed)
		if err != nil {
			return nil, nil, err
		}
		menuIDs = append(menuIDs, id)
		buttonIDs = append(buttonIDs, ids...)
	}
	return menuIDs, buttonIDs, nil
}

func (s *Store) ensureMenu(ctx context.Context, seed menuSeed) (int64, []int64, error) {
	var existing menuModel
	err := s.db.WithContext(ctx).Where("path = ?", seed.path).First(&existing).Error
	if err == nil {
		existing.Name = seed.name
		existing.Icon = seed.icon
		existing.Hidden = seed.hidden
		existing.Component = seed.component
		existing.ActiveName = seed.meta.ActiveName
		existing.KeepAlive = seed.meta.KeepAlive
		existing.DefaultMenu = seed.meta.DefaultMenu
		existing.CloseTab = seed.meta.CloseTab
		existing.TransitionType = seed.meta.TransitionType
		existing.Permission = seed.permission
		existing.Sort = seed.sort
		existing.Active = true
		if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			return 0, nil, apperr.WrapDatabase(saveErr, "update seed menu")
		}
		buttonIDs, buttonErr := s.ensureSeedButtons(ctx, existing.ID, seed.buttons)
		return existing.ID, buttonIDs, buttonErr
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil, apperr.WrapDatabase(err, "find seed menu")
	}
	now := time.Now().UTC()
	menu, err := domain.RestoreMenu(0, 0, seed.name, seed.path, seed.icon, seed.hidden, seed.component, seed.meta, seed.permission, seed.sort, true, nil, now, now)
	if err != nil {
		return 0, nil, err
	}
	model := menuModelFromDomain(menu)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return 0, nil, apperr.WrapDatabase(err, "create seed menu")
	}
	buttonIDs, buttonErr := s.ensureSeedButtons(ctx, model.ID, seed.buttons)
	return model.ID, buttonIDs, buttonErr
}

func (s *Store) ensureSeedButtons(ctx context.Context, menuID int64, seeds []menuButtonSeed) ([]int64, error) {
	buttonIDs := make([]int64, 0, len(seeds))
	for _, seed := range seeds {
		id, err := s.ensureSeedButton(ctx, menuID, seed)
		if err != nil {
			return nil, err
		}
		buttonIDs = append(buttonIDs, id)
	}
	return buttonIDs, nil
}

func (s *Store) ensureSeedButton(ctx context.Context, menuID int64, seed menuButtonSeed) (int64, error) {
	var existing menuButtonModel
	err := s.db.WithContext(ctx).Where("menu_id = ? AND name = ?", menuID, seed.name).First(&existing).Error
	if err == nil {
		existing.Description = seed.description
		if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			return 0, apperr.WrapDatabase(saveErr, "update seed menu button")
		}
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, apperr.WrapDatabase(err, "find seed menu button")
	}
	now := time.Now().UTC()
	button, err := domain.RestoreMenuButton(0, menuID, seed.name, seed.description, now, now)
	if err != nil {
		return 0, err
	}
	model := menuButtonModelFromDomain(button)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return 0, mapWriteError(err, "menu button already exists", "create seed menu button")
	}
	return model.ID, nil
}

func (s *Store) seedSuperAdminRole(ctx context.Context, menuIDs, apiIDs, buttonIDs []int64) error {
	roleIDs, err := s.allRoleIDs(ctx)
	if err != nil {
		return err
	}
	var existing roleModel
	err = s.db.WithContext(ctx).Where("code = ?", domain.RoleCodeSuperAdmin).First(&existing).Error
	if err == nil {
		existing.ParentID = 0
		existing.Name = "超级管理员"
		existing.Permissions = mysqljson.Strings(usecase.AllPermissions())
		existing.MenuIDs = mysqljson.Int64s(menuIDs)
		existing.APIIDs = mysqljson.Int64s(apiIDs)
		existing.ButtonIDs = mysqljson.Int64s(buttonIDs)
		existing.DataRoleIDs = mysqljson.Int64s(roleIDs)
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
	role, err := domain.RestoreRole(0, 0, domain.RoleCodeSuperAdmin, "超级管理员", usecase.AllPermissions(), menuIDs, apiIDs, buttonIDs, roleIDs, domain.DefaultRolePath, true, now, now)
	if err != nil {
		return err
	}
	model := roleModelFromDomain(role)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return apperr.WrapDatabase(err, "create seed role")
	}
	model.DataRoleIDs = mysqljson.Int64s(append(roleIDs, model.ID))
	if err := s.db.WithContext(ctx).Save(&model).Error; err != nil {
		return apperr.WrapDatabase(err, "update seed role data authority")
	}
	return nil
}

func (s *Store) allRoleIDs(ctx context.Context) ([]int64, error) {
	var roleIDs []int64
	if err := s.db.WithContext(ctx).Model(&roleModel{}).Pluck("id", &roleIDs).Error; err != nil {
		return nil, apperr.WrapDatabase(err, "list role ids")
	}
	return roleIDs, nil
}

type menuSeed struct {
	name       string
	path       string
	icon       string
	hidden     bool
	component  string
	meta       domain.MenuMeta
	permission string
	sort       int
	buttons    []menuButtonSeed
}

type menuButtonSeed struct {
	name        string
	description string
}

// defaultMenuSeeds is the boot-time navigation and page-button catalog. The
// root role is refreshed from these IDs on startup, so a new admin instance can
// log in with a complete gin-vue-admin-style back-office surface.
var defaultMenuSeeds = []menuSeed{
	{name: "工作台", path: "/dashboard", icon: "dashboard", component: "./Dashboard", sort: 10},
	{name: "管理员管理", path: "/admins", icon: "user", component: "./Admins", permission: domain.PermissionAdminRead, sort: 20, buttons: []menuButtonSeed{
		{name: "create", description: "新增管理员"},
		{name: "update", description: "编辑管理员"},
		{name: "delete", description: "删除管理员"},
	}},
	{name: "角色权限", path: "/roles", icon: "safety", component: "./Roles", permission: domain.PermissionRoleRead, sort: 30, buttons: []menuButtonSeed{
		{name: "create", description: "新增角色"},
		{name: "update", description: "编辑角色"},
		{name: "delete", description: "删除角色"},
		{name: "copy", description: "复制角色"},
		{name: "members", description: "授权角色成员"},
	}},
	{name: "菜单管理", path: "/menus", icon: "menu", component: "./Menus", permission: domain.PermissionMenuRead, meta: domain.MenuMeta{KeepAlive: true}, sort: 40, buttons: []menuButtonSeed{
		{name: "create", description: "新增菜单"},
		{name: "update", description: "编辑菜单"},
		{name: "delete", description: "删除菜单"},
		{name: "roles", description: "授权菜单角色"},
	}},
	{name: "API管理", path: "/apis", icon: "api", component: "./APIs", permission: domain.PermissionAPIRead, meta: domain.MenuMeta{KeepAlive: true}, sort: 50, buttons: []menuButtonSeed{
		{name: "create", description: "新增API"},
		{name: "update", description: "编辑API"},
		{name: "delete", description: "删除API"},
		{name: "roles", description: "授权API角色"},
	}},
	{name: "API Token", path: "/api-tokens", icon: "key", component: "./APITokens", permission: domain.PermissionAPITokenRead, meta: domain.MenuMeta{KeepAlive: true}, sort: 55, buttons: []menuButtonSeed{
		{name: "create", description: "新增API Token"},
		{name: "update", description: "编辑API Token"},
		{name: "delete", description: "删除API Token"},
	}},
	{name: "系统配置", path: "/configs", icon: "setting", component: "./Configs", permission: domain.PermissionConfigRead, sort: 60, buttons: []menuButtonSeed{
		{name: "update", description: "更新配置"},
		{name: "delete", description: "删除配置"},
	}},
	{name: "系统参数", path: "/params", icon: "control", component: "./Params", permission: domain.PermissionParamRead, meta: domain.MenuMeta{KeepAlive: true}, sort: 65, buttons: []menuButtonSeed{
		{name: "create", description: "新增参数"},
		{name: "update", description: "编辑参数"},
		{name: "delete", description: "删除参数"},
	}},
	{name: "版本管理", path: "/versions", icon: "server", component: "./Versions", permission: domain.PermissionVersionRead, sort: 68, buttons: []menuButtonSeed{
		{name: "create", description: "新增版本记录"},
		{name: "export", description: "导出版本包"},
		{name: "import", description: "导入版本包"},
		{name: "update", description: "编辑版本记录"},
		{name: "delete", description: "删除版本记录"},
	}},
	{name: "数据字典", path: "/dictionaries", icon: "profile", component: "./Dictionaries", permission: domain.PermissionDictRead, sort: 70, buttons: []menuButtonSeed{
		{name: "create", description: "新增字典"},
		{name: "export", description: "导出字典"},
		{name: "import", description: "导入字典"},
		{name: "update", description: "编辑字典"},
		{name: "delete", description: "删除字典"},
		{name: "item_create", description: "新增字典项"},
		{name: "item_update", description: "编辑字典项"},
		{name: "item_delete", description: "删除字典项"},
	}},
	{name: "文件上传", path: "/files", icon: "upload", component: "./Files", permission: domain.PermissionFileRead, sort: 80, buttons: []menuButtonSeed{
		{name: "upload", description: "上传文件"},
		{name: "update", description: "重命名文件"},
		{name: "delete", description: "删除文件"},
		{name: "category_create", description: "新增文件分类"},
		{name: "category_update", description: "编辑文件分类"},
		{name: "category_delete", description: "删除文件分类"},
	}},
	{name: "审计日志", path: "/logs", icon: "fileSearch", component: "./Logs", permission: domain.PermissionLogRead, sort: 90, buttons: []menuButtonSeed{
		{name: "resolve", description: "处理系统错误"},
		{name: "delete", description: "删除日志"},
	}},
}

type apiSeed struct {
	method      string
	path        string
	description string
	group       string
	permission  string
	public      bool
}

// apiSeeds is the boot-time API authorization catalog. Private handler routes
// must have a matching method/path row because auth checks both the permission
// token and the active role's assigned API ids.
var apiSeeds = []apiSeed{
	{method: "GET", path: "/api/health", description: "进程存活检查", group: "system", public: true},
	{method: "GET", path: "/api/info", description: "应用信息", group: "system", public: true},
	{method: "GET", path: "/api/ready", description: "整体 readiness", group: "system", public: true},
	{method: "GET", path: "/api/capabilities", description: "capability 状态", group: "system", public: true},
	{method: "POST", path: "/api/auth/login", description: "管理员登录", group: "auth", public: true},
	{method: "POST", path: "/api/auth/logout", description: "服务端退出当前登录会话", group: "auth"},
	{method: "POST", path: "/api/auth/logout-others", description: "撤销其他登录会话", group: "auth"},
	{method: "POST", path: "/api/auth/password", description: "当前管理员修改密码", group: "auth"},
	{method: "POST", path: "/api/auth/role", description: "切换当前角色", group: "auth"},
	{method: "GET", path: "/api/auth/me", description: "当前管理员", group: "auth"},
	{method: "PATCH", path: "/api/auth/me", description: "更新当前管理员资料", group: "auth"},
	{method: "GET", path: "/api/admins", description: "管理员列表", group: "admin", permission: domain.PermissionAdminRead},
	{method: "POST", path: "/api/admins", description: "创建管理员", group: "admin", permission: domain.PermissionAdminCreate},
	{method: "PATCH", path: "/api/admins/:id", description: "更新管理员", group: "admin", permission: domain.PermissionAdminUpdate},
	{method: "DELETE", path: "/api/admins/:id", description: "删除管理员", group: "admin", permission: domain.PermissionAdminDelete},
	{method: "GET", path: "/api/roles", description: "角色列表", group: "role", permission: domain.PermissionRoleRead},
	{method: "POST", path: "/api/roles", description: "创建角色", group: "role", permission: domain.PermissionRoleCreate},
	{method: "PATCH", path: "/api/roles/:id", description: "更新角色", group: "role", permission: domain.PermissionRoleUpdate},
	{method: "DELETE", path: "/api/roles/:id", description: "删除角色", group: "role", permission: domain.PermissionRoleDelete},
	{method: "POST", path: "/api/roles/:id/copy", description: "复制角色", group: "role", permission: domain.PermissionRoleCreate},
	{method: "GET", path: "/api/roles/:id/admins", description: "角色关联管理员", group: "role", permission: domain.PermissionRoleRead},
	{method: "PUT", path: "/api/roles/:id/admins", description: "更新角色关联管理员", group: "role", permission: domain.PermissionRoleUpdate},
	{method: "GET", path: "/api/permissions", description: "权限目录元数据", group: "access", permission: domain.PermissionRoleRead},
	{method: "GET", path: "/api/apis", description: "API列表", group: "api", permission: domain.PermissionAPIRead},
	{method: "GET", path: "/api/apis/groups", description: "API分组", group: "api", permission: domain.PermissionAPIRead},
	{method: "POST", path: "/api/apis", description: "创建API", group: "api", permission: domain.PermissionAPICreate},
	{method: "POST", path: "/api/apis/batch-delete", description: "批量删除API", group: "api", permission: domain.PermissionAPIDelete},
	{method: "GET", path: "/api/apis/:id", description: "API详情", group: "api", permission: domain.PermissionAPIRead},
	{method: "PATCH", path: "/api/apis/:id", description: "更新API", group: "api", permission: domain.PermissionAPIUpdate},
	{method: "DELETE", path: "/api/apis/:id", description: "删除API", group: "api", permission: domain.PermissionAPIDelete},
	{method: "GET", path: "/api/apis/:id/roles", description: "API授权角色", group: "api", permission: domain.PermissionAPIRead},
	{method: "PUT", path: "/api/apis/:id/roles", description: "更新API授权角色", group: "api", permission: domain.PermissionAPIUpdate},
	{method: "GET", path: "/api/api-tokens", description: "API Token列表", group: "api_token", permission: domain.PermissionAPITokenRead},
	{method: "POST", path: "/api/api-tokens", description: "创建API Token", group: "api_token", permission: domain.PermissionAPITokenCreate},
	{method: "PATCH", path: "/api/api-tokens/:id", description: "更新API Token", group: "api_token", permission: domain.PermissionAPITokenUpdate},
	{method: "DELETE", path: "/api/api-tokens/:id", description: "删除API Token", group: "api_token", permission: domain.PermissionAPITokenDelete},
	{method: "GET", path: "/api/menus", description: "菜单列表", group: "menu", permission: domain.PermissionMenuRead},
	{method: "POST", path: "/api/menus", description: "创建菜单", group: "menu", permission: domain.PermissionMenuCreate},
	{method: "GET", path: "/api/menus/:id", description: "菜单详情", group: "menu", permission: domain.PermissionMenuRead},
	{method: "PATCH", path: "/api/menus/:id", description: "更新菜单", group: "menu", permission: domain.PermissionMenuUpdate},
	{method: "DELETE", path: "/api/menus/:id", description: "删除菜单", group: "menu", permission: domain.PermissionMenuDelete},
	{method: "GET", path: "/api/menus/:id/roles", description: "菜单授权角色", group: "menu", permission: domain.PermissionMenuRead},
	{method: "PUT", path: "/api/menus/:id/roles", description: "更新菜单授权角色", group: "menu", permission: domain.PermissionMenuUpdate},
	{method: "GET", path: "/api/system/configs", description: "系统配置列表", group: "config", permission: domain.PermissionConfigRead},
	{method: "PUT", path: "/api/system/configs/:key", description: "创建或更新系统配置", group: "config", permission: domain.PermissionConfigUpdate},
	{method: "DELETE", path: "/api/system/configs/:key", description: "删除系统配置", group: "config", permission: domain.PermissionConfigDelete},
	{method: "GET", path: "/api/system/params", description: "系统参数列表", group: "param", permission: domain.PermissionParamRead},
	{method: "POST", path: "/api/system/params", description: "创建系统参数", group: "param", permission: domain.PermissionParamCreate},
	{method: "POST", path: "/api/system/params/batch-delete", description: "批量删除系统参数", group: "param", permission: domain.PermissionParamDelete},
	{method: "GET", path: "/api/system/params/key/:key", description: "按键获取系统参数", group: "param", permission: domain.PermissionParamRead},
	{method: "GET", path: "/api/system/params/:id", description: "系统参数详情", group: "param", permission: domain.PermissionParamRead},
	{method: "PATCH", path: "/api/system/params/:id", description: "更新系统参数", group: "param", permission: domain.PermissionParamUpdate},
	{method: "DELETE", path: "/api/system/params/:id", description: "删除系统参数", group: "param", permission: domain.PermissionParamDelete},
	{method: "GET", path: "/api/system/versions", description: "版本记录列表", group: "version", permission: domain.PermissionVersionRead},
	{method: "POST", path: "/api/system/versions", description: "创建版本记录", group: "version", permission: domain.PermissionVersionCreate},
	{method: "POST", path: "/api/system/versions/export", description: "导出版本包", group: "version", permission: domain.PermissionVersionCreate},
	{method: "POST", path: "/api/system/versions/import", description: "导入版本包", group: "version", permission: domain.PermissionVersionCreate},
	{method: "POST", path: "/api/system/versions/batch-delete", description: "批量删除版本记录", group: "version", permission: domain.PermissionVersionDelete},
	{method: "GET", path: "/api/system/versions/:id", description: "版本记录详情", group: "version", permission: domain.PermissionVersionRead},
	{method: "GET", path: "/api/system/versions/:id/download", description: "下载版本记录JSON", group: "version", permission: domain.PermissionVersionRead},
	{method: "PATCH", path: "/api/system/versions/:id", description: "更新版本记录", group: "version", permission: domain.PermissionVersionUpdate},
	{method: "DELETE", path: "/api/system/versions/:id", description: "删除版本记录", group: "version", permission: domain.PermissionVersionDelete},
	{method: "GET", path: "/api/dictionaries", description: "字典列表", group: "dictionary", permission: domain.PermissionDictRead},
	{method: "POST", path: "/api/dictionaries", description: "创建字典", group: "dictionary", permission: domain.PermissionDictCreate},
	{method: "GET", path: "/api/dictionaries/export", description: "导出字典", group: "dictionary", permission: domain.PermissionDictRead},
	{method: "POST", path: "/api/dictionaries/import", description: "导入字典", group: "dictionary", permission: domain.PermissionDictCreate},
	{method: "PATCH", path: "/api/dictionaries/:code", description: "更新字典", group: "dictionary", permission: domain.PermissionDictUpdate},
	{method: "DELETE", path: "/api/dictionaries/:code", description: "删除字典", group: "dictionary", permission: domain.PermissionDictDelete},
	{method: "POST", path: "/api/dictionaries/:code/items", description: "新增字典项", group: "dictionary", permission: domain.PermissionDictCreate},
	{method: "PATCH", path: "/api/dictionaries/:code/items/:item_id", description: "更新字典项", group: "dictionary", permission: domain.PermissionDictUpdate},
	{method: "DELETE", path: "/api/dictionaries/:code/items/:item_id", description: "删除字典项", group: "dictionary", permission: domain.PermissionDictDelete},
	{method: "GET", path: "/api/file-categories", description: "文件分类列表", group: "file", permission: domain.PermissionFileRead},
	{method: "POST", path: "/api/file-categories", description: "创建文件分类", group: "file", permission: domain.PermissionFileCategoryCreate},
	{method: "PATCH", path: "/api/file-categories/:id", description: "更新文件分类", group: "file", permission: domain.PermissionFileCategoryUpdate},
	{method: "DELETE", path: "/api/file-categories/:id", description: "删除文件分类", group: "file", permission: domain.PermissionFileCategoryDelete},
	{method: "GET", path: "/api/files", description: "文件列表", group: "file", permission: domain.PermissionFileRead},
	{method: "POST", path: "/api/files", description: "上传文件", group: "file", permission: domain.PermissionFileUpload},
	{method: "POST", path: "/api/files/import-url", description: "导入文件URL", group: "file", permission: domain.PermissionFileUpload},
	{method: "PATCH", path: "/api/files/:id/name", description: "重命名文件", group: "file", permission: domain.PermissionFileUpdate},
	{method: "DELETE", path: "/api/files/:id", description: "删除文件", group: "file", permission: domain.PermissionFileDelete},
	{method: "GET", path: "/api/uploads/*", description: "上传文件静态访问", group: "file"},
	{method: "GET", path: "/api/logs/operations", description: "操作日志", group: "log", permission: domain.PermissionLogRead},
	{method: "GET", path: "/api/logs/operations/:id", description: "操作日志详情", group: "log", permission: domain.PermissionLogRead},
	{method: "DELETE", path: "/api/logs/operations/:id", description: "删除操作日志", group: "log", permission: domain.PermissionLogDelete},
	{method: "POST", path: "/api/logs/operations/batch-delete", description: "批量删除操作日志", group: "log", permission: domain.PermissionLogDelete},
	{method: "GET", path: "/api/logs/logins", description: "登录日志", group: "log", permission: domain.PermissionLogRead},
	{method: "GET", path: "/api/logs/logins/:id", description: "登录日志详情", group: "log", permission: domain.PermissionLogRead},
	{method: "DELETE", path: "/api/logs/logins/:id", description: "删除登录日志", group: "log", permission: domain.PermissionLogDelete},
	{method: "POST", path: "/api/logs/logins/batch-delete", description: "批量删除登录日志", group: "log", permission: domain.PermissionLogDelete},
	{method: "GET", path: "/api/logs/errors", description: "系统错误日志", group: "log", permission: domain.PermissionLogRead},
	{method: "GET", path: "/api/logs/errors/:id", description: "系统错误日志详情", group: "log", permission: domain.PermissionLogRead},
	{method: "POST", path: "/api/logs/errors/:id/resolve", description: "处理系统错误日志", group: "log", permission: domain.PermissionLogResolve},
	{method: "DELETE", path: "/api/logs/errors/:id/resolve", description: "取消处理系统错误日志", group: "log", permission: domain.PermissionLogResolve},
	{method: "DELETE", path: "/api/logs/errors/:id", description: "删除系统错误日志", group: "log", permission: domain.PermissionLogDelete},
	{method: "POST", path: "/api/logs/errors/batch-delete", description: "批量删除系统错误日志", group: "log", permission: domain.PermissionLogDelete},
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
	ID          int64     `gorm:"primaryKey"`
	Method      string    `gorm:"type:varchar(12);not null;uniqueIndex:idx_access_api_method_path"`
	Path        string    `gorm:"type:varchar(180);not null;uniqueIndex:idx_access_api_method_path"`
	Description string    `gorm:"type:varchar(120);not null"`
	Group       string    `gorm:"column:api_group;type:varchar(80);not null;index"`
	Permission  string    `gorm:"type:varchar(80);not null;index"`
	Public      bool      `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (apiModel) TableName() string {
	return "access_apis"
}

func apiModelFromDomain(api domain.API) apiModel {
	return apiModel{
		ID:          api.ID,
		Method:      api.Method,
		Path:        api.Path,
		Description: api.Description,
		Group:       api.Group,
		Permission:  api.Permission,
		Public:      api.Public,
		CreatedAt:   api.CreatedAt,
		UpdatedAt:   api.UpdatedAt,
	}
}

func (m apiModel) toDomain() (domain.API, error) {
	return domain.RestoreAPI(m.ID, m.Method, m.Path, m.Description, m.Group, m.Permission, m.Public, m.CreatedAt, m.UpdatedAt)
}

type roleModel struct {
	ID          int64             `gorm:"primaryKey"`
	ParentID    int64             `gorm:"not null;index"`
	Code        string            `gorm:"type:varchar(64);not null;uniqueIndex"`
	Name        string            `gorm:"type:varchar(80);not null"`
	Permissions mysqljson.Strings `gorm:"type:json;not null"`
	MenuIDs     mysqljson.Int64s  `gorm:"type:json;not null"`
	APIIDs      mysqljson.Int64s  `gorm:"type:json;not null"`
	ButtonIDs   mysqljson.Int64s  `gorm:"type:json;not null"`
	DataRoleIDs mysqljson.Int64s  `gorm:"type:json"`
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
		APIIDs:      mysqljson.Int64s(role.APIIDs),
		ButtonIDs:   mysqljson.Int64s(role.ButtonIDs),
		DataRoleIDs: mysqljson.Int64s(role.DataRoleIDs),
		DefaultPath: role.DefaultPath,
		Active:      role.Active,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func (m roleModel) toDomain() (domain.Role, error) {
	return domain.RestoreRole(m.ID, m.ParentID, m.Code, m.Name, []string(m.Permissions), []int64(m.MenuIDs), []int64(m.APIIDs), []int64(m.ButtonIDs), []int64(m.DataRoleIDs), m.DefaultPath, m.Active, m.CreatedAt, m.UpdatedAt)
}

type menuModel struct {
	ID             int64             `gorm:"primaryKey"`
	ParentID       int64             `gorm:"not null"`
	Name           string            `gorm:"type:varchar(80);not null"`
	Path           string            `gorm:"type:varchar(160);not null;uniqueIndex"`
	Icon           string            `gorm:"type:varchar(80);not null"`
	Hidden         bool              `gorm:"not null"`
	Component      string            `gorm:"type:varchar(160);not null"`
	ActiveName     string            `gorm:"type:varchar(160);not null"`
	KeepAlive      bool              `gorm:"not null"`
	DefaultMenu    bool              `gorm:"not null"`
	CloseTab       bool              `gorm:"not null"`
	TransitionType string            `gorm:"type:varchar(80);not null"`
	Permission     string            `gorm:"type:varchar(80);not null"`
	Sort           int               `gorm:"not null"`
	Active         bool              `gorm:"not null"`
	Buttons        []menuButtonModel `gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time         `gorm:"not null"`
	UpdatedAt      time.Time         `gorm:"not null"`
}

func (menuModel) TableName() string {
	return "access_menus"
}

func menuModelFromDomain(menu domain.Menu) menuModel {
	return menuModel{
		ID:             menu.ID,
		ParentID:       menu.ParentID,
		Name:           menu.Name,
		Path:           menu.Path,
		Icon:           menu.Icon,
		Hidden:         menu.Hidden,
		Component:      menu.Component,
		ActiveName:     menu.Meta.ActiveName,
		KeepAlive:      menu.Meta.KeepAlive,
		DefaultMenu:    menu.Meta.DefaultMenu,
		CloseTab:       menu.Meta.CloseTab,
		TransitionType: menu.Meta.TransitionType,
		Permission:     menu.Permission,
		Sort:           menu.Sort,
		Active:         menu.Active,
		CreatedAt:      menu.CreatedAt,
		UpdatedAt:      menu.UpdatedAt,
	}
}

func (m menuModel) toDomain() (domain.Menu, error) {
	buttons := make([]domain.MenuButton, 0, len(m.Buttons))
	for _, buttonModel := range m.Buttons {
		button, err := buttonModel.toDomain()
		if err != nil {
			return domain.Menu{}, err
		}
		buttons = append(buttons, button)
	}
	meta := domain.MenuMeta{
		ActiveName:     m.ActiveName,
		KeepAlive:      m.KeepAlive,
		DefaultMenu:    m.DefaultMenu,
		CloseTab:       m.CloseTab,
		TransitionType: m.TransitionType,
	}
	return domain.RestoreMenu(m.ID, m.ParentID, m.Name, m.Path, m.Icon, m.Hidden, m.Component, meta, m.Permission, m.Sort, m.Active, buttons, m.CreatedAt, m.UpdatedAt)
}

type menuButtonModel struct {
	ID          int64     `gorm:"primaryKey"`
	MenuID      int64     `gorm:"not null;uniqueIndex:idx_access_menu_button_name"`
	Name        string    `gorm:"type:varchar(80);not null;uniqueIndex:idx_access_menu_button_name"`
	Description string    `gorm:"type:varchar(120);not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (menuButtonModel) TableName() string {
	return "access_menu_buttons"
}

func menuButtonModelFromDomain(button domain.MenuButton) menuButtonModel {
	return menuButtonModel{
		ID:          button.ID,
		MenuID:      button.MenuID,
		Name:        button.Name,
		Description: button.Description,
		CreatedAt:   button.CreatedAt,
		UpdatedAt:   button.UpdatedAt,
	}
}

func (m menuButtonModel) toDomain() (domain.MenuButton, error) {
	return domain.RestoreMenuButton(m.ID, m.MenuID, m.Name, m.Description, m.CreatedAt, m.UpdatedAt)
}

func loadMenu(ctx context.Context, tx *gorm.DB, id int64) (domain.Menu, error) {
	var model menuModel
	err := tx.WithContext(ctx).
		Preload("Buttons", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		First(&model, "id = ?", id).Error
	if err != nil {
		return domain.Menu{}, mapReadError(err, "menu", "load menu")
	}
	return model.toDomain()
}

func replaceMenuButtons(ctx context.Context, tx *gorm.DB, menuID int64, buttons []domain.MenuButton) error {
	var existing []menuButtonModel
	if err := tx.WithContext(ctx).Where("menu_id = ?", menuID).Find(&existing).Error; err != nil {
		return apperr.WrapDatabase(err, "list menu buttons")
	}
	byID := make(map[int64]menuButtonModel, len(existing))
	byName := make(map[string]menuButtonModel, len(existing))
	for _, button := range existing {
		byID[button.ID] = button
		byName[button.Name] = button
	}
	keptIDs := make([]int64, 0, len(buttons))
	for _, button := range buttons {
		model, ok := byID[button.ID]
		if !ok {
			model, ok = byName[button.Name]
		}
		if ok {
			model.Name = button.Name
			model.Description = button.Description
			if err := tx.WithContext(ctx).Save(&model).Error; err != nil {
				return mapWriteError(err, "menu button already exists", "update menu button")
			}
			keptIDs = append(keptIDs, model.ID)
			continue
		}
		now := time.Now().UTC()
		created, err := domain.RestoreMenuButton(0, menuID, button.Name, button.Description, now, now)
		if err != nil {
			return err
		}
		model = menuButtonModelFromDomain(created)
		if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
			return mapWriteError(err, "menu button already exists", "create menu button")
		}
		keptIDs = append(keptIDs, model.ID)
	}
	deleteQuery := tx.WithContext(ctx).Where("menu_id = ?", menuID)
	if len(keptIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", keptIDs)
	}
	if err := deleteQuery.Delete(&menuButtonModel{}).Error; err != nil {
		return apperr.WrapDatabase(err, "delete removed menu buttons")
	}
	return nil
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
