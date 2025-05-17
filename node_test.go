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

func TestUpdateNodeBattleVote(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("updateNodeBattleVote", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 3, 1, 1)
	require.Nil(t, err)

	// Set up the first user as logged in
	UpdateUserRoleAndReputation(db, users[1], true, 0)
	SetTestLoginUser(users[1])

	client := &http.Client{}

	// Test upvoting
	upvoteNode := openapi.NodeData{
		Topic:        topics[0],
		Id:           nodesAndEdges[0].SourceId,
		BattleTested: 1,
	}

	marshal, err := json.Marshal(upvoteNode)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/battleVote",
		bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var voteCount int32
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(1), voteCount)
	resp.Body.Close()

	// Verify node was updated in the database
	updatedNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(1), updatedNode.BattleTested)

	// Verify user's vote was recorded
	user, err := getUser(db, users[1])
	require.Nil(t, err)
	require.Equal(t, 1, len(user.BattleTestedUp))
	require.Equal(t, nodesAndEdges[0].SourceId, user.BattleTestedUp[0].NodeId)

	// Test removing upvote (sending the same vote again)
	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/battleVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(0), voteCount)
	resp.Body.Close()

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(0), updatedNode.BattleTested)

	// Verify user's vote was removed
	user, err = getUser(db, users[1])
	require.Nil(t, err)
	require.Equal(t, 0, len(user.BattleTestedUp))

	// Test downvoting
	downvoteNode := openapi.NodeData{
		Topic:        topics[0],
		Id:           nodesAndEdges[0].SourceId,
		BattleTested: -1,
	}

	marshal, err = json.Marshal(downvoteNode)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/battleVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(-1), voteCount)
	resp.Body.Close()

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(-1), updatedNode.BattleTested)

	// Verify user's vote was recorded
	user, err = getUser(db, users[1])
	require.Nil(t, err)
	require.Equal(t, 1, len(user.BattleTestedDown))
	require.Equal(t, nodesAndEdges[0].SourceId, user.BattleTestedDown[0].NodeId)

	// Test switching from downvote to upvote
	marshal, err = json.Marshal(upvoteNode)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/battleVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(1), voteCount)
	resp.Body.Close()

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(1), updatedNode.BattleTested)

	// Verify user's votes were updated correctly
	user, err = getUser(db, users[1])
	require.Nil(t, err)
	require.Equal(t, 0, len(user.BattleTestedDown))
	require.Equal(t, 1, len(user.BattleTestedUp))
	require.Equal(t, nodesAndEdges[0].SourceId, user.BattleTestedUp[0].NodeId)

	// Test reputation changes when another user votes
	// First, get the initial reputation of the node creator
	nodeCreator, err := getUser(db, updatedNode.CreatedBy.Id)
	require.Nil(t, err)
	initialReputation := nodeCreator.Reputation

	// Log in as the second user
	SetTestLoginUser(users[2])

	// Second user upvotes
	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/battleVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()

	// Verify the node creator's reputation increased
	nodeCreator, err = getUser(db, updatedNode.CreatedBy.Id)
	require.Nil(t, err)
	require.Equal(t, initialReputation+1, nodeCreator.Reputation, "Creator's reputation should increase after upvote")

	// Second user downvotes (switching from upvote)
	marshal, err = json.Marshal(downvoteNode)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/battleVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()

	// Verify the node creator's reputation decreased
	nodeCreator, err = getUser(db, updatedNode.CreatedBy.Id)
	require.Nil(t, err)
	require.Equal(t, initialReputation-1, nodeCreator.Reputation, "Creator's reputation should decrease after downvote")
}

