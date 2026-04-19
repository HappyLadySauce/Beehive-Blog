package svc

import (
	"context"
	"fmt"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
)

func (s *contentStore) ListTags(ctx context.Context) (*pb.ListTagsResponse, error) {
	query := `SELECT id, name, slug, color, description FROM tags ORDER BY id DESC`
	var rows []tagRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		return nil, err
	}
	list := make([]*pb.Tag, 0, len(rows))
	for i := range rows {
		list = append(list, &pb.Tag{
			Id:          rows[i].ID,
			Name:        rows[i].Name,
			Slug:        rows[i].Slug,
			Color:       rows[i].Color,
			Description: rows[i].Description,
		})
	}
	return &pb.ListTagsResponse{List: list}, nil
}

func (s *contentStore) CreateTag(ctx context.Context, in *pb.CreateTagRequest) (*pb.Tag, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	name := strings.TrimSpace(in.Name)
	slug := strings.TrimSpace(in.Slug)
	if name == "" || slug == "" {
		return nil, fmt.Errorf("name and slug are required")
	}
	query := `
INSERT INTO tags(name, slug, color, description)
VALUES ($1, $2, $3, $4)
RETURNING id, name, slug, color, description`
	var out tagRecord
	if err := s.conn.QueryRowCtx(ctx, &out, query, name, slug, strings.TrimSpace(in.Color), in.Description); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("tag already exists")
		}
		return nil, err
	}
	return &pb.Tag{
		Id:          out.ID,
		Name:        out.Name,
		Slug:        out.Slug,
		Color:       out.Color,
		Description: out.Description,
	}, nil
}

func (s *contentStore) DeleteTag(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	_, err := s.conn.ExecCtx(ctx, `DELETE FROM tags WHERE id = $1`, id)
	return err
}
