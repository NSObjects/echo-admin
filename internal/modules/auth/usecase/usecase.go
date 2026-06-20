package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

const tokenTTL = 12 * time.Hour

// casbinRBACModel maps UI permission tokens to Casbin's {subject, object, action}
// RBAC form. Users and roles are prefixed before insertion because Casbin treats
// both as plain strings.
const casbinRBACModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// Login authenticates an administrator and returns a bearer token.
func (u *Usecase) Login(ctx context.Context, input LoginInput) (LoginOutput, error) {
	if err := u.ready(); err != nil {
		return LoginOutput{}, err
	}
	username := strings.ToLower(strings.TrimSpace(input.Username))
	admin, err := u.admins.FindByUsername(ctx, username)
	if err != nil {
		if recordErr := u.recordLogin(ctx, loginRecordFromInput(0, username, input, false, "invalid credentials")); recordErr != nil {
			return LoginOutput{}, recordErr
		}
		return LoginOutput{}, apperr.New(apperr.ErrUnauthorized, "invalid username or password")
	}
	if !admin.Active {
		if recordErr := u.recordLogin(ctx, loginRecordFromInput(admin.ID, username, input, false, "account disabled")); recordErr != nil {
			return LoginOutput{}, recordErr
		}
		return LoginOutput{}, apperr.New(apperr.ErrAccountDisabled, "account disabled")
	}
	if compareErr := bcrypt.CompareHashAndPassword(admin.PasswordHash, []byte(input.Password)); compareErr != nil {
		if recordErr := u.recordLogin(ctx, loginRecordFromInput(admin.ID, username, input, false, "invalid credentials")); recordErr != nil {
			return LoginOutput{}, recordErr
		}
		return LoginOutput{}, apperr.New(apperr.ErrUnauthorized, "invalid username or password")
	}

	user, err := u.userSnapshot(ctx, admin, admin.ActiveRoleID)
	if err != nil {
		return LoginOutput{}, err
	}
	token, err := u.issueToken(admin.ID, admin.Username, user.ActiveRoleID, user.Permissions)
	if err != nil {
		return LoginOutput{}, err
	}
	if err := u.recordLogin(ctx, loginRecordFromInput(admin.ID, username, input, true, "login succeeded")); err != nil {
		return LoginOutput{}, err
	}
	return LoginOutput{Token: token, User: user}, nil
}

// CurrentUser returns the authenticated administrator profile and active-role grants.
func (u *Usecase) CurrentUser(ctx context.Context) (CurrentUser, error) {
	if err := u.ready(); err != nil {
		return CurrentUser{}, err
	}
	admin, activeRoleID, err := u.currentAdminAndRole(ctx)
	if err != nil {
		return CurrentUser{}, err
	}
	return u.userSnapshot(ctx, admin, activeRoleID)
}

// UpdateProfile changes the current administrator's display fields only.
func (u *Usecase) UpdateProfile(ctx context.Context, input UpdateProfileInput) (CurrentUser, error) {
	if err := u.ready(); err != nil {
		return CurrentUser{}, err
	}
	admin, activeRoleID, err := u.currentAdminAndRole(ctx)
	if err != nil {
		return CurrentUser{}, err
	}
	if !admin.Active {
		return CurrentUser{}, apperr.New(apperr.ErrAccountDisabled, "account disabled")
	}
	updated, err := admin.UpdateProfile(input.DisplayName, input.Email, admin.RoleIDs, admin.ActiveRoleID, admin.Active)
	if err != nil {
		return CurrentUser{}, apperr.NewBadRequest("invalid profile")
	}
	saved, err := u.admins.Update(ctx, updated)
	if err != nil {
		return CurrentUser{}, err
	}
	return u.userSnapshot(ctx, saved, activeRoleID)
}

// SwitchRole persists a new active role and returns a token scoped to that role.
func (u *Usecase) SwitchRole(ctx context.Context, input RoleSwitchInput) (RoleSwitchOutput, error) {
	if err := u.ready(); err != nil {
		return RoleSwitchOutput{}, err
	}
	adminID, err := currentAdminID(ctx)
	if err != nil {
		return RoleSwitchOutput{}, err
	}
	admin, err := u.admins.FindByID(ctx, adminID)
	if err != nil {
		return RoleSwitchOutput{}, err
	}
	if !admin.Active {
		return RoleSwitchOutput{}, apperr.New(apperr.ErrAccountDisabled, "account disabled")
	}
	switched, err := admin.SwitchActiveRole(input.RoleID)
	if err != nil {
		return RoleSwitchOutput{}, apperr.NewBadRequest("active role is not assigned")
	}
	saved, err := u.admins.Update(ctx, switched)
	if err != nil {
		return RoleSwitchOutput{}, err
	}
	user, err := u.userSnapshot(ctx, saved, input.RoleID)
	if err != nil {
		return RoleSwitchOutput{}, err
	}
	token, err := u.issueToken(saved.ID, saved.Username, user.ActiveRoleID, user.Permissions)
	if err != nil {
		return RoleSwitchOutput{}, err
	}
	return RoleSwitchOutput{Token: token, User: user}, nil
}

