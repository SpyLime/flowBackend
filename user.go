package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/auth/token"
	bolt "go.etcd.io/bbolt"
)

// UserAPIServiceImpl is a service that implements the logic for the UserAPIServicer
// This service should implement the business logic for every endpoint for the UserAPI API.
// Include any external packages or services that will be required by this service.
type UserAPIServiceImpl struct {
	db    *bolt.DB
	clock Clock
}

// NewUserAPIService creates a default api service
func NewUserAPIServiceImpl(db *bolt.DB, clock Clock) openapi.UserAPIServicer {
	return &UserAPIServiceImpl{
		db:    db,
		clock: clock,
	}
}

// LoginUser - Login to the system or create account
func (s *UserAPIServiceImpl) LoginUser(ctx context.Context, loginUserRequest openapi.LoginUserRequest) (openapi.ImplResponse, error) {
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
func (s *UserAPIServiceImpl) LogoutUser(ctx context.Context) (openapi.ImplResponse, error) {
	// We can't directly access the request or response writer from the context
	// in the OpenAPI generated code, so we'll just return a success response
	// The actual logout functionality is handled by our custom logout handler
	fmt.Println("LogoutUser method called")
	return openapi.Response(http.StatusOK, nil), nil
}

// UpdateUser - Update user
func (s *UserAPIServiceImpl) UpdateUser(ctx context.Context, updateUserRequest openapi.UpdateUserRequest) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	userDetails, err := getUser(s.db, user.ID)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	if userDetails.Role != KeyAdmin && updateUserRequest.Id != user.ID {
		return openapi.Response(401, nil), errors.New("unauthorized: user is not an admin or trying to update others")
	}

	err = updateUser(s.db, updateUserRequest)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, nil), nil

}

// GetUserByName - Get user by user name
func (s *UserAPIServiceImpl) GetUserByName(ctx context.Context, userId string) (openapi.ImplResponse, error) {
	_, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	response, err := getUser(s.db, userId)
	if err != nil {
		return openapi.Response(400, nil), err
	}

	return openapi.Response(200, response), nil

}

// DeleteUser - Delete user
func (s *UserAPIServiceImpl) DeleteUser(ctx context.Context, userId string) (openapi.ImplResponse, error) {
	user, ok := ctx.Value(userInfoKey).(token.User)
	if !ok {
		return openapi.Response(401, nil), errors.New("unauthorized: user not found in context")
	}

	userDetails, err := getUser(s.db, user.ID)
	if err != nil {
		return openapi.Response(401, nil), err
	}

	if userDetails.Role != KeyAdmin && userId != user.ID {
		return openapi.Response(401, nil), errors.New("unauthorized: user is not an admin or is trying to delete others")
	}

	err = deleteUser(s.db, userId)

	if err == nil {
		return openapi.Response(204, nil), nil
	}

	return openapi.Response(400, nil), err
}

// AuthUser - return authenticated user details
func (s *UserAPIServiceImpl) AuthUser(ctx context.Context) (openapi.ImplResponse, error) {
	// For the OpenAPI implementation, we'll just return a response indicating the user is not authenticated
	// The actual authentication check will be handled by our custom /users/auth endpoint
	response := openapi.AuthUser200Response{
		IsAuth: false,
		Role:   0,
	}

	// Return a successful response with isAuth=false
	return openapi.Response(http.StatusOK, response), nil
}
