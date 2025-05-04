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

	err = userNodeCreatedTx(tx, node.CreatedBy.Id, newNode)
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

// Helper function to save the created node to the user
func userNodeCreatedTx(tx *bolt.Tx, userId string, node openapi.NodeData) error {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return err
	}

	// Check for duplicates before appending
	isDuplicate := false
	for _, created := range user.Created {
		if created.NodeId == node.Id {
			isDuplicate = true
			break
		}
	}

	if !isDuplicate {
		user.Created = append(user.Created, openapi.UpdateUserRequestBattleTestedUpInner{
			Topic:  node.Topic,
			Title:  node.Title,
			NodeId: node.Id,
		})

		marshal, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return usersBucket.Put([]byte(userId), marshal)
	}

	return nil
}

func deleteNode(db *bolt.DB, nodeId, topicId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = deleteNodeTx(tx, nodeId, topicId)
		return err
	})

	return
}

// Helper function to remove the deleted node from all users who interacted with it
func removeNodeFromAllUsersTx(tx *bolt.Tx, nodeId string, topicId string) error {
	// First get the node to find all users who interacted with it
	_, nodeData, err := nodeDataFinderTx(tx, topicId, nodeId)
	if err != nil {
		return err
	}

	var node openapi.NodeData
	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return err
	}

	// Collect all unique user IDs who interacted with this node
	userIds := make(map[string]bool)

	// Add creator
	userIds[node.CreatedBy.Id] = true

	// Add editors
	for _, editor := range node.EditedBy {
		userIds[editor.Id] = true
	}

	// We need to scan all users to find those who voted on this node
	usersBucket := tx.Bucket([]byte(KeyUsers))
	if usersBucket == nil {
		return fmt.Errorf("can't find users bucket")
	}

	c := usersBucket.Cursor()
	for k, v := c.First(); k != nil && v != nil; k, v = c.Next() {
		var user openapi.UpdateUserRequest
		if err := json.Unmarshal(v, &user); err != nil {
			continue // Skip this user if we can't unmarshal
		}

		// Check if user interacted with this node in any way
		nodeIdStr := nodeId

		// Check battle votes
		for _, item := range user.BattleTestedUp {
			if item.NodeId.Format(time.RFC3339Nano) == nodeIdStr {
				userIds[string(k)] = true
				break
			}
		}

		for _, item := range user.BattleTestedDown {
			if item.NodeId.Format(time.RFC3339Nano) == nodeIdStr {
				userIds[string(k)] = true
				break
			}
		}

		// Check fresh votes
		for _, item := range user.FreshUp {
			if item.NodeId.Format(time.RFC3339Nano) == nodeIdStr {
				userIds[string(k)] = true
				break
			}
		}

		for _, item := range user.FreshDown {
			if item.NodeId.Format(time.RFC3339Nano) == nodeIdStr {
				userIds[string(k)] = true
				break
			}
		}

		// Check video interactions
		// We need to check if any of the node's videos are in the user's video votes
		for _, video := range node.YoutubeLinks {
			for _, userVideo := range user.VideoUp {
				if userVideo == video.Link {
					userIds[string(k)] = true
				}
			}

			for _, userVideo := range user.VideoDown {
				if userVideo == video.Link {
					userIds[string(k)] = true
				}
			}

			// Check linked videos
			for _, linked := range user.Linked {
				if linked.Link == video.Link {
					userIds[string(k)] = true
				}
			}
		}
	}

	// Process each user
	for userId := range userIds {
		err = removeNodeFromUserTx(tx, userId, nodeId, topicId, node.Topic, node.YoutubeLinks)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper function to remove a node from a specific user's data
func removeNodeFromUserTx(tx *bolt.Tx, userId string, nodeId string, topicId string, topic string, videos []openapi.AddTopic200ResponseNodeDataYoutubeLinksInner) error {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return err
	}

	// Remove the node from user's created list
	for i, created := range user.Created {
		if created.NodeId.Format(time.RFC3339Nano) == nodeId {
			user.Created = append(user.Created[:i], user.Created[i+1:]...)
			break
		}
	}

	// Remove from edited list
	for i, edited := range user.Edited {
		if edited.NodeId.Format(time.RFC3339Nano) == nodeId {
			user.Edited = append(user.Edited[:i], user.Edited[i+1:]...)
			break
		}
	}

	// Remove from battle tested votes
	for i, item := range user.BattleTestedUp {
		if item.NodeId.Format(time.RFC3339Nano) == nodeId {
			user.BattleTestedUp = append(user.BattleTestedUp[:i], user.BattleTestedUp[i+1:]...)
			break
		}
	}

	for i, item := range user.BattleTestedDown {
		if item.NodeId.Format(time.RFC3339Nano) == nodeId {
			user.BattleTestedDown = append(user.BattleTestedDown[:i], user.BattleTestedDown[i+1:]...)
			break
		}
	}

	// Remove from fresh votes
	for i, item := range user.FreshUp {
		if item.NodeId.Format(time.RFC3339Nano) == nodeId {
			user.FreshUp = append(user.FreshUp[:i], user.FreshUp[i+1:]...)
			break
		}
	}

	for i, item := range user.FreshDown {
		if item.NodeId.Format(time.RFC3339Nano) == nodeId {
			user.FreshDown = append(user.FreshDown[:i], user.FreshDown[i+1:]...)
			break
		}
	}

	// Remove video interactions
	for _, video := range videos {
		// Remove from video up votes
		for i, link := range user.VideoUp {
			if link == video.Link {
				user.VideoUp = append(user.VideoUp[:i], user.VideoUp[i+1:]...)
				break
			}
		}

		// Remove from video down votes
		for i, link := range user.VideoDown {
			if link == video.Link {
				user.VideoDown = append(user.VideoDown[:i], user.VideoDown[i+1:]...)
				break
			}
		}

		// Remove from linked videos
		for i, linked := range user.Linked {
			if linked.Link == video.Link {
				user.Linked = append(user.Linked[:i], user.Linked[i+1:]...)
				break
			}
		}
	}

	marshal, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return usersBucket.Put([]byte(userId), marshal)
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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

	// Remove the node from all users who interacted with it
	err = removeNodeFromAllUsersTx(tx, nodeId, topicId)
	if err != nil {
		return err
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

func updateNodeTitle(db *bolt.DB, request openapi.AddTopic200ResponseNodeData, editor openapi.User) (editorAdded bool, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		editorAdded, err = updateNodeTitleTx(tx, request, editor)
		return err
	})

	return
}

