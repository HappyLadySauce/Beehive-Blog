package attachments

import (
	"context"
	"errors"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"gorm.io/gorm"
)

// ListGroups 扁平列表附件分类。
func (s *Service) ListGroups(ctx context.Context) ([]v1.AttachmentGroupItem, int, error) {
	var rows []models.AttachmentGroup
	if err := s.svc.DB.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	out := make([]v1.AttachmentGroupItem, 0, len(rows))
	for i := range rows {
		g := &rows[i]
		out = append(out, v1.AttachmentGroupItem{
			ID:        g.ID,
			Name:      g.Name,
			ParentID:  g.ParentID,
			SortOrder: g.SortOrder,
		})
	}
	return out, http.StatusOK, nil
}

// CreateGroup 新建附件分类。
func (s *Service) CreateGroup(ctx context.Context, req *v1.CreateAttachmentGroupRequest) (*v1.AttachmentGroupItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if req.ParentID != nil && *req.ParentID > 0 {
		var c int64
		if err := s.svc.DB.WithContext(ctx).Model(&models.AttachmentGroup{}).Where("id = ?", *req.ParentID).Count(&c).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if c == 0 {
			return nil, http.StatusBadRequest, errors.New("parent group not found")
		}
	}
	g := models.AttachmentGroup{
		Name:      req.Name,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
	}
	if err := s.svc.DB.WithContext(ctx).Create(&g).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.AttachmentGroupItem{
		ID:        g.ID,
		Name:      g.Name,
		ParentID:  g.ParentID,
		SortOrder: g.SortOrder,
	}, http.StatusOK, nil
}

// UpdateGroup 更新附件分类。
func (s *Service) UpdateGroup(ctx context.Context, id int64, req *v1.UpdateAttachmentGroupRequest) (*v1.AttachmentGroupItem, int, error) {
	if req == nil || id <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	var g models.AttachmentGroup
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", id).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("group not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, http.StatusBadRequest, errors.New("cannot set parent to self")
		}
		if *req.ParentID > 0 {
			var c int64
			if err := s.svc.DB.WithContext(ctx).Model(&models.AttachmentGroup{}).Where("id = ?", *req.ParentID).Count(&c).Error; err != nil {
				return nil, http.StatusInternalServerError, errors.New("system error")
			}
			if c == 0 {
				return nil, http.StatusBadRequest, errors.New("parent group not found")
			}
		}
		updates["parent_id"] = req.ParentID
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) > 0 {
		if err := s.svc.DB.WithContext(ctx).Model(&g).Updates(updates).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", id).First(&g).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.AttachmentGroupItem{
		ID:        g.ID,
		Name:      g.Name,
		ParentID:  g.ParentID,
		SortOrder: g.SortOrder,
	}, http.StatusOK, nil
}

// DeleteGroup 删除附件分类（有附件引用时禁止）。
func (s *Service) DeleteGroup(ctx context.Context, id int64) (int, error) {
	if id <= 0 {
		return http.StatusBadRequest, errors.New("invalid id")
	}
	var c int64
	if err := s.svc.DB.WithContext(ctx).Model(&models.Attachment{}).Where("group_id = ?", id).Count(&c).Error; err != nil {
		return http.StatusInternalServerError, errors.New("system error")
	}
	if c > 0 {
		return http.StatusBadRequest, errors.New("group still has attachments")
	}
	var sub int64
	if err := s.svc.DB.WithContext(ctx).Model(&models.AttachmentGroup{}).Where("parent_id = ?", id).Count(&sub).Error; err != nil {
		return http.StatusInternalServerError, errors.New("system error")
	}
	if sub > 0 {
		return http.StatusBadRequest, errors.New("group has child groups")
	}
	res := s.svc.DB.WithContext(ctx).Delete(&models.AttachmentGroup{}, id)
	if res.Error != nil {
		return http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return http.StatusNotFound, errors.New("group not found")
	}
	return http.StatusOK, nil
}
