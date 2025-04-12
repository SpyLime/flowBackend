package main

import (
	"context"
	"errors"
	"net/http"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/auth/token"
	"github.com/go-pkgz/lgr"
	bolt "go.etcd.io/bbolt"
)

// TopicAPIServiceImpl is a service that implements the logic for the TopicAPIServicer
// This service should implement the business logic for every endpoint for the TopicAPI API.
// Include any external packages or services that will be required by this service.
type TopicAPIServiceImpl struct {
	db    *bolt.DB
	clock Clock
}

// NewTopicAPIService creates a default api service
func NewTopicAPIServiceImpl(db *bolt.DB, clock Clock) openapi.TopicAPIServicer {
	return &TopicAPIServiceImpl{
		db:    db,
		clock: clock,
	}
}

// GetTopics - get all topics
func (s *TopicAPIServiceImpl) GetTopics(ctx context.Context) (openapi.ImplResponse, error) {
	// Safely extract user from context
	userValue := ctx.Value("user")
	lgr.Printf("User value from context: %+v", userValue)
	if userValue == nil {
		lgr.Printf("WARNING: User value is nil")
	} else {
		lgr.Printf("User value type: %T", userValue)
	}

	_, ok := ctx.Value("user").(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	response, err := getTopics(s.db)
	if err != nil {
		return openapi.Response(404, nil), err
	}

	return openapi.Response(200, response), nil

}

// UpdateTopic - Update an existing topic
func (s *TopicAPIServiceImpl) UpdateTopic(ctx context.Context, getTopics200ResponseInner openapi.GetTopics200ResponseInner) (openapi.ImplResponse, error) {
	// TODO - update UpdateTopic with the required logic for this service method.
	// Add api_topic_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	// return Response(200, nil),nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	// TODO: Uncomment the next line to return response Response(404, {}) or use other options such as http.Ok ...
	// return Response(404, nil),nil

	// TODO: Uncomment the next line to return response Response(405, {}) or use other options such as http.Ok ...
	// return Response(405, nil),nil

	//this endpoint can't be right in the spec. The only way to update a topic is to change its title. there is no way to easily change a bucket name.
	//You must make a new bucket with the name you want and then copy all the contents into the new bucket and then delete the old one.
	//I don't see any reason to implement this. SysAdmin is the only one that can create topics and if you name it incorrectly then just delete and make a correct one.
	//Make sure you do it quickly before users add items to the bucket.

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("UpdateTopic method not implemented")
}

// AddTopic - Add a new topic
func (s *TopicAPIServiceImpl) AddTopic(ctx context.Context, getTopics200ResponseInner openapi.GetTopics200ResponseInner) (openapi.ImplResponse, error) {
	// Extract user information from context
	user, ok := ctx.Value("user").(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	userDetails, err := getUser(s.db, user.Name)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	if userDetails.Role != KeyAdmin && userDetails.Reputation < KeyReputationDeleter {
		return openapi.Response(401, nil), errors.New("unauthorized: user is not an admin or has low reputation(Deleter)")
	}

	responsePostTopic, err := postTopic(s.db, s.clock, getTopics200ResponseInner)
	if err != nil {
		return openapi.Response(405, nil), err
	}

	response := openapi.AddTopic200Response(responsePostTopic)
	return openapi.Response(200, response), nil
}

// DeleteTopic - Delete a node
func (s *TopicAPIServiceImpl) DeleteTopic(ctx context.Context, topicId string) (openapi.ImplResponse, error) {

	// Extract user information from context
	user, ok := ctx.Value("user").(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	userDetails, err := getUser(s.db, user.Name)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	if userDetails.Role != KeyAdmin {
		return openapi.Response(401, nil), errors.New("unauthorized: user is not an admin email a request for this topic to be deleted")
	}

	err = deleteTopic(s.db, topicId)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(204, nil), nil
}
