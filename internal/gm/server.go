package gm

import (
	"time"

	"github.com/kataras/iris/v12"

	"jseer/internal/config"
	"jseer/internal/storage"

	"go.uber.org/zap"
)

// Server holds GM HTTP API.
type Server struct {
	app    *iris.Application
	cfg    config.GMConfig
	store  storage.Store
	logger *zap.Logger
}

func NewServer(cfg config.GMConfig, store storage.Store, logger *zap.Logger) *Server {
	app := iris.New()
	app.UseRouter(iris.Compression)

	s := &Server{app: app, cfg: cfg, store: store, logger: logger}

	app.Get("/healthz", func(ctx iris.Context) { ctx.JSON(iris.Map{"status": "ok"}) })

	api := app.Party("/api")
	api.Post("/auth/login", s.handleLogin)
	api.Get("/auth/profile", s.jwtMiddleware, s.handleProfile)

	secured := api.Party("", s.jwtMiddleware)
	secured.Get("/config/keys", s.requirePermission("config.read"), s.handleConfigKeys)
	secured.Get("/config/{key:string}", s.requirePermission("config.read"), s.handleConfigGet)
	secured.Post("/config/{key:string}", s.requirePermission("config.write"), s.handleConfigSave)
	secured.Get("/config/{key:string}/versions", s.requirePermission("config.read"), s.handleConfigVersions)
	secured.Get("/config/{key:string}/version/{version:int}", s.requirePermission("config.read"), s.handleConfigVersionGet)
	secured.Post("/config/{key:string}/rollback/{version:int}", s.requirePermission("config.rollback"), s.handleConfigRollback)

	secured.Get("/users", s.requirePermission("user.read"), s.handleUserList)
	secured.Post("/users", s.requirePermission("user.write"), s.handleUserCreate)
	secured.Get("/users/{id:int}", s.requirePermission("user.read"), s.handleUserGet)
	secured.Put("/users/{id:int}", s.requirePermission("user.write"), s.handleUserUpdate)
	secured.Put("/users/{id:int}/password", s.requirePermission("user.write"), s.handleUserResetPassword)
	secured.Put("/users/{id:int}/status", s.requirePermission("user.write"), s.handleUserStatus)
	secured.Put("/users/{id:int}/roles", s.requirePermission("user.write"), s.handleUserRoles)

	secured.Get("/roles", s.requirePermission("role.read"), s.handleRoleList)
	secured.Post("/roles", s.requirePermission("role.write"), s.handleRoleCreate)
	secured.Get("/roles/{id:int}", s.requirePermission("role.read"), s.handleRoleGet)
	secured.Put("/roles/{id:int}", s.requirePermission("role.write"), s.handleRoleUpdate)
	secured.Delete("/roles/{id:int}", s.requirePermission("role.write"), s.handleRoleDelete)
	secured.Put("/roles/{id:int}/permissions", s.requirePermission("role.write"), s.handleRolePermissions)

	secured.Get("/permissions", s.requirePermission("permission.read"), s.handlePermissionList)
	secured.Post("/permissions", s.requirePermission("permission.write"), s.handlePermissionCreate)
	secured.Put("/permissions/{id:int}", s.requirePermission("permission.write"), s.handlePermissionUpdate)
	secured.Delete("/permissions/{id:int}", s.requirePermission("permission.write"), s.handlePermissionDelete)

	secured.Get("/audit", s.requirePermission("audit.read"), s.handleAuditList)

	s.bootstrap()
	return s
}

func (s *Server) Run(addr string) error {
	s.app.Configure(iris.WithConfiguration(iris.Configuration{
		TimeFormat: time.RFC3339,
	}))
	return s.app.Listen(addr)
}
