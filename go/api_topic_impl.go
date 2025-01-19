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
	// next 3 lines are probably redundant
	// foundTopic := topicsBucket.Get([]byte(topic.Title))
	// if foundTopic != nil {
	// 	return response, fmt.Errorf("There is already a topic with that title")
	// }

	//its tomorrow, you need node data returned as this is the first node of that topic or root node.
	//otherwise there will be nothing to attach future nodes to and no means to start the first one.

	topicBucket, err := topicsBucket.CreateBucket([]byte(topic.Title))
	if err != nil {
		return response, err
	}

	newNode := NodeData{
		Topic:     topic.Title,
		CreatedBy: "change hard code",
	}

	marshal, err := json.Marshal(newNode)
	if err != nil {
		return
	}

	idB, err := clock.Now().Truncate(time.Millisecond).MarshalText()
	if err != nil {
		return
	}

	err = topicBucket.Put(idB, marshal)

	//this is out of order I will need to also create its first node
	//I am not sure if the line below actually creates a bucket
	//I don't think it does as I think I need CreateBucket

	return
}
