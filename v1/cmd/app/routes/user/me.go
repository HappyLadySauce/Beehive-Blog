package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// GetMe loads the current user profile by primary key.
func (s *UserService) GetMe(ctx context.Context, userID int64) (*v1.MeResponse, int, error) {
	var u models.User
	if err := s.svc.DB.WithContext(ctx).First(&u, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("user not found")
		}
		klog.ErrorS(err, "Failed to load user for /me", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.MeResponse{
		ID:               u.ID,
		Username:         u.Username,
		Nickname:         u.Nickname,
		Email:            u.Email,
		Avatar:           u.Avatar,
		Role:             string(u.Role),
		Status:           string(u.Status),
		Level:            u.Level,
		Experience:       u.Experience,
		CommentCount:     u.CommentCount,
		ArticleViewCount: u.ArticleViewCount,
		LastLoginAt:      u.LastLoginAt,
		CreatedAt:        u.CreatedAt,
	}, http.StatusOK, nil
}
