package main

import (
	"context"
	"errors"
	"net/http"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

// UserAPIService is a service that implements the logic for the UserAPIServicer
// This service should implement the business logic for every endpoint for the UserAPI API.
// Include any external packages or services that will be required by this service.
type UserAPIService struct {
	db    *bolt.DB
	clock Clock
}

// NewUserAPIService creates a default api service
func NewUserAPIService(db *bolt.DB, clock Clock) *UserAPIService {
	return &UserAPIService{
		db:    db,
		clock: clock,
	}
}

// LoginUser - Login to the system or create account
func (s *UserAPIService) LoginUser(ctx context.Context, loginUserRequest openapi.LoginUserRequest) (openapi.ImplResponse, error) {
	// TODO - update LoginUser with the required logic for this service method.
	// Add api_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	// return Response(200, nil),nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	// TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	// return Response(401, nil),nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("LoginUser method not implemented")
}

// LogoutUser - Log the user out of the system
func (s *UserAPIService) LogoutUser(ctx context.Context) (openapi.ImplResponse, error) {
	// TODO - update LogoutUser with the required logic for this service method.
	// Add api_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	// return Response(200, nil),nil

	// TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	// return Response(401, nil),nil

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("LogoutUser method not implemented")
}

// UpdateUser - Update user
func (s *UserAPIService) UpdateUser(ctx context.Context, updateUserRequest openapi.UpdateUserRequest) (openapi.ImplResponse, error) {
	// err := updateUserResponse(updateUserRequest)

	// if err == nil {
	// 	return Response(200, nil), nil
	// }

	// return Response(400, nil), err

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("updateUser method not implemented")

}

// GetUserByName - Get user by user name
func (s *UserAPIService) GetUserByName(ctx context.Context, userId string) (openapi.ImplResponse, error) {

	response, err := getUser(s.db, userId)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, response), nil

}

// DeleteUser - Delete user
func (s *UserAPIService) DeleteUser(ctx context.Context, userId string) (openapi.ImplResponse, error) {
	// err := deleteUserResponse(userId)

	// if err == nil {
	// 	return Response(204, nil),nil
	// }

	// return Response(400, nil), err

	return openapi.Response(http.StatusNotImplemented, nil), errors.New("deleteUser method not implemented")

}