// ChangePassword verifies the current password, stores the new hash, and
// revokes the token used for the password-changing request.
func (u *Usecase) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	if err := u.ready(); err != nil {
		return err
	}
	adminID, err := currentAdminID(ctx)
	if err != nil {
		return err
	}
	admin, err := u.admins.FindByID(ctx, adminID)
	if err != nil {
		return err
	}
	if !admin.Active {
		return apperr.New(apperr.ErrAccountDisabled, "account disabled")
	}
	if compareErr := bcrypt.CompareHashAndPassword(admin.PasswordHash, []byte(input.CurrentPassword)); compareErr != nil {
		return apperr.New(apperr.ErrUnauthorized, "invalid current password")
	}
	hash, err := hashPassword(input.NewPassword)
	if err != nil {
		return err
	}
	updated, err := admin.ReplacePassword(hash)
	if err != nil {
		return apperr.NewBadRequest("invalid password")
	}
	if _, err := u.admins.Update(ctx, updated); err != nil {
		return err
	}
	return u.blacklistTokenForAdmin(ctx, input.RawToken, adminID)
}

// Logout revokes the current JWT so it cannot be reused after the client leaves.
func (u *Usecase) Logout(ctx context.Context, rawToken string) error {
	if err := u.ready(); err != nil {
		return err
	}
	adminID, err := currentAdminID(ctx)
	if err != nil {
		return err
	}
	return u.blacklistTokenForAdmin(ctx, rawToken, adminID)
}

func (u *Usecase) blacklistTokenForAdmin(ctx context.Context, rawToken string, adminID int64) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return apperr.NewUnauthorized()
	}
	claims := jwt.MapClaims{}
	token, err := jwt.NewParser(jwt.WithTimeFunc(u.now)).ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, apperr.NewUnauthorized()
		}
		return u.jwtSecret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return apperr.NewUnauthorized()
	}
	subject, err := claims.GetSubject()
	if err != nil || subject != strconv.FormatInt(adminID, 10) {
		return apperr.NewUnauthorized()
	}
	expiresAt, err := claims.GetExpirationTime()
	if err != nil || expiresAt == nil {
		return apperr.NewUnauthorized()
	}
	return u.jwtBlacklist.AddJWTBlacklist(ctx, JWTBlacklistEntry{
		TokenHash: jwtTokenHash(rawToken),
		ExpiresAt: expiresAt.UTC(),
	})
}

// TokenBlocked reports whether a raw JWT has been revoked and has not expired.
func (u *Usecase) TokenBlocked(ctx context.Context, rawToken string) (bool, error) {
	if err := u.ready(); err != nil {
		return false, err
	}
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return false, nil
	}
	return u.jwtBlacklist.JWTBlacklisted(ctx, jwtTokenHash(rawToken), u.now())
}

// RequirePermission verifies that the current admin has permission through the active role.
func (u *Usecase) RequirePermission(ctx context.Context, permission string) error {
	if err := u.ready(); err != nil {
		return err
	}
	admin, activeRoleID, err := u.currentAdminAndRole(ctx)
	if err != nil {
		return err
	}
	snapshot, err := u.casbinSnapshot(ctx, admin, activeRoleID)
	if err != nil {
		return err
	}
	return requirePermission(snapshot, permission)
}

// RequireRoutePermission verifies both the semantic permission token and the
// managed API route grant for the active role. The route check is intentionally
// tied to Echo's registered route pattern, not the raw URL, so IDs and other
// path parameters do not need to be persisted as concrete values.
func (u *Usecase) RequireRoutePermission(ctx context.Context, permission, method, path string) error {
	if err := u.ready(); err != nil {
		return err
	}
	admin, activeRoleID, err := u.currentAdminAndRole(ctx)
	if err != nil {
		return err
	}
	snapshot, err := u.casbinSnapshot(ctx, admin, activeRoleID)
	if err != nil {
		return err
	}
	api, err := u.findRouteAPI(ctx, method, path)
	if err != nil {
		return err
	}
	if err := requireDeclaredRoutePermissions(snapshot, permission, api.Permission); err != nil {
		return err
	}
	return requireAssignedRouteAPI(snapshot.activeRole, api)
}

