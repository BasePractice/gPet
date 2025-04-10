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

func (s *service) Validate(_ context.Context, cv *hasq.ChainValidate) (*hasq.ChainValidateReply, error) {
	tokenId, err := uuid.Parse(cv.TokenId)
	if err != nil {
		return nil, err
	}
	result, err := s.db.Validate(tokenId)
	if err != nil {
		return nil, err
	}
	return &hasq.ChainValidateReply{
		Successful: result.Successful,
		LastNum:    result.LastNum,
		OwnerId:    result.OwnerId.String(),
	}, nil
}

func (s *service) Owned(_ context.Context, own *hasq.OwnerCreate) (*hasq.OwnerCreateReply, error) {
	tokenId, err := uuid.Parse(own.TokenId)
	if err != nil {
		return nil, err
	}
	userId, err := uuid.Parse(own.UserId)
	if err != nil {
		return nil, err
	}
	err = s.db.Owner(userId, tokenId)
	if err != nil {
		return nil, err
	}
	return &hasq.OwnerCreateReply{
		Successful: true,
	}, nil
}

func (s *service) CreateKey(_ context.Context, kc *hasq.KeyCreate) (*hasq.KeyCreateReply, error) {
	tokenId, err := uuid.Parse(kc.TokenId)
	if err != nil {
		return nil, err
	}
	userId, err := uuid.Parse(kc.UserId)
	if err != nil {
		return nil, err
	}
	k, err := s.db.CreateKey(userId, tokenId, kc.Passphrase)
	if err != nil {
		return nil, err
	}
	return &hasq.KeyCreateReply{
		KeyId: k.Id.String(),
		Hash:  k.Hash,
	}, nil
}

func (s *service) CreateToken(_ context.Context, tc *hasq.TokenCreate) (*hasq.TokenReply, error) {
	t, err := s.db.CreateToken(tc.Title, tc.Data)
	if err != nil {
		return nil, err
	}
	return &hasq.TokenReply{
		TokenId: t.Id.String(),
		Title:   t.Title,
		Hash:    t.Hash,
	}, nil
}

func (s *service) SearchToken(_ context.Context, ts *hasq.TokenSearch) (*hasq.TokenReply, error) {
	var id *uuid.UUID
	var hash *string

	if ts.GetTokenId() != "" {
		uid, err := uuid.Parse(ts.GetTokenId())
		if err != nil {
			return nil, err
		}
		id = &uid
	} else if ts.GetTokenHash() != "" {
		hs := ts.GetTokenHash()
		hash = &hs
	} else {
		return nil, status.Error(codes.NotFound, "Token not found")
	}
	t, err := s.db.SearchToken(id, hash)
	if err != nil {
		return nil, err
	}
	return &hasq.TokenReply{
		TokenId: t.Id.String(),
		Title:   t.Title,
		Hash:    t.Hash,
		Data:    t.Data,
	}, nil
}
