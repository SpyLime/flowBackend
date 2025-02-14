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
	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return response, fmt.Errorf("can't find topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(topicId))
	if topicBucket == nil {
		return response, fmt.Errorf("can't find topic bucket")
	}

	nodesBucket := topicBucket.Bucket([]byte(KeyNodes))
	if nodesBucket == nil {
		return response, fmt.Errorf("can't find nodes bucket")
	}

	nodeData := nodesBucket.Get([]byte(nodeId))
	if nodeData == nil {
		return response, fmt.Errorf("%s", "cannot find node: "+nodeId)
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

func updateNode(db *bolt.DB, request openapi.AddTopic200ResponseNodeData) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		err = updateNodeTx(tx, request)
		return err
	})

	return
}

func updateNodeTx(tx *bolt.Tx, request openapi.AddTopic200ResponseNodeData) (err error) {
	topicsBucket := tx.Bucket([]byte(KeyTopics))
	if topicsBucket == nil {
		return fmt.Errorf("can't finb topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(request.Topic))
	if topicBucket == nil {
		return fmt.Errorf("can't finb topic bucket")
	}

	nodesBucket := topicBucket.Bucket([]byte(KeyNodes))
	if nodesBucket == nil {
		return fmt.Errorf("can't find nodes bucket")
	}

	nodeData := nodesBucket.Get([]byte(request.Id.Format(time.RFC3339Nano)))
	if nodeData == nil {
		return fmt.Errorf("can't find node data")
	}

	node := openapi.NodeData(request)

	marshal, err := json.Marshal(node)
	if err != nil {
		return
	}

	err = nodesBucket.Put([]byte(request.Id.Format(time.RFC3339Nano)), marshal)

	return
}
