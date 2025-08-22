package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

func backUpHandler(db *bolt.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		err := db.View(func(tx *bolt.Tx) error {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="my.db"`)
			w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
			_, err := tx.WriteTo(w)
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

type NewSchoolRequest struct {
	School    string
	FirstName string
	LastName  string
	Email     string
	City      string
	Zip       int
}

type NewSchoolResponse struct {
	AdminPassword string
}

type ResetPasswordRequest struct {
	Email string
}

type EventRequest struct {
	Positive    bool `json:",omitempty"`
	Description string
	Title       string `json:",omitempty"`
}

func UpdateUserRoleAndReputation(db *bolt.DB, userId string, isAdmin bool, reputation int32) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Get users bucket
		usersBucket := tx.Bucket([]byte(KeyUsers))
		if usersBucket == nil {
			return fmt.Errorf("users bucket not found")
		}

		// Get user data
		userData := usersBucket.Get([]byte(userId))
		if userData == nil {
			return fmt.Errorf("user %s not found", userId)
		}

		var user openapi.User
		if err := json.Unmarshal(userData, &user); err != nil {
			return fmt.Errorf("failed to unmarshal user data: %v", err)
		}

		// Update role
		if isAdmin {
			user.Role = KeyAdmin
		} else {
			user.Role = KeyUser
		}

		// Update reputation with direct value
		user.Reputation = reputation

		// Save updated user data
		updatedData, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("failed to marshal updated user data: %v", err)
		}

		return usersBucket.Put([]byte(userId), updatedData)
	})
}

func CreateTestData(db *bolt.DB, clock Clock, numUsers, numTopics, numNodes int) (users, topics []string, nodesAndEdges []openapi.ResponsePostNode, err error) {
	if numUsers == 0 && numNodes > 0 {
		return users, topics, nodesAndEdges, fmt.Errorf("you can't create nodes without a user")
	}

	if numUsers == 0 && numTopics > 0 {
		return users, topics, nodesAndEdges, fmt.Errorf("you can't create topics without a user")
	}

	if numTopics == 0 && numNodes > 0 {
		return users, topics, nodesAndEdges, fmt.Errorf("you can't create nodes without a topic")
	}

	err = db.Update(func(tx *bolt.Tx) error {

		for i := 0; i < numUsers; i++ {
			rep, _ := strconv.Atoi(RandomString(3))
			user := openapi.User{
				Username:    RandomString(2),
				FirstName:   "d",
				LastName:    "d",
				Email:       "d@d.com",
				Role:        KeyAdmin,
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
			topic := openapi.Topic{
				Title: RandomString(6),
			}

			user := openapi.User{
				Id:       users[0],
				Username: users[0],
			}

			response, err := postTopicTx(tx, clock, topic, user)
			if err != nil {
				return err
			}

			nodesAndEdges = append(nodesAndEdges, openapi.ResponsePostNode{SourceId: response.NodeData.Id})

			topics = append(topics, response.Topic.Title)

			for j := 0; j < numNodes; j++ {
				node := openapi.NodeData{
					Id:    response.NodeData.Id,
					Topic: topic.Title,
					CreatedBy: openapi.UserIdentifier{
						Id:       users[0],
						Username: "tester",
					},
				}

				clock.Tick()

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
