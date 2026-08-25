package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/modules/access/domain"
	"github.com/NSObjects/echo-admin/internal/platform/apperr"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/mysqljson"
)

const (
	retiredPermissionAPICreate = "api:create"
	retiredPermissionAPIUpdate = "api:update"
	retiredPermissionAPIDelete = "api:delete"
)

// UpgradeManagedAPIRouteCatalog explicitly reconciles persisted authorization
// data with the deployment-owned catalog. Applications must invoke this as an
// operator-controlled upgrade; normal process startup never repairs grants.
func (s *Store) UpgradeManagedAPIRouteCatalog(ctx context.Context) error {
	if ctx == nil {
		return errors.New("upgrade managed api route catalog: nil context")
	}
	if s == nil || s.db == nil {
		return errors.New("upgrade managed api route catalog: nil store")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return (&Store{db: tx}).upgradeManagedAPIRouteCatalogData(ctx)
	}); err != nil {
		return fmt.Errorf("upgrade managed api route catalog data: %w", err)
	}

	// MySQL DDL commits independently from the data transaction. Run it last:
	// a failed drop leaves authorization correct and the command safely rerunnable.
	migrator := s.db.WithContext(ctx).Migrator()
	if migrator.HasColumn(&apiModel{}, "public") {
		if err := migrator.DropColumn(&apiModel{}, "public"); err != nil {
			return apperr.WrapDatabase(err, "drop retired api public column")
		}
	}
	return nil
}

func (s *Store) upgradeManagedAPIRouteCatalogData(ctx context.Context) error {
	if err := s.seedPermissions(ctx); err != nil {
		return err
	}
	apiIDs, err := s.seedAPIs(ctx)
	if err != nil {
		return err
	}
	if _, _, err := s.seedMenus(ctx); err != nil {
		return err
	}
	permissionTokens := permissionCatalogTokens()
	if err := s.deleteRetiredAuthorizationRows(ctx, permissionTokens, apiIDs); err != nil {
		return err
	}
	menuIDs, err := s.catalogIDs(ctx, &menuModel{})
	if err != nil {
		return err
	}
	buttonIDs, err := s.catalogIDs(ctx, &menuButtonModel{})
	if err != nil {
		return err
	}
	return s.upgradeRoleAuthorization(ctx, permissionTokens, apiIDs, menuIDs, buttonIDs)
}

func (s *Store) deleteRetiredAuthorizationRows(ctx context.Context, permissionTokens []string, apiIDs []int64) error {
	if err := s.db.WithContext(ctx).Where("id NOT IN ?", apiIDs).Delete(&apiModel{}).Error; err != nil {
		return apperr.WrapDatabase(err, "delete retired api routes")
	}
	if err := s.db.WithContext(ctx).Where("token NOT IN ?", permissionTokens).Delete(&permissionModel{}).Error; err != nil {
		return apperr.WrapDatabase(err, "delete retired permissions")
	}
	return nil
}

func (s *Store) catalogIDs(ctx context.Context, model any) ([]int64, error) {
	var ids []int64
	if err := s.db.WithContext(ctx).Model(model).Order("id").Pluck("id", &ids).Error; err != nil {
		return nil, apperr.WrapDatabase(err, "list authorization catalog ids")
	}
	return ids, nil
}

func (s *Store) upgradeRoleAuthorization(ctx context.Context, permissionTokens []string, apiIDs, menuIDs, buttonIDs []int64) error {
	var roles []roleModel
	if err := s.db.WithContext(ctx).Order("id").Find(&roles).Error; err != nil {
		return apperr.WrapDatabase(err, "list roles for authorization upgrade")
	}
	for _, role := range roles {
		root := role.Code == domain.RoleCodeSuperAdmin
		permissions := upgradeRolePermissions([]string(role.Permissions), root)
		if len(permissions) == 0 {
			return fmt.Errorf("upgrade role authorization: role %q has no current permissions", role.Code)
		}
		roleAPIIDs := retainCatalogIDs([]int64(role.APIIDs), apiIDs)
		roleMenuIDs := retainCatalogIDs([]int64(role.MenuIDs), menuIDs)
		roleButtonIDs := retainCatalogIDs([]int64(role.ButtonIDs), buttonIDs)
		if root {
			permissions = append([]string(nil), permissionTokens...)
			roleAPIIDs = append([]int64(nil), apiIDs...)
			roleMenuIDs = append([]int64(nil), menuIDs...)
			roleButtonIDs = append([]int64(nil), buttonIDs...)
		}
		if err := s.updateRoleAuthorization(ctx, role, permissions, roleAPIIDs, roleMenuIDs, roleButtonIDs); err != nil {
			return err
		}
	}
	return nil
}

func permissionCatalogTokens() []string {
	definitions := domain.PermissionCatalog()
	tokens := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		tokens = append(tokens, definition.Token)
	}
	return tokens
}

func upgradeRolePermissions(current []string, root bool) []string {
	if root {
		return permissionCatalogTokens()
	}
	allowed := make(map[string]struct{})
	for _, token := range permissionCatalogTokens() {
		allowed[token] = struct{}{}
	}
	out := make([]string, 0, len(current)+2)
	seen := make(map[string]struct{}, len(current)+2)
	hadRetiredAPIManagement := false
	hadRetiredAPIUpdate := false
	for _, token := range current {
		switch token {
		case retiredPermissionAPICreate, retiredPermissionAPIDelete:
			hadRetiredAPIManagement = true
			continue
		case retiredPermissionAPIUpdate:
			hadRetiredAPIManagement = true
			hadRetiredAPIUpdate = true
			continue
		}
		if _, ok := allowed[token]; !ok {
			continue
		}
		out = appendUniqueToken(out, seen, token)
	}
	// The retired API management UI always operated on the visible catalog.
	// Preserve that least-privilege read capability; only former update holders
	// inherit the replacement role-grant capability.
	if hadRetiredAPIManagement {
		out = appendUniqueToken(out, seen, domain.PermissionAPIRead)
	}
	if hadRetiredAPIUpdate {
		out = appendUniqueToken(out, seen, domain.PermissionAPIGrant)
	}
	return out
}

func appendUniqueToken(tokens []string, seen map[string]struct{}, token string) []string {
	if _, ok := seen[token]; ok {
		return tokens
	}
	seen[token] = struct{}{}
	return append(tokens, token)
}

func retainCatalogIDs(current, catalog []int64) []int64 {
	allowed := make(map[int64]struct{}, len(catalog))
	for _, id := range catalog {
		allowed[id] = struct{}{}
	}
	out := make([]int64, 0, len(current))
	seen := make(map[int64]struct{}, len(current))
	for _, id := range current {
		if _, ok := allowed[id]; !ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *Store) updateRoleAuthorization(ctx context.Context, role roleModel, permissions []string, apiIDs, menuIDs, buttonIDs []int64) error {
	updates := map[string]any{
		"permissions": mysqljson.Strings(permissions),
		"api_ids":     mysqljson.Int64s(apiIDs),
		"menu_ids":    mysqljson.Int64s(menuIDs),
		"button_ids":  mysqljson.Int64s(buttonIDs),
		"updated_at":  time.Now().UTC(),
	}
	if err := s.db.WithContext(ctx).Model(&roleModel{}).Where("id = ?", role.ID).Updates(updates).Error; err != nil {
		return apperr.WrapDatabase(err, "upgrade role authorization")
	}
	return nil
}
