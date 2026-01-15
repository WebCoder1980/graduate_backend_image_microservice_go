package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"user_microservice/internal/model"

	"github.com/Nerzal/gocloak/v13"
)

type Service struct {
	ctx      context.Context
	keycloak *gocloak.GoCloak
}

func NewService(ctx context.Context) (*Service, error) {
	res := Service{
		ctx:      ctx,
		keycloak: gocloak.NewClient(os.Getenv("keycloak_address")),
	}

	err := res.init()
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *Service) init() error {
	adminToken, err := s.loginAdmin()
	if err != nil {
		return err
	}

	_, err = s.validateRole(adminToken, "user")
	if err != nil {
		return err
	}

	_, err = s.validateRole(adminToken, "admin")
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) validateRole(adminToken *gocloak.JWT, name string) (*gocloak.Role, error) {
	role := gocloak.Role{
		Name:        gocloak.StringP(name),
		Description: gocloak.StringP(fmt.Sprintf("${role_%s}", name)),
	}
	_, _ = s.keycloak.CreateRealmRole(s.ctx, adminToken.AccessToken, os.Getenv("keycloak_realm"), role)

	return s.keycloak.GetRealmRole(s.ctx, adminToken.AccessToken, os.Getenv("keycloak_realm"), name)
}

func (s *Service) UserRegisterPost(body *model.UserRegisterRequest) error {
	adminToken, err := s.loginAdmin()
	if err != nil {
		return err
	}

	newUserModel := gocloak.User{
		Enabled:  gocloak.BoolP(true),
		Username: gocloak.StringP(body.Username),
	}

	userId, err := s.keycloak.CreateUser(s.ctx, adminToken.AccessToken, os.Getenv("keycloak_realm"), newUserModel)
	if err != nil {
		return err
	}

	err = s.keycloak.SetPassword(s.ctx, adminToken.AccessToken, userId, os.Getenv("keycloak_realm"), body.Password, false)
	if err != nil {
		err2 := s.keycloak.DeleteUser(s.ctx, adminToken.AccessToken, os.Getenv("keycloak_realm"), userId)
		if err2 != nil {
			return errors.New(err.Error() + ";" + err2.Error())
		}

		return err
	}

	role, err := s.keycloak.GetRealmRole(s.ctx, adminToken.AccessToken, os.Getenv("keycloak_realm"), "user")
	if err != nil {
		err2 := s.keycloak.DeleteUser(s.ctx, adminToken.AccessToken, os.Getenv("keycloak_realm"), userId)
		if err2 != nil {
			return errors.New(err.Error() + ";" + err2.Error())
		}

		return err
	}

	err = s.keycloak.AddRealmRoleToUser(s.ctx, adminToken.AccessToken, os.Getenv("keycloak_realm"), userId, []gocloak.Role{*role})
	if err != nil {
		err2 := s.keycloak.DeleteUser(s.ctx, adminToken.AccessToken, os.Getenv("keycloak_realm"), userId)
		if err2 != nil {
			return errors.New(err.Error() + ";" + err2.Error())
		}

		return err
	}

	return nil
}

func (s *Service) loginAdmin() (*gocloak.JWT, error) {
	token, err := s.keycloak.LoginAdmin(s.ctx, os.Getenv("keycloak_username"), os.Getenv("keycloak_password"), os.Getenv("keycloak_realm"))
	if err != nil {
		return nil, err
	}
	return token, nil
}
