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

func postTopic(db *bolt.DB, clock Clock, topic openapi.GetTopics200ResponseInner) (response openapi.ResponsePostTopic, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		response, err = postTopicTx(tx, clock, topic)
		return err
	})

	return
}

func postTopicTx(tx *bolt.Tx, clock Clock, topic openapi.GetTopics200ResponseInner) (response openapi.ResponsePostTopic, err error) {

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

	response.Topic.Title = topic.Title

	newNode := openapi.NodeData{
		Topic:     topic.Title,
		CreatedBy: "change hard code",
	}

	marshal, err := json.Marshal(newNode)
	if err != nil {
		return
	}

	id := clock.Now()
	newNode.Id = id

	response.NodeData = openapi.AddTopic200ResponseNodeData(newNode)

	idB, err := id.MarshalText()
	if err != nil {
		return
	}

	err = nodesBucket.Put(idB, marshal)

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

	err = topicsBucket.DeleteBucket([]byte(topicId))

	return

}
