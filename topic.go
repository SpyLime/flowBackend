package main

import (
	"context"
	"errors"
	"net/http"

	openapi "github.com/SpyLime/flowBackend/go"
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

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("UpdateTopic method not implemented")
}

// AddTopic - Add a new topic
func (s *TopicAPIServiceImpl) AddTopic(ctx context.Context, getTopics200ResponseInner openapi.GetTopics200ResponseInner) (openapi.ImplResponse, error) {
	responsePostTopic, err := postTopic(s.db, s.clock, getTopics200ResponseInner)
	if err != nil {
		return openapi.Response(405, nil), err
	}

	response := openapi.AddTopic200Response(responsePostTopic)
	return openapi.Response(200, response), nil

}

// DeleteTopic - Delete a node
func (s *TopicAPIServiceImpl) DeleteTopic(ctx context.Context, topicId string) (openapi.ImplResponse, error) {
	// TODO - update DeleteTopic with the required logic for this service method.
	// Add api_topic_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	// return Response(204, nil),nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	// TODO: Uncomment the next line to return response Response(403, {}) or use other options such as http.Ok ...
	// return Response(403, nil),nil

	// TODO: Uncomment the next line to return response Response(404, {}) or use other options such as http.Ok ...
	// return Response(404, nil),nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("DeleteTopic method not implemented")
}
