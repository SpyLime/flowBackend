package main

import (
	"context"
	"errors"
	"net/http"

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
func (s *NodeAPIServiceImpl) UpdateNode(ctx context.Context, updateNodeRequest openapi.UpdateNodeRequest) (openapi.ImplResponse, error) {
	// TODO - update UpdateNode with the required logic for this service method.
	// Add api_node_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	// return Response(200, nil),nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	// TODO: Uncomment the next line to return response Response(404, {}) or use other options such as http.Ok ...
	// return Response(404, nil),nil

	// TODO: Uncomment the next line to return response Response(405, {}) or use other options such as http.Ok ...
	// return Response(405, nil),nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("UpdateNode method not implemented")
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
	// TODO - update DeleteNode with the required logic for this service method.
	// Add api_node_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	// return Response(204, nil),nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	// TODO: Uncomment the next line to return response Response(403, {}) or use other options such as http.Ok ...
	// return Response(403, nil),nil

	// TODO: Uncomment the next line to return response Response(404, {}) or use other options such as http.Ok ...
	// return Response(404, nil),nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("DeleteNode method not implemented")
}
