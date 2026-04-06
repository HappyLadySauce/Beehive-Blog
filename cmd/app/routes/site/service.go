// Package site 处理站点设置读写、SMTP 测试与数据统计。
package site

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/mailer"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/klog/v2"
)

// Service 站点设置与统计业务。
type Service struct {
	svc *svc.ServiceContext
}

// NewService constructs Service.
func NewService(svcCtx *svc.ServiceContext) *Service {
	return &Service{svc: svcCtx}
}

// smtpSensitiveKeys 需要脱敏的 SMTP key（读取时返回 "***"）。
var smtpSensitiveKeys = map[string]bool{
	"smtp.password": true,
}

// GetSettings 按 group 读取设置，SMTP 密码脱敏。
func (s *Service) GetSettings(ctx context.Context, group string) (*v1.SettingsResponse, int, error) {
	group = strings.TrimSpace(group)
	if group == "" {
		return nil, http.StatusBadRequest, errors.New("group is required")
	}
	var rows []models.Setting
	if err := s.svc.DB.WithContext(ctx).Where(`"group" = ?`, group).Find(&rows).Error; err != nil {
		klog.ErrorS(err, "GetSettings: find", "group", group)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	kv := make(map[string]string, len(rows))
	for _, r := range rows {
		if smtpSensitiveKeys[r.Key] {
			kv[r.Key] = "***"
		} else {
			kv[r.Key] = r.Value
		}
	}
	return &v1.SettingsResponse{Group: group, Settings: kv}, http.StatusOK, nil
}

// UpdateSettings 批量 upsert 设置；SMTP 配置更新后重建 Mailer。
func (s *Service) UpdateSettings(ctx context.Context, group string, req *v1.UpdateSettingsRequest) (*v1.SettingsResponse, int, error) {
	group = strings.TrimSpace(group)
	if group == "" {
		return nil, http.StatusBadRequest, errors.New("group is required")
	}
	if req == nil || len(req.Settings) == 0 {
		return s.GetSettings(ctx, group)
	}

	rows := make([]models.Setting, 0, len(req.Settings))
	for k, v := range req.Settings {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		// 忽略密码脱敏占位符（前端未修改时回传 "***"）
		if smtpSensitiveKeys[k] && v == "***" {
			continue
		}
		rows = append(rows, models.Setting{
			Key:   k,
			Value: v,
			Group: group,
		})
	}

	if len(rows) > 0 {
		if err := s.svc.DB.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "group"}),
			}).
			Create(&rows).Error; err != nil {
			klog.ErrorS(err, "UpdateSettings: upsert", "group", group)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}

	// SMTP 配置更新后重建 Mailer
	if group == models.SettingGroupSMTP {
		s.svc.RebuildMailer()
	}

	return s.GetSettings(ctx, group)
}

// TestSMTP 使用当前 SMTP 配置（或请求中临时覆盖）发送测试邮件。
// 不写库，仅验证配置可用性。
func (s *Service) TestSMTP(ctx context.Context, req *v1.TestSMTPRequest) (int, error) {
	if req == nil || strings.TrimSpace(req.To) == "" {
		return http.StatusBadRequest, errors.New("recipient email is required")
	}

	// 从 DB 读取最新 SMTP 配置（不走缓存，确保使用最新值）
	var rows []models.Setting
	if err := s.svc.DB.WithContext(ctx).Where(`"group" = ?`, models.SettingGroupSMTP).Find(&rows).Error; err != nil {
		return http.StatusInternalServerError, errors.New("system error")
	}
	kv := make(map[string]string, len(rows))
	for _, r := range rows {
		kv[r.Key] = r.Value
	}
	cfg := mailer.Config{
		Host:       kv["smtp.host"],
		Port:       kv["smtp.port"],
		Username:   kv["smtp.username"],
		Password:   kv["smtp.password"],
		FromName:   kv["smtp.fromName"],
		Encryption: kv["smtp.encryption"],
	}
	m, err := mailer.New(cfg)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("SMTP not configured: %w", err)
	}

	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := m.Send(sendCtx, req.To, "Beehive-Blog SMTP 测试邮件",
		"<p>这是一封 SMTP 配置测试邮件，收到即表示配置正确。</p>"); err != nil {
		return http.StatusBadGateway, fmt.Errorf("send test email failed: %w", err)
	}
	return http.StatusOK, nil
}

// GetStats 聚合查询站点统计数据。
func (s *Service) GetStats(ctx context.Context) (*v1.SiteStatsResponse, int, error) {
	var articleCount, userCount, commentCount, todayViews int64

	if err := s.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("deleted_at IS NULL").Count(&articleCount).Error; err != nil {
		klog.ErrorS(err, "GetStats: articleCount")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := s.svc.DB.WithContext(ctx).Model(&models.User{}).Count(&userCount).Error; err != nil {
		klog.ErrorS(err, "GetStats: userCount")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := s.svc.DB.WithContext(ctx).Model(&models.Comment{}).
		Where("status = ?", models.CommentStatusApproved).Count(&commentCount).Error; err != nil {
		klog.ErrorS(err, "GetStats: commentCount")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	today := time.Now().Truncate(24 * time.Hour)
	if err := s.svc.DB.WithContext(ctx).Model(&models.ArticleViewLog{}).
		Where("viewed_at >= ?", today).Count(&todayViews).Error; err != nil {
		klog.ErrorS(err, "GetStats: todayViews")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	var topRows []models.Article
	if err := s.svc.DB.WithContext(ctx).
		Select("id, title, view_count, like_count").
		Where("deleted_at IS NULL AND status = ?", models.ArticleStatusPublished).
		Order("view_count DESC").
		Limit(10).
		Find(&topRows).Error; err != nil {
		klog.ErrorS(err, "GetStats: topArticles")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	topArticles := make([]v1.ArticleStatItem, 0, len(topRows))
	for _, a := range topRows {
		topArticles = append(topArticles, v1.ArticleStatItem{
			ID:        a.ID,
			Title:     a.Title,
			ViewCount: a.ViewCount,
			LikeCount: a.LikeCount,
		})
	}

	return &v1.SiteStatsResponse{
		ArticleCount: articleCount,
		UserCount:    userCount,
		CommentCount: commentCount,
		TodayViews:   todayViews,
		TopArticles:  topArticles,
	}, http.StatusOK, nil
}

// validateGroup 检查 group 是否在允许范围内。
func validateGroup(group string) error {
	allowed := map[string]bool{
		models.SettingGroupGeneral:  true,
		models.SettingGroupSEO:      true,
		models.SettingGroupSMTP:     true,
		models.SettingGroupComment:  true,
		models.SettingGroupSecurity: true,
	}
	if !allowed[group] {
		return fmt.Errorf("unknown settings group %q; allowed: general, seo, smtp, comment, security", group)
	}
	return nil
}

// GetSettingsValidated 带 group 校验的读取。
func (s *Service) GetSettingsValidated(ctx context.Context, group string) (*v1.SettingsResponse, int, error) {
	if err := validateGroup(group); err != nil {
		return nil, http.StatusBadRequest, err
	}
	return s.GetSettings(ctx, group)
}

// UpdateSettingsValidated 带 group 校验的写入。
func (s *Service) UpdateSettingsValidated(ctx context.Context, group string, req *v1.UpdateSettingsRequest) (*v1.SettingsResponse, int, error) {
	if err := validateGroup(group); err != nil {
		return nil, http.StatusBadRequest, err
	}
	return s.UpdateSettings(ctx, group, req)
}

// ensure models.Comment is imported (used in GetStats)
var _ = (*gorm.DB)(nil)
