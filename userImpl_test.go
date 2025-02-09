package main

import (
	"testing"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/lgr"
	"github.com/stretchr/testify/require"
)

func TestPostGetUser(t *testing.T) {

	lgr.Printf("INFO TestPostGetUser")
	t.Log("INFO TestPostGetUser")
	db, dbTearDown := OpenTestDB("PostGetUser")
	defer dbTearDown()

	user := openapi.UpdateUserRequest{
		Username: "tommy",
	}

	userId, err := postUser(db, user)
	require.Nil(t, err)

	response, err := getUser(db, userId)
	require.Nil(t, err)

	require.NotEqual(t, response.Username, user.Username)

}

func TestUpdateUserImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateUserImpl")
	t.Log("INFO TestUpdateUserImpl")
	db, dbTearDown := OpenTestDB("UpdateUserImpl")
	defer dbTearDown()

	users, _, _, err := CreateTestData(db, &clock, 1, 0, 0)
	require.Nil(t, err)

	originalUser, err := getUser(db, users[0])
	require.Nil(t, err)

	modUser := openapi.UpdateUserRequest{
		LastName: "Reno",
		Username: users[0],
	}

	err = updateUser(db, modUser)
	require.Nil(t, err)

	updatedUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.NotEqual(t, originalUser.LastName, updatedUser.LastName)
	require.Equal(t, modUser.LastName, updatedUser.LastName)
}

func TestDeleteUserImpl(t *testing.T) {

	lgr.Printf("INFO TestDeleteUserImpl")
	t.Log("INFO TestDeleteUserImpl")
	clock := TestClock{}
	db, dbTearDown := OpenTestDB("DeleteUserImpl")
	defer dbTearDown()

	users, _, _, err := CreateTestData(db, &clock, 1, 0, 0)
	require.Nil(t, err)

	_, err = getUser(db, users[0])
	require.Nil(t, err)

	err = deleteUser(db, users[0])
	require.Nil(t, err)

	_, err = getUser(db, users[0])
	require.NotNil(t, err)

}