func TestUpdateNodeVideoVote(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("updateNodeVideoVote", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 2, 1, 1)
	require.Nil(t, err)

	// Set up the first user as logged in
	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	// First, add a video to the node directly using the function
	videoLink := "https://www.youtube.com/watch?v=testVideo"
	user, err := getUser(db, users[0])
	require.Nil(t, err)

	nodeWithVideo := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			{
				Link:  videoLink,
				Votes: 1,
				AddedBy: openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
					Id:       users[0],
					Username: "tester",
				},
				DateAdded: clock.Now(),
			},
		},
	}

	err = updateNodeVideoEdit(db, &clock, nodeWithVideo, user)
	require.Nil(t, err)

	client := &http.Client{}

	// Test upvoting the video
	upvoteVideo := openapi.NodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			{
				Link:  videoLink,
				Votes: 1,
			},
		},
	}

	marshal, err := json.Marshal(upvoteVideo)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/videoVote",
		bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var voteCount int32
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(1), voteCount)

	// Verify node was updated in the database
	updatedNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(1), updatedNode.YoutubeLinks[0].Votes)

	// Verify user's vote was recorded
	user, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 1, len(user.VideoUp))
	require.Equal(t, videoLink, user.VideoUp[0])

	// Test removing upvote (sending the same vote again)
	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/videoVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(0), voteCount)

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(0), updatedNode.YoutubeLinks[0].Votes)

	// Verify user's vote was removed
	user, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 0, len(user.VideoUp))

	// Test downvoting
	downvoteVideo := openapi.NodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			{
				Link:  videoLink,
				Votes: -1,
			},
		},
	}

	marshal, err = json.Marshal(downvoteVideo)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/videoVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(-1), voteCount)

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(-1), updatedNode.YoutubeLinks[0].Votes)

	// Verify user's vote was recorded
	user, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 1, len(user.VideoDown))
	require.Equal(t, videoLink, user.VideoDown[0])

	// Test switching from downvote to upvote
	marshal, err = json.Marshal(upvoteVideo)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/videoVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(1), voteCount)

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(1), updatedNode.YoutubeLinks[0].Votes)

	// Verify user's votes were updated correctly
	user, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 0, len(user.VideoDown))
	require.Equal(t, 1, len(user.VideoUp))
	require.Equal(t, videoLink, user.VideoUp[0])
}

func TestUpdateNodeFreshVote(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("updateNodeFreshVote", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 2, 1, 1)
	require.Nil(t, err)

	// Set up the first user as logged in
	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}

	// Test upvoting
	upvoteNode := openapi.NodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: 1,
	}

	marshal, err := json.Marshal(upvoteNode)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/freshVote",
		bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var voteCount int32
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(1), voteCount)

	// Verify node was updated in the database
	updatedNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(1), updatedNode.Fresh)

	// Verify user's vote was recorded
	user, err := getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 1, len(user.FreshUp))
	require.Equal(t, nodesAndEdges[0].SourceId, user.FreshUp[0].NodeId)

	// Test removing upvote (sending the same vote again)
	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/freshVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(0), voteCount)

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(0), updatedNode.Fresh)

	// Verify user's vote was removed
	user, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 0, len(user.FreshUp))

	// Test downvoting
	downvoteNode := openapi.NodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: -1,
	}

	marshal, err = json.Marshal(downvoteNode)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/freshVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(-1), voteCount)

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(-1), updatedNode.Fresh)

	// Verify user's vote was recorded
	user, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 1, len(user.FreshDown))
	require.Equal(t, nodesAndEdges[0].SourceId, user.FreshDown[0].NodeId)

	// Test switching from downvote to upvote
	marshal, err = json.Marshal(upvoteNode)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/freshVote",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	decoder = json.NewDecoder(resp.Body)
	_ = decoder.Decode(&voteCount)
	require.Equal(t, int32(1), voteCount)

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, int32(1), updatedNode.Fresh)

	// Verify user's votes were updated correctly
	user, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 0, len(user.FreshDown))
	require.Equal(t, 1, len(user.FreshUp))
	require.Equal(t, nodesAndEdges[0].SourceId, user.FreshUp[0].NodeId)
}

