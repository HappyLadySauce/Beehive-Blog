package attachment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// CategoryService orchestrates attachment category persistence.
// CategoryService 编排附件分类持久化逻辑。
type CategoryService struct {
	db *gorm.DB
}

// NewCategoryService builds a CategoryService bound to the given database handle.
// NewCategoryService 基于给定数据库句柄构造 CategoryService。
func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

// Create creates a new attachment category.
// Create 创建附件分类。
func (s *CategoryService) Create(ctx context.Context, actor Actor, in CategoryCreateInput) (model.AttachmentCategory, error) {
	if err := RequireAdmin(actor); err != nil {
		return model.AttachmentCategory{}, err
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Slug) == "" {
		return model.AttachmentCategory{}, fmt.Errorf("%w: name and slug are required", ErrInvalid)
	}
	status := in.Status
	if status == "" {
		status = CategoryStatusActive
	}
	if !CategoryStatusKnown(status) {
		return model.AttachmentCategory{}, fmt.Errorf("%w: invalid category status", ErrInvalid)
	}
	row := model.AttachmentCategory{
		ParentID:    in.ParentID,
		Name:        strings.TrimSpace(in.Name),
		Slug:        strings.TrimSpace(in.Slug),
		Description: CleanOptional(in.Description),
		Icon:        CleanOptional(in.Icon),
		SortOrder:   in.SortOrder,
		Status:      status,
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return model.AttachmentCategory{}, MapDBError(err)
	}
	return row, nil
}

// List returns live categories ordered for tree rendering.
// List 返回按树展示排序的未软删分类。
func (s *CategoryService) List(ctx context.Context, actor Actor) ([]model.AttachmentCategory, error) {
	if err := RequireAdmin(actor); err != nil {
		return nil, err
	}
	var rows []model.AttachmentCategory
	if err := s.db.WithContext(ctx).Order("path ASC").Find(&rows).Error; err != nil {
		return nil, MapDBError(err)
	}
	return rows, nil
}

// Get returns one category by ID.
// Get 按 ID 返回单个分类。
func (s *CategoryService) Get(ctx context.Context, actor Actor, id int64) (model.AttachmentCategory, error) {
	if err := RequireAdmin(actor); err != nil {
		return model.AttachmentCategory{}, err
	}
	var row model.AttachmentCategory
	if err := s.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return model.AttachmentCategory{}, MapDBError(err)
	}
	return row, nil
}

// Patch updates a category.
// Patch 更新分类。
func (s *CategoryService) Patch(ctx context.Context, actor Actor, id int64, in CategoryPatchInput) (model.AttachmentCategory, error) {
	if err := RequireAdmin(actor); err != nil {
		return model.AttachmentCategory{}, err
	}
	var out model.AttachmentCategory
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&out, "id = ?", id).Error; err != nil {
			return MapDBError(err)
		}
		updates := map[string]interface{}{"updated_at": time.Now()}
		if in.ParentID != nil {
			updates["parent_id"] = *in.ParentID
		}
		if in.Name != nil {
			updates["name"] = strings.TrimSpace(*in.Name)
		}
		if in.Slug != nil {
			updates["slug"] = strings.TrimSpace(*in.Slug)
		}
		if in.Description != nil {
			updates["description"] = strings.TrimSpace(*in.Description)
		}
		if in.Icon != nil {
			updates["icon"] = strings.TrimSpace(*in.Icon)
		}
		if in.SortOrder != nil {
			updates["sort_order"] = *in.SortOrder
		}
		if in.Status != nil {
			if !CategoryStatusKnown(*in.Status) {
				return fmt.Errorf("%w: invalid category status", ErrInvalid)
			}
			updates["status"] = strings.TrimSpace(*in.Status)
		}
		if err := tx.Model(&out).Updates(updates).Error; err != nil {
			return MapDBError(err)
		}
		return tx.First(&out, "id = ?", id).Error
	})
	if err != nil {
		return model.AttachmentCategory{}, err
	}
	return out, nil
}

// Delete soft-deletes a category.
// Delete 软删分类。
func (s *CategoryService) Delete(ctx context.Context, actor Actor, id int64) error {
	if err := RequireAdmin(actor); err != nil {
		return err
	}
	res := s.db.WithContext(ctx).Delete(&model.AttachmentCategory{}, "id = ?", id)
	if res.Error != nil {
		return MapDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
