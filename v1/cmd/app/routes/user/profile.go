package user

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"k8s.io/klog/v2"
)

// UpdateProfile updates nickname and/or avatar for the current user.
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, spec *v1.UpdateProfileRequest) (*v1.MeResponse, int, error) {
	if spec == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	nick := strings.TrimSpace(spec.Nickname)
	av := strings.TrimSpace(spec.Avatar)
	if nick == "" && av == "" {
		return nil, http.StatusBadRequest, errors.New("no fields to update")
	}
	updates := map[string]interface{}{}
	if nick != "" {
		updates["nickname"] = nick
	}
	if av != "" {
		updates["avatar"] = av
	}
	if err := s.svc.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		klog.ErrorS(err, "Failed to update profile", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return s.GetMe(ctx, userID)
}
