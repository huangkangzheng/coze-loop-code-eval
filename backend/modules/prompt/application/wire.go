// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

//go:build wireinject
// +build wireinject

package application

import (
	"github.com/google/wire"

	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/infra/external/audit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/external/benefit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/infra/limiter"
	"code.byted.org/flowdevops/cozeloop/backend/infra/metrics"
	"code.byted.org/flowdevops/cozeloop/backend/infra/redis"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth/authservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file/fileservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/user/userservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/llm/runtime/llmruntimeservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/prompt/debug"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/prompt/execute"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/prompt/manage"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/prompt/openapi"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/domain/service"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/infra/collector"
	promptconf "code.byted.org/flowdevops/cozeloop/backend/modules/prompt/infra/conf"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/infra/repo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/infra/repo/mysql"
	rediscache "code.byted.org/flowdevops/cozeloop/backend/modules/prompt/infra/repo/redis"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/infra/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

var (
	promptDomainSet = wire.NewSet(
		service.NewPromptService,
		repo.NewManageRepo,
		repo.NewLabelRepo,
		repo.NewDebugLogRepo,
		repo.NewDebugContextRepo,
		mysql.NewPromptBasicDAO,
		mysql.NewPromptCommitDAO,
		mysql.NewPromptUserDraftDAO,
		mysql.NewLabelDAO,
		mysql.NewCommitLabelMappingDAO,
		mysql.NewDebugLogDAO,
		mysql.NewDebugContextDAO,
		rediscache.NewPromptBasicDAO,
		rediscache.NewPromptDAO,
		rediscache.NewPromptLabelVersionDAO,
		promptconf.NewPromptConfigProvider,
		rpc.NewLLMRPCProvider,
		rpc.NewAuthRPCProvider,
		rpc.NewFileRPCProvider,
		rpc.NewUserRPCProvider,
		rpc.NewAuditRPCProvider,
		collector.NewEventCollectorProvider,
	)
	manageSet = wire.NewSet(
		NewPromptManageApplication,
		promptDomainSet,
	)
	debugSet = wire.NewSet(
		NewPromptDebugApplication,
		promptDomainSet,
	)
	executeSet = wire.NewSet(
		NewPromptExecuteApplication,
		promptDomainSet,
	)
	openAPISet = wire.NewSet(
		NewPromptOpenAPIApplication,
		promptDomainSet,
	)
)

func InitPromptManageApplication(
	idgen idgen.IIDGenerator,
	db db.Provider,
	redisCli redis.Cmdable,
	meter metrics.Meter,
	configFactory conf.IConfigLoaderFactory,
	llmClient llmruntimeservice.Client,
	authClient authservice.Client,
	fileClient fileservice.Client,
	userClient userservice.Client,
	auditClient audit.IAuditService,
) (manage.PromptManageService, error) {
	wire.Build(manageSet)
	return nil, nil
}

func InitPromptDebugApplication(
	idgen idgen.IIDGenerator,
	db db.Provider,
	redisCli redis.Cmdable,
	meter metrics.Meter,
	configFactory conf.IConfigLoaderFactory,
	llmClient llmruntimeservice.Client,
	authClient authservice.Client,
	fileClient fileservice.Client,
	benefitSvc benefit.IBenefitService,
) (debug.PromptDebugService, error) {
	wire.Build(debugSet)
	return nil, nil
}

func InitPromptExecuteApplication(
	idgen idgen.IIDGenerator,
	db db.Provider,
	redisCli redis.Cmdable,
	meter metrics.Meter,
	configFactory conf.IConfigLoaderFactory,
	llmClient llmruntimeservice.Client,
	fileClient fileservice.Client,
) (execute.PromptExecuteService, error) {
	wire.Build(executeSet)
	return nil, nil
}

func InitPromptOpenAPIApplication(
	idgen idgen.IIDGenerator,
	db db.Provider,
	redisCli redis.Cmdable,
	meter metrics.Meter,
	configFactory conf.IConfigLoaderFactory,
	limiterFactory limiter.IRateLimiterFactory,
	llmClient llmruntimeservice.Client,
	authClient authservice.Client,
	fileClient fileservice.Client,
) (openapi.PromptOpenAPIService, error) {
	wire.Build(openAPISet)
	return nil, nil
}
