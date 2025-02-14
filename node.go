package main

import (
	"context"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

// NodeAPIServiceImpl is a service that implements the logic for the NodeAPIServicer
// This service should implement the business logic for every endpoint for the NodeAPI API.
// Include any external packages or services that will be required by this service.
type NodeAPIServiceImpl struct {
	db    *bolt.DB
	clock Clock
}

// NewNodeAPIService creates a default api service
func NewNodeAPIServiceImpl(db *bolt.DB, clock Clock) openapi.NodeAPIServicer {
	return &NodeAPIServiceImpl{
		db:    db,
		clock: clock,
	}
}

// GetNode - get wiki node
func (s *NodeAPIServiceImpl) GetNode(ctx context.Context, nodeId string, tid string) (openapi.ImplResponse, error) {

	node, err := getNode(s.db, nodeId, tid)
	if err != nil {
		return openapi.Response(404, nil), err
	}

	return openapi.Response(200, node), nil

}

// UpdateNode - Update an node
func (s *NodeAPIServiceImpl) UpdateNode(ctx context.Context, updateNodeRequest openapi.AddTopic200ResponseNodeData) (openapi.ImplResponse, error) {
	err := updateNode(s.db, updateNodeRequest)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, nil), nil
}

// AddNode - Add a new node
func (s *NodeAPIServiceImpl) AddNode(ctx context.Context, addTopic200ResponseNodeData openapi.AddTopic200ResponseNodeData) (openapi.ImplResponse, error) {
	response, err := postNode(s.db, s.clock, addTopic200ResponseNodeData)
	if err != nil {
		return openapi.Response(405, nil), err
	}

	return openapi.Response(200, response), nil

}

// DeleteNode - Delete a node
func (s *NodeAPIServiceImpl) DeleteNode(ctx context.Context, nodeId string, tid string) (openapi.ImplResponse, error) {
	err := deleteNode(s.db, nodeId, tid)

	if err == nil {
		return openapi.Response(204, nil), nil
	}

	return openapi.Response(400, nil), err

}
