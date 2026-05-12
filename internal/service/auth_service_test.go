package service

import (
	"context"
	"testing"
	"time"

	"iam/internal/dto"
	"iam/internal/model"
	jwtpkg "iam/internal/pkg/jwt"
	"iam/internal/pkg/password"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type fakeUserRepo struct {
	user  *model.User
	roles []model.Role
}

func (f *fakeUserRepo) Create(context.Context, *model.User) error            { return nil }
func (f *fakeUserRepo) GetByID(context.Context, uint64) (*model.User, error) { return f.user, nil }
func (f *fakeUserRepo) GetByUsername(context.Context, string) (*model.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) List(context.Context, string, int) ([]model.User, error) {
	return []model.User{}, nil
}
func (f *fakeUserRepo) UpdateStatus(context.Context, uint64, int) error        { return nil }
func (f *fakeUserRepo) Delete(context.Context, uint64) error                   { return nil }
func (f *fakeUserRepo) SetRoles(context.Context, uint64, []model.Role) error   { return nil }
func (f *fakeUserRepo) GetRoles(context.Context, uint64) ([]model.Role, error) { return f.roles, nil }

type fakeIdentityRepo struct{ identity *model.AuthIdentity }

func (f *fakeIdentityRepo) Create(context.Context, *model.AuthIdentity) error { return nil }
func (f *fakeIdentityRepo) GetByIdentity(context.Context, string, string) (*model.AuthIdentity, error) {
	return f.identity, nil
}
func (f *fakeIdentityRepo) UpdateCredential(context.Context, uint64, string, string) error {
	return nil
}

func TestAuthServiceLogin(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	hash, err := password.Hash("123456", 4)
	if err != nil {
		t.Fatal(err)
	}
	service := NewAuthService(
		&fakeUserRepo{user: &model.User{Base: model.Base{ID: 1}, Username: "admin", DisplayName: "管理员", Status: 1, Roles: []model.Role{{Code: "admin"}}}},
		&fakeIdentityRepo{identity: &model.AuthIdentity{UserID: 1, IdentityType: "password", Identifier: "admin", Credential: hash}},
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		jwtpkg.NewManager("iam", "secret", int64(time.Hour.Seconds())),
		5,
		900,
	)

	_, err = service.Login(context.Background(), dto.LoginRequest{Username: "admin", Password: "123456"})
	if err == nil {
		return
	}
	t.Fatalf("expected login success, got %v", err)
}
