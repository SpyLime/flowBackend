package main

import (
	"context"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

type MapAPIServiceImpl struct {
	db    *bolt.DB
	clock Clock
}

func NewMapAPIServiceImpl(db *bolt.DB, clock Clock) openapi.MapAPIServicer {
	return &MapAPIServiceImpl{
		db:    db,
		clock: clock,
	}
}

// GetMapById - Find map by ID
func (s *MapAPIServiceImpl) GetMapById(ctx context.Context, topicId string) (openapi.ImplResponse, error) {

	response, err := getMapById(s.db, topicId)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, response), nil

}

// AddEdge - Add a new edge
func (s *MapAPIServiceImpl) AddEdge(ctx context.Context, topicId string, getMapById200ResponseEdgesInner openapi.GetMapById200ResponseEdgesInner) (openapi.ImplResponse, error) {

	_, err := postEdge(s.db, topicId, getMapById200ResponseEdgesInner)
	if err != nil {
		return openapi.Response(405, nil), err
	}

	return openapi.Response(200, nil), nil

}

// AddEdge - Add a new edge
func (s *MapAPIServiceImpl) DeleteEdge(ctx context.Context, topicId string, edgeId time.Time) (openapi.ImplResponse, error) {

	// _, err := deleteEdge(s.db, topicId, edgeId)
	// if err != nil {
	// return openapi.Response(405, nil), err
	return openapi.Response(405, nil), nil
	// }

	return openapi.Response(204, nil), nil

}