func (u *Usecase) ready() error {
	if u == nil || u.admins == nil || u.roles == nil || u.menus == nil || u.apis == nil || u.logins == nil || u.jwtBlacklist == nil {
		return apperr.New(apperr.ErrInternalServer, "auth dependencies are not configured")
	}
	if len(u.jwtSecret) == 0 {
		return apperr.New(apperr.ErrInternalServer, "jwt secret is not configured")
	}
	return nil
}

func (u *Usecase) currentAdminAndRole(ctx context.Context) (identitydomain.Admin, int64, error) {
	adminID, err := currentAdminID(ctx)
	if err != nil {
		return identitydomain.Admin{}, 0, err
	}
	admin, err := u.admins.FindByID(ctx, adminID)
	if err != nil {
		return identitydomain.Admin{}, 0, err
	}
	if !admin.Active {
		return identitydomain.Admin{}, 0, apperr.New(apperr.ErrAccountDisabled, "account disabled")
	}
	activeRoleID := admin.ActiveRoleID
	if roleID, parseErr := currentRoleID(ctx); parseErr == nil && roleID > 0 {
		activeRoleID = roleID
	}
	if !admin.HasRole(activeRoleID) {
		return identitydomain.Admin{}, 0, apperr.NewUnauthorized()
	}
	return admin, activeRoleID, nil
}

func (u *Usecase) userSnapshot(ctx context.Context, admin identitydomain.Admin, activeRoleID int64) (CurrentUser, error) {
	snapshot, err := u.casbinSnapshot(ctx, admin, activeRoleID)
	if err != nil {
		return CurrentUser{}, err
	}
	menus, err := u.visibleMenus(ctx, snapshot)
	if err != nil {
		return CurrentUser{}, err
	}
	return CurrentUser{
		ID:           admin.ID,
		Username:     admin.Username,
		DisplayName:  admin.DisplayName,
		Email:        admin.Email,
		ActiveRoleID: snapshot.activeRole.ID,
		ActiveRole:   snapshot.activeRole,
		DefaultPath:  snapshot.activeRole.DefaultPath,
		Roles:        snapshot.roles,
		Permissions:  snapshot.permissions,
		Menus:        menus,
	}, nil
}

func (u *Usecase) casbinSnapshot(ctx context.Context, admin identitydomain.Admin, activeRoleID int64) (rbacSnapshot, error) {
	roleIDs := admin.RoleIDs
	roles := make([]Role, 0, len(roleIDs))
	var menuIDSet map[int64]struct{}
	var buttonIDSet map[int64]struct{}
	enforcer, err := newCasbinEnforcer()
	if err != nil {
		return rbacSnapshot{}, err
	}
	user := userSubject(admin.ID)
	var activeRole Role
	activeFound := false
	for _, roleID := range roleIDs {
		role, findErr := u.roles.FindRoleByID(ctx, roleID)
		if findErr != nil {
			return rbacSnapshot{}, findErr
		}
		if !role.Active {
			continue
		}
		dto := fromRole(role)
		roles = append(roles, dto)
		if role.ID != activeRoleID {
			continue
		}
		activeFound = true
		activeRole = dto
		menuIDSet, buttonIDSet, err = addActiveRoleGrants(enforcer, user, role)
		if err != nil {
			return rbacSnapshot{}, err
		}
	}
	if !activeFound {
		return rbacSnapshot{}, apperr.NewPermissionDenied("role", strconv.FormatInt(activeRoleID, 10))
	}
	permissions, err := implicitPermissions(enforcer, user)
	if err != nil {
		return rbacSnapshot{}, err
	}
	return rbacSnapshot{enforcer: enforcer, user: user, activeRole: activeRole, roles: roles, permissions: permissions, menuIDSet: menuIDSet, buttonIDSet: buttonIDSet}, nil
}

