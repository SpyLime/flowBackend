package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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
	lgr.Printf("DEBUG current time - %v", t.Current)
	return t.Current
}

func (t *TestClock) Tick() {
	t.Current = t.Current.Add(time.Millisecond)
}
func (t *TestClock) TickOne(d time.Duration) {
	t.Current = t.Current.Add(d)
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

func SetTestLoginUser(id string) {
	testLoginUser = id
}

func InitTestServer(port int, db *bolt.DB, id string, clock Clock) (teardown func()) {
	SetTestLoginUser(id)
	mux := createRouterClock(db, clock)
	mux.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			user := token.User{
				ID: testLoginUser,
			}
			ctx := request.Context()
			ctx = context.WithValue(ctx, userInfoKey, user)
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
//
// when a topic is made a root node is made, if numNodes > 0 then all new nodes will be connected to the root node
//
// the root nodes is included in nodesAndEdges and that is why there is 1 element that has a source but no target
//
// nodesAndEdges[0].targetId is invalid
//
// nodesAndEdges[x].sourceId is the id of the current node while targetId is the id of the node its forward point is connecting to

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
	require.Equal(t, len(nodes), (numNodes*numTopics)+numTopics) //3 topics each get 4 nodes then add the root node for each topic
}
