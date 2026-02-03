package gm

import "github.com/kataras/iris/v12"

func (s *Server) handleAuditList(ctx iris.Context) {
	limit, _ := ctx.URLParamInt("limit")
	if limit <= 0 {
		limit = 100
	}
	items, err := s.store.ListAuditLogs(ctx.Request().Context(), limit)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.ok(ctx, iris.Map{"items": items})
}
