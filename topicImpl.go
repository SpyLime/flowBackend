package main

import (
	"encoding/json"
	"fmt"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

func getTopics(db *bolt.DB) (response []openapi.GetTopics200ResponseInner, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		response, err = getTopicsRx(tx)
		return err
	})

	return
}

func getTopicsRx(tx *bolt.Tx) (response []openapi.GetTopics200ResponseInner, err error) {
	response = []openapi.GetTopics200ResponseInner{}
	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return response, fmt.Errorf("cannot find topics bucket")
	}

	c := topicsBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		response = append(response, openapi.GetTopics200ResponseInner{
			Title: string(k),
		})
	}

	return
}

func postTopic(db *bolt.DB, clock Clock, topic openapi.Topic, user openapi.User) (response openapi.ResponsePostTopic, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		response, err = postTopicTx(tx, clock, topic, user)
		return err
	})

	return
}

func postTopicTx(tx *bolt.Tx, clock Clock, topic openapi.Topic, user openapi.User) (response openapi.ResponsePostTopic, err error) {

	topicsBucket, err := tx.CreateBucketIfNotExists([]byte(KeyTopics))
	if err != nil {
		return
	}

	topicBucket, err := topicsBucket.CreateBucket([]byte(topic.Title))
	if err != nil {
		return response, err
	}

	nodesBucket, err := topicBucket.CreateBucket([]byte(KeyNodes))
	if err != nil {
		return
	}

	_, err = topicBucket.CreateBucket([]byte(KeyEdges))
	if err != nil {
		return
	}

	response.Topic.Title = topic.Title

	newNode := openapi.NodeData{
		Topic: topic.Title,
		CreatedBy: openapi.UserIdentifier{
			Id:       user.Id,
			Username: user.Username,
		},
	}

	id := clock.Now()
	newNode.Id = id

	marshal, err := json.Marshal(newNode)
	if err != nil {
		return
	}

	response.NodeData = newNode

	idB, err := id.MarshalText()
	if err != nil {
		return
	}

	err = nodesBucket.Put(idB, marshal)
	if err != nil {
		return
	}

	err = userNodeCreatedTx(tx, user.Id, newNode)

	return
}

func deleteTopic(db *bolt.DB, topicId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = deleteTopicTx(tx, topicId)
		return err
	})

	return
}

func deleteTopicTx(tx *bolt.Tx, topicId string) (err error) {
	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return fmt.Errorf("can't find topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(topicId))
	if topicBucket == nil {
		return fmt.Errorf("can't find topic bucket")
	}

	// Get the nodes bucket to process all nodes
	nodesBucket := topicBucket.Bucket([]byte(KeyNodes))
	if nodesBucket != nil {
		// Iterate through all nodes and remove them from user records
		c := nodesBucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			nodeId := string(k)
			// Remove this node from all users who interacted with it
			err = removeNodeFromAllUsersTx(tx, nodeId, topicId)
			if err != nil {
				return err
			}
		}
	}

	// Now delete the topic bucket
	err = topicsBucket.DeleteBucket([]byte(topicId))
	return
}
