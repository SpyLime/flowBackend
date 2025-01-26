package main

import (
	"encoding/json"
	"fmt"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

func getMapById(db *bolt.DB, topicId string) (response openapi.GetMapById200Response, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		response, err = getMapByIdRx(tx, topicId)
		return err
	})

	return
}

func getMapByIdRx(tx *bolt.Tx, topicId string) (response openapi.GetMapById200Response, err error) {
	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return response, fmt.Errorf("can't find topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(topicId))
	if topicBucket == nil {
		return response, fmt.Errorf("can't find topics bucket")
	}

	nodesBucket := topicBucket.Bucket([]byte(KeyNodes))
	if nodesBucket == nil {
		return response, fmt.Errorf("can't find nodes bucket")
	}

	edgesBucket := topicBucket.Bucket([]byte(KeyEdges))
	if edgesBucket == nil {
		return response, fmt.Errorf("can't find edges bucket")
	}

	nodes := make([]openapi.GetMapById200ResponseNodesInner, 0)
	c := nodesBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			continue
		}

		var retrievedNode openapi.AddTopic200ResponseNodeData
		err = json.Unmarshal(v, &retrievedNode)
		if err != nil {
			return
		}

		var id time.Time
		err = id.UnmarshalText(k)
		if err != nil {
			return
		}

		newNode := openapi.GetMapById200ResponseNodesInner{
			Id: id,
			Data: openapi.GetMapById200ResponseNodesInnerData{
				Title:        retrievedNode.Title,
				BattleTested: retrievedNode.BattleTested,
				Fresh:        retrievedNode.Fresh,
				Speed:        retrievedNode.Speed,
			},
		}

		nodes = append(nodes, newNode)
	}

	edges := make([]openapi.GetMapById200ResponseEdgesInner, 0)
	c = edgesBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			continue
		}

		var newEdge openapi.GetMapById200ResponseEdgesInner
		err = json.Unmarshal(v, &newEdge)
		if err != nil {
			return
		}

		newEdge.Id = string(k)

		edges = append(edges, newEdge)
	}

	response.Nodes = nodes
	response.Edges = edges

	return
}

func postEdge(db *bolt.DB, topic string, edge openapi.GetMapById200ResponseEdgesInner) (newId string, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		topicsBucket := tx.Bucket([]byte(KeyTopics))
		if topicsBucket == nil {
			return fmt.Errorf("can't find topics bucket")
		}
		topicBucket := topicsBucket.Bucket([]byte(topic))
		if topicBucket == nil {
			return fmt.Errorf("can't find topic bucket")
		}

		newId, err = postEdgeTx(topicBucket, edge)
		if err != nil {
			return err
		}

		return nil
	})

	return
}

func postEdgeTx(topicBucket *bolt.Bucket, edge openapi.GetMapById200ResponseEdgesInner) (newId string, err error) {
	edgesBucket, err := topicBucket.CreateBucketIfNotExists([]byte(KeyEdges))
	if err != nil {
		return
	}

	id := edge.Id
	edge.Id = "" //I hope this is considered empty and is omitted

	marshal, err := json.Marshal(edge)
	if err != nil {
		return
	}

	err = edgesBucket.Put([]byte(id), marshal)
	return
}
