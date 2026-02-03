package gm

import (
	"strconv"

	"github.com/kataras/iris/v12"
)

type permissionRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) handlePermissionList(ctx iris.Context) {
	filter := parseGMPermissionFilter(ctx)
	perms, err := s.store.ListGMPermissions(ctx.Request().Context(), filter)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.ok(ctx, iris.Map{"items": perms})
}

func (s *Server) handlePermissionCreate(ctx iris.Context) {
	var req permissionRequest
	if err := ctx.ReadJSON(&req); err != nil || req.Code == "" {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	perm, err := s.store.CreateGMPermission(ctx.Request().Context(), &GMPermission{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.permission.create", "gm_permission", strconv.FormatInt(perm.ID, 10), "permission created")
	s.ok(ctx, iris.Map{"id": perm.ID})
}

func (s *Server) handlePermissionUpdate(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	var req permissionRequest
	if err := ctx.ReadJSON(&req); err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	perm, err := s.store.UpdateGMPermission(ctx.Request().Context(), &GMPermission{
		ID:          id,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.permission.update", "gm_permission", strconv.FormatInt(perm.ID, 10), "permission updated")
	s.ok(ctx, iris.Map{"id": perm.ID})
}

func (s *Server) handlePermissionDelete(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	if err := s.store.DeleteGMPermission(ctx.Request().Context(), id); err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.permission.delete", "gm_permission", strconv.FormatInt(id, 10), "permission deleted")
	s.ok(ctx, iris.Map{"status": "ok"})
}

func parseGMPermissionFilter(ctx iris.Context) GMPermissionFilter {
	limit, _ := ctx.URLParamInt("limit")
	offset, _ := ctx.URLParamInt("offset")
	return GMPermissionFilter{
		Limit:  limit,
		Offset: offset,
		Search: ctx.URLParam("search"),
	}
}
