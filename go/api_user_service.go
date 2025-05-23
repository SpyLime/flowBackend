// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Flow Learning - OpenAPI 3.1
 *
 * api for flow learning
 *
 * API version: 1.0.0
 * Contact: floTeam@gmail.com
 */

package openapi

import (
	"context"
	"net/http"
	"errors"
)

// UserAPIService is a service that implements the logic for the UserAPIServicer
// This service should implement the business logic for every endpoint for the UserAPI API.
// Include any external packages or services that will be required by this service.
type UserAPIService struct {
}

// NewUserAPIService creates a default api service
func NewUserAPIService() *UserAPIService {
	return &UserAPIService{}
}

// AuthUser - return authenticated user details
func (s *UserAPIService) AuthUser(ctx context.Context) (ImplResponse, error) {
	// TODO - update AuthUser with the required logic for this service method.
	// Add api_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, AuthUser200Response{}) or use other options such as http.Ok ...
	// return Response(200, AuthUser200Response{}), nil

	// TODO: Uncomment the next line to return response Response(404, AuthUser404Response{}) or use other options such as http.Ok ...
	// return Response(404, AuthUser404Response{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("AuthUser method not implemented")
}

// LoginUser - Login to the system or create account
func (s *UserAPIService) LoginUser(ctx context.Context, loginUserRequest LoginUserRequest) (ImplResponse, error) {
	// TODO - update LoginUser with the required logic for this service method.
	// Add api_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	// return Response(200, nil),nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	// TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	// return Response(401, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("LoginUser method not implemented")
}

// LogoutUser - Log the user out of the system
func (s *UserAPIService) LogoutUser(ctx context.Context) (ImplResponse, error) {
	// TODO - update LogoutUser with the required logic for this service method.
	// Add api_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	// return Response(200, nil),nil

	// TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	// return Response(401, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("LogoutUser method not implemented")
}

// UpdateUser - Update user
func (s *UserAPIService) UpdateUser(ctx context.Context, updateUserRequest UpdateUserRequest) (ImplResponse, error) {
	// TODO - update UpdateUser with the required logic for this service method.
	// Add api_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, {}) or use other options such as http.Ok ...
	// return Response(200, nil),nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("UpdateUser method not implemented")
}

// GetUserByName - Get user by user name
func (s *UserAPIService) GetUserByName(ctx context.Context, userId string) (ImplResponse, error) {
	// TODO - update GetUserByName with the required logic for this service method.
	// Add api_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, UpdateUserRequest{}) or use other options such as http.Ok ...
	// return Response(200, UpdateUserRequest{}), nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	// TODO: Uncomment the next line to return response Response(404, {}) or use other options such as http.Ok ...
	// return Response(404, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("GetUserByName method not implemented")
}

// DeleteUser - Delete user
func (s *UserAPIService) DeleteUser(ctx context.Context, userId string) (ImplResponse, error) {
	// TODO - update DeleteUser with the required logic for this service method.
	// Add api_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	// return Response(204, nil),nil

	// TODO: Uncomment the next line to return response Response(400, {}) or use other options such as http.Ok ...
	// return Response(400, nil),nil

	// TODO: Uncomment the next line to return response Response(404, {}) or use other options such as http.Ok ...
	// return Response(404, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("DeleteUser method not implemented")
}
