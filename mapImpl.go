package main

import (
	"encoding/json"
	"fmt"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

func getMapById(db *bolt.DB, topicId string) (response openapi.MapData, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		response, err = getMapByIdRx(tx, topicId)
		return err
	})

	return
}

func getMapByIdRx(tx *bolt.Tx, topicId string) (response openapi.MapData, err error) {
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

	nodes := make([]openapi.FlowNode, 0)
	c := nodesBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			continue
		}

		var retrievedNode openapi.NodeData
		err = json.Unmarshal(v, &retrievedNode)
		if err != nil {
			return
		}

		var id time.Time
		err = id.UnmarshalText(k)
		if err != nil {
			return
		}

		newNode := openapi.FlowNode{
			Id: id,
			Data: openapi.FlowNodeData{
				Title:        retrievedNode.Title,
				BattleTested: retrievedNode.BattleTested,
				Fresh:        retrievedNode.Fresh,
				Speed:        retrievedNode.Speed,
			},
		}

		nodes = append(nodes, newNode)
	}

	edges := make([]openapi.Edge, 0)
	c = edgesBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			continue
		}

		var newEdge openapi.Edge
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

func postEdge(db *bolt.DB, topic string, edge openapi.Edge) (newId string, err error) {
	if edge.Source == edge.Target {
		return newId, fmt.Errorf("your trying to connect a node to itself")
	}

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

func postEdgeTx(topicBucket *bolt.Bucket, edge openapi.Edge) (newId string, err error) {
	edgesBucket, err := topicBucket.CreateBucketIfNotExists([]byte(KeyEdges))
	if err != nil {
		return
	}

	foundEdge := edgesBucket.Get([]byte(edge.Id))
	if foundEdge != nil {
		return newId, fmt.Errorf("your trying to connect nodes that are already connected")
	}

	reverseEdge := edgesBucket.Get([]byte(edge.Target.Format(time.RFC3339Nano) + "-" + edge.Source.Format(time.RFC3339Nano)))
	if reverseEdge != nil {
		return newId, fmt.Errorf("your trying to connect nodes that are already connected")
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

func deleteEdge(db *bolt.DB, topicId string, edgeId string) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = deleteEdgeTx(tx, topicId, edgeId)
		return err
	})

	return
}

func deleteEdgeTx(tx *bolt.Tx, topicId string, edgeId string) (err error) {

	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return fmt.Errorf("can't find topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(topicId))
	if topicBucket == nil {
		return fmt.Errorf("can't find topic bucket")
	}

	edgesBucket := topicBucket.Bucket([]byte(KeyEdges))
	if edgesBucket == nil {
		return fmt.Errorf("can't find edges bucket")
	}

	err = edgesBucket.Delete([]byte(edgeId))

	return
}