// updates the title and description
func updateNodeTitleTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData, editor openapi.User) (editorAdded bool, err error) {
	nodesBucket, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
	if err != nil {
		return
	}

	var node openapi.NodeData

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return
	}

	isEdited := false
	if request.Title != "" && request.Title != node.Title {
		node.Title = request.Title
		isEdited = true
	}
	if request.Description != "" && request.Title != node.Description {
		node.Description = request.Description
		isEdited = true
	}

	if !isEdited {
		return
	}

	// Check if the editor already exists in the slice
	editorExists := false

	if node.CreatedBy.Id == editor.Id {
		editorExists = true
	}

	if !editorExists {
		for _, existingEditor := range node.EditedBy {
			if existingEditor.Id == editor.Id {
				editorExists = true
				break
			}
		}
	}

	// Only append if the editor doesn't already exist
	if !editorExists {
		node.EditedBy = append(node.EditedBy, openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
			Id:       editor.Id,
			Username: editor.Username,
		})
		editorAdded = true

		// Update the user's record to indicate they edited this node
		err = userNodeEditedTx(tx, editor.Id, node)
		if err != nil {
			return false, err
		}
	}

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
	if err != nil {
		return
	}

	// Update the title in all user records
	err = updateUserNodeTitleTx(tx, request.Id, request.Topic, request.Title)
	if err != nil {
		return false, err
	}

	return
}

// Helper function to save the edited node to the user's record
func userNodeEditedTx(tx *bolt.Tx, userId string, node openapi.NodeData) error {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return err
	}

	// Check for duplicates before appending to edited list
	isDuplicate := false
	for _, edited := range user.Edited {
		if edited.NodeId.Format(time.RFC3339Nano) == node.Id.Format(time.RFC3339Nano) {
			isDuplicate = true
			// Update the title in case it changed
			edited.Title = node.Title
			break
		}
	}

	if !isDuplicate {
		user.Edited = append(user.Edited, openapi.UpdateUserRequestBattleTestedUpInner{
			Topic:  node.Topic,
			Title:  node.Title,
			NodeId: node.Id,
		})

		marshal, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return usersBucket.Put([]byte(userId), marshal)
	}

	return nil
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

