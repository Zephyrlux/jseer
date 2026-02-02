package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kataras/iris/v12"

	"jseer/internal/config"
	"jseer/internal/logging"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load(config.ResolvePath("configs/config.yaml"))
	if err != nil {
		panic(err)
	}
	logger, err := logging.New(cfg.Log.Level)
	if err != nil {
		panic(err)
	}

	app := iris.New()
	app.Get("/ip.txt", func(ctx iris.Context) {
		_, _ = ctx.WriteString(cfg.HTTP.IPTxt)
	})
	app.Get("/healthz", func(ctx iris.Context) { ctx.JSON(iris.Map{"status": "ok"}) })
	app.Get("/{path:path}", func(ctx iris.Context) {
		serveResource(ctx, cfg.HTTP.StaticRoot, cfg.HTTP.ProxyRoot)
	})

	if err := app.Listen(cfg.HTTP.Address); err != nil {
		logger.Error("resource server stopped", zap.Error(err))
		os.Exit(1)
	}
}

func serveResource(ctx iris.Context, staticRoot, proxyRoot string) {
	reqPath := ctx.Path()
	rel := strings.TrimPrefix(reqPath, "/")
	rel = path.Clean(rel)
	if rel == "." || rel == "/" {
		ctx.StatusCode(404)
		return
	}
	if strings.Contains(rel, "..") {
		ctx.StatusCode(400)
		return
	}

	if proxyRoot != "" {
		if filePath, ok := safeJoin(proxyRoot, rel); ok && isFile(filePath) {
			http.ServeFile(ctx.ResponseWriter(), ctx.Request(), filePath)
			return
		}
	}
	if staticRoot != "" {
		if filePath, ok := safeJoin(staticRoot, rel); ok && isFile(filePath) {
			http.ServeFile(ctx.ResponseWriter(), ctx.Request(), filePath)
			return
		}
	}
	ctx.StatusCode(404)
}

func safeJoin(root, rel string) (string, bool) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	absPath, err := filepath.Abs(filepath.Join(absRoot, rel))
	if err != nil {
		return "", false
	}
	if absPath == absRoot {
		return "", false
	}
	if !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return "", false
	}
	return absPath, true
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
