package gm

import (
	"strconv"

	"github.com/kataras/iris/v12"
)

type roleRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type rolePermissionRequest struct {
	PermissionIDs []int64 `json:"permission_ids"`
}

func (s *Server) handleRoleList(ctx iris.Context) {
	filter := parseGMRoleFilter(ctx)
	roles, err := s.store.ListGMRoles(ctx.Request().Context(), filter)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	items := make([]iris.Map, 0, len(roles))
	for _, role := range roles {
		perms, _ := s.store.ListPermissionsByRole(ctx.Request().Context(), role.ID)
		items = append(items, iris.Map{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
			"created_at":  role.CreatedAt,
			"permissions": perms,
		})
	}
	s.ok(ctx, iris.Map{"items": items})
}

func (s *Server) handleRoleGet(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	role, err := s.store.GetGMRoleByID(ctx.Request().Context(), id)
	if err != nil {
		s.fail(ctx, iris.StatusNotFound, "not found")
		return
	}
	perms, _ := s.store.ListPermissionsByRole(ctx.Request().Context(), role.ID)
	s.ok(ctx, iris.Map{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description,
		"created_at":  role.CreatedAt,
		"permissions": perms,
	})
}

func (s *Server) handleRoleCreate(ctx iris.Context) {
	var req roleRequest
	if err := ctx.ReadJSON(&req); err != nil || req.Name == "" {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	role, err := s.store.CreateGMRole(ctx.Request().Context(), &GMRole{
		Name:        req.Name,
		Description: req.Description,
	}, req.PermissionIDs)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.role.create", "gm_role", strconv.FormatInt(role.ID, 10), "role created")
	s.ok(ctx, iris.Map{"id": role.ID})
}

func (s *Server) handleRoleUpdate(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	var req roleRequest
	if err := ctx.ReadJSON(&req); err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	role, err := s.store.UpdateGMRole(ctx.Request().Context(), &GMRole{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}, req.PermissionIDs)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.role.update", "gm_role", strconv.FormatInt(role.ID, 10), "role updated")
	s.ok(ctx, iris.Map{"id": role.ID})
}

func (s *Server) handleRoleDelete(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	role, err := s.store.GetGMRoleByID(ctx.Request().Context(), id)
	if err == nil && role.Name == "super_admin" {
		s.fail(ctx, iris.StatusBadRequest, "cannot delete super_admin")
		return
	}
	if err := s.store.DeleteGMRole(ctx.Request().Context(), id); err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.role.delete", "gm_role", strconv.FormatInt(id, 10), "role deleted")
	s.ok(ctx, iris.Map{"status": "ok"})
}

func (s *Server) handleRolePermissions(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	var req rolePermissionRequest
	if err := ctx.ReadJSON(&req); err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	if err := s.store.SetRolePermissions(ctx.Request().Context(), id, req.PermissionIDs); err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.role.permissions", "gm_role", strconv.FormatInt(id, 10), "permissions updated")
	s.ok(ctx, iris.Map{"status": "ok"})
}

func parseGMRoleFilter(ctx iris.Context) GMRoleFilter {
	limit, _ := ctx.URLParamInt("limit")
	offset, _ := ctx.URLParamInt("offset")
	return GMRoleFilter{
		Limit:  limit,
		Offset: offset,
		Search: ctx.URLParam("search"),
	}
}
