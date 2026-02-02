package gm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/kataras/iris/v12"

	"jseer/internal/storage"
)

type configPayload struct {
	Value json.RawMessage `json:"value"`
}

func (s *Server) handleConfigKeys(ctx iris.Context) {
	keys, err := s.store.ListConfigKeys(ctx.Request().Context())
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, iris.Map{"error": err.Error()})
		return
	}
	ctx.JSON(iris.Map{"keys": keys})
}

func (s *Server) handleConfigGet(ctx iris.Context) {
	key := ctx.Params().Get("key")
	entry, err := s.store.GetConfig(ctx.Request().Context(), key)
	if err != nil {
		ctx.StopWithJSON(iris.StatusNotFound, iris.Map{"error": "not found"})
		return
	}
	ctx.JSON(iris.Map{
		"key":      entry.Key,
		"value":    json.RawMessage(entry.Value),
		"version":  entry.Version,
		"checksum": entry.Checksum,
	})
}

func (s *Server) handleConfigSave(ctx iris.Context) {
	key := ctx.Params().Get("key")
	var payload configPayload
	if err := ctx.ReadJSON(&payload); err != nil {
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "invalid payload"})
		return
	}
	checksum := sha256.Sum256(payload.Value)
	entry := &storage.ConfigEntry{
		Key:      key,
		Value:    payload.Value,
		Checksum: hex.EncodeToString(checksum[:]),
	}
	operator := s.operatorFromCtx(ctx)
	version, err := s.store.SaveConfig(ctx.Request().Context(), entry, operator)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, iris.Map{"error": err.Error()})
		return
	}
	ctx.JSON(iris.Map{
		"key":       key,
		"version":   version.Version,
		"operator":  version.Operator,
		"created_at": time.Unix(version.CreatedAt, 0).UTC(),
	})
}

func (s *Server) handleConfigVersions(ctx iris.Context) {
	key := ctx.Params().Get("key")
	versions, err := s.store.ListConfigVersions(ctx.Request().Context(), key, 50)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, iris.Map{"error": err.Error()})
		return
	}
	ctx.JSON(iris.Map{"versions": versions})
}

func (s *Server) operatorFromCtx(ctx iris.Context) string {
	claims := ctx.Values().Get("gm_claims")
	if claims == nil {
		return "unknown"
	}
	if m, ok := claims.(map[string]any); ok {
		if sub, ok := m["sub"].(string); ok {
			return sub
		}
	}
	return "unknown"
}
