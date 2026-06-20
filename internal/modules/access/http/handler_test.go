package accesshttp_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	accesshttp "github.com/NSObjects/echo-admin/internal/modules/access/http"
	accessusecase "github.com/NSObjects/echo-admin/internal/modules/access/usecase"
	auditusecase "github.com/NSObjects/echo-admin/internal/modules/audit/usecase"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
	"github.com/NSObjects/echo-admin/internal/platform/server/middlewares"
)

func TestCreateRoleRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)

	rec := doJSON(t, e, http.MethodPost, "/api/roles", `{"code":"operator","name":"运营","permissions":["log:read"],"menu_ids":[1],"active":true}`, "42")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create role status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionRoleCreate {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionRoleCreate)
	}
	if store.createRoleCalls != 1 {
		t.Fatalf("createRoleCalls = %d, want 1", store.createRoleCalls)
	}
	if got := store.createdRole.Code; got != "operator" {
		t.Fatalf("created role code = %q, want operator", got)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].ActorID; got != 42 {
		t.Fatalf("operation actor id = %d, want 42", got)
	}
	if got := recorder.records[0].Resource; got != "role" {
		t.Fatalf("operation resource = %q, want role", got)
	}
}

func TestCopyRoleRequiresCreatePermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)
	store.roles = twoRoles(t)

	rec := doJSON(t, e, http.MethodPost, "/api/roles/2/copy", `{"code":"operator_copy","name":"运营副本"}`, "42")
	if rec.Code != http.StatusCreated {
		t.Fatalf("copy role status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionRoleCreate {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionRoleCreate)
	}
	if store.createRoleCalls != 1 {
		t.Fatalf("createRoleCalls = %d, want 1", store.createRoleCalls)
	}
	if got := store.createdRole.Code; got != "operator_copy" {
		t.Fatalf("created role code = %q, want operator_copy", got)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].Action; got != "copy" {
		t.Fatalf("operation action = %q, want copy", got)
	}
}

func TestListMenusRejectsUnauthorizedBeforeStore(t *testing.T) {
	e, store, recorder, _ := newAccessEcho(apperr.NewPermissionDenied("menu", "read"))

	rec := doJSON(t, e, http.MethodGet, "/api/menus", "", "42")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("list menus status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if store.listMenuCalls != 0 {
		t.Fatalf("listMenuCalls = %d, want 0", store.listMenuCalls)
	}
	if len(recorder.records) != 0 {
		t.Fatalf("operation records = %d, want 0", len(recorder.records))
	}
}

func TestCreateMenuAcceptsMetaAndButtons(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)

	rec := doJSON(t, e, http.MethodPost, "/api/menus", `{"name":"菜单管理","path":"/menus","icon":"menu","component":"./Menus","meta":{"keep_alive":true},"permission":"menu:read","sort":40,"active":true,"buttons":[{"name":"create","description":"新增菜单"}]}`, "42")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create menu status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionMenuCreate {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionMenuCreate)
	}
	if store.createdMenu.Component != "./Menus" {
		t.Fatalf("created menu component = %q, want ./Menus", store.createdMenu.Component)
	}
	if !store.createdMenu.Meta.KeepAlive {
		t.Fatal("created menu KeepAlive = false, want true")
	}
	if len(store.createdMenu.Buttons) != 1 || store.createdMenu.Buttons[0].Name != "create" {
		t.Fatalf("created menu buttons = %#v, want create button", store.createdMenu.Buttons)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
}

func TestReadMenuRequiresPermission(t *testing.T) {
	e, store, _, auth := newAccessEcho(nil)
	menu, err := accessdomain.RestoreMenu(2, 0, "菜单管理", "/menus", "menu", false, "./Menus", accessdomain.MenuMeta{}, accessdomain.PermissionMenuRead, 20, true, nil, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreMenu() error = %v", err)
	}
	store.menus = []accessdomain.Menu{menu}

	rec := doJSON(t, e, http.MethodGet, "/api/menus/2", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("get menu status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionMenuRead {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionMenuRead)
	}
}

func TestDeleteRoleRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)
	store.roles = twoRoles(t)

	rec := doJSON(t, e, http.MethodDelete, "/api/roles/2", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete role status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionRoleDelete {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionRoleDelete)
	}
	if store.deletedRoleID != 2 {
		t.Fatalf("deletedRoleID = %d, want 2", store.deletedRoleID)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].Action; got != "delete" {
		t.Fatalf("operation action = %q, want delete", got)
	}
}

func TestDeleteMenuRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)
	menu, err := accessdomain.RestoreMenu(2, 0, "临时菜单", "/scratch", "menu", false, "./Scratch", accessdomain.MenuMeta{}, "", 20, true, nil, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreMenu() error = %v", err)
	}
	store.menus = []accessdomain.Menu{menu}

	rec := doJSON(t, e, http.MethodDelete, "/api/menus/2", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete menu status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionMenuDelete {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionMenuDelete)
	}
	if store.deletedMenuID != 2 {
		t.Fatalf("deletedMenuID = %d, want 2", store.deletedMenuID)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].Resource; got != "menu" {
		t.Fatalf("operation resource = %q, want menu", got)
	}
}

func TestSetMenuRolesRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)
	store.roles = twoRoles(t)
	menu, err := accessdomain.RestoreMenu(2, 0, "菜单管理", "/menus", "menu", false, "./Menus", accessdomain.MenuMeta{}, accessdomain.PermissionMenuRead, 20, true, nil, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreMenu() error = %v", err)
	}
	store.menus = []accessdomain.Menu{menu}

	rec := doJSON(t, e, http.MethodPut, "/api/menus/2/roles", `{"role_ids":[2]}`, "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("set menu roles status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionMenuUpdate {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionMenuUpdate)
	}
	if len(store.updatedRoles) != 1 || store.updatedRoles[0].ID != 2 {
		t.Fatalf("updatedRoles = %#v, want role 2 update", store.updatedRoles)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].Action; got != "set_roles" {
		t.Fatalf("operation action = %q, want set_roles", got)
	}
}

func TestCreateAPIRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)

	rec := doJSON(t, e, http.MethodPost, "/api/apis", `{"method":"GET","path":"/api/example","description":"示例API","group":"example","permission":"log:read","public":false}`, "42")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create api status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionAPICreate {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionAPICreate)
	}
	if got := store.createdAPI.Path; got != "/api/example" {
		t.Fatalf("created api path = %q, want /api/example", got)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].Resource; got != "api" {
		t.Fatalf("operation resource = %q, want api", got)
	}
}

