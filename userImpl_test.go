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
