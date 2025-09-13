package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

// areSameYouTubeVideo returns true if two YouTube URLs point to the same video.
func areSameYouTubeVideo(link1, link2 string) bool {
	return link1 == link2
}

func getNode(db *bolt.DB, nodeId, topicId string) (response openapi.NodeData, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		response, err = getNodeRx(tx, nodeId, topicId)
		return err
	})

	return
}

// getNextNode returns the ID of the next node
func getNextNode(db *bolt.DB, nodeId, topicId, search string) (Id string, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		Id, err = getNextNodeTx(tx, nodeId, topicId, search)
		return err
	})

	return
}

func getNextNodeTx(tx *bolt.Tx, nodeId, topicId, search string) (Id string, err error) {
	// go through every edge and select any edge that has the current node as the source save a list of all the targets
	var targetIds []string
	edgesBucket := tx.Bucket([]byte(KeyTopics)).Bucket([]byte(topicId)).Bucket([]byte(KeyEdges))
	c := edgesBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			continue
		}

		var edge openapi.Edge
		err = json.Unmarshal(v, &edge)
		if err != nil {
			return
		}

		if edge.Source.Format(time.RFC3339Nano) == nodeId {
			targetIds = append(targetIds, edge.Target.Format(time.RFC3339Nano))
		}
	}

	// go through each target and find the one with the highest battle tested score
	var highestScore int32 = -1000000
	for _, targetId := range targetIds {
		node, err := getNodeRx(tx, targetId, topicId)
		if err != nil {
			return Id, err
		}

		if search == "battleTested" {
			if node.BattleTested > highestScore {
				highestScore = node.BattleTested
				Id = targetId
			}
		}
		if search == "fresh" {
			if node.Fresh > highestScore {
				highestScore = node.Fresh
				Id = targetId
			}
		}
	}

	return

}

func getNodeRx(tx *bolt.Tx, nodeId, topicId string) (response openapi.NodeData, err error) {
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
func postNode(db *bolt.DB, clock Clock, node openapi.NodeData) (response openapi.ResponsePostNode, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		response, err = postNodeTx(tx, clock, node)
		return err
	})

	return
}

