package main

import (
	"context"
	"errors"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/auth/token"
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

// UpdateNode - Update an node
func (s *NodeAPIServiceImpl) UpdateNodeBattleVote(ctx context.Context, updateNodeRequest openapi.NodeData) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	vote, err := updateNodeBattleVote(s.db, updateNodeRequest, user.ID)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, vote), nil
}

// UpdateNode - Update an node
func (s *NodeAPIServiceImpl) UpdateNodeTitle(ctx context.Context, updateNodeRequest openapi.NodeData) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	userDetails, err := getUser(s.db, user.ID)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	// Get the node to check creation time and creator
	node, err := getNode(s.db, updateNodeRequest.Id.Format(time.RFC3339Nano), updateNodeRequest.Topic)
	if err != nil {
		return openapi.Response(404, nil), err
	}

	// Check if user is the creator and within 15 minutes of creation
	isCreator := node.CreatedBy.Id == userDetails.Id
	creationTime := updateNodeRequest.Id
	withinEditWindow := time.Since(creationTime) <= 15*time.Minute

	// Allow edit if user is creator and within edit window, or has sufficient reputation
	if !(isCreator && withinEditWindow) && userDetails.Role != KeyAdmin && userDetails.Reputation < KeyReputationEditor {
		return openapi.Response(401, nil), errors.New("unauthorized: user is not an admin or has low reputation(Editor)")
	}

	editorAdded, err := updateNodeTitle(s.db, updateNodeRequest, userDetails)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, editorAdded), nil
}

// UpdateNode - Update an node
func (s *NodeAPIServiceImpl) UpdateNodeVideoEdit(ctx context.Context, updateNodeRequest openapi.NodeData) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	userDetails, err := getUser(s.db, user.ID)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	err = updateNodeVideoEdit(s.db, s.clock, updateNodeRequest, userDetails)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, nil), nil
}

// UpdateNode - Update an node
func (s *NodeAPIServiceImpl) UpdateNodeVideoVote(ctx context.Context, updateNodeRequest openapi.NodeData) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	vote, err := updateNodeVideoVote(s.db, updateNodeRequest, user.ID)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, vote), nil
}

// UpdateNode - Update an node
func (s *NodeAPIServiceImpl) UpdateNodeFlag(ctx context.Context, updateNodeRequest openapi.NodeData) (openapi.ImplResponse, error) {
	_, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	err := updateNodeFlag(s.db, updateNodeRequest)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, nil), nil
}

// UpdateNode - Update an node
func (s *NodeAPIServiceImpl) UpdateNodeFreshVote(ctx context.Context, updateNodeRequest openapi.NodeData) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	vote, err := updateNodeFreshVote(s.db, updateNodeRequest, user.ID)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, vote), nil
}

// GetNode - get wiki node
func (s *NodeAPIServiceImpl) GetNode(ctx context.Context, nodeId string, tid string) (openapi.ImplResponse, error) {
	node, err := getNode(s.db, nodeId, tid)
	if err != nil {
		return openapi.Response(404, nil), err
	}

	return openapi.Response(200, node), nil

}

// AddNode - Add a new node
func (s *NodeAPIServiceImpl) AddNode(ctx context.Context, nodeData openapi.NodeData) (openapi.ImplResponse, error) {
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

	nodeData.CreatedBy = openapi.UserIdentifier{
		Id:       user.ID,
		Username: user.Name,
	}

	response, err := postNode(s.db, s.clock, nodeData)
	if err != nil {
		return openapi.Response(405, nil), err
	}

	return openapi.Response(200, response), nil

}

// DeleteNode - Delete a node
func (s *NodeAPIServiceImpl) DeleteNode(ctx context.Context, nodeId string, tid string) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	userDetails, err := getUser(s.db, user.ID)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	// Get the node to check creation time and creator
	node, err := getNode(s.db, nodeId, tid)
	if err != nil {
		return openapi.Response(404, nil), err
	}

	// Check if user is the creator and within 15 minutes of creation
	isCreator := node.CreatedBy.Id == userDetails.Id
	creationTime := node.Id
	withinEditWindow := time.Since(creationTime) <= 15*time.Minute

	// Allow edit if user is creator and within edit window, or has sufficient reputation
	if !(isCreator && withinEditWindow) && userDetails.Role != KeyAdmin && userDetails.Reputation < KeyReputationDeleter {
		return openapi.Response(401, nil), errors.New("unauthorized: user is not an admin or has low reputation(Deleter)")
	}

	err = deleteNode(s.db, nodeId, tid)

	if err == nil {
		return openapi.Response(204, nil), nil
	}

	return openapi.Response(400, nil), err

}
