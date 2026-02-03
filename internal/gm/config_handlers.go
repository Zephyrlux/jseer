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
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.ok(ctx, iris.Map{"keys": keys})
}

func (s *Server) handleConfigGet(ctx iris.Context) {
	key := ctx.Params().Get("key")
	entry, err := s.store.GetConfig(ctx.Request().Context(), key)
	if err != nil {
		s.fail(ctx, iris.StatusNotFound, "not found")
		return
	}
	s.ok(ctx, iris.Map{
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
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
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
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.ok(ctx, iris.Map{
		"key":        key,
		"version":    version.Version,
		"operator":   version.Operator,
		"created_at": time.Unix(version.CreatedAt, 0).UTC(),
	})
}

func (s *Server) handleConfigVersions(ctx iris.Context) {
	key := ctx.Params().Get("key")
	versions, err := s.store.ListConfigVersions(ctx.Request().Context(), key, 50)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.ok(ctx, iris.Map{"versions": versions})
}

func (s *Server) handleConfigVersionGet(ctx iris.Context) {
	key := ctx.Params().Get("key")
	version, err := ctx.Params().GetInt64("version")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid version")
		return
	}
	entry, err := s.store.GetConfigVersion(ctx.Request().Context(), key, version)
	if err != nil {
		s.fail(ctx, iris.StatusNotFound, "not found")
		return
	}
	s.ok(ctx, iris.Map{
		"key":        entry.Key,
		"value":      json.RawMessage(entry.Value),
		"version":    entry.Version,
		"checksum":   entry.Checksum,
		"operator":   entry.Operator,
		"created_at": entry.CreatedAt,
	})
}

func (s *Server) handleConfigRollback(ctx iris.Context) {
	key := ctx.Params().Get("key")
	version, err := ctx.Params().GetInt64("version")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid version")
		return
	}
	operator := s.operatorFromCtx(ctx)
	cv, err := s.store.RollbackConfig(ctx.Request().Context(), key, version, operator)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.ok(ctx, iris.Map{
		"key":        key,
		"version":    cv.Version,
		"operator":   cv.Operator,
		"created_at": time.Unix(cv.CreatedAt, 0).UTC(),
	})
}

func (s *Server) operatorFromCtx(ctx iris.Context) string {
	claims, ok := s.claimsFromCtx(ctx)
	if !ok {
		return "unknown"
	}
	if claims.Subject != "" {
		return claims.Subject
	}
	return "unknown"
}
