package svc

import (
	"context"
	"fmt"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
)

func (s *contentStore) ListRelations(ctx context.Context, contentID int64) (*pb.ListRelationsResponse, error) {
	if contentID <= 0 {
		return nil, fmt.Errorf("invalid content_id")
	}
	query := `
SELECT id, source_content_id, target_content_id, relation_type, weight, note
FROM content_relations
WHERE source_content_id = $1
ORDER BY id DESC`
	var rows []relationRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, contentID); err != nil {
		return nil, err
	}
	list := make([]*pb.Relation, 0, len(rows))
	for i := range rows {
		list = append(list, &pb.Relation{
			Id:              rows[i].ID,
			SourceContentId: rows[i].SourceContentID,
			TargetContentId: rows[i].TargetContentID,
			RelationType:    rows[i].RelationType,
			Weight:          rows[i].Weight,
			Note:            rows[i].Note,
		})
	}
	return &pb.ListRelationsResponse{List: list}, nil
}

func (s *contentStore) CreateRelation(ctx context.Context, in *pb.CreateRelationRequest) (*pb.Relation, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	if in.SourceContentId <= 0 || in.TargetContentId <= 0 || strings.TrimSpace(in.RelationType) == "" {
		return nil, fmt.Errorf("source_content_id,target_content_id,relation_type are required")
	}
	weight := in.Weight
	if weight <= 0 {
		weight = 1
	}
	query := `
INSERT INTO content_relations (source_content_id, target_content_id, relation_type, weight, note)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, source_content_id, target_content_id, relation_type, weight, note`
	var out relationRecord
	if err := s.conn.QueryRowCtx(ctx, &out, query, in.SourceContentId, in.TargetContentId, strings.TrimSpace(in.RelationType), weight, in.Note); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("relation already exists")
		}
		return nil, err
	}
	return &pb.Relation{
		Id:              out.ID,
		SourceContentId: out.SourceContentID,
		TargetContentId: out.TargetContentID,
		RelationType:    out.RelationType,
		Weight:          out.Weight,
		Note:            out.Note,
	}, nil
}

func (s *contentStore) DeleteRelation(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	_, err := s.conn.ExecCtx(ctx, `DELETE FROM content_relations WHERE id = $1`, id)
	return err
}
