// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

//go:build wireinject
// +build wireinject

package application

import (
	"github.com/google/wire"

	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/infra/fileserver"
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth/authservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/authn"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/openapi"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/space"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/user"
	"code.byted.org/flowdevops/cozeloop/backend/modules/foundation/domain/user/service"
	auth2 "code.byted.org/flowdevops/cozeloop/backend/modules/foundation/infra/auth"
	"code.byted.org/flowdevops/cozeloop/backend/modules/foundation/infra/repo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/foundation/infra/repo/mysql"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

var (
	userDomainSet = wire.NewSet(
		service.NewUserService,
		repo.NewUserRepo,
		mysql.NewUserDAOImpl,
		mysql.NewSpaceDAOImpl,
		mysql.NewSpaceUserDAOImpl,
	)

	userSet = wire.NewSet(
		NewUserApplication,
		userDomainSet,
	)

	spaceSet = wire.NewSet(
		NewSpaceApplication,
		userDomainSet,
	)

	authSet = wire.NewSet(
		NewAuthApplication,
		userDomainSet,
	)

	authNSet = wire.NewSet(
		NewAuthNApplication,
		repo.NewAuthNRepo,
		mysql.NewAuthNDAOImpl,
	)

	fileSet = wire.NewSet(
		NewFileApplication,
		auth2.NewAuthProvider,
	)

	openAPISet = wire.NewSet(
		NewFoundationOpenAPIApplication,
		auth2.NewAuthProvider,
	)
)

func InitAuthApplication(idgen idgen.IIDGenerator,
	db db.Provider,
) (auth.AuthService, error) {
	wire.Build(authSet)
	return nil, nil
}

func InitAuthNApplication(
	idgen idgen.IIDGenerator,
	db db.Provider,
) (authn.AuthNService, error) {
	wire.Build(authNSet)
	return nil, nil
}

func InitSpaceApplication(
	idgen idgen.IIDGenerator,
	db db.Provider,
) (space.SpaceService, error) {
	wire.Build(spaceSet)
	return nil, nil
}

func InitUserApplication(
	idgen idgen.IIDGenerator,
	db db.Provider,
	configFactory conf.IConfigLoaderFactory,
) (user.UserService, error) {
	wire.Build(userSet)
	return nil, nil
}

func InitFileApplication(
	objectStorage fileserver.BatchObjectStorage,
	authClient authservice.Client,
) (file.FileService, error) {
	wire.Build(fileSet)
	return nil, nil
}

func InitFoundationOpenAPIApplication(
	objectStorage fileserver.BatchObjectStorage,
	authClient authservice.Client,
) (openapi.FoundationOpenAPIService, error) {
	wire.Build(openAPISet)
	return nil, nil
}
