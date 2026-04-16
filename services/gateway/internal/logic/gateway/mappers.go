package gateway

import (
	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	identityrpc "github.com/HappyLadySauce/Beehive-Blog/services/identity/identity"
	searchrpc "github.com/HappyLadySauce/Beehive-Blog/services/search/search"
)

func toTokenData(in *identityrpc.TokenReply) *types.TokenData {
	if in == nil {
		return &types.TokenData{}
	}
	return &types.TokenData{
		AccessToken:  in.AccessToken,
		RefreshToken: in.RefreshToken,
		ExpiresIn:    in.ExpiresIn,
		User: types.UserProfile{
			Id:       in.GetUser().GetId(),
			Username: in.GetUser().GetUsername(),
			Nickname: in.GetUser().GetNickname(),
			Email:    in.GetUser().GetEmail(),
			Role:     in.GetUser().GetRole(),
		},
	}
}

func toUserProfile(in *identityrpc.UserProfile) *types.UserProfile {
	if in == nil {
		return &types.UserProfile{}
	}
	return &types.UserProfile{
		Id:       in.Id,
		Username: in.Username,
		Nickname: in.Nickname,
		Email:    in.Email,
		Role:     in.Role,
	}
}

func toContentDetail(in *contentrpc.ContentDetail) *types.ContentDetail {
	if in == nil {
		return &types.ContentDetail{}
	}
	return &types.ContentDetail{
		Id:           in.Id,
		ContentType:  in.Type,
		Title:        in.Title,
		Slug:         in.Slug,
		Summary:      in.Summary,
		BodyMarkdown: in.BodyMarkdown,
		Status:       in.Status,
		Visibility:   in.Visibility,
		AiAccess:     in.AiAccess,
	}
}

func toContentListResponse(in *contentrpc.ListContentsResponse) *types.ContentListResponse {
	resp := &types.ContentListResponse{List: []types.ContentSummary{}}
	if in == nil {
		return resp
	}

	for _, item := range in.List {
		resp.List = append(resp.List, types.ContentSummary{
			Id:          item.Id,
			ContentType: item.Type,
			Title:       item.Title,
			Slug:        item.Slug,
			Summary:     item.Summary,
			Status:      item.Status,
			Visibility:  item.Visibility,
			AiAccess:    item.AiAccess,
			PublishedAt: item.PublishedAt,
		})
	}
	return resp
}

func toSearchResponse(in *searchrpc.SearchResponse) *types.SearchResponse {
	resp := &types.SearchResponse{List: []types.SearchResultItem{}}
	if in == nil {
		return resp
	}
	for _, item := range in.List {
		resp.List = append(resp.List, types.SearchResultItem{
			ContentId:   item.ContentId,
			ContentType: item.Type,
			Title:       item.Title,
			Slug:        item.Slug,
			Summary:     item.Summary,
			Highlight:   item.Highlight,
			Score:       item.Score,
		})
	}
	return resp
}

func toSearchIndexDocument(in *searchrpc.IndexDocument) *types.SearchIndexDocument {
	if in == nil {
		return &types.SearchIndexDocument{}
	}
	return &types.SearchIndexDocument{
		ContentId:   in.ContentId,
		ContentType: in.Type,
		Title:       in.Title,
		Slug:        in.Slug,
		Status:      in.Status,
		Visibility:  in.Visibility,
		IndexedAt:   in.IndexedAt,
	}
}

func toTagListResponse(in *contentrpc.ListTagsResponse) *types.TagListResponse {
	resp := &types.TagListResponse{List: []types.Tag{}}
	if in == nil {
		return resp
	}
	for _, item := range in.List {
		resp.List = append(resp.List, types.Tag{
			Id:          item.Id,
			Name:        item.Name,
			Slug:        item.Slug,
			Color:       item.Color,
			Description: item.Description,
		})
	}
	return resp
}

func toTag(in *contentrpc.Tag) *types.Tag {
	if in == nil {
		return &types.Tag{}
	}
	return &types.Tag{
		Id:          in.Id,
		Name:        in.Name,
		Slug:        in.Slug,
		Color:       in.Color,
		Description: in.Description,
	}
}

func toRelationListResponse(in *contentrpc.ListRelationsResponse) *types.RelationListResponse {
	resp := &types.RelationListResponse{List: []types.Relation{}}
	if in == nil {
		return resp
	}
	for _, item := range in.List {
		resp.List = append(resp.List, types.Relation{
			Id:              item.Id,
			SourceContentId: item.SourceContentId,
			TargetContentId: item.TargetContentId,
			RelationType:    item.RelationType,
			Weight:          item.Weight,
			Note:            item.Note,
		})
	}
	return resp
}

func toRelation(in *contentrpc.Relation) *types.Relation {
	if in == nil {
		return &types.Relation{}
	}
	return &types.Relation{
		Id:              in.Id,
		SourceContentId: in.SourceContentId,
		TargetContentId: in.TargetContentId,
		RelationType:    in.RelationType,
		Weight:          in.Weight,
		Note:            in.Note,
	}
}

func toAttachmentListResponse(in *contentrpc.ListAttachmentsResponse) *types.AttachmentListResponse {
	resp := &types.AttachmentListResponse{List: []types.Attachment{}}
	if in == nil {
		return resp
	}
	for _, item := range in.List {
		resp.List = append(resp.List, types.Attachment{
			Id:              item.Id,
			ContentId:       item.ContentId,
			StorageProvider: item.StorageProvider,
			Bucket:          item.Bucket,
			ObjectKey:       item.ObjectKey,
			FileName:        item.FileName,
			MimeType:        item.MimeType,
			Ext:             item.Ext,
			SizeBytes:       item.SizeBytes,
			UsageType:       item.UsageType,
		})
	}
	return resp
}

func toAttachment(in *contentrpc.Attachment) *types.Attachment {
	if in == nil {
		return &types.Attachment{}
	}
	return &types.Attachment{
		Id:              in.Id,
		ContentId:       in.ContentId,
		StorageProvider: in.StorageProvider,
		Bucket:          in.Bucket,
		ObjectKey:       in.ObjectKey,
		FileName:        in.FileName,
		MimeType:        in.MimeType,
		Ext:             in.Ext,
		SizeBytes:       in.SizeBytes,
		UsageType:       in.UsageType,
	}
}

func toCommentListResponse(in *contentrpc.ListCommentsResponse) *types.CommentListResponse {
	resp := &types.CommentListResponse{List: []types.Comment{}}
	if in == nil {
		return resp
	}
	for _, item := range in.List {
		resp.List = append(resp.List, types.Comment{
			Id:              item.Id,
			ContentId:       item.ContentId,
			ParentCommentId: item.ParentCommentId,
			AuthorName:      item.AuthorName,
			AuthorEmail:     item.AuthorEmail,
			BodyMarkdown:    item.BodyMarkdown,
			Status:          item.Status,
			ModerationNote:  item.ModerationNote,
		})
	}
	return resp
}

func toComment(in *contentrpc.Comment) *types.Comment {
	if in == nil {
		return &types.Comment{}
	}
	return &types.Comment{
		Id:              in.Id,
		ContentId:       in.ContentId,
		ParentCommentId: in.ParentCommentId,
		AuthorName:      in.AuthorName,
		AuthorEmail:     in.AuthorEmail,
		BodyMarkdown:    in.BodyMarkdown,
		Status:          in.Status,
		ModerationNote:  in.ModerationNote,
	}
}