func TestListAPIGroupsRequiresPermission(t *testing.T) {
	e, store, _, auth := newAccessEcho(nil)
	firstAPI, err := accessdomain.RestoreAPI(1, "GET", "/api/a", "A", "admin", accessdomain.PermissionAdminRead, false, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreAPI(first) error = %v", err)
	}
	secondAPI, err := accessdomain.RestoreAPI(2, "GET", "/api/b", "B", "log", accessdomain.PermissionLogRead, false, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreAPI(second) error = %v", err)
	}
	store.apis = []accessdomain.API{secondAPI, firstAPI}

	rec := doJSON(t, e, http.MethodGet, "/api/apis/groups", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("list api groups status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionAPIRead {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionAPIRead)
	}
}

func TestReadAPIRequiresPermission(t *testing.T) {
	e, store, _, auth := newAccessEcho(nil)
	api, err := accessdomain.RestoreAPI(3, "GET", "/api/example", "示例API", "example", accessdomain.PermissionLogRead, false, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreAPI() error = %v", err)
	}
	store.apis = []accessdomain.API{api}

	rec := doJSON(t, e, http.MethodGet, "/api/apis/3", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("get api status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionAPIRead {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionAPIRead)
	}
}

func TestReadAPIRolesRequiresPermission(t *testing.T) {
	e, store, _, auth := newAccessEcho(nil)
	api, err := accessdomain.RestoreAPI(3, "GET", "/api/example", "示例API", "example", accessdomain.PermissionLogRead, false, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreAPI() error = %v", err)
	}
	store.apis = []accessdomain.API{api}
	store.roles = twoRoles(t)

	rec := doJSON(t, e, http.MethodGet, "/api/apis/3/roles", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("get api roles status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionAPIRead {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionAPIRead)
	}
}

func TestBatchDeleteAPIsRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)
	firstAPI, err := accessdomain.RestoreAPI(3, "GET", "/api/example", "示例API", "example", accessdomain.PermissionLogRead, false, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreAPI(first) error = %v", err)
	}
	secondAPI, err := accessdomain.RestoreAPI(4, "POST", "/api/example", "创建示例", "example", accessdomain.PermissionLogRead, false, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreAPI(second) error = %v", err)
	}
	store.apis = []accessdomain.API{firstAPI, secondAPI}

	rec := doJSON(t, e, http.MethodPost, "/api/apis/batch-delete", `{"ids":[3,4]}`, "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("batch delete api status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionAPIDelete {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionAPIDelete)
	}
	if !sameInt64s(store.deletedAPIIDs, []int64{3, 4}) {
		t.Fatalf("deletedAPIIDs = %v, want [3 4]", store.deletedAPIIDs)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].ResourceID; got != "batch" {
		t.Fatalf("operation resource id = %q, want batch", got)
	}
}

func TestDeleteAPIRequiresPermissionAndRecordsOperation(t *testing.T) {
	e, store, recorder, auth := newAccessEcho(nil)
	api, err := accessdomain.RestoreAPI(3, "GET", "/api/example", "示例API", "example", accessdomain.PermissionLogRead, false, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreAPI() error = %v", err)
	}
	store.apis = []accessdomain.API{api}

	rec := doJSON(t, e, http.MethodDelete, "/api/apis/3", "", "42")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete api status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(auth.permissions) != 1 || auth.permissions[0] != accessdomain.PermissionAPIDelete {
		t.Fatalf("permissions = %v, want [%q]", auth.permissions, accessdomain.PermissionAPIDelete)
	}
	if !sameInt64s(store.deletedAPIIDs, []int64{3}) {
		t.Fatalf("deletedAPIIDs = %v, want [3]", store.deletedAPIIDs)
	}
	if len(recorder.records) != 1 {
		t.Fatalf("operation records = %d, want 1", len(recorder.records))
	}
	if got := recorder.records[0].Action; got != "delete" {
		t.Fatalf("operation action = %q, want delete", got)
	}
}

func newAccessEcho(authErr error) (*echo.Echo, *accessStore, *operationRecorder, *accessAuthorizer) {
	store := &accessStore{}
	uc := accessusecase.New(store, accessAdminRoleReader{})
	auth := &accessAuthorizer{err: authErr}
	recorder := &operationRecorder{}
	handler := accesshttp.New(uc, auth, recorder)

	e := echo.New()
	e.Validator = &middlewares.Validator{Validator: validator.New()}
	e.HTTPErrorHandler = middlewares.ErrorHandler
	accesshttp.Register(e.Group("/api"), handler)
	return e, store, recorder, auth
}

func doJSON(t *testing.T, e *echo.Echo, method, path, body, userID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	req = req.WithContext(requestctx.WithUserID(req.Context(), userID))
	req = req.WithContext(requestctx.WithRoleID(req.Context(), "1"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

type accessStore struct {
	createRoleCalls int
	listMenuCalls   int
	createdRole     accessdomain.Role
	createdAPI      accessdomain.API
	createdMenu     accessdomain.Menu
	updatedRoles    []accessdomain.Role
	roles           []accessdomain.Role
	menus           []accessdomain.Menu
	apis            []accessdomain.API
	deletedRoleID   int64
	deletedMenuID   int64
	deletedAPIIDs   []int64
}

func (s *accessStore) FindRoleByID(_ context.Context, id int64) (accessdomain.Role, error) {
	for _, role := range s.roles {
		if role.ID == id {
			return role, nil
		}
	}
	return accessdomain.Role{}, apperr.NewNotFound("role")
}

func (s *accessStore) FindRoleByCode(context.Context, string) (accessdomain.Role, error) {
	return accessdomain.Role{}, apperr.NewNotFound("role")
}

func (s *accessStore) ListAllRoles(ctx context.Context) ([]accessdomain.Role, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(s.roles) > 0 {
		return s.roles, nil
	}
	role, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", accessusecase.AllPermissions(), []int64{1}, []int64{1}, []int64{1}, []int64{1}, accessdomain.DefaultRolePath, true, fixedTime(), fixedTime())
	if err != nil {
		return nil, err
	}
	return []accessdomain.Role{role}, nil
}

func (s *accessStore) CreateRole(ctx context.Context, role accessdomain.Role) (accessdomain.Role, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Role{}, err
	}
	s.createRoleCalls++
	s.createdRole = role
	return accessdomain.RestoreRole(10, role.ParentID, role.Code, role.Name, role.Permissions, role.MenuIDs, role.APIIDs, role.ButtonIDs, role.DataRoleIDs, role.DefaultPath, role.Active, fixedTime(), fixedTime())
}

func (s *accessStore) UpdateRole(ctx context.Context, role accessdomain.Role) (accessdomain.Role, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Role{}, err
	}
	s.updatedRoles = append(s.updatedRoles, role)
	for index, existing := range s.roles {
		if existing.ID == role.ID {
			s.roles[index] = role
			return role, nil
		}
	}
	s.roles = append(s.roles, role)
	return role, nil
}

func (s *accessStore) DeleteRole(_ context.Context, id int64) error {
	s.deletedRoleID = id
	return nil
}

func (s *accessStore) FindAPIByID(ctx context.Context, id int64) (accessdomain.API, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.API{}, err
	}
	for _, api := range s.apis {
		if api.ID == id {
			return api, nil
		}
	}
	return accessdomain.RestoreAPI(id, "GET", "/api/existing", "Existing", "example", accessdomain.PermissionLogRead, false, fixedTime(), fixedTime())
}

func (s *accessStore) ListAPIs(ctx context.Context) ([]accessdomain.API, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.apis, nil
}

func (s *accessStore) CreateAPI(ctx context.Context, api accessdomain.API) (accessdomain.API, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.API{}, err
	}
	s.createdAPI = api
	return accessdomain.RestoreAPI(11, api.Method, api.Path, api.Description, api.Group, api.Permission, api.Public, fixedTime(), fixedTime())
}

