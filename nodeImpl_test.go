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
		CreatedBy: openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
			Id: users[0],
		},
		Title: "turbo",
		Topic: topics[0],
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
		Topic: topics[0],
		Title: "tester",
		CreatedBy: openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
			Id: users[0],
		},
		Id: nodesAndEdges[0].SourceId,
	}

	for i := 0; i < 5; i++ {
		clock.Tick()
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

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	originalNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	modNode := originalNode

	modNode.Title = "Jack"
	modNode.Description = "turbo"

	user, err := getUser(db, users[0])
	require.Nil(t, err)

	_, err = updateNodeTitle(db, modNode, user)
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

	_, err = updateNodeBattleVote(db, battleUp, users[0]) // should cause +1
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.BattleTested, int32(1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.BattleTestedUp), 1)

	_, err = updateNodeBattleVote(db, battleUp, users[0]) // should cause -1
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

	_, err = updateNodeBattleVote(db, battleDown, users[0])
	require.Nil(t, err)

	downNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, downNode.BattleTested, int32(-1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.BattleTestedDown), 1)

	_, err = updateNodeBattleVote(db, battleDown, users[0])
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

	_, err = updateNodeBattleVote(db, battleUp, users[0])
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

	_, err = updateNodeBattleVote(db, battleDown, users[0])
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

	_, err = updateNodeBattleVote(db, battleUp, users[0])
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

	_, err = updateNodeBattleVote(db, battleDown, users[0])
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

	_, err = updateNodeFreshVote(db, freshUp, users[0]) // should cause +1
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.Fresh, int32(1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.FreshUp), 1)

	_, err = updateNodeFreshVote(db, freshUp, users[0]) // should cause -1
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

	_, err = updateNodeFreshVote(db, freshDown, users[0])
	require.Nil(t, err)

	downNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, downNode.Fresh, int32(-1))

	DownUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(DownUser.FreshDown), 1)

	_, err = updateNodeFreshVote(db, freshDown, users[0])
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

	_, err = updateNodeFreshVote(db, freshUp, users[0])
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

	_, err = updateNodeFreshVote(db, freshDown, users[0])
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

	_, err = updateNodeFreshVote(db, freshUp, users[0])
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

	_, err = updateNodeFreshVote(db, freshDown, users[0])
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

	user, err := getUser(db, users[0])
	require.Nil(t, err)

	err = updateNodeVideoEdit(db, &clock, vidUp, user)
	require.Nil(t, err)

	_, err = updateNodeVideoVote(db, vidUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.YoutubeLinks[0].Votes, int32(1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.VideoUp), 1)

	_, err = updateNodeVideoVote(db, vidUp, users[0]) // should cause -1
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Zero(t, upNode.YoutubeLinks[0].Votes)

	upUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(upUser.VideoUp))
}

func TestUpdateNodeVideoVoteDownImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideohVoteDownImpl")
	t.Log("INFO TestUpdateNodeVideoVoteDownImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoVoteDownImpl")
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

	user, err := getUser(db, users[0])
	require.Nil(t, err)

	err = updateNodeVideoEdit(db, &clock, vidUp, user)
	require.Nil(t, err)

	vidDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "www.youtube.com",
			Votes: -1,
		}},
	}

	_, err = updateNodeVideoVote(db, vidDown, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.YoutubeLinks[0].Votes, int32(-1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.VideoDown), 1)

	_, err = updateNodeVideoVote(db, vidDown, users[0]) // should cause -1
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Zero(t, upNode.YoutubeLinks[0].Votes)

	upUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(upUser.VideoDown))
}

func TestUpdateNodeVideoVoteUpDownImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideoVoteUpDownImpl")
	t.Log("INFO TestUpdateNodeVideoVoteUpDownImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoVoteUpDownImpl")
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

	user, err := getUser(db, users[0])
	require.Nil(t, err)

	err = updateNodeVideoEdit(db, &clock, vidUp, user)
	require.Nil(t, err)

	vidDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "www.youtube.com",
			Votes: -1,
		}},
	}

	_, err = updateNodeVideoVote(db, vidUp, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.YoutubeLinks[0].Votes, int32(1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.VideoUp), 1)

	_, err = updateNodeVideoVote(db, vidDown, users[0]) // should cause -1
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.YoutubeLinks[0].Votes, int32(-1))

	upUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(upUser.VideoUp))
	require.Equal(t, len(upUser.VideoDown), 1)
}

func TestUpdateNodeVideoVoteDownUpImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideoVoteDownUpImpl")
	t.Log("INFO TestUpdateNodeVideoVoteDownUpImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoVoteDownUpImpl")
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

	user, err := getUser(db, users[0])
	require.Nil(t, err)

	err = updateNodeVideoEdit(db, &clock, vidUp, user)
	require.Nil(t, err)

	vidDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "www.youtube.com",
			Votes: -1,
		}},
	}

	_, err = updateNodeVideoVote(db, vidDown, users[0])
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.YoutubeLinks[0].Votes, int32(-1))

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	require.Equal(t, len(upUser.VideoDown), 1)

	_, err = updateNodeVideoVote(db, vidUp, users[0]) // should cause -1
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.YoutubeLinks[0].Votes, int32(1))

	upUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(upUser.VideoDown))
	require.Equal(t, len(upUser.VideoUp), 1)
}

func TestUpdateNodeFlagImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeFlagImpl")
	t.Log("INFO TestUpdateNodeFlagImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeFlagImpl")
	defer dbTearDown()

	_, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	vidUp := openapi.AddTopic200ResponseNodeData{
		Topic:     topics[0],
		Id:        nodesAndEdges[0].SourceId,
		IsFlagged: true,
	}

	err = updateNodeFlag(db, vidUp)
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.True(t, upNode.IsFlagged)

	err = updateNodeFlag(db, vidUp)
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.False(t, upNode.IsFlagged)
}

func TestUpdateNodeVideoEditImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideoEditImpl")
	t.Log("INFO TestUpdateNodeVideoEditImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoEditImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	vidAdd := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "www.youtube.com",
			Votes: 1,
		}},
	}

	user, err := getUser(db, users[0])
	require.Nil(t, err)

	err = updateNodeVideoEdit(db, &clock, vidAdd, user)
	require.Nil(t, err)

	upNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Equal(t, upNode.YoutubeLinks[0].Link, vidAdd.YoutubeLinks[0].Link)

	upUser, err := getUser(db, users[0])
	require.Nil(t, err)

	err = updateNodeVideoEdit(db, &clock, vidAdd, upUser)
	require.NotNil(t, err)

	require.Equal(t, upUser.Linked[0].Link, vidAdd.YoutubeLinks[0].Link)

	vidSub := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "www.youtube.com",
			Votes: -1,
		}},
	}

	err = updateNodeVideoEdit(db, &clock, vidSub, upUser)
	require.Nil(t, err)

	upNode, err = getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)

	require.Zero(t, len(upNode.YoutubeLinks))

	upUser, err = getUser(db, users[0])
	require.Nil(t, err)

	require.Zero(t, len(upUser.VideoDown))

	err = updateNodeVideoEdit(db, &clock, vidSub, upUser)
	require.NotNil(t, err)

}

