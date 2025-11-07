// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

//go:build wireinject
// +build wireinject

package application

import (
	"context"

	"github.com/google/wire"

	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/infra/limiter"
	"code.byted.org/flowdevops/cozeloop/backend/infra/redis"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth/authservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/llm/manage"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/llm/runtime"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/service"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/service/llmfactory"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/infra/config"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/infra/repo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/infra/repo/dao"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/infra/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

var (
	llmDomainSet = wire.NewSet(
		llmfactory.NewFactory,
		config.NewManage,
		config.NewRuntime,
		service.NewRuntime,
		service.NewManage,
		repo.NewRuntimeRepo,
		dao.NewModelRequestRecordDao,
		rpc.NewAuthRPCProvider,
	)
	runtimeSet = wire.NewSet(
		NewRuntimeApplication,
		llmDomainSet,
	)
	manageSet = wire.NewSet(
		NewManageApplication,
		llmDomainSet,
	)
)

func InitRuntimeApplication(
	ctx context.Context,
	idGen idgen.IIDGenerator,
	configFactory conf.IConfigLoaderFactory,
	db db.Provider,
	redis redis.Cmdable,
	factory limiter.IRateLimiterFactory) (runtime.LLMRuntimeService, error) {
	wire.Build(runtimeSet)
	return nil, nil
}

func InitManageApplication(
	ctx context.Context,
	configFactory conf.IConfigLoaderFactory,
	authClient authservice.Client) (manage.LLMManageService, error) {
	wire.Build(manageSet)
	return nil, nil
}
