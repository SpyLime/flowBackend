package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

func getNode(db *bolt.DB, nodeId, topicId string) (response openapi.AddTopic200ResponseNodeData, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		response, err = getNodeRx(tx, nodeId, topicId)
		return err
	})

	return
}

func getNodeRx(tx *bolt.Tx, nodeId, topicId string) (response openapi.AddTopic200ResponseNodeData, err error) {
	_, nodeData, err := nodeDataFinderTx(tx, topicId, nodeId)
	if err != nil {
		return
	}

	err = json.Unmarshal(nodeData, &response)
	if err != nil {
		return
	}

	newTime, err := time.Parse(time.RFC3339, nodeId)
	if err != nil {
		return
	}

	id := newTime

	response.Id = id

	return
}

// do not supply a target id
//
// supply topic, title, createdBy, ID which is the source ID
func postNode(db *bolt.DB, clock Clock, node openapi.AddTopic200ResponseNodeData) (response openapi.ResponsePostNode, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		response, err = postNodeTx(tx, clock, node)
		return err
	})

	return
}

// do not supply a target id
//
// supply topic, title, createdBy, ID which is the source ID
func postNodeTx(tx *bolt.Tx, clock Clock, node openapi.AddTopic200ResponseNodeData) (response openapi.ResponsePostNode, err error) {

	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return response, fmt.Errorf("can't find topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(node.Topic))
	if topicBucket == nil {
		return response, fmt.Errorf("can't find topic bucket")
	}

	nodesBucket := topicBucket.Bucket([]byte(KeyNodes))
	if nodesBucket == nil {
		return response, fmt.Errorf("can't find nodes bucket")
	}

	newNode := openapi.NodeData{
		Topic:     node.Topic,
		Title:     node.Title,
		CreatedBy: node.CreatedBy,
	}

	if newNode.CreatedBy == "" {
		newNode.CreatedBy = "change hardcode"
	}

	marshal, err := json.Marshal(newNode)
	if err != nil {
		return
	}

	id := clock.Now()
	newNode.Id = id

	idB, err := id.MarshalText()
	if err != nil {
		return
	}

	err = nodesBucket.Put(idB, marshal)
	if err != nil {
		return
	}

	response.SourceId = node.Id
	response.TargetId = id

	edge := openapi.GetMapById200ResponseEdgesInner{
		Id:     response.SourceId.Format(time.RFC3339Nano) + "-" + response.TargetId.Format(time.RFC3339Nano),
		Source: response.SourceId,
		Target: response.TargetId,
	}
	_, err = postEdgeTx(topicBucket, edge)
	if err != nil {
		return
	}

	return
}

func deleteNode(db *bolt.DB, nodeId, topicId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = deleteNodeTx(tx, nodeId, topicId)
		return err
	})

	return
}

func deleteNodeTx(tx *bolt.Tx, nodeId, topicId string) (err error) {

	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return fmt.Errorf("can't find topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(topicId))
	if topicBucket == nil {
		return fmt.Errorf("can't find topic bucket")
	}

	nodesBucket := topicBucket.Bucket([]byte(KeyNodes))
	if nodesBucket == nil {
		return fmt.Errorf("can't find nodes bucket")
	}

	//this might not be good enough
	//I might need to also search for orphaned nodes
	//that is nodes that after deletion of this node are left with no source or target
	//also what happens when the chain is broken
	//should I link the 2 nodes that were connected to the deleted node
	//probably not that simple as those 2 nodes may have multiple links and it may no longer be logical to link them without
	//the hub node that was deleted
	err = nodesBucket.Delete([]byte(nodeId))
	if err != nil {
		return err
	}

	edgesBucket := topicBucket.Bucket([]byte(KeyEdges))
	c := edgesBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		if strings.Contains(string(k), nodeId) {
			err = edgesBucket.Delete(k)
			if err != nil {
				return err
			}
		}
	}

	return

}

func updateNodeTitle(db *bolt.DB, request openapi.AddTopic200ResponseNodeData) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeTitleTx(tx, request)
		return err
	})

	return
}

// updates the title and description
func updateNodeTitleTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData) (err error) {

	nodesBucket, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
	if err != nil {
		return
	}

	var node openapi.NodeData

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return
	}

	if request.Title != "" && request.Title != node.Title {
		node.Title = request.Title
	}
	if request.Description != "" && request.Title != node.Description {
		node.Description = request.Description
	}

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)

	return
}

func nodeDataFinderTx(tx *bolt.Tx, TopicId, NodeId string) (nodesBucket *bolt.Bucket, nodeData []byte, err error) {
	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return nil, nil, fmt.Errorf("can't finb topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(TopicId))
	if topicBucket == nil {
		return nil, nil, fmt.Errorf("can't finb topic bucket")
	}

	nodesBucket = topicBucket.Bucket([]byte(KeyNodes))
	if nodesBucket == nil {
		return nil, nil, fmt.Errorf("can't find nodes bucket")
	}

	nodeData = nodesBucket.Get([]byte(NodeId))
	if nodeData == nil {
		return nil, nil, fmt.Errorf("can't find node data")
	}

	return
}

func updateNodeBattleVote(db *bolt.DB, request openapi.AddTopic200ResponseNodeData, userId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeBattleVoteTx(tx, request, userId)
		return err
	})

	return
}

func updateNodeBattleVoteTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData, userId string) (err error) {

	nodesBucket, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
	if err != nil {
		return
	}

	var node openapi.NodeData

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return
	}

	if request.BattleTested != 0 {
		vote, err := userBattleVoteTx(tx, userId, request)
		if err != nil {
			return err
		}

		node.BattleTested += vote //vote will either be a -2,-1,1,2
	}

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)

	return
}

func userBattleVoteTx(tx *bolt.Tx, userId string, request openapi.AddTopic200ResponseNodeData) (vote int32, err error) {

	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	if request.BattleTested > 0 {
		for i, item := range user.BattleTestedUp {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.BattleTestedUp = append(user.BattleTestedUp[:i], user.BattleTestedUp[i+1:]...) //already voted up so subtract 1 to unvote
				vote--
				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}

				err = usersBucket.Put([]byte(userId), marshal)

				return vote, err
			}
		}

		user.BattleTestedUp = append(user.BattleTestedUp, openapi.UpdateUserRequestBattleTestedUpInner{
			Topic:  request.Topic,
			Title:  request.Title,
			NodeId: request.Id,
		})

		vote++

		for i, item := range user.BattleTestedDown {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.BattleTestedDown = append(user.BattleTestedDown[:i], user.BattleTestedDown[i+1:]...) //already voted down so add 1 to unvote
				vote++
				break
			}
		}

	} else {
		for i, item := range user.BattleTestedDown {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.BattleTestedDown = append(user.BattleTestedDown[:i], user.BattleTestedDown[i+1:]...)
				vote++
				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}

				err = usersBucket.Put([]byte(userId), marshal)

				return vote, err
			}
		}

		user.BattleTestedDown = append(user.BattleTestedDown, openapi.UpdateUserRequestBattleTestedUpInner{
			Topic:  request.Topic,
			Title:  request.Title,
			NodeId: request.Id,
		})

		vote--

		for i, item := range user.BattleTestedDown {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.BattleTestedDown = append(user.BattleTestedDown[:i], user.BattleTestedDown[i+1:]...) //already voted down so add 1 to unvote
				vote--
				break
			}
		}

	}

	marshal, err := json.Marshal(user)
	if err != nil {
		return vote, err
	}

	err = usersBucket.Put([]byte(userId), marshal)

	return
}

func updateNodeVideoEdit(db *bolt.DB, clock Clock, request openapi.AddTopic200ResponseNodeData, userId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeVideoEditTx(tx, clock, request, userId)
		return err
	})

	return
}

func updateNodeVideoEditTx(tx *bolt.Tx, clock Clock, request openapi.AddTopic200ResponseNodeData, userId string) (err error) {

	nodesBucket, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
	if err != nil {
		return
	}

	var node openapi.NodeData

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return
	}

	for i, item := range node.YoutubeLinks {
		if item.Link == request.YoutubeLinks[0].Link { // Check if ID matches
			if request.YoutubeLinks[0].Votes > 0 {
				return fmt.Errorf("this video is already added")
			} else {
				node.YoutubeLinks = append(node.YoutubeLinks[:i], node.YoutubeLinks[i+1:]...)
			}

			marshal, err := json.Marshal(node)
			if err != nil {
				return err
			}

			err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)

			//userupdate here

			return err
		}
	}

	if request.YoutubeLinks[0].Votes > 0 {
		node.YoutubeLinks = append(node.YoutubeLinks, openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			Link:      request.YoutubeLinks[0].Link,
			Votes:     0,
			AddedBy:   userId,
			DateAdded: clock.Now(),
		})
	} else {
		return fmt.Errorf("could not find that video to delete")
	}

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
	if err != nil {
		return
	}

	err = userVideoEditTx(tx, clock, userId, request)

	return
}

func userVideoEditTx(tx *bolt.Tx, clock Clock, userId string, request openapi.AddTopic200ResponseNodeData) (err error) {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	for i, item := range user.Linked {
		if item.Link == request.YoutubeLinks[0].Link {
			if request.YoutubeLinks[0].Votes > 0 {
				return fmt.Errorf("this video is already added")
			} else {
				user.Linked = append(user.Linked[:i], user.Linked[i+1:]...)
			}

			marshal, err := json.Marshal(user)
			if err != nil {
				return err
			}

			err = usersBucket.Put([]byte(userId), marshal)

			return err
		}
	}

	if request.YoutubeLinks[0].Votes > 0 {
		user.Linked = append(user.Linked, openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			Link:      request.YoutubeLinks[0].Link,
			Votes:     0,
			AddedBy:   userId,
			DateAdded: clock.Now(),
		})

		marshal, err := json.Marshal(user)
		if err != nil {
			return err
		}

		err = usersBucket.Put([]byte(userId), marshal)

		return err

	} else {
		return fmt.Errorf("could not find the video to remove on the user")
	}

}

