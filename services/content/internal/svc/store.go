package svc

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
)

type contentRecord struct {
	ID           int64
	Type         string
	Title        string
	Slug         string
	Summary      string
	BodyMarkdown string
	Status       string
	Visibility   string
	AiAccess     string
	PublishedAt  string
}

type memoryStore struct {
	mu       sync.RWMutex
	nextID   int64
	items    map[int64]*contentRecord
	slugIdx  map[string]int64
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		nextID:  1,
		items:   make(map[int64]*contentRecord),
		slugIdx: make(map[string]int64),
	}
}

func (s *memoryStore) Create(in *pb.CreateContentRequest) (*pb.ContentDetail, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	typ := strings.TrimSpace(in.Type)
	title := strings.TrimSpace(in.Title)
	slug := strings.TrimSpace(in.Slug)
	if typ == "" || title == "" || slug == "" {
		return nil, fmt.Errorf("type/title/slug are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.slugIdx[slug]; ok {
		return nil, fmt.Errorf("slug already exists")
	}

	id := s.nextID
	s.nextID++

	visibility := defaultIfEmpty(in.Visibility, "private")
	aiAccess := defaultIfEmpty(in.AiAccess, "denied")
	status := "draft"
	publishedAt := ""
	if status == "published" {
		publishedAt = time.Now().Format(time.RFC3339)
	}

	record := &contentRecord{
		ID:           id,
		Type:         typ,
		Title:        title,
		Slug:         slug,
		Summary:      in.Summary,
		BodyMarkdown: in.BodyMarkdown,
		Status:       status,
		Visibility:   visibility,
		AiAccess:     aiAccess,
		PublishedAt:  publishedAt,
	}
	s.items[id] = record
	s.slugIdx[slug] = id

	return toDetail(record), nil
}

func (s *memoryStore) Get(id int64) (*pb.ContentDetail, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	item := s.items[id]
	if item == nil {
		return nil, fmt.Errorf("content not found")
	}
	return toDetail(item), nil
}

func (s *memoryStore) Update(in *pb.UpdateContentRequest) (*pb.ContentDetail, error) {
	if in == nil || in.Id <= 0 {
		return nil, fmt.Errorf("invalid request")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item := s.items[in.Id]
	if item == nil {
		return nil, fmt.Errorf("content not found")
	}
	if strings.TrimSpace(in.Title) != "" {
		item.Title = strings.TrimSpace(in.Title)
	}
	if in.Summary != "" {
		item.Summary = in.Summary
	}
	if in.BodyMarkdown != "" {
		item.BodyMarkdown = in.BodyMarkdown
	}
	if strings.TrimSpace(in.Visibility) != "" {
		item.Visibility = strings.TrimSpace(in.Visibility)
	}
	if strings.TrimSpace(in.AiAccess) != "" {
		item.AiAccess = strings.TrimSpace(in.AiAccess)
	}
	return toDetail(item), nil
}

func (s *memoryStore) UpdateStatus(in *pb.UpdateStatusRequest) (*pb.ContentDetail, error) {
	if in == nil || in.Id <= 0 || strings.TrimSpace(in.Status) == "" {
		return nil, fmt.Errorf("invalid request")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item := s.items[in.Id]
	if item == nil {
		return nil, fmt.Errorf("content not found")
	}
	item.Status = strings.TrimSpace(in.Status)
	if item.Status == "published" && item.PublishedAt == "" {
		item.PublishedAt = time.Now().Format(time.RFC3339)
	}
	return toDetail(item), nil
}

func (s *memoryStore) List(in *pb.ListContentsRequest, publicOnly bool) *pb.ListContentsResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*pb.ContentSummary
	for _, item := range s.items {
		if publicOnly && (item.Visibility != "public" || item.Status != "published") {
			continue
		}
		if in != nil {
			if strings.TrimSpace(in.Type) != "" && item.Type != strings.TrimSpace(in.Type) {
				continue
			}
			if strings.TrimSpace(in.Status) != "" && item.Status != strings.TrimSpace(in.Status) {
				continue
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				if !strings.Contains(strings.ToLower(item.Title), strings.ToLower(kw)) &&
					!strings.Contains(strings.ToLower(item.Summary), strings.ToLower(kw)) {
					continue
				}
			}
		}
		list = append(list, toSummary(item))
	}
	return &pb.ListContentsResponse{List: list}
}

func toSummary(in *contentRecord) *pb.ContentSummary {
	return &pb.ContentSummary{
		Id:          in.ID,
		Type:        in.Type,
		Title:       in.Title,
		Slug:        in.Slug,
		Summary:     in.Summary,
		Status:      in.Status,
		Visibility:  in.Visibility,
		AiAccess:    in.AiAccess,
		PublishedAt: in.PublishedAt,
	}
}

func toDetail(in *contentRecord) *pb.ContentDetail {
	return &pb.ContentDetail{
		Id:           in.ID,
		Type:         in.Type,
		Title:        in.Title,
		Slug:         in.Slug,
		Summary:      in.Summary,
		BodyMarkdown: in.BodyMarkdown,
		Status:       in.Status,
		Visibility:   in.Visibility,
		AiAccess:     in.AiAccess,
	}
}

func defaultIfEmpty(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}
