package gm

import "github.com/kataras/iris/v12"

func (s *Server) handleAuditList(ctx iris.Context) {
	items, err := s.store.ListAuditLogs(ctx.Request().Context(), 100)
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, iris.Map{"error": err.Error()})
		return
	}
	ctx.JSON(iris.Map{"items": items})
}
