package openapi

import (
	"fmt"

	"github.com/SpyLime/flowBackend/utility"
	bolt "go.etcd.io/bbolt"
)

func PostEdge(db *bolt.DB, topic string, edge GetMapById200ResponseEdgesInner) (newId string, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		topicsBucket := tx.Bucket([]byte(utility.KeyTopics))
		if topicsBucket == nil {
			return fmt.Errorf("can't find topics bucket")
		}
		topicBucket := topicsBucket.Bucket([]byte(topic))
		if topicBucket == nil {
			return fmt.Errorf("can't find topic bucket")
		}

		newId, err = postEdgeTx(tx, topicBucket, edge)
		if err != nil {
			return err
		}

		return nil
	})

	return
}

func postEdgeTx(tx *bolt.Tx, topicBucket *bolt.Bucket, edge GetMapById200ResponseEdgesInner) (newId string, err error) {
	//need to make this edge inside edges bucked.
	//I already have topic bucket just need to make edges
	//then use id as key and input the source and target object as the value
	edge.Id = "" //I hope this is considered empty and is omitted
	return
}
