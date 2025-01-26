package main

import (
	"context"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

type MapAPIService struct {
	db    *bolt.DB
	clock Clock
}

// NewMapAPIService creates a default api service
func NewMapAPIService(db *bolt.DB, clock Clock) *MapAPIService {
	return &MapAPIService{
		db:    db,
		clock: clock,
	}
}

// GetMapById - Find map by ID
func (s *MapAPIService) GetMapById(ctx context.Context, topicId string) (openapi.ImplResponse, error) {

	response, err := getMapById(s.db, topicId)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, response), nil

}

// AddEdge - Add a new edge
func (s *MapAPIService) AddEdge(ctx context.Context, topicId string, getMapById200ResponseEdgesInner openapi.GetMapById200ResponseEdgesInner) (openapi.ImplResponse, error) {

	_, err := postEdge(s.db, topicId, getMapById200ResponseEdgesInner)
	if err != nil {
		return openapi.Response(405, nil), err
	}

	return openapi.Response(200, nil), nil

}
