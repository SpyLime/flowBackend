package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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

	CreateTestData(db, &clock)

	// SetTestLoginUser(teachers[0])

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet,
		"http://127.0.0.1:8088/api/users",
		nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	var data []openapi.UserNoHistory
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&data)

	require.Equal(t, 1, len(data))
	require.NotZero(t, data[0].NetWorth)
}