func (s *accessStore) UpdateAPI(ctx context.Context, api accessdomain.API) (accessdomain.API, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.API{}, err
	}
	return api, nil
}

func (s *accessStore) DeleteAPI(_ context.Context, id int64) error {
	s.deletedAPIIDs = append(s.deletedAPIIDs, id)
	return nil
}

func (s *accessStore) ListMenus(ctx context.Context) ([]accessdomain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.listMenuCalls++
	return s.menus, nil
}

func (s *accessStore) FindMenuByID(ctx context.Context, id int64) (accessdomain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Menu{}, err
	}
	for _, menu := range s.menus {
		if menu.ID == id {
			return menu, nil
		}
	}
	return accessdomain.RestoreMenu(id, 0, "Existing", "/existing", "menu", false, "./Existing", accessdomain.MenuMeta{}, "log:read", 10, true, nil, fixedTime(), fixedTime())
}

func (s *accessStore) CreateMenu(ctx context.Context, menu accessdomain.Menu) (accessdomain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Menu{}, err
	}
	s.createdMenu = menu
	return menu, nil
}

func (s *accessStore) UpdateMenu(ctx context.Context, menu accessdomain.Menu) (accessdomain.Menu, error) {
	if err := ctx.Err(); err != nil {
		return accessdomain.Menu{}, err
	}
	return menu, nil
}

func (s *accessStore) DeleteMenu(_ context.Context, id int64) error {
	s.deletedMenuID = id
	return nil
}

type accessAuthorizer struct {
	err         error
	permissions []string
}

func (a *accessAuthorizer) RequireRoutePermission(_ context.Context, permission, _, _ string) error {
	a.permissions = append(a.permissions, permission)
	return a.err
}

type operationRecorder struct {
	records []auditusecase.OperationInput
}

func (r *operationRecorder) RecordOperation(_ context.Context, input auditusecase.OperationInput) (auditusecase.OperationLog, error) {
	r.records = append(r.records, input)
	return auditusecase.OperationLog{}, nil
}

type accessAdminRoleReader struct{}

func (accessAdminRoleReader) AdminRoleState(context.Context, int64) (accessusecase.AdminRoleState, error) {
	return accessusecase.AdminRoleState{RoleIDs: []int64{1}, ActiveRoleID: 1}, nil
}

func (accessAdminRoleReader) RoleAssigned(context.Context, int64) (bool, error) {
	return false, nil
}

func twoRoles(t *testing.T) []accessdomain.Role {
	t.Helper()
	root, err := accessdomain.RestoreRole(1, 0, accessdomain.RoleCodeSuperAdmin, "超级管理员", accessusecase.AllPermissions(), []int64{1}, []int64{1}, []int64{1}, []int64{1, 2}, accessdomain.DefaultRolePath, true, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreRole(root) error = %v", err)
	}
	operator, err := accessdomain.RestoreRole(2, 1, "operator", "运营", []string{accessdomain.PermissionAdminRead}, []int64{1}, []int64{1}, []int64{1}, []int64{2}, "/admins", true, fixedTime(), fixedTime())
	if err != nil {
		t.Fatalf("RestoreRole(operator) error = %v", err)
	}
	return []accessdomain.Role{root, operator}
}

func fixedTime() time.Time {
	return time.Unix(1_800_000_000, 0).UTC()
}

func sameInt64s(got, want []int64) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