func TestVideoDeletedFromAnotherUser(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestVideoDeletedFromAnotherUser")
	t.Log("INFO TestVideoDeletedFromAnotherUser")

	db, dbTearDown := OpenTestDB("VideoDeletedFromAnotherUser")
	defer dbTearDown()

	// Create 2 users, 1 topic, 1 node
	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 2, 1, 1)
	require.Nil(t, err)

	userAId := users[0]
	userBId := users[1]
	topic := topics[0]
	nodeId := nodesAndEdges[0].SourceId

	// Get userB (to simulate video add)
	userB, err := getUser(db, userBId)
	require.Nil(t, err)

	// User B adds a video to the node
	addVideo := openapi.AddTopic200ResponseNodeData{
		Topic: topic,
		Id:    nodeId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "https://www.youtube.com/watch?v=abc123",
			Votes: 1,
		}},
	}

	err = updateNodeVideoEdit(db, &clock, addVideo, userB)
	require.Nil(t, err)

	// Verify video added
	nodeAfterAdd, err := getNode(db, nodeId.Format(time.RFC3339Nano), topic)
	require.Nil(t, err)
	require.Equal(t, 1, len(nodeAfterAdd.YoutubeLinks))
	require.Equal(t, "https://www.youtube.com/watch?v=abc123", nodeAfterAdd.YoutubeLinks[0].Link)

	// User A upvotes the video

	upvoteVideo := openapi.AddTopic200ResponseNodeData{
		Topic: topic,
		Id:    nodeId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "https://www.youtube.com/watch?v=abc123",
			Votes: 1,
		}},
	}

	_, err = updateNodeVideoVote(db, upvoteVideo, userAId)
	require.Nil(t, err)

	// User A deletes the video
	deleteVideo := openapi.AddTopic200ResponseNodeData{
		Topic: topic,
		Id:    nodeId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "https://www.youtube.com/watch?v=abc123",
			Votes: -1,
		}},
	}

	userA, err := getUser(db, userAId)
	require.Nil(t, err)

	err = updateNodeVideoEdit(db, &clock, deleteVideo, userA)
	require.Nil(t, err)

	// Verify video is removed from node
	nodeAfterDelete, err := getNode(db, nodeId.Format(time.RFC3339Nano), topic)
	require.Nil(t, err)
	require.Zero(t, len(nodeAfterDelete.YoutubeLinks))

	// Optional: check userA doesn't have it in downvoted list
	userAAfter, err := getUser(db, userAId)
	require.Nil(t, err)
	require.Zero(t, len(userAAfter.VideoDown))
}

func TestUpdateNodeVideoVoteReputationImpl(t *testing.T) {
	clock := TestClock{}
	lgr.Printf("INFO TestUpdateNodeVideoVoteReputationImpl")
	t.Log("INFO TestUpdateNodeVideoVoteReputationImpl")
	db, dbTearDown := OpenTestDB("UpdateNodeVideoVoteReputationImpl")
	defer dbTearDown()

	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 2, 1, 1)
	require.Nil(t, err)

	// First user adds a video
	vidUp := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "www.youtube.com/test-reputation",
			Votes: 1,
		}},
	}

	user1, err := getUser(db, users[0])
	require.Nil(t, err)

	// Get initial reputation
	initialReputation := user1.Reputation

	// User1 adds a video
	err = updateNodeVideoEdit(db, &clock, vidUp, user1)
	require.Nil(t, err)

	// User2 upvotes the video
	_, err = updateNodeVideoVote(db, vidUp, users[1])
	require.Nil(t, err)

	// Check that user1's reputation increased
	updatedUser1, err := getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, initialReputation+1, updatedUser1.Reputation, "Reputation should increase by 1 after upvote")

	// User2 removes upvote
	_, err = updateNodeVideoVote(db, vidUp, users[1])
	require.Nil(t, err)

	// Check that user1's reputation decreased back
	updatedUser1Again, err := getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, initialReputation, updatedUser1Again.Reputation, "Reputation should decrease back after removing upvote")

	// User2 downvotes the video
	vidDown := openapi.AddTopic200ResponseNodeData{
		Topic: topics[0],
		Id:    nodesAndEdges[0].SourceId,
		YoutubeLinks: []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{{
			Link:  "www.youtube.com/test-reputation",
			Votes: -1,
		}},
	}
	_, err = updateNodeVideoVote(db, vidDown, users[1])
	require.Nil(t, err)

	// Check that user1's reputation decreased
	updatedUser1AfterDownvote, err := getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, initialReputation-1, updatedUser1AfterDownvote.Reputation, "Reputation should decrease by 1 after downvote")

	// User2 removes downvote
	_, err = updateNodeVideoVote(db, vidDown, users[1])
	require.Nil(t, err)

	// Check that user1's reputation inreased
	updatedUser1AfterDownvote, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, initialReputation, updatedUser1AfterDownvote.Reputation, "Reputation should decrease by 1 after downvote")

	// User2 upvotes the video
	_, err = updateNodeVideoVote(db, vidUp, users[1])
	require.Nil(t, err)

	// Check that user1's reputation increased
	updatedUser1, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, initialReputation+1, updatedUser1.Reputation, "Reputation should increase by 1 after upvote")

	// User2 downvotes the video
	_, err = updateNodeVideoVote(db, vidDown, users[1])
	require.Nil(t, err)

	// Check that user1's reputation is now -1 from initial (after switching from upvote to downvote)
	// The change is -2 from the upvoted state, but compared to initial it's -1
	updatedUser1AfterDownvote, err = getUser(db, users[0])
	require.Nil(t, err)
	require.Equal(t, initialReputation-1, updatedUser1AfterDownvote.Reputation, "Reputation should be initialReputation-1 after switching from upvote to downvote")
}

