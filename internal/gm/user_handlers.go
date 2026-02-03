package gm

import (
	"strconv"

	"github.com/kataras/iris/v12"
	"golang.org/x/crypto/bcrypt"
)

type userCreateRequest struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Status   string  `json:"status"`
	RoleIDs  []int64 `json:"role_ids"`
}

type userUpdateRequest struct {
	Username string  `json:"username"`
	Status   string  `json:"status"`
	RoleIDs  []int64 `json:"role_ids"`
}

type userPasswordRequest struct {
	Password string `json:"password"`
}

type userStatusRequest struct {
	Status string `json:"status"`
}

type userRoleRequest struct {
	RoleIDs []int64 `json:"role_ids"`
}

func (s *Server) handleUserList(ctx iris.Context) {
	filter := parseGMUserFilter(ctx)
	users, err := s.store.ListGMUsers(ctx.Request().Context(), filter)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	items := make([]iris.Map, 0, len(users))
	for _, user := range users {
		roles, _ := s.store.ListGMRolesByUser(ctx.Request().Context(), user.ID)
		items = append(items, iris.Map{
			"id":            user.ID,
			"username":      user.Username,
			"status":        user.Status,
			"last_login_at": user.LastLoginAt,
			"created_at":    user.CreatedAt,
			"roles":         roles,
		})
	}
	s.ok(ctx, iris.Map{"items": items})
}

func (s *Server) handleUserGet(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	user, err := s.store.GetGMUserByID(ctx.Request().Context(), id)
	if err != nil {
		s.fail(ctx, iris.StatusNotFound, "not found")
		return
	}
	roles, _ := s.store.ListGMRolesByUser(ctx.Request().Context(), user.ID)
	s.ok(ctx, iris.Map{
		"id":            user.ID,
		"username":      user.Username,
		"status":        user.Status,
		"last_login_at": user.LastLoginAt,
		"created_at":    user.CreatedAt,
		"roles":         roles,
	})
}

func (s *Server) handleUserCreate(ctx iris.Context) {
	var req userCreateRequest
	if err := ctx.ReadJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, "hash failed")
		return
	}
	user, err := s.store.CreateGMUser(ctx.Request().Context(), &GMUser{
		Username:     req.Username,
		PasswordHash: string(hash),
		Status:       req.Status,
	}, req.RoleIDs)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.user.create", "gm_user", strconv.FormatInt(user.ID, 10), "user created")
	s.ok(ctx, iris.Map{"id": user.ID})
}

func (s *Server) handleUserUpdate(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	var req userUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	user, err := s.store.UpdateGMUser(ctx.Request().Context(), &GMUser{
		ID:       id,
		Username: req.Username,
		Status:   req.Status,
	}, req.RoleIDs)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.user.update", "gm_user", strconv.FormatInt(user.ID, 10), "user updated")
	s.ok(ctx, iris.Map{"id": user.ID})
}

func (s *Server) handleUserResetPassword(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	var req userPasswordRequest
	if err := ctx.ReadJSON(&req); err != nil || req.Password == "" {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, "hash failed")
		return
	}
	if err := s.store.SetGMUserPassword(ctx.Request().Context(), id, string(hash)); err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.user.password", "gm_user", strconv.FormatInt(id, 10), "password reset")
	s.ok(ctx, iris.Map{"status": "ok"})
}

func (s *Server) handleUserStatus(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	var req userStatusRequest
	if err := ctx.ReadJSON(&req); err != nil || req.Status == "" {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	if err := s.store.SetGMUserStatus(ctx.Request().Context(), id, req.Status); err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.user.status", "gm_user", strconv.FormatInt(id, 10), "status updated")
	s.ok(ctx, iris.Map{"status": "ok"})
}

func (s *Server) handleUserRoles(ctx iris.Context) {
	id, err := ctx.Params().GetInt64("id")
	if err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid id")
		return
	}
	var req userRoleRequest
	if err := ctx.ReadJSON(&req); err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	if err := s.store.SetGMUserRoles(ctx.Request().Context(), id, req.RoleIDs); err != nil {
		s.fail(ctx, iris.StatusInternalServerError, err.Error())
		return
	}
	s.audit(ctx, "gm.user.roles", "gm_user", strconv.FormatInt(id, 10), "roles updated")
	s.ok(ctx, iris.Map{"status": "ok"})
}

func parseGMUserFilter(ctx iris.Context) GMUserFilter {
	limit, _ := ctx.URLParamInt("limit")
	offset, _ := ctx.URLParamInt("offset")
	return GMUserFilter{
		Limit:  limit,
		Offset: offset,
		Search: ctx.URLParam("search"),
	}
}
