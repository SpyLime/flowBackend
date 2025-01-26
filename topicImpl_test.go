package main

import (
	"testing"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/lgr"
	"github.com/stretchr/testify/require"
)

func TestPostGetTopic(t *testing.T) {

	lgr.Printf("INFO TestPostGetTopic")
	t.Log("INFO TestPostGetTopic")
	clock := TestClock{}
	db, dbTearDown := OpenTestDB("PostGetTopic")
	defer dbTearDown()

	_, _, _, err := CreateTestData(db, &clock, 0, 2, 0)
	require.Nil(t, err)

	topic := openapi.GetTopics200ResponseInner{
		Title: "zzz",
	}

	topicData, err := postTopic(db, &clock, topic)
	require.Nil(t, err)

	require.Equal(t, topic.Title, topicData.Topic.Title)

	response, err := getTopics(db)
	require.Nil(t, err)

	require.Equal(t, 3, len(response))
	require.Equal(t, topic.Title, response[len(response)-1].Title)

}
