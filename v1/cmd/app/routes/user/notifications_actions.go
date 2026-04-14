package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// MarkNotificationRead marks one notification as read for the current user.
func (s *UserService) MarkNotificationRead(ctx context.Context, userID, notificationID int64) (*v1.MarkNotificationReadResponse, int, error) {
	if userID <= 0 || notificationID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	var n models.Notification
	if err := s.svc.DB.WithContext(ctx).Where("id = ? AND user_id = ?", notificationID, userID).First(&n).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("notification not found")
		}
		klog.ErrorS(err, "MarkNotificationRead load", "id", notificationID, "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	now := time.Now().UTC()
	if err := s.svc.DB.WithContext(ctx).Model(&n).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": now,
	}).Error; err != nil {
		klog.ErrorS(err, "MarkNotificationRead update", "id", notificationID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.MarkNotificationReadResponse{ID: n.ID, IsRead: true}, http.StatusOK, nil
}

// DeleteNotification removes a notification owned by the current user.
func (s *UserService) DeleteNotification(ctx context.Context, userID, notificationID int64) (*v1.DeleteNotificationResponse, int, error) {
	if userID <= 0 || notificationID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	res := s.svc.DB.WithContext(ctx).Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{})
	if res.Error != nil {
		klog.ErrorS(res.Error, "DeleteNotification", "id", notificationID, "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("notification not found")
	}
	return &v1.DeleteNotificationResponse{ID: notificationID}, http.StatusOK, nil
}
