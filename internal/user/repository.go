package user

import (
	"errors"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(u *User) error
	GetByID(id string) (*User, error)
	GetByProviderID(providerID string) (*User, error)
	GetUserEncryptedGoogleCalendarAccessToken(id string) (string, error)
	Update(u *User) error
	Delete(id string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(u *User) error {
	return r.db.Create(u).Error
}

func (r *userRepository) GetByID(id string) (*User, error) {
	var u User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetUserEncryptedGoogleCalendarAccessToken(id string) (string, error) {
	var u User
	if err := r.db.Select("encrypted_google_access_token").First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.EncryptedGoogleAccessToken, nil
}

func (r *userRepository) GetByProviderID(providerID string) (*User, error) {
	var u User
	if err := r.db.First(&u, "provider_id = ?", providerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Update(u *User) error {
	return r.db.Save(u).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}
