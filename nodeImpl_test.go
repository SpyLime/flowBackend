package main

import (
	"testing"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/lgr"
	"github.com/stretchr/testify/require"
)

func TestPostGetNode(t *testing.T) {

	lgr.Printf("INFO TestPostGetNode")
	t.Log("INFO TestPostGetNode")
	db, dbTearDown := OpenTestDB("PostGetNode")
	defer dbTearDown()
	clock := TestClock{}

	users, topics, _, err := CreateTestData(db, &clock, 1, 1, 0)
	require.Nil(t, err)

	node := openapi.AddTopic200ResponseNodeData{
		CreatedBy: users[0],
		Title:     "turbo",
		Topic:     topics[0],
	}

	nodeInfo, err := postNode(db, &clock, node)
	require.Nil(t, err)

	response, err := getNode(db, nodeInfo.TargetId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, node.CreatedBy, response.CreatedBy)
	require.Equal(t, node.Topic, response.Topic)
	require.Equal(t, node.Title, response.Title)

}
