package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/stretchr/testify/require"
)

func TestGetTopics(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("getUserByName", 8088, "")
	defer tearDown()

	_, topics, _, err := CreateTestData(db, &clock, 1, 4, 0)
	require.Nil(t, err)

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet,
		"http://127.0.0.1:8088/api/v1/topic",
		nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var data []openapi.Topic
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&data)

	require.Equal(t, 4, len(data))
	require.Equal(t, topics[0], data[0].Title)
}

func TestDeleteTopic(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("DeleteTopic", 8088, "")
	defer tearDown()

	users, topics, _, err := CreateTestData(db, &clock, 1, 1, 0)
	require.Nil(t, err)

	UpdateUserRoleAndReputation(db, users[0], true, 0)
	SetTestLoginUser(users[0])

	nonEmptyTopics, err := getTopics(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(nonEmptyTopics))

	require.Nil(t, err)

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodDelete, "http://127.0.0.1:8088/api/v1/topic/"+topics[0], nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 204, resp.StatusCode)

	emptyTopics, err := getTopics(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(emptyTopics))

}

func TestAddTopic(t *testing.T) {

	db, tearDown := FullStartTestServer("AddTopic", 8088, "")
	defer tearDown()
	clock := TestClock{}
	users, _, _, err := CreateTestData(db, &clock, 1, 0, 0)
	require.Nil(t, err)

	err = UpdateUserRoleAndReputation(db, users[0], true, 0)
	require.Nil(t, err)
	SetTestLoginUser(users[0])

	emptyTopics, err := getTopics(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(emptyTopics))

	require.Nil(t, err)

	client := &http.Client{}

	newTopic := openapi.Topic{
		Title: "test1",
	}

	marshal, err := json.Marshal(newTopic)
	require.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8088/api/v1/topic", bytes.NewBuffer(marshal))

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	nonEmptyTopics, err := getTopics(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(nonEmptyTopics))

}
