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

func TestGetNode(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("getNode", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 2)
	require.Nil(t, err)

	SetTestLoginUser(users[0])

	client := &http.Client{}

	baseURL := "http://127.0.0.1:8088/api/v1/node"
	params := url.Values{}
	params.Add("nodeId", nodesAndEdges[0].SourceId.Format(time.RFC3339Nano))
	params.Add("tid", topics[0])

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, _ := http.NewRequest(http.MethodGet, url, nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var data openapi.NodeData
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&data)

	require.Equal(t, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), data.Id.Format(time.RFC3339Nano))

}

func TestPostNode(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("PostNode", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}

	newNode := openapi.NodeData{
		Id:    nodesAndEdges[0].TargetId,
		Topic: topics[0],
		Title: "turbo",
		CreatedBy: openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
			Id:       users[0],
			Username: users[0],
		},
	}

	marshal, err := json.Marshal(newNode)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8088/api/v1/node", bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var data openapi.ResponsePostNode
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&data)

	node, err := getNode(db, data.TargetId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.NotEqual(t, nodesAndEdges[0].TargetId.Format(time.RFC3339Nano), node.Id.Format(time.RFC3339Nano))

}

func TestDeleteNode(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("DeleteNode", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}

	baseURL := "http://127.0.0.1:8088/api/v1/node"
	params := url.Values{}
	params.Add("nodeId", nodesAndEdges[0].SourceId.Format(time.RFC3339Nano))
	params.Add("tid", topics[0])

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 204, resp.StatusCode)

	_, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.NotNil(t, err)

}

func TestUpdateNode(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("udpateNode", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	originalNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	modNode := originalNode

	modNode.Title = "Jack"
	modNode.Description = "turbo"

	client := &http.Client{}

	marshal, err := json.Marshal(modNode)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/title",
		bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	updatedNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, modNode.Title, updatedNode.Title)
	require.NotEqual(t, originalNode.Description, updatedNode.Description)
}

//I can't test all these endpoints until SSO is complete because I currently have no way to grab the user
// func TestUpdateBattleVote(t *testing.T) {
// 	clock := TestClock{}
// 	db, tearDown := FullStartTestServer("udpateNode", 8088, "")
// 	defer tearDown()

// 	_, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
// 	require.Nil(t, err)

// 	client := &http.Client{}

// 	modNode := openapi.NodeData{
// 		BattleTested: 1,
// 		Topic:        topics[0],
// 		Id:           nodesAndEdges[0].SourceId,
// 	}

// 	marshal, err := json.Marshal(modNode)
// 	require.Nil(t, err)

// 	req, _ := http.NewRequest(http.MethodPut,
// 		"http://127.0.0.1:8088/api/v1/node/battleVote",
// 		bytes.NewBuffer(marshal))

// 	resp, err := client.Do(req)
// 	require.Nil(t, err)
// 	defer resp.Body.Close()
// 	require.Equal(t, 200, resp.StatusCode)

// }
