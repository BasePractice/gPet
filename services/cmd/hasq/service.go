package main

import (
	"context"

	"pet/middleware/hasq"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service struct {
	hasq.UnimplementedServiceServer
	db DatabaseToken
}

func (s *service) CreateToken(_ context.Context, tc *hasq.TokenCreate) (*hasq.TokenReply, error) {
	err, t := s.db.CreateToken(tc.Title, tc.Data)
	if err != nil {
		return nil, err
	}
	return &hasq.TokenReply{
		Id:    t.Id.String(),
		Title: t.Title,
		Hash:  t.Hash,
	}, nil
}

func (s *service) SearchToken(_ context.Context, ts *hasq.TokenSearch) (*hasq.TokenReply, error) {
	var id *uuid.UUID
	var hash *string

	if ts.GetId() != "" {
		uid, err := uuid.Parse(ts.GetId())
		if err != nil {
			return nil, err
		}
		id = &uid
	} else if ts.GetHash() != "" {
		hs := ts.GetHash()
		hash = &hs
	} else {
		return nil, status.Error(codes.NotFound, "Token not found")
	}
	err, t := s.db.SearchToken(id, hash)
	if err != nil {
		return nil, err
	}
	return &hasq.TokenReply{
		Id:    t.Id.String(),
		Title: t.Title,
		Hash:  t.Hash,
		Data:  t.Data,
	}, nil
}
