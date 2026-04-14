package gateway

import (
	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	identityrpc "github.com/HappyLadySauce/Beehive-Blog/services/identity/identity"
	searchrpc "github.com/HappyLadySauce/Beehive-Blog/services/search/search"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
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
