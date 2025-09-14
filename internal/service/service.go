package service

import "inHack/internal/repository"

type Service struct {
	repository *repository.Repository
}

func NewService(reposit *repository.Repository) *Service {
	return &Service{repository: reposit}
}
