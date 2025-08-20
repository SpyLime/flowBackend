package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/stretchr/testify/require"
)

func TestGetMapById(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("GetMapById", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 2)
	require.Nil(t, err)

	SetTestLoginUser(users[0])

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:8088/api/v1/map/"+topics[0], nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var data openapi.MapData
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&data)

	//root + the two created
	require.Equal(t, len(nodesAndEdges), len(data.Nodes))
	//the first node does not have an edge, only the 2 aditional ones
	require.Equal(t, len(nodesAndEdges)-1, len(data.Edges))

}

func TestAddEdge(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("AddEdge", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 2)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}

	newTopic := openapi.Edge{
		Id:     nodesAndEdges[1].TargetId.Format(time.RFC3339Nano) + "-" + nodesAndEdges[2].TargetId.Format(time.RFC3339Nano),
		Source: nodesAndEdges[1].TargetId,
		Target: nodesAndEdges[2].TargetId,
	}

	marshal, err := json.Marshal(newTopic)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8088/api/v1/map/"+topics[0]+"/edge", bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	newMap, err := getMapById(db, topics[0])
	require.Nil(t, err)

	require.Equal(t, 3, len(newMap.Edges)) //root to the 2 other nodes plus 1 more that was just created between the second and third node

}

func TestDeleteEdge(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("deleteEdge", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 2)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}

	edge := openapi.Edge{
		Id:     nodesAndEdges[1].TargetId.Format(time.RFC3339Nano) + "-" + nodesAndEdges[2].TargetId.Format(time.RFC3339Nano),
		Source: nodesAndEdges[1].TargetId,
		Target: nodesAndEdges[2].TargetId,
	}

	_, err = postEdge(db, topics[0], edge)
	require.Nil(t, err)

	baseURL := "http://127.0.0.1:8088/api/v1/map/" + topics[0] + "/edge"
	params := url.Values{}
	params.Add("edgeId", edge.Id)

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 204, resp.StatusCode)

}
