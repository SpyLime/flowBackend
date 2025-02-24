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

func TestUpdateBattleVoteUpImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateBattleVoteUpImpl")
	t.Log("INFO TestUpdateBattleVoteUpImpl")
	db, dbTearDown := OpenTestDB("UpdateBattleVoteUPImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	battleUp := openapi.AddTopic200ResponseNodeData{
		Topic:        topics[0],
		Id:           nodesAndEdges[0].SourceId,
		BattleTested: 1,
	}

	err = updateNodeBattleVote(db, battleUp, users[0]) // should cause +1
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.BattleTested, int32(1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.BattleTestedUp), 1)

	err = updateNodeBattleVote(db, battleUp, users[0]) // should cause -1
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Zero(t, upNode.BattleTested)

	upUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(upUser.BattleTestedUp))
}

func TestUpdateBattleVoteDownImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateBattleVoteDownImpl")
	t.Log("INFO TestUpdateBattleVoteDownImpl")
	db, dbTearDown := OpenTestDB("UpdateBattleVoteDownImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	battleDown := openapi.AddTopic200ResponseNodeData{
		Topic:        topics[0],
		Id:           nodesAndEdges[0].SourceId,
		BattleTested: -1,
	}

	err = updateNodeBattleVote(db, battleDown, users[0])
	require.Nil(t, err)

	downNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, downNode.BattleTested, int32(-1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.BattleTestedDown), 1)

	err = updateNodeBattleVote(db, battleDown, users[0])
	require.Nil(t, err)

	downNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, downNode.BattleTested, int32(0))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(DownUser.BattleTestedDown))
}

func TestUpdateBattleVoteUpDownImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateBattleVoteUpDownImpl")
	t.Log("INFO TestUpdateBattleVoteUpDownImpl")
	db, dbTearDown := OpenTestDB("UpdateBattleVoteUpDownImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	battleUp := openapi.AddTopic200ResponseNodeData{
		Topic:        topics[0],
		Id:           nodesAndEdges[0].SourceId,
		BattleTested: 1,
	}

	err = updateNodeBattleVote(db, battleUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.BattleTested, int32(1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.BattleTestedUp), 1)

	battleDown := openapi.AddTopic200ResponseNodeData{
		Topic:        topics[0],
		Id:           nodesAndEdges[0].SourceId,
		BattleTested: -1,
	}

	err = updateNodeBattleVote(db, battleDown, users[0])
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.BattleTested, int32(-1))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.BattleTestedDown), 1)
	require.Zero(t, len(DownUser.BattleTestedUp))
}

func TestUpdateBattleVoteDownUpImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateBattleVoteDownUpImpl")
	t.Log("INFO TestUpdateBattleVoteDownUpImpl")
	db, dbTearDown := OpenTestDB("UpdateBattleVoteDownUpImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	battleUp := openapi.AddTopic200ResponseNodeData{
		Topic:        topics[0],
		Id:           nodesAndEdges[0].SourceId,
		BattleTested: -1,
	}

	err = updateNodeBattleVote(db, battleUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.BattleTested, int32(-1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.BattleTestedDown), 1)

	battleDown := openapi.AddTopic200ResponseNodeData{
		Topic:        topics[0],
		Id:           nodesAndEdges[0].SourceId,
		BattleTested: 1,
	}

	err = updateNodeBattleVote(db, battleDown, users[0])
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.BattleTested, int32(1))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(DownUser.BattleTestedDown))
	require.Equal(t, len(DownUser.BattleTestedUp), 1)
}

func TestUpdateFreshVoteUpImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateFreshVoteUpImpl")
	t.Log("INFO TestUpdateFreshVoteUpImpl")
	db, dbTearDown := OpenTestDB("UpdateFreshVoteUPImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	freshUp := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: 1,
	}

	err = updateNodeFreshVote(db, freshUp, users[0]) // should cause +1
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.FreshUp), 1)

	err = updateNodeFreshVote(db, freshUp, users[0]) // should cause -1
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Zero(t, upNode.Fresh)

	upUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(upUser.FreshUp))
}

func TestUpdateFreshVoteDownImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateFreshVoteDownImpl")
	t.Log("INFO TestUpdateFreshVoteDownImpl")
	db, dbTearDown := OpenTestDB("UpdateFreshVoteDownImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	freshDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: -1,
	}

	err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	downNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, downNode.Fresh, int32(-1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshDown), 1)

	err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	downNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, downNode.Fresh, int32(0))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(DownUser.FreshDown))
}

func TestUpdateFreshVoteUpDownImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateFreshVoteUpDownImpl")
	t.Log("INFO TestUpdateFreshVoteUpDownImpl")
	db, dbTearDown := OpenTestDB("UpdateFreshVoteUpDownImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	freshUp := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: 1,
	}

	err = updateNodeFreshVote(db, freshUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshUp), 1)

	freshDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: -1,
	}

	err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(-1))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshDown), 1)
	require.Zero(t, len(DownUser.FreshUp))
}

func TestUpdateFreshVoteDownUpImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateFreshVoteDownUpImpl")
	t.Log("INFO TestUpdateFreshVoteDownUpImpl")
	db, dbTearDown := OpenTestDB("UpdateFreshVoteDownUpImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	freshUp := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: -1,
	}

	err = updateNodeFreshVote(db, freshUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(-1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshDown), 1)

	freshDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: 1,
	}

	err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(1))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(DownUser.FreshDown))
	require.Equal(t, len(DownUser.FreshUp), 1)
}

func TestUpdateNodeVideoVoteUpImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideoVoteUpImpl")
	t.Log("INFO TestUpdateNodeVideoVoteUpImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoVoteUpImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	vidUp := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "www.youtube.com",
			Votes: 1,
		}},
	}

	err = updateNodeVideoEdit(db, &clock, vidUp, users[0])
	require.Nil(t, err)

	err = updateNodeVideoVote(db, vidUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.YoutubeLinks[0].Votes, int32(1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.VideoUp), 1)

	err = updateNodeVideoVote(db, vidUp, users[0]) // should cause -1
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Zero(t, upNode.YoutubeLinks[0].Votes)

	upUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(upUser.VideoUp))
}

// above works, now copy to the next 3
// needs to be rewritten for node video
func TestUpdateNodeVideoVoteDownImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideohVoteDownImpl")
	t.Log("INFO TestUpdateNodeVideoVoteDownImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoVoteDownImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	freshDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: -1,
	}

	err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	downNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, downNode.Fresh, int32(-1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshDown), 1)

	err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	downNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, downNode.Fresh, int32(0))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(DownUser.FreshDown))
}

// needs to be rewritten for node video
func TestUpdateNodeVideoVoteUpDownImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideoVoteUpDownImpl")
	t.Log("INFO TestUpdateNodeVideoVoteUpDownImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoVoteUpDownImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	freshUp := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: 1,
	}

	err = updateNodeFreshVote(db, freshUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshUp), 1)

	freshDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: -1,
	}

	err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(-1))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshDown), 1)
	require.Zero(t, len(DownUser.FreshUp))
}

// needs to be rewritten for node video
func TestUpdateNodeVideoVoteDownUpImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideoVoteDownUpImpl")
	t.Log("INFO TestUpdateNodeVideoVoteDownUpImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoVoteDownUpImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	freshUp := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: -1,
	}

	err = updateNodeFreshVote(db, freshUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(-1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshDown), 1)

	freshDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		Fresh: 1,
	}

	err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(1))

	DownUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(DownUser.FreshDown))
	require.Equal(t, len(DownUser.FreshUp), 1)
}

//need series of test to add and subtract video link to node
//need to test node flag
