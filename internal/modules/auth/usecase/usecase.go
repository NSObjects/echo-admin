package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"golang.org/x/crypto/bcrypt"

	accessdomain "github.com/NSObjects/echo-admin/internal/modules/access/domain"
	authdomain "github.com/NSObjects/echo-admin/internal/modules/auth/domain"
	identitydomain "github.com/NSObjects/echo-admin/internal/modules/identity/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/logging"
	"github.com/NSObjects/echo-admin/internal/platform/requestctx"
)

const (
	loginSessionTokenBytes      = 32
	loginSessionIdleTTL         = 2 * time.Hour
	loginSessionAbsoluteTTL     = 12 * time.Hour
	loginSessionRefreshInterval = 5 * time.Minute

	loginSessionRevokedLogout         = "logout"
	loginSessionRevokedLogoutOthers   = "logout_others"
	loginSessionRevokedPasswordChange = "password_changed"
	loginSessionRevokedSecurityEvent  = "security_event"
)

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

// Login authenticates an administrator and creates a browser login session.
func (u *Usecase) Login(ctx context.Context, input LoginInput) (LoginOutput, error) {
	if err := u.ready(); err != nil {
		return LoginOutput{}, err
	}
	username := strings.ToLower(strings.TrimSpace(input.Username))
	now := u.now()
	attemptKey := loginAttemptKey(username, input.IP)
	if err := u.ensureLoginAttemptAllowed(ctx, attemptKey, now, username, input); err != nil {
		return LoginOutput{}, err
	}
	admin, err := u.admins.FindByUsername(ctx, username)
	if err != nil {
		record := loginRecordFromInput(0, username, input, false, "invalid credentials")
		return u.rejectLogin(ctx, attemptKey, now, record, apperr.New(apperr.ErrUnauthorized, "invalid username or password"))
	}
	if !admin.Active {
		record := loginRecordFromInput(admin.ID, username, input, false, "account disabled")
		return u.rejectLogin(ctx, attemptKey, now, record, apperr.New(apperr.ErrAccountDisabled, "account disabled"))
	}
	if compareErr := bcrypt.CompareHashAndPassword(admin.PasswordHash, []byte(input.Password)); compareErr != nil {
		record := loginRecordFromInput(admin.ID, username, input, false, "invalid credentials")
		return u.rejectLogin(ctx, attemptKey, now, record, apperr.New(apperr.ErrUnauthorized, "invalid username or password"))
	}

	if resetErr := u.loginLimiter.ResetLoginAttempts(ctx, attemptKey); resetErr != nil {
		return LoginOutput{}, resetErr
	}
	user, err := u.userSnapshot(ctx, admin, admin.ActiveRoleID)
	if err != nil {
		return LoginOutput{}, err
	}
	rawToken, tokenHash, err := newLoginSessionToken()
	if err != nil {
		return LoginOutput{}, err
	}
	session, err := authdomain.NewLoginSession(authdomain.LoginSessionInput{
		AdminID:           admin.ID,
		ActiveRoleID:      user.ActiveRoleID,
		TokenHash:         tokenHash,
		IP:                input.IP,
		UserAgent:         input.UserAgent,
		CreatedAt:         now,
		IdleExpiresAt:     now.Add(loginSessionIdleTTL),
		AbsoluteExpiresAt: now.Add(loginSessionAbsoluteTTL),
	})
	if err != nil {
		return LoginOutput{}, err
	}
	created, err := u.sessions.CreateLoginSession(ctx, session)
	if err != nil {
		return LoginOutput{}, err
	}
	if err := u.recordLogin(ctx, loginRecordFromInput(admin.ID, username, input, true, "login succeeded")); err != nil {
		return LoginOutput{}, err
	}
	return LoginOutput{SessionToken: rawToken, SessionExpiresAt: created.AbsoluteExpiresAt, User: user}, nil
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

// SwitchRole changes the active role for the current login session.
func (u *Usecase) SwitchRole(ctx context.Context, input RoleSwitchInput) (RoleSwitchOutput, error) {
	if err := u.ready(); err != nil {
		return RoleSwitchOutput{}, err
	}
	sessionID, err := currentLoginSessionID(ctx)
	if err != nil {
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
	if err := u.sessions.UpdateLoginSessionRole(ctx, sessionID, switched.ActiveRoleID, u.now()); err != nil {
		return RoleSwitchOutput{}, err
	}
	user, err := u.userSnapshot(ctx, admin, input.RoleID)
	if err != nil {
		return RoleSwitchOutput{}, err
	}
	return RoleSwitchOutput{User: user}, nil
}

// ChangePassword verifies the current password, stores the new hash, and
// revokes other login sessions for the same administrator.
func (u *Usecase) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	if err := u.ready(); err != nil {
		return err
	}
	sessionID, err := currentLoginSessionID(ctx)
	if err != nil {
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
	return u.sessions.RevokeOtherLoginSessions(ctx, adminID, sessionID, loginSessionRevokedPasswordChange, u.now())
}

// Logout revokes the current login session.
func (u *Usecase) Logout(ctx context.Context) error {
	if err := u.ready(); err != nil {
		return err
	}
	sessionID, err := currentLoginSessionID(ctx)
	if err != nil {
		return err
	}
	return u.sessions.RevokeLoginSession(ctx, sessionID, loginSessionRevokedLogout, u.now())
}

// LogoutOthers revokes every other login session for the current administrator.
func (u *Usecase) LogoutOthers(ctx context.Context) error {
	if err := u.ready(); err != nil {
		return err
	}
	sessionID, err := currentLoginSessionID(ctx)
	if err != nil {
		return err
	}
	adminID, err := currentAdminID(ctx)
	if err != nil {
		return err
	}
	return u.sessions.RevokeOtherLoginSessions(ctx, adminID, sessionID, loginSessionRevokedLogoutOthers, u.now())
}

// AuthenticateLoginSession validates a raw browser session token and returns
// the identity that middleware should attach to request context.
func (u *Usecase) AuthenticateLoginSession(ctx context.Context, rawToken string) (LoginSessionIdentity, error) {
	if err := u.ready(); err != nil {
		return LoginSessionIdentity{}, err
	}
	tokenHash := loginSessionTokenHash(rawToken)
	if tokenHash == "" {
		return LoginSessionIdentity{}, apperr.NewUnauthorized()
	}
	session, found, err := u.sessions.FindLoginSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return LoginSessionIdentity{}, err
	}
	if !found {
		return LoginSessionIdentity{}, apperr.NewUnauthorized()
	}
	now := u.now()
	if err := session.AvailabilityError(now); err != nil {
		return LoginSessionIdentity{}, mapSessionAvailabilityError(err)
	}
	admin, err := u.admins.FindByID(ctx, session.AdminID)
	if err != nil {
		return LoginSessionIdentity{}, apperr.NewUnauthorized()
	}
	if !admin.Active || !admin.HasRole(session.ActiveRoleID) {
		return LoginSessionIdentity{}, apperr.NewUnauthorized()
	}
	if session.NeedsRefresh(now, loginSessionRefreshInterval) {
		refreshed := session.Refreshed(now, loginSessionIdleTTL)
		if err := u.sessions.RefreshLoginSession(ctx, refreshed); err != nil {
			logging.FromContext(ctx).Warn().Err(err).Int64("session_id", session.ID).Msg("refresh login session")
		}
	}
	return LoginSessionIdentity{
		SessionID: session.ID,
		AdminID:   session.AdminID,
		RoleID:    session.ActiveRoleID,
	}, nil
}

// RevokeLoginSessions revokes all login sessions for one administrator after a
// security event such as disabling or deleting the account.
func (u *Usecase) RevokeLoginSessions(ctx context.Context, adminID int64) error {
	if err := u.ready(); err != nil {
		return err
	}
	if adminID <= 0 {
		return apperr.NewBadRequest("invalid admin id")
	}
	return u.sessions.RevokeLoginSessions(ctx, adminID, loginSessionRevokedSecurityEvent, u.now())
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

// AuthorizeRoute verifies that the current active role may call the managed
// API route. The path must be Echo's registered route pattern rather than the
// raw URL, so IDs and other path parameters are authorized by one catalog row.
func (u *Usecase) AuthorizeRoute(ctx context.Context, method, path string) error {
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
	return requireAssignedRouteAPI(snapshot.activeRole, api)
}

func (u *Usecase) ready() error {
	if u == nil || u.admins == nil || u.roles == nil || u.menus == nil || u.apis == nil || u.logins == nil || u.sessions == nil || u.loginLimiter == nil {
		return apperr.New(apperr.ErrInternalServer, "auth dependencies are not configured")
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
	api, err := u.apis.FindAPIByRoute(ctx, method, path)
	if err != nil {
		if appErr, ok := apperr.Parse(err); ok && appErr.Code() == apperr.ErrNotFound {
			return accessdomain.API{}, apperr.NewPermissionDenied("api", path)
		}
		return accessdomain.API{}, err
	}
	return api, nil
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

func (u *Usecase) recordLogin(ctx context.Context, record LoginRecord) error {
	return u.logins.RecordLogin(ctx, record)
}

func (u *Usecase) ensureLoginAttemptAllowed(ctx context.Context, attemptKey string, now time.Time, username string, input LoginInput) error {
	err := u.loginLimiter.CheckLoginAttempt(ctx, attemptKey, now)
	if err == nil {
		return nil
	}
	if !isTooManyLoginAttempts(err) {
		return err
	}
	record := loginRecordFromInput(0, username, input, false, "too many login attempts")
	if recordErr := u.recordLogin(ctx, record); recordErr != nil {
		return recordErr
	}
	return err
}

func (u *Usecase) rejectLogin(ctx context.Context, attemptKey string, now time.Time, record LoginRecord, loginErr error) (LoginOutput, error) {
	if recordErr := u.recordLogin(ctx, record); recordErr != nil {
		return LoginOutput{}, recordErr
	}
	if limitErr := u.loginLimiter.RecordLoginFailure(ctx, attemptKey, now); limitErr != nil {
		return LoginOutput{}, limitErr
	}
	return LoginOutput{}, loginErr
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

func newLoginSessionToken() (string, string, error) {
	token := make([]byte, loginSessionTokenBytes)
	if _, err := rand.Read(token); err != nil {
		return "", "", fmt.Errorf("generate login session token: %w", err)
	}
	rawToken := base64.RawURLEncoding.EncodeToString(token)
	return rawToken, loginSessionTokenHash(rawToken), nil
}

func loginSessionTokenHash(rawToken string) string {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func mapSessionAvailabilityError(err error) error {
	if errors.Is(err, authdomain.ErrLoginSessionExpired) || errors.Is(err, authdomain.ErrLoginSessionRevoked) {
		return apperr.NewUnauthorized()
	}
	return err
}

func isTooManyLoginAttempts(err error) bool {
	appErr, ok := apperr.Parse(err)
	return ok && appErr.Code() == apperr.ErrTooManyAttempts
}

func loginAttemptKey(username, ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "unknown"
	}
	sum := sha256.Sum256([]byte(username + "\x00" + ip))
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

func currentLoginSessionID(ctx context.Context) (int64, error) {
	raw := requestctx.GetLoginSessionID(ctx)
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
