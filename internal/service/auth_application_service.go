package service

import (
	"context"

	"iam/internal/dto"
	"iam/internal/model"
	"iam/internal/repository"
)

type AuthApplicationService interface {
	Create(ctx context.Context, req dto.CreateAuthApplicationRequest) (*dto.AuthApplicationResponse, error)
	Update(ctx context.Context, id uint64, req dto.UpdateAuthApplicationRequest) (*dto.AuthApplicationResponse, error)
	Delete(ctx context.Context, id uint64) error
	Get(ctx context.Context, id uint64) (*dto.AuthApplicationResponse, error)
	List(ctx context.Context, query dto.AuthApplicationListQuery) ([]dto.AuthApplicationResponse, error)
}

type authApplicationService struct {
	clients repository.OAuthClientRepository
}

func NewAuthApplicationService(clients repository.OAuthClientRepository) AuthApplicationService {
	return &authApplicationService{clients: clients}
}

func (s *authApplicationService) Create(ctx context.Context, req dto.CreateAuthApplicationRequest) (*dto.AuthApplicationResponse, error) {
	client := &model.OAuthClient{
		Name:         req.Name,
		Code:         req.Code,
		ClientID:     req.ClientID,
		ClientSecret: req.SecretKey,
		ResponseType: req.ResponseType,
		RedirectURI:  req.RedirectURI,
		Status:       req.Status,
		Remark:       req.Remark,
	}
	if err := s.clients.Create(ctx, client); err != nil {
		return nil, err
	}
	return toAuthApplicationResponse(client), nil
}

func (s *authApplicationService) Update(ctx context.Context, id uint64, req dto.UpdateAuthApplicationRequest) (*dto.AuthApplicationResponse, error) {
	client, err := s.clients.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	client.Name = req.Name
	client.Code = req.Code
	client.ClientID = req.ClientID
	client.ClientSecret = req.SecretKey
	client.ResponseType = req.ResponseType
	client.RedirectURI = req.RedirectURI
	client.Status = req.Status
	client.Remark = req.Remark
	if err := s.clients.Update(ctx, client); err != nil {
		return nil, err
	}
	return toAuthApplicationResponse(client), nil
}

func (s *authApplicationService) Delete(ctx context.Context, id uint64) error {
	return s.clients.Delete(ctx, id)
}

func (s *authApplicationService) Get(ctx context.Context, id uint64) (*dto.AuthApplicationResponse, error) {
	client, err := s.clients.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toAuthApplicationResponse(client), nil
}

func (s *authApplicationService) List(ctx context.Context, query dto.AuthApplicationListQuery) ([]dto.AuthApplicationResponse, error) {
	clients, err := s.clients.List(ctx, query.Keyword, query.Status)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AuthApplicationResponse, 0, len(clients))
	for _, client := range clients {
		result = append(result, *toAuthApplicationResponse(&client))
	}
	return result, nil
}

func toAuthApplicationResponse(client *model.OAuthClient) *dto.AuthApplicationResponse {
	return &dto.AuthApplicationResponse{
		ID:           client.ID,
		Name:         client.Name,
		Code:         client.Code,
		ClientID:     client.ClientID,
		SecretKey:    client.ClientSecret,
		ResponseType: client.ResponseType,
		RedirectURI:  client.RedirectURI,
		Status:       client.Status,
		Remark:       client.Remark,
	}
}
