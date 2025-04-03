package main

import (
	"context"
	"log"

	"pet/middleware/class"
)

type service struct {
	class.UnimplementedServiceServer
	db DatabaseClass
}

func classStatusOfString(key string) class.ClassStatus {
	switch key {
	case "DRAFT":
		return class.ClassStatus_CLASS_DRAFT
	case "PUBLISHED":
		return class.ClassStatus_CLASS_PUBLISHED
	case "ARCHIVED":
		return class.ClassStatus_CLASS_ARCHIVED
	default:
		return class.ClassStatus_CLASS_NONE
	}
}

func elementStatusOfString(key string) class.ClassElementStatus {
	switch key {
	case "DRAFT":
		return class.ClassElementStatus_ITEM_DRAFT
	case "PUBLISHED":
		return class.ClassElementStatus_ITEM_PUBLISHED
	case "SKIP":
		return class.ClassElementStatus_ITEM_SKIP
	default:
		return class.ClassElementStatus_ITEM_NONE
	}
}

func classStringOfStatus(key class.ClassStatus) string {
	switch key {
	case class.ClassStatus_CLASS_DRAFT:
		return "DRAFT"
	case class.ClassStatus_CLASS_PUBLISHED:
		return "PUBLISHED"
	case class.ClassStatus_CLASS_ARCHIVED:
		return "ARCHIVED"
	default:
		return "NONE"
	}
}

func elementStringOfStatus(key class.ClassElementStatus) string {
	switch key {
	case class.ClassElementStatus_ITEM_DRAFT:
		return "DRAFT"
	case class.ClassElementStatus_ITEM_PUBLISHED:
		return "PUBLISHED"
	case class.ClassElementStatus_ITEM_SKIP:
		return "SKIP"
	default:
		return "NONE"
	}
}

func (s *service) Classes(_ context.Context, request *class.ClassRequest) (*class.ClassReply, error) {
	var status *string = nil
	if request.Status != nil {
		s2 := classStringOfStatus(request.GetStatus())
		status = &s2
	}
	classes, err := s.db.Classes(request.NameFilter, status, request.Version)
	if err != nil {
		log.Println(err)
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
			Status: classStatusOfString(element.Status),
		})
	}
	return &reply, nil
}

func (s *service) Elements(_ context.Context, request *class.ClassElementRequest) (*class.ClassElementReply, error) {
	c, err := s.db.Class(request.Name)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var status *string = nil
	if request.Status != nil {
		s2 := elementStringOfStatus(request.GetStatus())
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
			Status:  elementStatusOfString(element.Status),
			Version: element.Version,
		})
	}
	reply.NextOffset = uint32(next)
	reply.Eof = len(elements) < limit
	return &reply, nil
}
