package main

import (
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kataras/iris/v12"

	"jseer/internal/config"
	"jseer/internal/logging"

	"go.uber.org/zap"
)

func main() {
	cfgPath := config.ResolvePath("configs/config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		panic(err)
	}
	logger, err := logging.New(cfg.Log.Level)
	if err != nil {
		panic(err)
	}

	var cfgVal atomic.Value
	cfgVal.Store(cfg)
	if cfg.App.ReloadIntervalS > 0 {
		config.Watch(cfgPath, time.Duration(cfg.App.ReloadIntervalS)*time.Second, func(newCfg *config.Config) {
			cfgVal.Store(newCfg)
		})
	}

	app := iris.New()
	app.Get("/healthz", func(ctx iris.Context) { ctx.JSON(iris.Map{"status": "ok"}) })
	app.Get("/{path:path}", func(ctx iris.Context) {
		current := cfgVal.Load().(*config.Config)
		serveResource(ctx, current.HTTP.StaticRoot, current.HTTP.ProxyRoot, current.HTTP.Upstream)
	})

	if cfg.HTTP.LoginIPAddress != "" {
		startLoginIPServer(cfg.HTTP.LoginIPAddress, &cfgVal, logger)
	}

	if err := app.Listen(cfg.HTTP.Address); err != nil {
		logger.Error("resource server stopped", zap.Error(err))
		os.Exit(1)
	}
}

func startLoginIPServer(addr string, cfgVal *atomic.Value, logger *zap.Logger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ip.txt", func(w http.ResponseWriter, r *http.Request) {
		current := cfgVal.Load().(*config.Config)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(current.HTTP.IPTxt))
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"status\":\"ok\"}"))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("login ip server stopped", zap.Error(err))
			os.Exit(1)
		}
	}()
}

func serveResource(ctx iris.Context, staticRoot, proxyRoot, upstream string) {
	reqPath := ctx.Path()
	rel := strings.TrimPrefix(reqPath, "/")
	rel = path.Clean(rel)
	if rel == "." || rel == "/" {
		rel = "index.html"
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
	if upstream != "" {
		if fetchAndServe(ctx, upstream, proxyRoot, rel) {
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

func fetchAndServe(ctx iris.Context, upstream, proxyRoot, rel string) bool {
	base := strings.TrimRight(upstream, "/")
	url := base + "/" + rel
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	for k, v := range resp.Header {
		if len(v) > 0 {
			ctx.ResponseWriter().Header()[k] = v
		}
	}
	ctx.StatusCode(resp.StatusCode)
	_, _ = ctx.Write(body)

	if proxyRoot != "" {
		if filePath, ok := safeJoin(proxyRoot, rel); ok {
			_ = os.MkdirAll(filepath.Dir(filePath), 0o755)
			_ = os.WriteFile(filePath, body, 0o644)
		}
	}
	return true
}
