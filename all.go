package main

import (
	"context"

	openapi "github.com/SpyLime/flowBackend/go"
	bolt "go.etcd.io/bbolt"
)

type AllAPIServiceImpl struct {
	db    *bolt.DB
	clock Clock
}

func NewAllAPIServiceImpl(db *bolt.DB, clock Clock) openapi.AllAPIServicer {
	return &AllAPIServiceImpl{
		db:    db,
		clock: clock,
	}
}

func (s *AllAPIServiceImpl) ClipImage(ctx context.Context, clipUrl string) (openapi.ImplResponse, error) {
	info, err := fetchClipTitle(clipUrl)
	if err != nil {
		return openapi.Response(400, nil), err
	}
	return openapi.Response(200, info), nil
}
