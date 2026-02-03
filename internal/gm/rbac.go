package gm

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type gmClaims struct {
	UID   int64    `json:"uid"`
	Roles []string `json:"roles"`
	Perms []string `json:"perms"`
	jwt.RegisteredClaims
}

var defaultPermissions = []struct {
	Code string
	Name string
	Desc string
}{
	{"config.read", "配置读取", "查看配置数据"},
	{"config.write", "配置发布", "新增与发布配置"},
	{"config.rollback", "配置回滚", "回滚配置版本"},
	{"audit.read", "审计查看", "查看操作日志"},
	{"user.read", "用户查看", "查看 GM 用户"},
	{"user.write", "用户管理", "新增/修改 GM 用户"},
	{"role.read", "角色查看", "查看角色"},
	{"role.write", "角色管理", "新增/修改角色"},
	{"permission.read", "权限查看", "查看权限"},
	{"permission.write", "权限管理", "新增/修改权限"},
}

func (s *Server) requirePermission(code string) iris.Handler {
	return func(ctx iris.Context) {
		if s.hasPermission(ctx, code) {
			ctx.Next()
			return
		}
		s.fail(ctx, iris.StatusForbidden, "forbidden")
	}
}

func (s *Server) hasPermission(ctx iris.Context, code string) bool {
	claims, ok := s.claimsFromCtx(ctx)
	if !ok {
		return false
	}
	if contains(claims.Roles, "super_admin") {
		return true
	}
	if contains(claims.Perms, "*") {
		return true
	}
	return contains(claims.Perms, code)
}

func (s *Server) claimsFromCtx(ctx iris.Context) (*gmClaims, bool) {
	raw := ctx.Values().Get("gm_claims")
	if raw == nil {
		return nil, false
	}
	claims, ok := raw.(*gmClaims)
	return claims, ok
}

func (s *Server) bootstrap() {
	ctx := context.Background()

	permIDs := make([]int64, 0, len(defaultPermissions))
	for _, perm := range defaultPermissions {
		existing, err := s.store.GetGMPermissionByCode(ctx, perm.Code)
		if err != nil {
			created, createErr := s.store.CreateGMPermission(ctx, &GMPermission{
				Code:        perm.Code,
				Name:        perm.Name,
				Description: perm.Desc,
			})
			if createErr != nil {
				s.logger.Warn("gm bootstrap permission failed", zap.Error(createErr))
				continue
			}
			permIDs = append(permIDs, created.ID)
		} else {
			permIDs = append(permIDs, existing.ID)
		}
	}

	superRole, err := s.store.GetGMRoleByName(ctx, "super_admin")
	if err != nil {
		role, createErr := s.store.CreateGMRole(ctx, &GMRole{
			Name:        "super_admin",
			Description: "默认超级管理员",
		}, permIDs)
		if createErr != nil {
			s.logger.Warn("gm bootstrap role failed", zap.Error(createErr))
		} else {
			superRole = role
		}
	} else if len(permIDs) > 0 {
		_ = s.store.SetRolePermissions(ctx, superRole.ID, permIDs)
	}

	if s.cfg.DefaultAdminUser == "" || s.cfg.DefaultAdminPass == "" {
		return
	}
	admin, err := s.store.GetGMUserByUsername(ctx, s.cfg.DefaultAdminUser)
	if err == nil {
		if superRole != nil {
			_ = s.store.SetGMUserRoles(ctx, admin.ID, []int64{superRole.ID})
		}
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.DefaultAdminPass), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Warn("gm bootstrap password hash failed", zap.Error(err))
		return
	}
	user, createErr := s.store.CreateGMUser(ctx, &GMUser{
		Username:     s.cfg.DefaultAdminUser,
		PasswordHash: string(hash),
		Status:       "active",
		CreatedAt:    time.Now().Unix(),
	}, nil)
	if createErr != nil {
		s.logger.Warn("gm bootstrap user failed", zap.Error(createErr))
		return
	}
	if superRole != nil {
		_ = s.store.SetGMUserRoles(ctx, user.ID, []int64{superRole.ID})
	}

	seedDefaultConfigs(ctx, s.store, s.logger)
}

func (s *Server) buildToken(user *GMUser) (string, int64, error) {
	roles, perms := s.collectRolesAndPerms(user.ID)
	expires := time.Now().Add(time.Duration(s.cfg.TokenTTLMinutes) * time.Minute)
	claims := gmClaims{
		UID:   user.ID,
		Roles: roles,
		Perms: perms,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}
	return signed, expires.Unix(), nil
}

func (s *Server) collectRolesAndPerms(userID int64) ([]string, []string) {
	ctx := context.Background()
	roles, err := s.store.ListGMRolesByUser(ctx, userID)
	if err != nil {
		return nil, nil
	}
	roleNames := make([]string, 0, len(roles))
	permSet := make(map[string]struct{})
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
		perms, permErr := s.store.ListPermissionsByRole(ctx, role.ID)
		if permErr != nil {
			continue
		}
		for _, p := range perms {
			permSet[p.Code] = struct{}{}
		}
	}
	if contains(roleNames, "super_admin") {
		permSet["*"] = struct{}{}
	}
	permList := make([]string, 0, len(permSet))
	for code := range permSet {
		permList = append(permList, code)
	}
	return roleNames, permList
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}
