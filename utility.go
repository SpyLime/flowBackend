package main

import (
	"crypto/rand"
	"math/big"
	"time"
)

// Clock is an interface that provides the current time.
type Clock interface {
	Now() time.Time
}

const (
	KeyUsers                 = "users"
	KeyTopics                = "topics"
	KeyNodes                 = "nodes"
	KeyEdges                 = "edges"
	KeyUser                  = 0
	KeyAdmin                 = 1
	KeyReputationDeleter     = 200
	KeyReputationEditor      = 100
	KeyReputationContributor = 50
)

func RandomString(n int) string {
	letters := "0123456789"
	//
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}
