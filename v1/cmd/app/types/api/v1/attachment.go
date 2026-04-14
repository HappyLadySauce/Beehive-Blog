package v1

import "time"

// AttachmentItem 单个附件详情。
type AttachmentItem struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	OriginalName string    `json:"originalName"`
	URL          string    `json:"url"`
	ThumbURL     string    `json:"thumbUrl,omitempty"`
	Type         string    `json:"type"`
	MimeType     string    `json:"mimeType"`
	Size         int64     `json:"size"`
	Width        int       `json:"width,omitempty"`
	Height       int       `json:"height,omitempty"`
	ParentID     *int64    `json:"parentId,omitempty"`
	Variant      string    `json:"variant,omitempty"`
	GroupID      *int64    `json:"groupId,omitempty"`
	GroupName    string    `json:"groupName,omitempty"`
	RefCount     int64     `json:"refArticleCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

// AttachmentListQuery 管理员附件列表查询。
type AttachmentListQuery struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Type      string `form:"type" binding:"omitempty,oneof=image document video audio other"`
	Keyword   string `form:"keyword" binding:"omitempty,max=200"`
	GroupID   *int64 `form:"groupId"`
	RootsOnly *bool  `form:"rootsOnly"` // 默认 true：仅根附件
}

// AttachmentListResponse 分页列表。
type AttachmentListResponse struct {
	Items    []AttachmentItem `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// DeleteAttachmentResponse 删除结果。
type DeleteAttachmentResponse struct {
	ID int64 `json:"id"`
}

// UploadImageResponse 编辑器图片上传结果（仅含 url 和 alt）。
type UploadImageResponse struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

// AttachmentFamilyArticleRef 引用家族内任一附件的文章。
type AttachmentFamilyArticleRef struct {
	ArticleID int64  `json:"articleId"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
}

// AttachmentMemberArticleRefs 单个家族成员在 article_attachments 中关联的文章列表。
type AttachmentMemberArticleRefs struct {
	AttachmentID int64                        `json:"attachmentId"`
	Articles     []AttachmentFamilyArticleRef `json:"articles"`
}

// AttachmentFamilyResponse 根附件及其派生子附件。
type AttachmentFamilyResponse struct {
	Root             AttachmentItem                `json:"root"`
	Children         []AttachmentItem              `json:"children"`
	MemberReferences []AttachmentMemberArticleRefs `json:"memberReferences"`
}

// AttachmentProcessingSettings 全局默认处理参数（存 settings 表 JSON）。
type AttachmentProcessingSettings struct {
	DefaultQuality int      `json:"defaultQuality"`
	AllowedFormats []string `json:"allowedFormats"`
}

// ProcessAttachmentRequest 从根附件生成派生子附件（压缩质量 + 可选输出格式）。
type ProcessAttachmentRequest struct {
	Quality int    `json:"quality" binding:"omitempty,min=1,max=100"`
	Format  string `json:"format" binding:"omitempty,oneof=jpeg jpg png gif webp ico"`
}

// UpdateAttachmentRequest 更新附件：可选物理重命名（移动磁盘文件、更新 path/url/name/original_name，并替换正文中的旧 URL）与可选分类。
// 至少提供 name 或 groupId 之一。groupId 省略表示不改分类；0 或负数表示清除分类；正数为有效分类 ID。根附件变更分类时，子附件 group_id 与根一致。
type UpdateAttachmentRequest struct {
	Name    *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	GroupID *int64  `json:"groupId,omitempty"`
}

// ReplaceAttachmentInArticlesRequest 将正文中 from 附件 URL 批量替换为 to。
type ReplaceAttachmentInArticlesRequest struct {
	FromAttachmentID int64   `json:"fromAttachmentId" binding:"required"`
	ToAttachmentID   int64   `json:"toAttachmentId" binding:"required"`
	ArticleIDs       []int64 `json:"articleIds" binding:"required,min=1"`
}

// ReplaceAttachmentInArticlesResponse 替换结果。
type ReplaceAttachmentInArticlesResponse struct {
	Updated int `json:"updated"`
}

// AttachmentGroupItem 附件分类（独立于文章分类）。
type AttachmentGroupItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ParentID  *int64 `json:"parentId,omitempty"`
	SortOrder int    `json:"sortOrder"`
}

// CreateAttachmentGroupRequest 创建附件分类。
type CreateAttachmentGroupRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=50"`
	ParentID  *int64 `json:"parentId"`
	SortOrder int    `json:"sortOrder" binding:"omitempty,min=0,max=1000000"`
}

// UpdateAttachmentGroupRequest 更新附件分类。
type UpdateAttachmentGroupRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=50"`
	ParentID  *int64  `json:"parentId"`
	SortOrder *int    `json:"sortOrder" binding:"omitempty,min=0,max=1000000"`
}
