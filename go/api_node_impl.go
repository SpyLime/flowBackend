package openapi

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/SpyLime/flowBackend/utility"
	bolt "go.etcd.io/bbolt"
)

func PostNode(db *bolt.DB, clock utility.Clock, node AddTopic200ResponseNodeData) (response ResponsePostNode, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		response, err = postNodeTx(tx, clock, node)
		if err != nil {
			return err
		}

		return nil
	})

	return
}

func postNodeTx(tx *bolt.Tx, clock utility.Clock, node AddTopic200ResponseNodeData) (response ResponsePostNode, err error) {

	topicsBucket := tx.Bucket([]byte(utility.KeyTopics))
	if topicsBucket == nil {
		return response, fmt.Errorf("can't find topics bucket")
	}

	topicBucket := topicsBucket.Bucket([]byte(node.Topic))
	if topicBucket == nil {
		return response, fmt.Errorf("can't find topic bucket")
	}

	nodesBucket := topicBucket.Bucket([]byte(utility.KeyNodes))
	if nodesBucket == nil {
		return response, fmt.Errorf("can't find nodes bucket")
	}

	newNode := NodeData{
		Topic:     node.Title,
		CreatedBy: "change hard code",
	}

	marshal, err := json.Marshal(newNode)
	if err != nil {
		return
	}

	id := clock.Now().Truncate(time.Millisecond)
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

	edge := GetMapById200ResponseEdgesInner{
		Id:     response.SourceId.String() + "-" + response.TargetId.String(),
		Source: response.SourceId,
		Target: response.TargetId,
	}
	_, err = postEdgeTx(tx, topicBucket, edge)
	if err != nil {
		return
	}

	return
}