func updateNodeVideoVote(db *bolt.DB, request openapi.AddTopic200ResponseNodeData, userId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeVideoVoteTx(tx, request, userId)
		return err
	})

	return
}

func updateNodeVideoVoteTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData, userId string) (err error) {

	nodesBucket, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
	if err != nil {
		return
	}

	var node openapi.NodeData

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return
	}

	if request.YoutubeLinks != nil && request.YoutubeLinks[0].Votes != 0 {
		vote, err := userVideoVoteTx(tx, userId, request)
		if err != nil {
			return err
		}

		found := false
		for i, item := range node.YoutubeLinks {
			if item.Link == request.YoutubeLinks[0].Link {
				node.YoutubeLinks[i].Votes += vote
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("could not find the video on node")
		}

	}

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)

	return
}

func userVideoVoteTx(tx *bolt.Tx, userId string, request openapi.AddTopic200ResponseNodeData) (vote int32, err error) {

	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	if request.YoutubeLinks[0].Votes > 0 {
		for i, item := range user.VideoUp {
			if item == request.YoutubeLinks[0].Link {
				user.VideoUp = append(user.VideoUp[:i], user.VideoUp[i+1:]...)
				vote--
				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}

				err = usersBucket.Put([]byte(userId), marshal)

				return vote, err
			}
		}

		user.VideoUp = append(user.VideoUp, request.YoutubeLinks[0].Link)

		vote++

		for i, item := range user.VideoDown {
			if item == request.YoutubeLinks[0].Link {
				user.VideoDown = append(user.VideoDown[:i], user.VideoDown[i+1:]...)
				vote++
				break
			}
		}

	} else {
		for i, item := range user.VideoDown {
			if item == request.YoutubeLinks[0].Link {
				user.VideoDown = append(user.VideoDown[:i], user.VideoDown[i+1:]...)
				vote++
				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}

				err = usersBucket.Put([]byte(userId), marshal)

				return vote, err
			}
		}

		user.VideoDown = append(user.VideoDown, request.YoutubeLinks[0].Link)

		vote--

		for i, item := range user.VideoUp {
			if item == request.YoutubeLinks[0].Link {
				user.VideoUp = append(user.VideoUp[:i], user.VideoUp[i+1:]...)
				vote--
				break
			}
		}

	}

	marshal, err := json.Marshal(user)
	if err != nil {
		return vote, err
	}

	err = usersBucket.Put([]byte(userId), marshal)

	return
}

func updateNodeFlag(db *bolt.DB, request openapi.AddTopic200ResponseNodeData) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeFlagTx(tx, request)
		return err
	})

	return
}

// updates the title and description
func updateNodeFlagTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData) (err error) {

	nodesBucket, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
	if err != nil {
		return
	}

	var node openapi.NodeData

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return
	}

	node.IsFlagged = !node.IsFlagged

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)

	return
}

func updateNodeFreshVote(db *bolt.DB, request openapi.AddTopic200ResponseNodeData, userId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeFreshVoteTx(tx, request, userId)
		return err
	})

	return
}

// updates the title and description
func updateNodeFreshVoteTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData, userId string) (err error) {

	nodesBucket, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
	if err != nil {
		return
	}

	var node openapi.NodeData

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return
	}

	if request.Fresh != 0 {
		vote, err := userFreshVoteTx(tx, userId, request)
		if err != nil {
			return err
		}

		node.Fresh += vote //vote will either be a +1 or -1
	}

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)

	return
}

func userFreshVoteTx(tx *bolt.Tx, userId string, request openapi.AddTopic200ResponseNodeData) (vote int32, err error) {

	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	if request.Fresh > 0 {
		for i, item := range user.FreshUp {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.FreshUp = append(user.FreshUp[:i], user.FreshUp[i+1:]...) //already voted up so subtract 1 to unvote
				vote--
				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}

				err = usersBucket.Put([]byte(userId), marshal)

				return vote, err
			}
		}

		user.FreshUp = append(user.FreshUp, openapi.UpdateUserRequestBattleTestedUpInner{
			Topic:  request.Topic,
			Title:  request.Title,
			NodeId: request.Id,
		})

		vote++

		for i, item := range user.FreshDown {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.FreshDown = append(user.FreshDown[:i], user.FreshDown[i+1:]...) //already voted down so add 1 to unvote
				vote++
				break
			}
		}

	} else {
		for i, item := range user.FreshDown {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.FreshDown = append(user.FreshDown[:i], user.FreshDown[i+1:]...)
				vote++
				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}

				err = usersBucket.Put([]byte(userId), marshal)

				return vote, err
			}
		}

		user.FreshDown = append(user.FreshDown, openapi.UpdateUserRequestBattleTestedUpInner{
			Topic:  request.Topic,
			Title:  request.Title,
			NodeId: request.Id,
		})

		vote--

		for i, item := range user.FreshUp {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.FreshUp = append(user.FreshUp[:i], user.FreshUp[i+1:]...) //already voted down so add 1 to unvote
				vote--
				break
			}
		}

	}

	marshal, err := json.Marshal(user)
	if err != nil {
		return vote, err
	}

	err = usersBucket.Put([]byte(userId), marshal)

	return
}