// do not supply a target id
//
// supply topic, title, createdBy, ID which is the source ID
func postNodeTx(tx *bolt.Tx, clock Clock, node openapi.NodeData) (response openapi.ResponsePostNode, err error) {

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

	// Generate the ID first
	id := clock.Now()
	newNode.Id = id // Set the ID in the node object

	// Marshal the node with the ID included
	marshal, err := json.Marshal(newNode)
	if err != nil {
		return
	}

	// Convert ID to bytes for the bucket key
	idB, err := id.MarshalText()
	if err != nil {
		return
	}

	// Store the node in the bucket
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

	edge := openapi.Edge{
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
		user.Created = append(user.Created, openapi.ResponseUserInfoInner{
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
		var user openapi.User
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
				if areSameYouTubeVideo(userVideo, video.Link) {
					userIds[string(k)] = true
				}
			}

			for _, userVideo := range user.VideoDown {
				if areSameYouTubeVideo(userVideo, video.Link) {
					userIds[string(k)] = true
				}
			}

			// Check linked videos
			for _, linked := range user.Linked {
				if areSameYouTubeVideo(linked.Link, video.Link) {
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
func removeNodeFromUserTx(tx *bolt.Tx, userId string, nodeId string, topicId string, topic string, videos []openapi.LinkData) error {
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
			if areSameYouTubeVideo(link, video.Link) {
				user.VideoUp = append(user.VideoUp[:i], user.VideoUp[i+1:]...)
				break
			}
		}

		// Remove from video down votes
		for i, link := range user.VideoDown {
			if areSameYouTubeVideo(link, video.Link) {
				user.VideoDown = append(user.VideoDown[:i], user.VideoDown[i+1:]...)
				break
			}
		}

		// Remove from linked videos
		for i, linked := range user.Linked {
			if areSameYouTubeVideo(linked.Link, video.Link) {
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

func updateNodeTitle(db *bolt.DB, request openapi.NodeData, editor openapi.User) (editorAdded bool, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		editorAdded, err = updateNodeTitleTx(tx, request, editor)
		return err
	})

	return
}

// updates the title and description
func updateNodeTitleTx(tx *bolt.Tx, request openapi.NodeData, editor openapi.User) (editorAdded bool, err error) {
	fmt.Printf(request.Id.Format(time.RFC3339Nano))
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
	if request.Description != "" && request.Description != node.Description {
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
		node.EditedBy = append(node.EditedBy, openapi.UserIdentifier{
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

	// Update all users who have this node in their lists
	if isEdited {
		err = updateUserNodeTitleTx(tx, node.Id, node.Topic, node.Title)
		if err != nil {
			return editorAdded, err
		}
	}

	return
}

// Wrapper function that opens a transaction
func updateUserNodeEdited(db *bolt.DB, userId string, node openapi.NodeData) error {
	return db.Update(func(tx *bolt.Tx) error {
		return userNodeEditedTx(tx, userId, node)
	})
}

// Helper function to save the edited node to the user's record
func userNodeEditedTx(tx *bolt.Tx, userId string, node openapi.NodeData) error {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return err
	}

	// Check for duplicates before appending to edited list
	isDuplicate := false
	for i, edited := range user.Edited {
		if edited.NodeId.Format(time.RFC3339Nano) == node.Id.Format(time.RFC3339Nano) {
			isDuplicate = true
			// Update the title in case it changed
			user.Edited[i].Title = node.Title
			break
		}
	}

	if !isDuplicate {
		user.Edited = append(user.Edited, openapi.ResponseUserInfoInner{
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

	// Even if it's a duplicate, we need to save any updates
	marshal, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return usersBucket.Put([]byte(userId), marshal)
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

func updateNodeBattleVote(db *bolt.DB, request openapi.NodeData, userId string) (vote int32, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		vote, err = updateNodeBattleVoteTx(tx, request, userId)
		return err
	})

	return
}

func updateNodeBattleVoteTx(tx *bolt.Tx, request openapi.NodeData, userId string) (vote int32, err error) {
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
				err = updateCreatorReputation(tx, node.CreatedBy.Id, vote)
				if err != nil {
					return vote, err
				}
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

func userBattleVoteTx(tx *bolt.Tx, userId string, request openapi.NodeData) (int32, error) {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return 0, err
	}

	// Resolve title if needed
	nodeTitle := request.Title
	if nodeTitle == "" {
		if _, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano)); err == nil {
			var node openapi.NodeData
			if json.Unmarshal(nodeData, &node) == nil {
				nodeTitle = node.Title
			}
		}
	}

	// --- Normalize current state (no duplicates, and not present in both lists) ---
	hadUp := containsNode(user.BattleTestedUp, request.Id)
	hadDown := containsNode(user.BattleTestedDown, request.Id)

	// If inconsistent (both present), clean to neutral as a defensive fix.
	if hadUp && hadDown {
		user.BattleTestedUp = removeNodeAll(user.BattleTestedUp, request.Id)
		user.BattleTestedDown = removeNodeAll(user.BattleTestedDown, request.Id)
		hadUp, hadDown = false, false
	}

	// Old state
	var oldVote int32
	switch {
	case hadUp:
		oldVote = +1
	case hadDown:
		oldVote = -1
	default:
		oldVote = 0
	}

	// Requested direction (only sign matters)
	reqDir := int32(0)
	if request.BattleTested > 0 {
		reqDir = +1
	} else if request.BattleTested < 0 {
		reqDir = -1
	} else {
		// If you ever send 0, treat as no-op to avoid accidental 0 deltas.
		// You could also choose to "unvote" here if that's desired.
		return 0, fmt.Errorf("invalid request: BattleTested must be +/-1")
	}

	// --- State machine producing ONLY {-2,-1,+1,+2} ---
	var delta int32
	switch {
	// Request UP
	case reqDir == +1 && oldVote == +1:
		// up -> unvote
		delta = -1
		user.BattleTestedUp = removeNodeAll(user.BattleTestedUp, request.Id)

	case reqDir == +1 && oldVote == 0:
		// none -> up
		delta = +1
		user.BattleTestedUp = append(user.BattleTestedUp, openapi.ResponseUserInfoInner{
			Topic:  request.Topic,
			Title:  nodeTitle,
			NodeId: request.Id,
		})

	case reqDir == +1 && oldVote == -1:
		// down -> up (unvote down, then up)
		delta = +2
		user.BattleTestedDown = removeNodeAll(user.BattleTestedDown, request.Id)
		user.BattleTestedUp = append(user.BattleTestedUp, openapi.ResponseUserInfoInner{
			Topic:  request.Topic,
			Title:  nodeTitle,
			NodeId: request.Id,
		})

	// Request DOWN
	case reqDir == -1 && oldVote == -1:
		// down -> unvote
		delta = +1
		user.BattleTestedDown = removeNodeAll(user.BattleTestedDown, request.Id)

	case reqDir == -1 && oldVote == 0:
		// none -> down
		delta = -1
		user.BattleTestedDown = append(user.BattleTestedDown, openapi.ResponseUserInfoInner{
			Topic:  request.Topic,
			Title:  nodeTitle,
			NodeId: request.Id,
		})

	case reqDir == -1 && oldVote == +1:
		// up -> down (unvote up, then down)
		delta = -2
		user.BattleTestedUp = removeNodeAll(user.BattleTestedUp, request.Id)
		user.BattleTestedDown = append(user.BattleTestedDown, openapi.ResponseUserInfoInner{
			Topic:  request.Topic,
			Title:  nodeTitle,
			NodeId: request.Id,
		})
	}

	// Persist user
	b, err := json.Marshal(user)
	if err != nil {
		return delta, err
	}
	if err := usersBucket.Put([]byte(userId), b); err != nil {
		return delta, err
	}
	return delta, nil
}

func containsNode(list []openapi.ResponseUserInfoInner, id time.Time) bool {
	for _, item := range list {
		if item.NodeId.Equal(id) {
			return true
		}
	}
	return false
}

func removeNodeAll(list []openapi.ResponseUserInfoInner, id time.Time) []openapi.ResponseUserInfoInner {
	out := list[:0]
	for _, item := range list {
		if !item.NodeId.Equal(id) {
			out = append(out, item)
		}
	}
	return out
}

// want to add a video to a node
//
// if votes are greater than zero then trying to add a video
func updateNodeVideoEdit(db *bolt.DB, clock Clock, request openapi.NodeData, user openapi.User) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeVideoEditTx(tx, clock, request, user)
		return err
	})

	return
}

func updateNodeVideoEditTx(tx *bolt.Tx, clock Clock, request openapi.NodeData, user openapi.User) (err error) {

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

		if areSameYouTubeVideo(item.Link, request.YoutubeLinks[0].Link) { // Check if ID matches

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

			err = userVideoEditTx(tx, clock, item.AddedBy.Id, request) //remove the video from the user that added it linked field
			if err != nil {
				return err
			}
			err = removeVideoFromUsersVotersTx(tx, request.YoutubeLinks[0].Link) //remove the video from every user that voted on it
			if err != nil {
				return err
			}

			return err
		}
	}

	if request.YoutubeLinks[0].Votes > 0 { //video was not found and you want to add
		node.YoutubeLinks = append(node.YoutubeLinks, openapi.LinkData{
			Link:  request.YoutubeLinks[0].Link,
			Votes: 0,
			AddedBy: openapi.UserIdentifier{
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

	err = userVideoEditTx(tx, clock, user.Id, request) //add the video to the user that added it linked field

	return
}

func removeVideoFromUsersVotersTx(tx *bolt.Tx, videoLink string) (err error) {
	usersBucket := tx.Bucket([]byte(KeyUsers))
	if usersBucket == nil {
		return fmt.Errorf("users bucket not found")
	}

	c := usersBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var user openapi.User

		// Unmarshal the user JSON
		if err := json.Unmarshal(v, &user); err != nil {
			return err
		}

		// Remove from video up votes
		for i, link := range user.VideoUp {
			if areSameYouTubeVideo(link, videoLink) {
				user.VideoUp = append(user.VideoUp[:i], user.VideoUp[i+1:]...)
				break
			}
		}

		// Remove from video down votes
		for i, link := range user.VideoDown {
			if areSameYouTubeVideo(link, videoLink) {
				user.VideoDown = append(user.VideoDown[:i], user.VideoDown[i+1:]...)
				break
			}
		}

		// Marshal back to JSON
		marshaled, err := json.Marshal(user)
		if err != nil {
			return err
		}

		// Save updated user
		if err := usersBucket.Put(k, marshaled); err != nil {
			return err
		}
	}

	return
}

func userVideoEditTx(tx *bolt.Tx, clock Clock, userId string, request openapi.NodeData) (err error) {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	for i, item := range user.Linked {
		if areSameYouTubeVideo(item.Link, request.YoutubeLinks[0].Link) {

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
		user.Linked = append(user.Linked, openapi.LinkData{
			Link:  request.YoutubeLinks[0].Link,
			Votes: 0,
			AddedBy: openapi.UserIdentifier{
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
func updateNodeVideoVote(db *bolt.DB, request openapi.NodeData, userId string) (vote int32, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		vote, err = updateNodeVideoVoteTx(tx, request, userId)
		return err
	})

	return
}

func updateNodeVideoVoteTx(tx *bolt.Tx, request openapi.NodeData, userId string) (vote int32, err error) {
	nodesBucket, nodeData, err := nodeDataFinderTx(tx, request.Topic, request.Id.Format(time.RFC3339Nano))
	if err != nil {
		return
	}

	var node openapi.NodeData
	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return
	}

	// Find the video link
	var videoIndex = -1
	for i, video := range node.YoutubeLinks {
		if areSameYouTubeVideo(video.Link, request.YoutubeLinks[0].Link) {
			videoIndex = i
			break
		}
	}

	if videoIndex == -1 {
		return 0, fmt.Errorf("video link not found")
	}

	// Get user data
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	nodeTitle := node.Title
	if nodeTitle == "" {
		nodeTitle = "Untitled"
	}

	// Calculate reputation changes
	var reputationChange int32 = 0

	// Handle upvote
	if request.YoutubeLinks[0].Votes > 0 {
		// Check if user already upvoted
		for i, item := range user.VideoUp {
			if areSameYouTubeVideo(item, request.YoutubeLinks[0].Link) {
				// Remove upvote
				user.VideoUp = append(user.VideoUp[:i], user.VideoUp[i+1:]...)
				node.YoutubeLinks[videoIndex].Votes--

				// Update reputation of video creator (if not the voter)
				if node.YoutubeLinks[videoIndex].AddedBy.Id != "" && node.YoutubeLinks[videoIndex].AddedBy.Id != userId {
					reputationChange = -1
				}

				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}
				err = usersBucket.Put([]byte(userId), marshal)
				if err != nil {
					return vote, err
				}
				vote = node.YoutubeLinks[videoIndex].Votes

				// Update creator reputation
				if reputationChange != 0 {
					updateCreatorReputation(tx, node.YoutubeLinks[videoIndex].AddedBy.Id, reputationChange)
				}

				marshal, err = json.Marshal(node)
				if err != nil {
					return vote, err
				}
				return vote, nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
			}
		}

		// Check if user already downvoted
		for i, item := range user.VideoDown {
			if areSameYouTubeVideo(item, request.YoutubeLinks[0].Link) {
				// Remove downvote and add upvote (switch vote)
				user.VideoDown = append(user.VideoDown[:i], user.VideoDown[i+1:]...)
				node.YoutubeLinks[videoIndex].Votes += 2 // +1 for removing downvote, +1 for adding upvote

				// Update reputation
				if node.YoutubeLinks[videoIndex].AddedBy.Id != "" && node.YoutubeLinks[videoIndex].AddedBy.Id != userId {
					reputationChange = 2 // +1 for removing downvote, +1 for adding upvote
				}

				user.VideoUp = append(user.VideoUp, request.YoutubeLinks[0].Link)

				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}
				err = usersBucket.Put([]byte(userId), marshal)
				if err != nil {
					return vote, err
				}
				vote = node.YoutubeLinks[videoIndex].Votes

				// Update creator reputation
				if reputationChange != 0 {
					updateCreatorReputation(tx, node.YoutubeLinks[videoIndex].AddedBy.Id, reputationChange)
				}

				marshal, err = json.Marshal(node)
				if err != nil {
					return vote, err
				}
				return vote, nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
			}
		}

		// New upvote
		user.VideoUp = append(user.VideoUp, request.YoutubeLinks[0].Link)
		node.YoutubeLinks[videoIndex].Votes++

		// Update reputation
		if node.YoutubeLinks[videoIndex].AddedBy.Id != "" && node.YoutubeLinks[videoIndex].AddedBy.Id != userId {
			reputationChange = 1
		}
	} else if request.YoutubeLinks[0].Votes < 0 {
		// Handle downvote
		for i, item := range user.VideoDown {
			if areSameYouTubeVideo(item, request.YoutubeLinks[0].Link) {
				// Remove downvote
				user.VideoDown = append(user.VideoDown[:i], user.VideoDown[i+1:]...)
				node.YoutubeLinks[videoIndex].Votes++

				// Update reputation
				if node.YoutubeLinks[videoIndex].AddedBy.Id != "" && node.YoutubeLinks[videoIndex].AddedBy.Id != userId {
					reputationChange = 1
				}

				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}
				err = usersBucket.Put([]byte(userId), marshal)
				if err != nil {
					return vote, err
				}
				vote = node.YoutubeLinks[videoIndex].Votes

				// Update creator reputation
				if reputationChange != 0 {
					updateCreatorReputation(tx, node.YoutubeLinks[videoIndex].AddedBy.Id, reputationChange)
				}

				marshal, err = json.Marshal(node)
				if err != nil {
					return vote, err
				}
				return vote, nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
			}
		}

		// Check if user already upvoted
		for i, item := range user.VideoUp {
			if areSameYouTubeVideo(item, request.YoutubeLinks[0].Link) {
				// Remove upvote and add downvote (switch vote)
				user.VideoUp = append(user.VideoUp[:i], user.VideoUp[i+1:]...)
				node.YoutubeLinks[videoIndex].Votes -= 2 // -1 for removing upvote, -1 for adding downvote

				// Update reputation
				if node.YoutubeLinks[videoIndex].AddedBy.Id != "" && node.YoutubeLinks[videoIndex].AddedBy.Id != userId {
					reputationChange = -2 // -1 for removing upvote, -1 for adding downvote
				}

				user.VideoDown = append(user.VideoDown, request.YoutubeLinks[0].Link)

				marshal, err := json.Marshal(user)
				if err != nil {
					return vote, err
				}
				err = usersBucket.Put([]byte(userId), marshal)
				if err != nil {
					return vote, err
				}
				vote = node.YoutubeLinks[videoIndex].Votes

				// Update creator reputation
				if reputationChange != 0 {
					updateCreatorReputation(tx, node.YoutubeLinks[videoIndex].AddedBy.Id, reputationChange)
				}

				marshal, err = json.Marshal(node)
				if err != nil {
					return vote, err
				}
				return vote, nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
			}
		}

		// New downvote
		user.VideoDown = append(user.VideoDown, request.YoutubeLinks[0].Link)
		node.YoutubeLinks[videoIndex].Votes--

		// Update reputation
		if node.YoutubeLinks[videoIndex].AddedBy.Id != "" && node.YoutubeLinks[videoIndex].AddedBy.Id != userId {
			reputationChange = -1
		}
	}

	marshal, err := json.Marshal(user)
	if err != nil {
		return vote, err
	}
	err = usersBucket.Put([]byte(userId), marshal)
	if err != nil {
		return
	}

	// Update creator reputation
	if reputationChange != 0 {
		updateCreatorReputation(tx, node.YoutubeLinks[videoIndex].AddedBy.Id, reputationChange)
	}

	marshal, err = json.Marshal(node)
	if err != nil {
		return
	}
	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)
	vote = node.YoutubeLinks[videoIndex].Votes

	return
}

func userVideoVoteTx(tx *bolt.Tx, userId string, request openapi.NodeData) (vote int32, err error) {
	usersBucket, user, err := getUserAndBucketRx(tx, userId)
	if err != nil {
		return
	}

	if request.YoutubeLinks[0].Votes > 0 {
		for i, item := range user.VideoUp {
			if areSameYouTubeVideo(item, request.YoutubeLinks[0].Link) {
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
			if areSameYouTubeVideo(item, request.YoutubeLinks[0].Link) {
				user.VideoDown = append(user.VideoDown[:i], user.VideoDown[i+1:]...)
				vote++
				break
			}
		}

	} else {
		for i, item := range user.VideoDown {
			if areSameYouTubeVideo(item, request.YoutubeLinks[0].Link) {
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
			if areSameYouTubeVideo(item, request.YoutubeLinks[0].Link) {
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

func updateNodeFlag(db *bolt.DB, request openapi.NodeData) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeFlagTx(tx, request)
		return err
	})

	return
}

// updates the title and description
func updateNodeFlagTx(tx *bolt.Tx, request openapi.NodeData) (err error) {

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

func updateNodeFreshVote(db *bolt.DB, request openapi.NodeData, userId string) (vote int32, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		vote, err = updateNodeFreshVoteTx(tx, request, userId)
		return err
	})

	return
}

// updates the title and description
func updateNodeFreshVoteTx(tx *bolt.Tx, request openapi.NodeData, userId string) (vote int32, err error) {
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

func userFreshVoteTx(tx *bolt.Tx, userId string, request openapi.NodeData) (vote int32, err error) {
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

		user.FreshUp = append(user.FreshUp, openapi.ResponseUserInfoInner{
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

		user.FreshDown = append(user.FreshDown, openapi.ResponseUserInfoInner{
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
		var user openapi.User
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

	// Use the actual vote value for reputation change
	// This allows for -2 when switching from upvote to downvote
	// and +2 when switching from downvote to upvote
	creator.Reputation += voteValue

	// Save updated user
	marshal, err := json.Marshal(creator)
	if err != nil {
		return err
	}

	err = usersBucket.Put([]byte(creatorId), marshal)
	if err != nil {
		return err
	}

	// Debugging re-read
	saved := usersBucket.Get([]byte(creatorId))
	var savedUser openapi.User
	if err := json.Unmarshal(saved, &savedUser); err == nil {
		log.Printf("Reputation persisted for %s: %d", creatorId, savedUser.Reputation)
	} else {
		log.Printf("Failed to re-unmarshal user after save: %v", err)
	}

	return nil
}
