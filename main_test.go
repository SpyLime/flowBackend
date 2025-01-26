package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/auth/token"
	"github.com/go-pkgz/lgr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

type TestClock struct {
	Current time.Time
}

func (t *TestClock) Now() time.Time {
	if t.Current.IsZero() {
		t.Current, _ = time.Parse(time.RFC3339, "2020-01-02T15:04:05Z")
	}
	t.Tick()
	lgr.Printf("DEBUG current time - %v", t.Current)
	return t.Current
}

func (t *TestClock) Tick() {
	t.Current = t.Current.Add(time.Millisecond)
}

func OpenTestDB(suffix string) (db *bolt.DB, teardown func()) {
	_ = os.Mkdir("testdata", 0755)
	ldb, err := bolt.Open("testdata/db"+suffix+".db", 0666, nil)
	if err != nil {
		lgr.Printf("FATAL cannot open %v", err)
		return nil, func() {}
	}

	return ldb, func() {
		ldb.Close()
		os.Remove("testdata/db" + suffix + ".db")
	}
}

var testLoginUser string

func SetTestLoginUser(username string) {
	testLoginUser = username
}

func InitTestServer(port int, db *bolt.DB, userName string, clock Clock) (teardown func()) {
	SetTestLoginUser(userName)
	mux := createRouterClock(db, clock)
	mux.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			user := token.User{
				Name: testLoginUser,
			}
			ctx := request.Context()
			ctx = context.WithValue(ctx, "user", user)
			handler.ServeHTTP(writer, request.WithContext(ctx))
		})
	})
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	l, _ := net.Listen("tcp", addr)
	ts := httptest.NewUnstartedServer(mux)

	ts.Listener = l
	ts.Start()
	return func() {
		ts.Close()
	}
}

func FullStartTestServer(dbSuffix string, port int, userName string) (db *bolt.DB, teardown func()) {
	return FullStartTestServerClock(dbSuffix, port, userName, &TestClock{})
}

func FullStartTestServerClock(dbSuffix string, port int, userName string, clock Clock) (db *bolt.DB, teardown func()) {
	db, dbTearDown := OpenTestDB(dbSuffix)
	InitDB(db, clock)
	netTearDown := InitTestServer(port, db, userName, clock)

	return db, func() {
		dbTearDown()
		netTearDown()
	}

}

func InitDB(db *bolt.DB, clock Clock) {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(KeyUsers))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucket([]byte(KeyTopics))

		return err
	})
}

// creates test data
func CreateTestData(db *bolt.DB, clock Clock, numUsers, numTopics, numNodes int) (users, topics []string, nodesAndEdges []openapi.ResponsePostNode, err error) {
	err = db.Update(func(tx *bolt.Tx) error {

		for i := 0; i < numUsers; i++ {
			rep, _ := strconv.Atoi(RandomString(3))
			user := openapi.UpdateUserRequest{
				Username:    RandomString(2),
				FirstName:   "d",
				LastName:    "d",
				Email:       "d@d.com",
				Role:        0,
				Reputation:  int32(rep),
				Description: "d",
			}
			userId, err := postUserTx(tx, user)
			if err != nil {
				return err
			}
			users = append(users, userId)
		}

		for i := 0; i < numTopics; i++ {
			topic := openapi.GetTopics200ResponseInner{
				Title: RandomString(6),
			}

			response, err := postTopicTx(tx, clock, topic)
			if err != nil {
				return err
			}

			topics = append(topics, response.Topic.Title)

			for j := 0; j < numNodes; j++ {
				node := openapi.AddTopic200ResponseNodeData{
					Id:        response.NodeData.Id,
					Topic:     topic.Title,
					CreatedBy: users[0],
				}
				nodeIds, err := postNodeTx(tx, clock, node)
				if err != nil {
					return err
				}

				nodesAndEdges = append(nodesAndEdges, nodeIds)
			}
		}

		return err
	})

	sort.Strings(users)
	sort.Strings(topics)

	return

}

func TestFirst(t *testing.T) {

	db, teardown := OpenTestDB("")
	defer teardown()

	if db == nil {
		t.Fatalf("db not opened")
	}
}

func TestSchema(t *testing.T) {
	db, teardown := OpenTestDB("")
	defer teardown()
	clock := TestClock{}
	InitDB(db, &clock)
	users, _, _, err := CreateTestData(db, &clock, 3, 2, 2)
	assert.Nil(t, err)

	response, err := getUser(db, users[0])
	assert.Nil(t, err)

	assert.NotNil(t, response.Username)
}

func TestCreateTestData(t *testing.T) {
	db, teardown := OpenTestDB("createTestData")
	clock := TestClock{}
	defer teardown()

	numUsers := 2
	numTopics := 3
	numNodes := 4

	users, topics, nodes, err := CreateTestData(db, &clock, numUsers, numTopics, numNodes)
	require.Nil(t, err)

	require.Equal(t, len(users), numUsers)
	require.Equal(t, len(topics), numTopics)
	require.Equal(t, len(nodes), numNodes*numTopics) //3 topics each get 4 nodes
}

func TestSeedDbSecured(t *testing.T) {

	db, teardown := OpenTestDB("-integration")
	defer teardown()

	// InitDefaultAccounts(db, &clock)
	// auth := initAuth(db, ServerConfig{
	// 	AdminPassword: "test1",
	// })
	mux, clock := createRouter(db)

	// m := auth.Middleware()
	// mux.Use(buildAuthMiddleware(m))
	mux.Handle("/admin/seedDb", seedDbHandler(db, clock))

	l, _ := net.Listen("tcp", "127.0.0.1:8088")

	ts := httptest.NewUnstartedServer(mux)
	assert.NoError(t, ts.Listener.Close())
	ts.Listener = l
	ts.Start()
	defer func() {
		ts.Close()
	}()

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodPost,
		"http://localhost:8088/admin/seedDb",
		nil)
	// req.Header.Add("Authorization", "Basic YWRtaW46dGVzdDE=")
	resp, err := client.Do(req)
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	topics, err := getTopics(db)
	require.Nil(t, err)

	require.NotEmpty(t, topics)

}
