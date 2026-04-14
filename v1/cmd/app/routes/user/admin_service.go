package user

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	authutil "github.com/HappyLadySauce/Beehive-Blog/v1/pkg/utils/auth"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/utils/passwd"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

func toAdminUserItem(u *models.User) v1.AdminUserItem {
	return v1.AdminUserItem{
		ID:          u.ID,
		Username:    u.Username,
		Nickname:    u.Nickname,
		Email:       u.Email,
		Avatar:      u.Avatar,
		Role:        string(u.Role),
		Status:      string(u.Status),
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func (s *UserService) adminUsersBaseQuery(ctx context.Context) *gorm.DB {
	return s.svc.DB.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NULL")
}

func (s *UserService) countOtherAdmins(ctx context.Context, excludeID int64) (int64, error) {
	q := s.adminUsersBaseQuery(ctx).Where("role = ?", models.UserRoleAdmin)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var c int64
	if err := q.Count(&c).Error; err != nil {
		return 0, err
	}
	return c, nil
}

func (s *UserService) usernameTaken(ctx context.Context, username string, excludeID int64) (bool, error) {
	q := s.adminUsersBaseQuery(ctx).Where("username = ?", username)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var c int64
	if err := q.Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}

func (s *UserService) emailTaken(ctx context.Context, email string, excludeID int64) (bool, error) {
	q := s.adminUsersBaseQuery(ctx).Where("email = ?", email)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var c int64
	if err := q.Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}

func (s *UserService) getActiveUserByID(ctx context.Context, id int64) (*models.User, int, error) {
	var u models.User
	if err := s.svc.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("user not found")
		}
		klog.ErrorS(err, "load user failed", "userID", id)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &u, http.StatusOK, nil
}

func (s *UserService) syncUserAuthSnapshot(ctx context.Context, userID int64) (int, error) {
	if err := authutil.SyncUserAuthSnapshot(
		ctx,
		s.svc.Redis,
		s.svc.DB,
		s.svc.Config.JWTOptions.ExpireDuration,
		userID,
	); err != nil {
		klog.ErrorS(err, "sync user auth snapshot failed", "userID", userID)
		return http.StatusInternalServerError, errors.New("auth service unavailable")
	}
	return http.StatusOK, nil
}

// AdminCreateUser 管理员创建用户。
func (s *UserService) AdminCreateUser(ctx context.Context, req *v1.AdminCreateUserRequest) (*v1.AdminUserItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	if username == "" || email == "" {
		return nil, http.StatusBadRequest, errors.New("username and email are required")
	}

	if taken, err := s.usernameTaken(ctx, username, 0); err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	} else if taken {
		return nil, http.StatusConflict, errors.New("username already exists")
	}
	if taken, err := s.emailTaken(ctx, email, 0); err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	} else if taken {
		return nil, http.StatusConflict, errors.New("email already exists")
	}

	hash, err := passwd.HashPassword(req.Password)
	if err != nil {
		klog.ErrorS(err, "hash password failed when create user")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	role := models.UserRoleUser
	if strings.TrimSpace(req.Role) != "" {
		role = models.UserRole(strings.TrimSpace(req.Role))
	}
	status := models.UserStatusActive
	if strings.TrimSpace(req.Status) != "" {
		status = models.UserStatus(strings.TrimSpace(req.Status))
	}

	u := &models.User{
		Username: username,
		Nickname: strings.TrimSpace(req.Nickname),
		Email:    email,
		Password: hash,
		Avatar:   strings.TrimSpace(req.Avatar),
		Role:     role,
		Status:   status,
	}
	if err := s.svc.DB.WithContext(ctx).Create(u).Error; err != nil {
		klog.ErrorS(err, "create user failed", "username", username)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	item := toAdminUserItem(u)
	return &item, http.StatusOK, nil
}

// AdminListUsers 管理员用户分页列表。
func (s *UserService) AdminListUsers(ctx context.Context, req *v1.AdminUserListQuery) (*v1.AdminUserListResponse, int, error) {
	if req == nil {
		req = &v1.AdminUserListQuery{}
	}
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := s.adminUsersBaseQuery(ctx)
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("(username ILIKE ? OR nickname ILIKE ? OR email ILIKE ?)", like, like, like)
	}
	if req.Role != "" {
		q = q.Where("role = ?", req.Role)
	}
	if req.Status != "" {
		q = q.Where("status = ?", req.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		klog.ErrorS(err, "count users failed")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	var rows []models.User
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		klog.ErrorS(err, "list users failed")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	items := make([]v1.AdminUserItem, 0, len(rows))
	for i := range rows {
		items = append(items, toAdminUserItem(&rows[i]))
	}

	return &v1.AdminUserListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// AdminGetUser 获取用户详情。
func (s *UserService) AdminGetUser(ctx context.Context, userID int64) (*v1.AdminUserItem, int, error) {
	u, code, err := s.getActiveUserByID(ctx, userID)
	if err != nil {
		return nil, code, err
	}
	item := toAdminUserItem(u)
	return &item, http.StatusOK, nil
}

// AdminUpdateUser 更新用户资料。
func (s *UserService) AdminUpdateUser(ctx context.Context, userID int64, req *v1.AdminUpdateUserRequest) (*v1.AdminUserItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}

	target, code, err := s.getActiveUserByID(ctx, userID)
	if err != nil {
		return nil, code, err
	}

	updates := map[string]interface{}{}
	if req.Nickname != nil {
		updates["nickname"] = strings.TrimSpace(*req.Nickname)
	}
	if req.Avatar != nil {
		updates["avatar"] = strings.TrimSpace(*req.Avatar)
	}
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" {
			return nil, http.StatusBadRequest, errors.New("invalid email")
		}
		taken, err := s.emailTaken(ctx, email, userID)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if taken {
			return nil, http.StatusConflict, errors.New("email already exists")
		}
		updates["email"] = email
	}
	if req.Role != nil {
		newRole := models.UserRole(strings.TrimSpace(*req.Role))
		if target.Role == models.UserRoleAdmin && newRole != models.UserRoleAdmin {
			otherAdmins, err := s.countOtherAdmins(ctx, userID)
			if err != nil {
				return nil, http.StatusInternalServerError, errors.New("system error")
			}
			if otherAdmins == 0 {
				return nil, http.StatusConflict, errors.New("cannot downgrade last admin")
			}
		}
		updates["role"] = newRole
	}

	if len(updates) > 0 {
		if err := s.svc.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			klog.ErrorS(err, "update user failed", "userID", userID)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if code, err := s.syncUserAuthSnapshot(ctx, userID); err != nil {
			return nil, code, err
		}
	}

	return s.AdminGetUser(ctx, userID)
}

