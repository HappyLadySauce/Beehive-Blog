package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	authutil "github.com/HappyLadySauce/Beehive-Blog/v1/pkg/utils/auth"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/utils/passwd"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// UpdatePassword verifies the old password, sets a new hash, and invalidates the auth Redis snapshot.
func (s *UserService) UpdatePassword(ctx context.Context, userID int64, spec *v1.UpdatePasswordRequest) (*v1.UpdatePasswordResponse, int, error) {
	if spec == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if spec.OldPassword == spec.NewPassword {
		return nil, http.StatusBadRequest, errors.New("new password must differ from old password")
	}
	if s.svc.Redis == nil {
		klog.Error("Redis client is not configured")
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}
	var u models.User
	if err := s.svc.DB.WithContext(ctx).First(&u, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("user not found")
		}
		klog.ErrorS(err, "Failed to load user for password change", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if !passwd.VerifyPassword(spec.OldPassword, u.Password) {
		return nil, http.StatusUnauthorized, errors.New("invalid old password")
	}
	hashed, err := passwd.HashPassword(spec.NewPassword)
	if err != nil {
		klog.ErrorS(err, "Failed to hash new password", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := s.svc.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password", hashed).Error; err != nil {
		klog.ErrorS(err, "Failed to save new password", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	key := authutil.UserAuthCacheKey(userID)
	if err := s.svc.Redis.Del(ctx, key).Err(); err != nil {
		klog.ErrorS(err, "Failed to invalidate auth snapshot after password change", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("auth service unavailable")
	}
	return &v1.UpdatePasswordResponse{Message: "password updated"}, http.StatusOK, nil
}