func TestUserNodeEdited(t *testing.T) {
	lgr.Printf("INFO TestUserNodeEdited")
	t.Log("INFO TestUserNodeEdited")
	db, dbTearDown := OpenTestDB("UserNodeEdited")
	defer dbTearDown()
	clock := TestClock{}

	// Create test data
	users, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 1)
	require.Nil(t, err)

	// Get the original node
	originalNode, err := getNode(db, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), topics[0])
	require.Nil(t, err)
	require.NotNil(t, originalNode.Id, "Original node ID should not be nil")

	// Verify the node ID is set correctly
	require.Equal(t, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), originalNode.Id.Format(time.RFC3339Nano),
		"Node ID should match the source ID")

	// Get the user
	user, err := getUser(db, users[0])
	require.Nil(t, err)

	// Count initial edited nodes
	initialEditedCount := len(user.Edited)

	// Call updateUserNodeEdited instead of directly using a transaction
	err = updateUserNodeEdited(db, users[0], originalNode)
	require.Nil(t, err)

	// Get the updated user
	updatedUser, err := getUser(db, users[0])
	require.Nil(t, err)

	// Verify the edited node was added
	require.Equal(t, initialEditedCount+1, len(updatedUser.Edited),
		"User should have one more edited node")

	// Find the edited node
	var editedNode *openapi.UpdateUserRequestBattleTestedUpInner
	for i := range updatedUser.Edited {
		if updatedUser.Edited[i].Topic == topics[0] &&
			updatedUser.Edited[i].Title == originalNode.Title {
			editedNode = &updatedUser.Edited[i]
			break
		}
	}

	require.NotNil(t, editedNode, "Edited node should be found in user's edited list")

	// Verify the node ID is set correctly in the user's edited list
	require.False(t, editedNode.NodeId.IsZero(), "Edited node ID should not be zero time")
	require.Equal(t, originalNode.Id.Format(time.RFC3339Nano), editedNode.NodeId.Format(time.RFC3339Nano),
		"Edited node ID should match the original node ID")

	// Test editing the same node again (should not add duplicate)
	err = updateUserNodeEdited(db, users[0], originalNode)
	require.Nil(t, err)

	// Get the user again
	updatedUserAgain, err := getUser(db, users[0])
	require.Nil(t, err)

	// Verify no duplicate was added
	require.Equal(t, len(updatedUser.Edited), len(updatedUserAgain.Edited),
		"No duplicate edited node should be added")
}
