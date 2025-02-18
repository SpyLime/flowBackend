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

func TestDeleteNodeImpl(t *testing.T) {

	lgr.Printf("INFO TestDeleteNodeImpl")
	t.Log("INFO TestDeleteNodeImpl")
	clock := TestClock{}
	db, dbTearDown := OpenTestDB("DeleteNodeImpl")
	defer dbTearDown()

	//look into how this creates 2 nodes and 1 edge and document
	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	data := openapi.AddTopic200ResponseNodeData{
		Topic:     topics[0],
		Title:     "tester",
		CreatedBy: users[0],
		Id:        nodesAndEdges[0].SourceId,
	}

	for i := 0; i < 5; i++ {
		_, err = postNode(db, &clock, data)
		require.Nil(t, err)
	}

	oldMap, err := getMapById(db, topics[0])
	require.Nil(t, err)

	require.Equal(t, 6, len(oldMap.Edges))

	err = deleteNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	_, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.NotNil(t, err)

	newMap, err := getMapById(db, topics[0])
	require.Nil(t, err)

	require.Zero(t, len(newMap.Edges))

}

func TestUpdateNodeImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeImpl")
	t.Log("INFO TestUpdateNodeImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeImpl")
	defer dbTearDown()

	_, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	originalNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	modNode := originalNode

	modNode.Title = "Jack"
	modNode.Description = "turbo"

	err = updateNodeTitle(db, modNode)
	require.Nil(t, err)

	updatedNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, modNode.Title, updatedNode.Title)
	require.NotEqual(t, originalNode.Description, updatedNode.Description)
}