func updateNodeBattleVote(db *bolt.DB, request openapi.AddTopic200ResponseNodeData, userId string) (vote int32, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		vote, err = updateNodeBattleVoteTx(tx, request, userId)
		return err
	})

	return
}

func updateNodeBattleVoteTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData, userId string) (vote int32, err error) {
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
			return vote, err
		}

		node.BattleTested += vote //vote will either be a -2,-1,1,2

		// Update reputation of the node creator if this is a vote (not an unvote)
		if vote != 0 && node.CreatedBy.Id != "" {
			// Only update reputation if the voter is not the creator
			if node.CreatedBy.Id != userId {
				updateCreatorReputation(tx, node.CreatedBy.Id, vote)
			}
		}
	}

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
	vote = node.BattleTested

	return
}

func userBattleVoteTx(tx *bolt.Tx, userId string, request openapi.AddTopic200ResponseNodeData) (vote int32, err error) {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	// Get the node to retrieve its title if not provided
	var nodeTitle string
	if request.Title == "" {
		// Fetch the node to get its title
		_, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
		if err == nil {
			var node openapi.NodeData
			if err = json.Unmarshal(nodeData, &node); err == nil {
				nodeTitle = node.Title
			}
		}
	} else {
		nodeTitle = request.Title
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
			Title:  nodeTitle,
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
			Title:  nodeTitle,
			NodeId: request.Id,
		})

		vote--

		for i, item := range user.BattleTestedUp {
			if item.NodeId.Equal(request.Id) { // Check if ID matches
				user.BattleTestedUp = append(user.BattleTestedUp[:i], user.BattleTestedUp[i+1:]...) //already voted down so add 1 to unvote
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

// want to add a video to a node
//
// if votes are greater than zero then trying to add a video
func updateNodeVideoEdit(db *bolt.DB, clock Clock, request openapi.AddTopic200ResponseNodeData, user openapi.User) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeVideoEditTx(tx, clock, request, user)
		return err
	})

	return
}

func updateNodeVideoEditTx(tx *bolt.Tx, clock Clock, request openapi.AddTopic200ResponseNodeData, user openapi.User) (err error) {

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
				node.YoutubeLinks = append(node.YoutubeLinks[:i], node.YoutubeLinks[i+1:]...) //subtract video because votes are less than zero
			}

			marshal, err := json.Marshal(node)
			if err != nil {
				return err
			}

			err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
			if err != nil {
				return err
			}

			err = userVideoEditTx(tx, clock, user.Id, request)

			return err
		}
	}

	if request.YoutubeLinks[0].Votes > 0 { //video was not found and you want to add
		node.YoutubeLinks = append(node.YoutubeLinks, openapi.AddTopic200ResponseNodeDataYoutubeLinksInner{
			Link:  request.YoutubeLinks[0].Link,
			Votes: 0,
			AddedBy: openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
				Id:       user.Id,
				Username: user.Username,
			},
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

	err = userVideoEditTx(tx, clock, user.Id, request)

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
				//already added
				return nil
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
			Link:  request.YoutubeLinks[0].Link,
			Votes: 0,
			AddedBy: openapi.AddTopic200ResponseNodeDataYoutubeLinksInnerAddedBy{
				Id:       user.Id,
				Username: user.Username,
			},
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

// vote on a video
//
// if votes are greater than zero then trying to add a vote
func updateNodeVideoVote(db *bolt.DB, request openapi.AddTopic200ResponseNodeData, userId string) (vote int32, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		vote, err = updateNodeVideoVoteTx(tx, request, userId)
		return err
	})

	return
}

func updateNodeVideoVoteTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData, userId string) (vote int32, err error) {

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
		vote, err = userVideoVoteTx(tx, userId, request)
		if err != nil {
			return vote, err
		}

		found := false
		for i, item := range node.YoutubeLinks {
			if item.Link == request.YoutubeLinks[0].Link {
				node.YoutubeLinks[i].Votes += vote
				found = true
				vote = node.YoutubeLinks[i].Votes
				break
			}
		}

		if !found {
			return vote, fmt.Errorf("could not find the video on node")
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
	//I might need to add this to the user

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)

	return
}

