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
	CreatedAt    time.Time `json:"createdAt"`
}

// AttachmentListQuery 管理员附件列表查询。
type AttachmentListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Type     string `form:"type" binding:"omitempty,oneof=image document video audio other"`
	Keyword  string `form:"keyword" binding:"omitempty,max=200"`
	GroupID  *int64 `form:"groupId"`
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
