package main

import (
	"context"
	"errors"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/auth/token"
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
	_, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	response, err := getMapById(s.db, topicId)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, response), nil

}

// AddEdge - Add a new edge
func (s *MapAPIServiceImpl) AddEdge(ctx context.Context, topicId string, getMapById200ResponseEdgesInner openapi.GetMapById200ResponseEdgesInner) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}
	userDetails, err := getUser(s.db, user.ID)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	if userDetails.Role != KeyAdmin && userDetails.Reputation < KeyReputationContributor {
		return openapi.Response(401, nil), errors.New("unauthorized: user is not an admin or has low reputation(Contributor)")
	}

	_, err = postEdge(s.db, topicId, getMapById200ResponseEdgesInner)
	if err != nil {
		return openapi.Response(405, nil), err
	}

	return openapi.Response(200, nil), nil

}

// AddEdge - Add a new edge
func (s *MapAPIServiceImpl) DeleteEdge(ctx context.Context, topicId string, edgeId string) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}
	userDetails, err := getUser(s.db, user.ID)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	if userDetails.Role != KeyAdmin && userDetails.Reputation < KeyReputationEditor {
		return openapi.Response(401, nil), errors.New("unauthorized: user is not an admin or has low reputation(Editor)")
	}

	err = deleteEdge(s.db, topicId, edgeId)
	if err != nil {
		return openapi.Response(405, nil), err
	}

	return openapi.Response(204, nil), nil

}
