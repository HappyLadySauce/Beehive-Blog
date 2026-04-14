// Package likes 处理文章点赞/取消点赞业务。
package likes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Service 点赞业务。
type Service struct {
	svc *svc.ServiceContext
}

// NewService constructs Service.
func NewService(svcCtx *svc.ServiceContext) *Service {
	return &Service{svc: svcCtx}
}

// LikeArticle 为文章点赞；已点赞返回 409；成功后异步发邮件通知作者。
func (s *Service) LikeArticle(ctx context.Context, articleID, userID int64) (*v1.LikeArticleResponse, int, error) {
	if articleID <= 0 || userID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid id")
	}

	var art models.Article
	if err := s.svc.DB.WithContext(ctx).Select("id, title, author_id, like_count").
		Where("id = ? AND deleted_at IS NULL", articleID).First(&art).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("article not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	like := &models.ArticleLike{ArticleID: articleID, UserID: userID}
	if err := s.svc.DB.WithContext(ctx).Create(like).Error; err != nil {
		if isDuplicateError(err) {
			return nil, http.StatusConflict, errors.New("already liked")
		}
		klog.ErrorS(err, "LikeArticle: create", "articleID", articleID, "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 重新读取最新 like_count（触发器已更新）
	var updated models.Article
	if err := s.svc.DB.WithContext(ctx).Select("like_count").Where("id = ?", articleID).First(&updated).Error; err != nil {
		klog.ErrorS(err, "LikeArticle: reload like_count", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	// 异步发邮件通知文章作者（不阻塞请求）
	if art.AuthorID != userID {
		go s.notifyAuthorLiked(art.AuthorID, art.Title, articleID)
	}

	return &v1.LikeArticleResponse{ArticleID: articleID, LikeCount: updated.LikeCount}, http.StatusOK, nil
}

// UnlikeArticle 取消文章点赞；未点赞返回 404。
func (s *Service) UnlikeArticle(ctx context.Context, articleID, userID int64) (*v1.UnlikeArticleResponse, int, error) {
	if articleID <= 0 || userID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid id")
	}

	res := s.svc.DB.WithContext(ctx).
		Where("article_id = ? AND user_id = ?", articleID, userID).
		Delete(&models.ArticleLike{})
	if res.Error != nil {
		klog.ErrorS(res.Error, "UnlikeArticle: delete", "articleID", articleID, "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("like not found")
	}

	var updated models.Article
	if err := s.svc.DB.WithContext(ctx).Select("like_count").Where("id = ?", articleID).First(&updated).Error; err != nil {
		klog.ErrorS(err, "UnlikeArticle: reload like_count", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	return &v1.UnlikeArticleResponse{ArticleID: articleID, LikeCount: updated.LikeCount}, http.StatusOK, nil
}

// notifyAuthorLiked 异步发邮件通知文章作者被点赞。
func (s *Service) notifyAuthorLiked(authorID int64, articleTitle string, articleID int64) {
	if s.svc.Mailer == nil {
		return
	}
	var author models.User
	if err := s.svc.DB.Select("email, nickname, username").
		Where("id = ?", authorID).First(&author).Error; err != nil {
		klog.V(4).InfoS("notifyAuthorLiked: load author", "authorID", authorID, "err", err)
		return
	}
	if strings.TrimSpace(author.Email) == "" {
		return
	}
	name := author.Nickname
	if name == "" {
		name = author.Username
	}
	subject := fmt.Sprintf("您的文章《%s》被点赞了", articleTitle)
	body := fmt.Sprintf(`<p>Hi %s，</p><p>您的文章《<strong>%s</strong>》刚刚被点赞了！</p>`,
		name, articleTitle)
	ctx := context.Background()
	if err := s.svc.Mailer.Send(ctx, author.Email, subject, body); err != nil {
		klog.V(4).InfoS("notifyAuthorLiked: send email", "authorID", authorID, "err", err)
	}
}

// isDuplicateError 检查是否为唯一约束冲突错误（PostgreSQL）。
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