func TestUpdateNodeFlag(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("updateNodeFlag", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 2, 1, 1)
	require.Nil(t, err)

	// Set up the first user as logged in
	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}

	// Test flagging a node
	flagNode := openapi.NodeData{
		Topic:     topics[0],
		Id:        nodesAndEdges[0].SourceId,
		IsFlagged: true,
	}

	marshal, err := json.Marshal(flagNode)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/flag",
		bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	// Verify node was updated in the database
	updatedNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.True(t, updatedNode.IsFlagged)

	// Test unflagging a node
	unflagNode := openapi.NodeData{
		Topic:     topics[0],
		Id:        nodesAndEdges[0].SourceId,
		IsFlagged: false,
	}

	marshal, err = json.Marshal(unflagNode)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/flag",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	// Verify node was updated in the database
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.False(t, updatedNode.IsFlagged)

	// Test that only moderators can unflag
	// First, flag the node again
	// ****************currently anyone can unflag***********************
	// marshal, err = json.Marshal(flagNode)
	// require.Nil(t, err)

	// req, _ = http.NewRequest(http.MethodPut,
	// 	"http://127.0.0.1:8088/api/v1/node/flag",
	// 	bytes.NewBuffer(marshal))

	// resp, err = client.Do(req)
	// require.Nil(t, err)
	// defer resp.Body.Close()
	// require.NotNil(t, resp)
	// require.Equal(t, 200, resp.StatusCode)

	// // Set up a non-moderator user
	// UpdateUserRoleAndReputation(db, users[1], false, 0)
	// SetTestLoginUser(users[1])

	// // Try to unflag as non-moderator
	// marshal, err = json.Marshal(unflagNode)
	// require.Nil(t, err)

	// req, _ = http.NewRequest(http.MethodPut,
	// 	"http://127.0.0.1:8088/api/v1/node/flag",
	// 	bytes.NewBuffer(marshal))

	// resp, err = client.Do(req)
	// require.Nil(t, err)
	// defer resp.Body.Close()

	// // This should either fail with 403 or the node should remain flagged
	// updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	// require.Nil(t, err)

	// if resp.StatusCode == 200 {
	// 	// If the API allows the request but doesn't actually unflag
	// 	require.True(t, updatedNode.IsFlagged, "Non-moderator should not be able to unflag content")
	// } else {
	// 	// If the API rejects the request
	// 	require.Equal(t, 403, resp.StatusCode, "Non-moderator should get 403 when trying to unflag")
	// }
}

func TestUpdateNodeVideoEdit(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("updateNodeVideoEdit", 8088, "")
	defer tearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 2, 1, 1)
	require.Nil(t, err)

	// Set up the first user as logged in
	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	client := &http.Client{}

	// Test adding a video
	videoLink := "https://www.youtube.com/watch?v=testVideo"
	nodeWithVideo := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			{
				Link:  videoLink,
				Votes: 1,
				AddedBy: openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
					Id:       users[0],
					Username: "tester",
				},
				DateAdded: clock.Now(),
			},
		},
	}

	marshal, err := json.Marshal(nodeWithVideo)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/videoEdit",
		bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	// Verify video was added to the node
	updatedNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, 1, len(updatedNode.YoutubeLinks))
	require.Equal(t, videoLink, updatedNode.YoutubeLinks[0].Link)
	require.Equal(t, users[0], updatedNode.YoutubeLinks[0].AddedBy.Id)

	// Verify user's linked videos were updated
	user, err := getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, 1, len(user.Linked))
	require.Equal(t, videoLink, user.Linked[0].Link)

	// Test removing a video (by setting votes to -1)
	removeVideo := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			{
				Link:  videoLink,
				Votes: -1,
			},
		},
	}

	marshal, err = json.Marshal(removeVideo)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/videoEdit",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	// Verify video was removed from the node
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, 0, len(updatedNode.YoutubeLinks))

	// Test adding a second video
	secondVideoLink := "https://www.youtube.com/watch?v=anotherVideo"
	nodeWithSecondVideo := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			{
				Link:  secondVideoLink,
				Votes: 1,
				AddedBy: openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
					Id:       users[0],
					Username: "tester",
				},
				DateAdded: clock.Now(),
			},
		},
	}

	marshal, err = json.Marshal(nodeWithSecondVideo)
	require.Nil(t, err)

	req, _ = http.NewRequest(http.MethodPut,
		"http://127.0.0.1:8088/api/v1/node/videoEdit",
		bytes.NewBuffer(marshal))

	resp, err = client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	// Verify second video was added
	updatedNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.Equal(t, 1, len(updatedNode.YoutubeLinks))
	require.Equal(t, secondVideoLink, updatedNode.YoutubeLinks[0].Link)
}