func updateNodeFreshVote(db *bolt.DB, request openapi.AddTopic200ResponseNodeData, userId string) (vote int32, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		vote, err = updateNodeFreshVoteTx(tx, request, userId)
		return err
	})

	return
}

// updates the title and description
func updateNodeFreshVoteTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData, userId string) (vote int32, err error) {
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
		vote, err = userFreshVoteTx(tx, userId, request)
		if err != nil {
			return vote, err
		}

		node.Fresh += vote //vote will either be a +1 or -1

		// Update reputation of the node creator if this is a vote (not an unvote)
		if vote != 0 && node.CreatedBy.Id != "" {
			// Only update reputation if the voter is not the creator
			if node.CreatedBy.Id != userId {
				updateCreatorReputation(tx, node.CreatedBy.Id, vote)
			}
		}
	}

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
	vote = node.Fresh

	return
}

func userFreshVoteTx(tx *bolt.Tx, userId string, request openapi.AddTopic200ResponseNodeData) (vote int32, err error) {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	// Get the node to retrieve its title if not provided
	var nodeTitle string
	if request.Title == "" {
		// Fetch the node to get its title
		_, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
		if err == nil {
			var node openapi.NodeData
			if err = json.Unmarshal(nodeData, &node); err == nil {
				nodeTitle = node.Title
			}
		}
	} else {
		nodeTitle = request.Title
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
			Title:  nodeTitle,
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
			Title:  nodeTitle,
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

// Helper function to update node title in user records
func updateUserNodeTitleTx(tx *bolt.Tx, nodeId time.Time, topic string, newTitle string) error {
	// Get all users
	usersBucket := tx.Bucket([]byte(KeyUsers))
	if usersBucket == nil {
		return fmt.Errorf("can't find users bucket")
	}

	c := usersBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var user openapi.UpdateUserRequest
		if err := json.Unmarshal(v, &user); err != nil {
			continue // Skip this user if we can't unmarshal
		}

		updated := false

		// Update in Created list
		for i, created := range user.Created {
			if created.NodeId.Equal(nodeId) {
				user.Created[i].Title = newTitle
				updated = true
			}
		}

		// Update in Edited list
		for i, edited := range user.Edited {
			if edited.NodeId.Equal(nodeId) {
				user.Edited[i].Title = newTitle
				updated = true
			}
		}

		// Update in BattleTestedUp list
		for i, item := range user.BattleTestedUp {
			if item.NodeId.Equal(nodeId) {
				user.BattleTestedUp[i].Title = newTitle
				updated = true
			}
		}

		// Update in BattleTestedDown list
		for i, item := range user.BattleTestedDown {
			if item.NodeId.Equal(nodeId) {
				user.BattleTestedDown[i].Title = newTitle
				updated = true
			}
		}

		// Update in FreshUp list
		for i, item := range user.FreshUp {
			if item.NodeId.Equal(nodeId) {
				user.FreshUp[i].Title = newTitle
				updated = true
			}
		}

		// Update in FreshDown list
		for i, item := range user.FreshDown {
			if item.NodeId.Equal(nodeId) {
				user.FreshDown[i].Title = newTitle
				updated = true
			}
		}

		// Save the user if any updates were made
		if updated {
			marshal, err := json.Marshal(user)
			if err != nil {
				return err
			}

			if err := usersBucket.Put(k, marshal); err != nil {
				return err
			}
		}
	}

	return nil
}

// updateCreatorReputation updates the reputation of a node creator based on votes
func updateCreatorReputation(tx *bolt.Tx, creatorId string, voteValue int32) error {
	usersBucket, creator, err := getUserAndBucketRx(tx, creatorId)
	if err != nil {
		return err
	}

	// Normalize vote value to +1 or -1
	reputationChange := int32(1)
	if voteValue < 0 {
		reputationChange = -1
	}

	// Update reputation
	creator.Reputation += reputationChange

	// Save updated user
	marshal, err := json.Marshal(creator)
	if err != nil {
		return err
	}

	return usersBucket.Put([]byte(creatorId), marshal)
}
