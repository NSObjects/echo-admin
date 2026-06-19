package usecase

import (
	"context"
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
	allowed, err := enforcePermission(snapshot.enforcer, snapshot.user, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return apperr.NewPermissionDenied("admin", permission)
	}
	return nil
}

func (u *Usecase) ready() error {
	if u == nil || u.admins == nil || u.roles == nil || u.menus == nil || u.logins == nil {
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
		menuIDSet, err = addActiveRoleGrants(enforcer, user, role)
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
	return rbacSnapshot{enforcer: enforcer, user: user, activeRole: activeRole, roles: roles, permissions: permissions, menuIDSet: menuIDSet}, nil
}

func addActiveRoleGrants(enforcer *casbin.Enforcer, user string, role accessdomain.Role) (map[int64]struct{}, error) {
	roleName := roleSubject(role.Code)
	if _, err := enforcer.AddRoleForUser(user, roleName); err != nil {
		return nil, fmt.Errorf("add casbin role: %w", err)
	}
	for _, permission := range role.Permissions {
		obj, act, splitErr := splitPermission(permission)
		if splitErr != nil {
			return nil, splitErr
		}
		if _, err := enforcer.AddPolicy(roleName, obj, act); err != nil {
			return nil, fmt.Errorf("add casbin policy: %w", err)
		}
	}
	menuIDSet := make(map[int64]struct{}, len(role.MenuIDs))
	for _, menuID := range role.MenuIDs {
		menuIDSet[menuID] = struct{}{}
	}
	return menuIDSet, nil
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
		out = append(out, fromMenu(menu))
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
		DefaultPath: role.DefaultPath,
		Active:      role.Active,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func fromMenu(menu accessdomain.Menu) Menu {
	return Menu{
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