func addActiveRoleGrants(enforcer *casbin.Enforcer, user string, role accessdomain.Role) (map[int64]struct{}, map[int64]struct{}, error) {
	roleName := roleSubject(role.Code)
	if _, err := enforcer.AddRoleForUser(user, roleName); err != nil {
		return nil, nil, fmt.Errorf("add casbin role: %w", err)
	}
	for _, permission := range role.Permissions {
		obj, act, splitErr := splitPermission(permission)
		if splitErr != nil {
			return nil, nil, splitErr
		}
		if _, err := enforcer.AddPolicy(roleName, obj, act); err != nil {
			return nil, nil, fmt.Errorf("add casbin policy: %w", err)
		}
	}
	menuIDSet := make(map[int64]struct{}, len(role.MenuIDs))
	for _, menuID := range role.MenuIDs {
		menuIDSet[menuID] = struct{}{}
	}
	buttonIDSet := make(map[int64]struct{}, len(role.ButtonIDs))
	for _, buttonID := range role.ButtonIDs {
		buttonIDSet[buttonID] = struct{}{}
	}
	return menuIDSet, buttonIDSet, nil
}

func (u *Usecase) findRouteAPI(ctx context.Context, method, path string) (accessdomain.API, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	if method == "" || path == "" {
		return accessdomain.API{}, apperr.NewPermissionDenied("api", "route")
	}
	apis, err := u.apis.ListAPIs(ctx)
	if err != nil {
		return accessdomain.API{}, err
	}
	for _, api := range apis {
		if api.Method == method && api.Path == path {
			return api, nil
		}
	}
	return accessdomain.API{}, apperr.NewPermissionDenied("api", path)
}

func requireDeclaredRoutePermissions(snapshot rbacSnapshot, handlerPermission, apiPermission string) error {
	if apiPermission != "" {
		if err := requirePermission(snapshot, apiPermission); err != nil {
			return err
		}
	}
	if handlerPermission == "" || handlerPermission == apiPermission {
		return nil
	}
	return requirePermission(snapshot, handlerPermission)
}

func requireAssignedRouteAPI(role Role, api accessdomain.API) error {
	if api.Public || role.Code == accessdomain.RoleCodeSuperAdmin {
		return nil
	}
	if containsInt64(role.APIIDs, api.ID) {
		return nil
	}
	return apperr.NewPermissionDenied("api", api.Path)
}

func (u *Usecase) visibleMenus(ctx context.Context, snapshot rbacSnapshot) ([]Menu, error) {
	menus, err := u.menus.ListMenus(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Menu, 0, len(menus))
	for _, menu := range menus {
		if !menu.Active {
			continue
		}
		if _, ok := snapshot.menuIDSet[menu.ID]; !ok {
			continue
		}
		if permission := menu.Permission; permission != "" {
			allowed, err := enforcePermission(snapshot.enforcer, snapshot.user, permission)
			if err != nil {
				return nil, err
			}
			if !allowed {
				continue
			}
		}
		out = append(out, fromMenu(menu, visibleButtons(menu.Buttons, snapshot)))
	}
	return out, nil
}

