package service

import (
	"errors"

	"github.com/google/uuid"
	umonocrypto "github.com/umono-cms/crypto"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/gorm"
)

var ErrSecretNotFound = errors.New("secret not found")

type SecretService struct {
	repo   *repository.SecretRepository
	crypto *umonocrypto.Secret
}

func NewSecretService(repo *repository.SecretRepository, crypto *umonocrypto.Secret) *SecretService {
	return &SecretService{
		repo:   repo,
		crypto: crypto,
	}
}

func (s *SecretService) Create(plaintext []byte) (models.Secret, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return models.Secret{}, err
	}

	secret := models.Secret{ID: id.String()}
	ciphertext, err := s.crypto.Encrypt(plaintext, []byte(secret.ID))
	if err != nil {
		return models.Secret{}, err
	}
	secret.Ciphertext = ciphertext

	return s.repo.Create(secret)
}

func (s *SecretService) GetByID(id string) (models.Secret, error) {
	secret, err := s.repo.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Secret{}, ErrSecretNotFound
	}
	if err != nil {
		return models.Secret{}, err
	}

	return secret, nil
}

func (s *SecretService) Decrypt(secret models.Secret) ([]byte, error) {
	return s.crypto.Decrypt(secret.Ciphertext, []byte(secret.ID))
}

func (s *SecretService) DecryptByID(id string) ([]byte, error) {
	secret, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.Decrypt(secret)
}

func (s *SecretService) Update(id string, plaintext []byte) (models.Secret, error) {
	secret, err := s.GetByID(id)
	if err != nil {
		return models.Secret{}, err
	}

	ciphertext, err := s.crypto.Encrypt(plaintext, []byte(secret.ID))
	if err != nil {
		return models.Secret{}, err
	}
	secret.Ciphertext = ciphertext

	return s.repo.Update(secret)
}

func (s *SecretService) Delete(id string) error {
	return s.repo.Delete(id)
}
