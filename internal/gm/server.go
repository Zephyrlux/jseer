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

	secured := api.Party("", s.jwtMiddleware)
	secured.Get("/config/keys", s.handleConfigKeys)
	secured.Get("/config/{key:string}", s.handleConfigGet)
	secured.Post("/config/{key:string}", s.handleConfigSave)
	secured.Get("/config/{key:string}/versions", s.handleConfigVersions)
	secured.Get("/audit", s.handleAuditList)

	return s
}

func (s *Server) Run(addr string) error {
	s.app.Configure(iris.WithConfiguration(iris.Configuration{
		TimeFormat: time.RFC3339,
	}))
	return s.app.Listen(addr)
}
