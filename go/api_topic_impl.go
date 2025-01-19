package openapi

import (
	"encoding/json"
	"time"

	"github.com/SpyLime/flowBackend/utility"
	bolt "go.etcd.io/bbolt"
)

func PostTopic(db *bolt.DB, clock utility.Clock, topic GetTopics200ResponseInner) (response ResponsePostTopic, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		response, err = postTopicTx(tx, clock, topic)
		if err != nil {
			return err
		}

		return nil
	})

	return
}

func postTopicTx(tx *bolt.Tx, clock utility.Clock, topic GetTopics200ResponseInner) (response ResponsePostTopic, err error) {

	topicsBucket, err := tx.CreateBucketIfNotExists([]byte(utility.KeyTopics))
	if err != nil {
		return
	}

	topicBucket, err := topicsBucket.CreateBucket([]byte(topic.Title))
	if err != nil {
		return response, err
	}

	nodesBucket, err := topicBucket.CreateBucket([]byte(utility.KeyNodes))
	if err != nil {
		return
	}

	response.Topic.Title = topic.Title

	newNode := NodeData{
		Topic:     topic.Title,
		CreatedBy: "change hard code",
	}

	marshal, err := json.Marshal(newNode)
	if err != nil {
		return
	}

	id := clock.Now().Truncate(time.Millisecond)
	newNode.Id = id

	response.NodeData = AddTopic200ResponseNodeData(newNode)

	idB, err := id.MarshalText()
	if err != nil {
		return
	}

	err = nodesBucket.Put(idB, marshal)

	return
}
