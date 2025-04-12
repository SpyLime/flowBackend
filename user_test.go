package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/stretchr/testify/require"
)

// func TestAuth(t *testing.T) {
// 	_, tearDown := FullStartTestServer("auth", 8088, "test@admin.com")
// 	defer tearDown()
// 	client := &http.Client{}

// 	req, _ := http.NewRequest(http.MethodGet,
// 		"http://127.0.0.1:8088/api/users/auth",
// 		nil)

// 	resp, err := client.Do(req)
// 	require.Nil(t, err)
// 	defer resp.Body.Close()
// 	require.NotNil(t, resp)
// 	assert.Equal(t, 200, resp.StatusCode)

// 	var data openapi.ResponseAuth2
// 	decoder := json.NewDecoder(resp.Body)
// 	_ = decoder.Decode(&data)

// 	require.Equal(t, UserRoleAdmin, data.Role)
// 	require.Equal(t, "test@admin.com", data.Email)
// 	require.Equal(t, "test@admin.com", data.Id)
// 	require.True(t, data.IsAuth)
// 	require.True(t, data.IsAdmin)
// 	require.False(t, len(data.SchoolId) == 0)
// }

func TestGetUserByName(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("getUserByName", 8088, "")
	defer tearDown()

	users, _, _, err := CreateTestData(db, &clock, 2, 0, 0)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], false, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}
	userID := url.QueryEscape(users[0])

	req, _ := http.NewRequest(http.MethodGet,
		"http://127.0.0.1:8088/api/v1/user/"+userID,
		nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var data openapi.User
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&data)

	require.Equal(t, int32(0), data.Role)
	require.Equal(t, users[0], data.Username)
}

func TestUpdateUser(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("udpateUser", 8088, "")
	defer tearDown()

	users, _, _, err := CreateTestData(db, &clock, 1, 0, 0)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	originalUser, err := getUser(db, users[0])
	require.Nil(t, err)

	modUser := originalUser

	modUser.FirstName = "Jack"
	modUser.Email = "jack@jack.com"

	client := &http.Client{}

	marshal, err := json.Marshal(modUser)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/user",
		bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	updatedUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, modUser.FirstName, updatedUser.FirstName)
	require.NotEqual(t, originalUser.Email, updatedUser.Email)
}

func TestDeleteUser(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("DeleteUser", 8088, "")
	defer tearDown()

	users, _, _, err := CreateTestData(db, &clock, 1, 0, 0)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}

	userID := url.QueryEscape(users[0])

	req, _ := http.NewRequest(http.MethodDelete, "http://127.0.0.1:8088/api/v1/user/"+userID, nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 204, resp.StatusCode)

	_, err = getUser(db, users[0])
	require.NotNil(t, err)

}
