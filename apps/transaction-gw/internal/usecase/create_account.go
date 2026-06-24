package usecase

import (
	"context"
	"transaction-gw/internal/domain"
	"transaction-gw/internal/ports"
)

type CreateAccountUseCase struct {
	repo ports.AccountRepository
}

func NewCreateAccountUseCase(repo ports.AccountRepository) *CreateAccountUseCase {
	return &CreateAccountUseCase{repo: repo}
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, account *domain.Account) error {
	if err := account.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, account)
}