func (u *Usecase) issueToken(adminID int64, username string, activeRoleID int64, permissions []string) (string, error) {
	now := u.now()
	claims := jwt.MapClaims{
		"sub":         strconv.FormatInt(adminID, 10),
		"username":    username,
		"role_id":     strconv.FormatInt(activeRoleID, 10),
		"permissions": permissions,
		"iat":         now.Unix(),
		"exp":         now.Add(tokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign jwt token: %w", err)
	}
	return signed, nil
}

func (u *Usecase) recordLogin(ctx context.Context, record LoginRecord) error {
	return u.logins.RecordLogin(ctx, record)
}

func hashPassword(password string) ([]byte, error) {
	if len(password) < 8 || len(password) > 72 {
		return nil, apperr.NewBadRequest("invalid password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	return hash, nil
}

func jwtTokenHash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func currentAdminID(ctx context.Context) (int64, error) {
	raw := requestctx.GetUserID(ctx)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.NewUnauthorized()
	}
	return id, nil
}

func currentRoleID(ctx context.Context) (int64, error) {
	raw := requestctx.GetRoleID(ctx)
	if raw == "" {
		return 0, apperr.NewUnauthorized()
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.NewUnauthorized()
	}
	return id, nil
}

func fromRole(role accessdomain.Role) Role {
	return Role{
		ID:          role.ID,
		ParentID:    role.ParentID,
		Code:        role.Code,
		Name:        role.Name,
		Permissions: role.Permissions,
		MenuIDs:     role.MenuIDs,
		APIIDs:      role.APIIDs,
		ButtonIDs:   role.ButtonIDs,
		DataRoleIDs: role.DataRoleIDs,
		DefaultPath: role.DefaultPath,
		Active:      role.Active,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func fromMenu(menu accessdomain.Menu, buttons []accessdomain.MenuButton) Menu {
	return Menu{
		ID:        menu.ID,
		ParentID:  menu.ParentID,
		Name:      menu.Name,
		Path:      menu.Path,
		Icon:      menu.Icon,
		Hidden:    menu.Hidden,
		Component: menu.Component,
		Meta: MenuMeta{
			ActiveName:     menu.Meta.ActiveName,
			KeepAlive:      menu.Meta.KeepAlive,
			DefaultMenu:    menu.Meta.DefaultMenu,
			CloseTab:       menu.Meta.CloseTab,
			TransitionType: menu.Meta.TransitionType,
		},
		Permission: menu.Permission,
		Sort:       menu.Sort,
		Active:     menu.Active,
		Buttons:    fromButtons(buttons),
		CreatedAt:  menu.CreatedAt,
		UpdatedAt:  menu.UpdatedAt,
	}
}

func visibleButtons(buttons []accessdomain.MenuButton, snapshot rbacSnapshot) []accessdomain.MenuButton {
	if snapshot.activeRole.Code == accessdomain.RoleCodeSuperAdmin {
		return buttons
	}
	out := make([]accessdomain.MenuButton, 0, len(buttons))
	for _, button := range buttons {
		if _, ok := snapshot.buttonIDSet[button.ID]; ok {
			out = append(out, button)
		}
	}
	return out
}

func fromButtons(buttons []accessdomain.MenuButton) []Button {
	out := make([]Button, 0, len(buttons))
	for _, button := range buttons {
		out = append(out, Button{
			ID:          button.ID,
			MenuID:      button.MenuID,
			Name:        button.Name,
			Description: button.Description,
			CreatedAt:   button.CreatedAt,
			UpdatedAt:   button.UpdatedAt,
		})
	}
	return out
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func loginRecordFromInput(adminID int64, username string, input LoginInput, success bool, reason string) LoginRecord {
	if username == "" {
		username = "unknown"
	}
	return LoginRecord{
		AdminID:   adminID,
		Username:  username,
		IP:        input.IP,
		UserAgent: input.UserAgent,
		Success:   success,
		Reason:    reason,
	}
}

type rbacSnapshot struct {
	enforcer    *casbin.Enforcer
	user        string
	activeRole  Role
	roles       []Role
	permissions []string
	menuIDSet   map[int64]struct{}
	buttonIDSet map[int64]struct{}
}

func newCasbinEnforcer() (*casbin.Enforcer, error) {
	rbacModel, err := model.NewModelFromString(casbinRBACModel)
	if err != nil {
		return nil, fmt.Errorf("create casbin model: %w", err)
	}
	enforcer, err := casbin.NewEnforcer(rbacModel)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}
	enforcer.EnableAutoSave(false)
	return enforcer, nil
}

func implicitPermissions(enforcer *casbin.Enforcer, user string) ([]string, error) {
	policies, err := enforcer.GetImplicitPermissionsForUser(user)
	if err != nil {
		return nil, fmt.Errorf("get casbin permissions: %w", err)
	}
	set := make(map[string]struct{}, len(policies))
	for _, policy := range policies {
		if len(policy) < 3 {
			continue
		}
		set[policy[1]+":"+policy[2]] = struct{}{}
	}
	return sortedKeys(set), nil
}

func enforcePermission(enforcer *casbin.Enforcer, user, permission string) (bool, error) {
	obj, act, err := splitPermission(permission)
	if err != nil {
		return false, err
	}
	allowed, err := enforcer.Enforce(user, obj, act)
	if err != nil {
		return false, fmt.Errorf("enforce casbin permission: %w", err)
	}
	return allowed, nil
}

func requirePermission(snapshot rbacSnapshot, permission string) error {
	allowed, err := enforcePermission(snapshot.enforcer, snapshot.user, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return apperr.NewPermissionDenied("admin", permission)
	}
	return nil
}

func containsInt64(values []int64, want int64) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func splitPermission(permission string) (string, string, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(permission)), ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", apperr.NewBadRequest("invalid permission")
	}
	return parts[0], parts[1], nil
}

func userSubject(adminID int64) string {
	return "user:" + strconv.FormatInt(adminID, 10)
}

func roleSubject(code string) string {
	return "role:" + code
}
