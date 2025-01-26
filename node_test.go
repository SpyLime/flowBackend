package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/stretchr/testify/require"
)

func TestGetNode(t *testing.T) {
	clock := TestClock{}
	db, tearDown := FullStartTestServer("getNode", 8088, "")
	defer tearDown()

	_, topics, nodesAndEdges, err := CreateTestData(db, &clock, 1, 1, 2)
	require.Nil(t, err)

	// SetTestLoginUser(teachers[0])

	client := &http.Client{}

	baseURL := "http://127.0.0.1:8088/api/v1/node"
	params := url.Values{}
	params.Add("nodeId", nodesAndEdges[0].SourceId.Format(time.RFC3339Nano))
	params.Add("tid", topics[0])

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, _ := http.NewRequest(http.MethodGet, url, nil)

	resp, err := client.Do(req)
	require.Nil(t, err)
	defer resp.Body.Close()
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	var data openapi.NodeData
	decoder := json.NewDecoder(resp.Body)
	_ = decoder.Decode(&data)

	require.Equal(t, nodesAndEdges[0].SourceId.Format(time.RFC3339Nano), data.Id.Format(time.RFC3339Nano))

}
