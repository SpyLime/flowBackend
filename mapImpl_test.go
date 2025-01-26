package main

import (
	"testing"

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