// AdminUpdateUserStatus 更新用户状态。
func (s *UserService) AdminUpdateUserStatus(ctx context.Context, userID int64, req *v1.AdminUpdateUserStatusRequest) (*v1.AdminUserItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}

	if _, code, err := s.getActiveUserByID(ctx, userID); err != nil {
		return nil, code, err
	}

	if err := s.svc.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("status", req.Status).Error; err != nil {
		klog.ErrorS(err, "update user status failed", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if code, err := s.syncUserAuthSnapshot(ctx, userID); err != nil {
		return nil, code, err
	}
	return s.AdminGetUser(ctx, userID)
}

// AdminResetUserPassword 管理员设置新密码。
func (s *UserService) AdminResetUserPassword(ctx context.Context, userID int64, req *v1.AdminResetUserPasswordRequest) (*v1.AdminResetUserPasswordResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if _, code, err := s.getActiveUserByID(ctx, userID); err != nil {
		return nil, code, err
	}

	hashed, err := passwd.HashPassword(req.NewPassword)
	if err != nil {
		klog.ErrorS(err, "hash reset password failed", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := s.svc.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password", hashed).Error; err != nil {
		klog.ErrorS(err, "reset password failed", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 使会话快照失效，强制该用户重新登录。
	if s.svc.Redis != nil {
		key := authutil.UserAuthCacheKey(userID)
		if err := s.svc.Redis.Del(ctx, key).Err(); err != nil {
			klog.ErrorS(err, "invalidate auth snapshot failed", "userID", userID)
			return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
		}
	}

	return &v1.AdminResetUserPasswordResponse{
		ID:      userID,
		Message: "password reset success",
	}, http.StatusOK, nil
}

// AdminDeleteUser 软删除用户。
func (s *UserService) AdminDeleteUser(ctx context.Context, operatorID, userID int64) (*v1.AdminDeleteUserResponse, int, error) {
	if operatorID == userID {
		return nil, http.StatusBadRequest, errors.New("cannot delete yourself")
	}
	target, code, err := s.getActiveUserByID(ctx, userID)
	if err != nil {
		return nil, code, err
	}
	if target.Role == models.UserRoleAdmin {
		otherAdmins, err := s.countOtherAdmins(ctx, userID)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if otherAdmins == 0 {
			return nil, http.StatusConflict, errors.New("cannot delete last admin")
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":     models.UserStatusDeleted,
		"deleted_at": &now,
	}
	if err := s.svc.DB.WithContext(ctx).Model(&models.User{}).Where("id = ? AND deleted_at IS NULL", userID).Updates(updates).Error; err != nil {
		klog.ErrorS(err, "soft delete user failed", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if code, err := s.syncUserAuthSnapshot(ctx, userID); err != nil {
		return nil, code, err
	}
	return &v1.AdminDeleteUserResponse{ID: userID}, http.StatusOK, nil
}
