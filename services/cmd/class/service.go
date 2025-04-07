package main

import (
	"context"
	"log"
	"log/slog"

	"pet/middleware/class"
	"pet/services"
)

type service struct {
	class.UnimplementedServiceServer
	db    DatabaseClass
	cache services.Cache
}

func (s *service) Classes(ctx context.Context, request *class.ClassRequest) (*class.ClassReply, error) {
	services.PrintMetadata(ctx)
	var status *string = nil
	if request.Status != nil {
		s2 := request.GetStatus().String()
		status = &s2
	}
	classes, err := s.db.Classes(request.NameFilter, status, request.Version)
	if err != nil {
		slog.Error("Get classes error", slog.String("err", err.Error()))
		return nil, err
	}
	var reply class.ClassReply
	for _, element := range classes {
		if reply.Classes == nil {
			reply.Classes = make([]*class.Class, 0)
		}
		reply.Classes = append(reply.Classes, &class.Class{
			Name:   element.Name,
			Title:  element.Title,
			Status: class.ClassStatusFromSql(element.Status),
		})
	}
	return &reply, nil
}

func (s *service) Elements(ctx context.Context, request *class.ClassElementRequest) (*class.ClassElementReply, error) {
	services.PrintMetadata(ctx)
	c, err := s.db.Class(request.Name)
	if err != nil {
		slog.Error("Get class error ", slog.String("err", err.Error()))
		return nil, err
	}
	var status *string = nil
	if request.Status != nil {
		s2 := request.GetStatus().String()
		status = &s2
	}
	var offset = 0
	if request.Offset != nil {
		offset = int(*request.Offset)
	}
	var limit = 100
	if request.Limit != nil {
		limit = int(*request.Limit)
	}
	elements, next, err := s.db.Elements(*c, request.Version, status, offset, limit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var reply class.ClassElementReply
	for _, element := range elements {
		if reply.Elements == nil {
			reply.Elements = make([]*class.ClassElement, 0)
		}
		reply.Elements = append(reply.Elements, &class.ClassElement{
			Key:     element.Key,
			Value:   element.Value,
			Status:  class.ElementStatusFromSql(element.Status),
			Version: element.Version,
		})
	}
	reply.NextOffset = uint32(next)
	reply.Eof = len(elements) < limit
	return &reply, nil
}
