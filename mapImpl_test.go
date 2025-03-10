package main

import (
	"testing"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/lgr"
	"github.com/stretchr/testify/require"
)

func TestGetMapByImpl(t *testing.T) {

	lgr.Printf("INFO TestGetMapByImpl")
	t.Log("INFO TestGetMapByImpl")
	db, dbTearDown := OpenTestDB("GetMapByImpl")
	defer dbTearDown()
	clock := TestClock{}

	_, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 2)
	require.Nil(t, err)

	response, err := getMapById(db, topics[0])
	require.Nil(t, err)

	require.Equal(t, nodesAndEdges[0].SourceId, response.Nodes[0].Id)
	require.NotEqual(t, nodesAndEdges[0].TargetId, response.Nodes[0].Id)

}

func TestPostEdge(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestPostEdge")
	t.Log("INFO TestPostEdge")
	db, dbTearDown := OpenTestDB("PostEdge")
	defer dbTearDown()

	_, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 3)
	require.Nil(t, err)

	oldMap, err := getMapById(db, topics[0])
	require.Nil(t, err)

	require.Equal(t, 3, len(oldMap.Edges))

	edge := openapi.GetMapById200ResponseEdgesInner{
		Id:     nodesAndEdges[1].TargetId.Format(time.RFC3339Nano) + "-" + nodesAndEdges[3].TargetId.Format(time.RFC3339Nano),
		Source: nodesAndEdges[1].TargetId,
		Target: nodesAndEdges[3].TargetId,
	}

	_, err = postEdge(db, topics[0], edge)
	require.Nil(t, err)

	newMap, err := getMapById(db, topics[0])
	require.Nil(t, err)

	require.Equal(t, 4, len(newMap.Edges))

	_, err = postEdge(db, topics[0], edge)
	require.NotNil(t, err)

}

func TestDeleteEdgeImpl(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("deleteEdgeImpl", 8088, "")
	defer tearDown()

	_, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 2)
	require.Nil(t, err)

	edge := openapi.GetMapById200ResponseEdgesInner{
		Id:     nodesAndEdges[1].TargetId.Format(time.RFC3339Nano) + "-" + nodesAndEdges[2].TargetId.Format(time.RFC3339Nano),
		Source: nodesAndEdges[1].TargetId,
		Target: nodesAndEdges[2].TargetId,
	}

	_, err = postEdge(db, topics[0], edge)
	require.Nil(t, err)

	response, err := getMapById(db, topics[0])
	require.Nil(t, err)

	require.Equal(t, len(response.Edges), 3)

	err = deleteEdge(db, topics[0], edge.Id)
	require.Nil(t, err)

	response, err = getMapById(db, topics[0])
	require.Nil(t, err)

	require.Equal(t, len(response.Edges), 2)

}
