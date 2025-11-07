// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

//go:build wireinject
// +build wireinject

package apis

import (
	"context"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/google/wire"

	"code.byted.org/flowdevops/cozeloop/backend/infra/ck"
	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/infra/external/audit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/external/benefit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/fileserver"
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/infra/limiter"
	"code.byted.org/flowdevops/cozeloop/backend/infra/metrics"
	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/infra/redis"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/apis/promptexecuteservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/dataset/datasetservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/tag/tagservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/evaluation/evaluationsetservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/evaluation/evaluatorservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth/authservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file/fileservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/user/userservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/llm/runtime/llmruntimeservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/prompt/promptmanageservice"
	"code.byted.org/flowdevops/cozeloop/backend/loop_gen/coze/loop/foundation/loauth"
	dataapp "code.byted.org/flowdevops/cozeloop/backend/modules/data/application"
	conf2 "code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/conf"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/rpc/foundation"
	evaluationapp "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/application"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/data"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/prompt"
	foundationapp "code.byted.org/flowdevops/cozeloop/backend/modules/foundation/application"
	llmapp "code.byted.org/flowdevops/cozeloop/backend/modules/llm/application"
	obapp "code.byted.org/flowdevops/cozeloop/backend/modules/observability/application"
	promptapp "code.byted.org/flowdevops/cozeloop/backend/modules/prompt/application"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

var (
	foundationSet = wire.NewSet(
		NewFoundationHandler,
		foundationapp.InitAuthApplication,
		foundationapp.InitAuthNApplication,
		foundationapp.InitSpaceApplication,
		foundationapp.InitUserApplication,
		foundationapp.InitFileApplication,
		foundationapp.InitFoundationOpenAPIApplication,
		wire.Value([]endpoint.Middleware(nil)),
		wire.Bind(new(authservice.Client), new(*loauth.LocalAuthService)),
		loauth.NewLocalAuthService,
	)
	llmSet = wire.NewSet(
		NewLLMHandler,
		llmapp.InitManageApplication,
		llmapp.InitRuntimeApplication,
	)
	promptSet = wire.NewSet(
		NewPromptHandler,
		promptapp.InitPromptManageApplication,
		promptapp.InitPromptDebugApplication,
		promptapp.InitPromptExecuteApplication,
		promptapp.InitPromptOpenAPIApplication,
	)
	evaluationSet = wire.NewSet(
		NewEvaluationHandler,
		data.NewDatasetRPCAdapter,
		prompt.NewPromptRPCAdapter,
		evaluationapp.InitExperimentApplication,
		evaluationapp.InitEvaluatorApplication,
		evaluationapp.InitEvaluationSetApplication,
		evaluationapp.InitEvalTargetApplication,
	)
	dataSet = wire.NewSet(
		NewDataHandler,
		dataapp.InitDatasetApplication,
		dataapp.InitTagApplication,
		foundation.NewAuthRPCProvider,
		conf2.NewConfigerFactory,
	)
	observabilitySet = wire.NewSet(
		NewObservabilityHandler,
		obapp.InitTraceApplication,
		obapp.InitTraceIngestionApplication,
		obapp.InitOpenAPIApplication,
	)
)

func InitFoundationHandler(
	idgen idgen.IIDGenerator,
	db db.Provider,
	objectStorage fileserver.BatchObjectStorage,
	configFactory conf.IConfigLoaderFactory,
) (*FoundationHandler, error) {
	wire.Build(
		foundationSet,
	)
	return nil, nil
}

func InitPromptHandler(
	ctx context.Context,
	idgen idgen.IIDGenerator,
	db db.Provider,
	redisCli redis.Cmdable,
	meter metrics.Meter,
	configFactory conf.IConfigLoaderFactory,
	limiterFactory limiter.IRateLimiterFactory,
	benefitSvc benefit.IBenefitService,
	llmClient llmruntimeservice.Client,
	authClient authservice.Client,
	fileClient fileservice.Client,
	userClient userservice.Client,
	auditClient audit.IAuditService,
) (*PromptHandler, error) {
	wire.Build(
		promptSet,
	)
	return nil, nil
}

func InitLLMHandler(
	ctx context.Context,
	idgen idgen.IIDGenerator,
	db db.Provider,
	cmdable redis.Cmdable,
	configFactory conf.IConfigLoaderFactory,
	limiterFactory limiter.IRateLimiterFactory,
	authClient authservice.Client,
) (*LLMHandler, error) {
	wire.Build(
		llmSet,
	)
	return nil, nil
}

func InitEvaluationHandler(
	ctx context.Context,
	idgen idgen.IIDGenerator,
	db db.Provider,
	ckDb ck.Provider,
	cmdable redis.Cmdable,
	configFactory conf.IConfigLoaderFactory,
	mqFactory mq.IFactory,
	client datasetservice.Client,
	promptClient promptmanageservice.Client,
	pec promptexecuteservice.Client,
	authClient authservice.Client,
	meter metrics.Meter,
	auditClient audit.IAuditService,
	llmClient llmruntimeservice.Client,
	userClient userservice.Client,
	benefitSvc benefit.IBenefitService,
	limiterFactory limiter.IRateLimiterFactory,
	fileClient fileservice.Client,
	tagClient tagservice.Client,
	objectStorage fileserver.ObjectStorage,
) (*EvaluationHandler, error) {
	wire.Build(
		evaluationSet,
	)
	return nil, nil
}

func InitDataHandler(
	ctx context.Context,
	idgen idgen.IIDGenerator,
	db db.Provider,
	redisCli redis.Cmdable,
	configFactory conf.IConfigLoaderFactory,
	mqFactory mq.IFactory,
	objectStorage fileserver.ObjectStorage,
	batchObjectStorage fileserver.BatchObjectStorage,
	auditClient audit.IAuditService,
	auth authservice.Client,
	userClient userservice.Client,
) (*DataHandler, error) {
	wire.Build(
		dataSet,
	)
	return nil, nil
}

func InitObservabilityHandler(
	ctx context.Context,
	db db.Provider,
	ckDb ck.Provider,
	meter metrics.Meter,
	mqFactory mq.IFactory,
	configFactory conf.IConfigLoaderFactory,
	idgen idgen.IIDGenerator,
	benefit benefit.IBenefitService,
	fileClient fileservice.Client,
	authCli authservice.Client,
	userClient userservice.Client,
	evalClient evaluatorservice.Client,
	evalSetClient evaluationsetservice.Client,
	tagClient tagservice.Client,
	limiterFactory limiter.IRateLimiterFactory,
	datasetClient datasetservice.Client,
) (*ObservabilityHandler, error) {
	wire.Build(
		observabilitySet,
	)
	return nil, nil
}
