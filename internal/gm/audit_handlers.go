package gm

import "github.com/kataras/iris/v12"

func (s *Server) handleAuditList(ctx iris.Context) {
	ctx.JSON(iris.Map{"items": []any{}})
}
