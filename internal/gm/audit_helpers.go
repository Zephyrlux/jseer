package gm

import (
	"github.com/kataras/iris/v12"

	"jseer/internal/storage"
)

func (s *Server) audit(ctx iris.Context, action, resource, resourceID, detail string) {
	operator := s.operatorFromCtx(ctx)
	_, _ = s.store.CreateAuditLog(ctx.Request().Context(), &storage.AuditLog{
		Operator:   operator,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
	})
}
