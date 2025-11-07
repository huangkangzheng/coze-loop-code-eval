// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

//go:build wireinject
// +build wireinject

package application

import (
	"context"

	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/notify"
	"github.com/google/wire"

	"code.byted.org/flowdevops/cozeloop/backend/infra/ck"
	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/infra/external/audit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/external/benefit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/fileserver"
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/infra/limiter"
	"code.byted.org/flowdevops/cozeloop/backend/infra/lock"
	"code.byted.org/flowdevops/cozeloop/backend/infra/metrics"
	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/infra/platestwrite"
	"code.byted.org/flowdevops/cozeloop/backend/infra/redis"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/apis/promptexecuteservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/dataset/datasetservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/tag/tagservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/evaluation"
	evaluationservice "code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/evaluation"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth/authservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file/fileservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/user/userservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/llm/runtime/llmruntimeservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/prompt/promptmanageservice"
	mtr "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/component/metrics"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/component/rpc"
	componentrpc "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/component/userinfo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/service"
	domainservice "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/service"
	evaltargetmtr "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/metrics/eval_target"
	evalsetmtr "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/metrics/evaluation_set"
	evaluatormtr "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/metrics/evaluator"
	exptmtr "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/metrics/experiment"
	rmqproducer "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/mq/rocket/producer"
	evaluatorrepo "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/evaluator"
	evaluatormysql "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/evaluator/mysql"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/experiment"
	exptck "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/experiment/ck"
	exptmysql "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/experiment/mysql"
	exptredis "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/experiment/redis/dao"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/idem"
	iredis "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/idem/redis"
	targetrepo "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/target"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/target/mysql"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/agent"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/data"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/foundation"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/llm"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/prompt"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/rpc/tag"
	evalconf "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/pkg/conf"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

var (
	flagSet = wire.NewSet(
		platestwrite.NewLatestWriteTracker,
	)

	experimentSet = wire.NewSet(
		NewExperimentApplication,
		domainservice.NewExptManager,
		domainservice.NewExptResultService,
		domainservice.NewExptAggrResultService,
		domainservice.NewExptSchedulerSvc,
		domainservice.NewExptRecordEvalService,
		domainservice.NewExptAnnotateService,
		domainservice.NewExptResultExportService,
		domainservice.NewInsightAnalysisService,
		domainservice.NewSchedulerModeFactory,
		experiment.NewExptRepo,
		experiment.NewExptStatsRepo,
		experiment.NewExptAggrResultRepo,
		experiment.NewExptItemResultRepo,
		experiment.NewExptTurnResultRepo,
		experiment.NewExptRunLogRepo,
		experiment.NewExptTurnResultFilterRepo,
		experiment.NewExptAnnotateRepo,
		experiment.NewExptResultExportRecordRepo,
		experiment.NewExptInsightAnalysisRecordRepo,
		experiment.NewQuotaService,
		idem.NewIdempotentService,
		exptmysql.NewExptDAO,
		exptmysql.NewExptEvaluatorRefDAO,
		exptmysql.NewExptRunLogDAO,
		exptmysql.NewExptStatsDAO,
		exptmysql.NewExptTurnResultDAO,
		exptmysql.NewExptItemResultDAO,
		exptmysql.NewExptTurnEvaluatorResultRefDAO,
		exptmysql.NewExptTurnResultFilterKeyMappingDAO,
		exptmysql.NewExptAggrResultDAO,
		exptmysql.NewExptTurnAnnotateRecordRefDAO,
		exptmysql.NewAnnotateRecordDAO,
		exptmysql.NewExptTurnResultTagRefDAO,
		exptmysql.NewExptResultExportRecordDAO,
		exptmysql.NewExptInsightAnalysisRecordDAO,
		exptmysql.NewExptInsightAnalysisFeedbackVoteDAO,
		exptmysql.NewExptInsightAnalysisFeedbackCommentDAO,
		exptredis.NewQuotaDAO,
		iredis.NewIdemDAO,
		exptck.NewExptTurnResultFilterDAO,
		evalconf.NewExptConfiger,
		rmqproducer.NewExptEventPublisher,
		exptmtr.NewExperimentMetric,
		evaltargetmtr.NewEvalTargetMetrics,
		foundation.NewAuthRPCProvider,
		foundation.NewUserRPCProvider,
		tag.NewTagRPCProvider,
		agent.NewAgentAdapter,
		notify.NewNotifyRPCAdapter,
		userinfo.NewUserInfoServiceImpl,
		NewLock,
		evalSetDomainService,
		targetDomainService,
		evaluatorDomainService,
		flagSet,
	)

	evaluatorDomainService = wire.NewSet(
		domainservice.NewEvaluatorServiceImpl,
		domainservice.NewEvaluatorRecordServiceImpl,
		NewEvaluatorSourceServices,
		llm.NewLLMRPCProvider,
		evaluatorrepo.NewEvaluatorRepo,
		evaluatorrepo.NewEvaluatorRecordRepo,
		evaluatormysql.NewEvaluatorDAO,
		evaluatormysql.NewEvaluatorVersionDAO,
		evaluatormysql.NewEvaluatorRecordDAO,
		evaluatorrepo.NewRateLimiterImpl,
		evalconf.NewEvaluatorConfiger,
		evaluatormtr.NewEvaluatorMetrics,
		rmqproducer.NewEvaluatorEventPublisher,
	)

	evaluatorSet = wire.NewSet(
		NewEvaluatorHandlerImpl,
		foundation.NewAuthRPCProvider,
		foundation.NewFileRPCProvider,
		foundation.NewUserRPCProvider,
		userinfo.NewUserInfoServiceImpl,
		idem.NewIdempotentService,
		iredis.NewIdemDAO,
		rmqproducer.NewExptEventPublisher,
		evaluatorDomainService,
		flagSet,
		experiment.NewExptRepo,
		exptmysql.NewExptDAO,
		exptmysql.NewExptEvaluatorRefDAO,
	)

	evalSetDomainService = wire.NewSet(
		domainservice.NewEvaluationSetVersionServiceImpl,
		domainservice.NewEvaluationSetItemServiceImpl,
		data.NewDatasetRPCAdapter,
		domainservice.NewEvaluationSetServiceImpl,
	)

	evaluationSetSet = wire.NewSet(
		NewEvaluationSetApplicationImpl,
		evalSetDomainService,
		evalsetmtr.NewEvaluationSetMetrics,
		domainservice.NewEvaluationSetSchemaServiceImpl,
		foundation.NewAuthRPCProvider,
		foundation.NewUserRPCProvider,
		userinfo.NewUserInfoServiceImpl,
	)

	targetDomainService = wire.NewSet(
		domainservice.NewEvalTargetServiceImpl,
		NewSourceTargetOperators,
		prompt.NewPromptRPCAdapter,
		targetrepo.NewEvalTargetRepo,
		mysql.NewEvalTargetDAO,
		mysql.NewEvalTargetRecordDAO,
		mysql.NewEvalTargetVersionDAO,
	)

	evalTargetSet = wire.NewSet(
		NewEvalTargetHandlerImpl,
		evaltargetmtr.NewEvalTargetMetrics,
		foundation.NewAuthRPCProvider,
		targetDomainService,
		flagSet,
	)
)

func NewSourceTargetOperators(adapter rpc.IPromptRPCAdapter) map[entity.EvalTargetType]service.ISourceEvalTargetOperateService {
	return map[entity.EvalTargetType]service.ISourceEvalTargetOperateService{
		entity.EvalTargetTypeLoopPrompt: service.NewPromptSourceEvalTargetServiceImpl(adapter),
	}
}

func NewLock(cmdable redis.Cmdable) lock.ILocker {
	return lock.NewRedisLockerWithHolder(cmdable, "evaluation")
}

func InitExperimentApplication(
	ctx context.Context,
	idgen idgen.IIDGenerator,
	db db.Provider,
	configFactory conf.IConfigLoaderFactory,
	rmqFactory mq.IFactory,
	cmdable redis.Cmdable,
	auditClient audit.IAuditService,
	meter metrics.Meter,
	authClient authservice.Client,
	evalSetService evaluationservice.EvaluationSetService,
	evaluatorService evaluationservice.EvaluatorService,
	targetService evaluationservice.EvalTargetService,
	uc userservice.Client,
	pms promptmanageservice.Client,
	pes promptexecuteservice.Client,
	sds datasetservice.Client,
	limiterFactory limiter.IRateLimiterFactory,
	llmcli llmruntimeservice.Client,
	benefitSvc benefit.IBenefitService,
	ckDb ck.Provider,
	tagClient tagservice.Client,
	objectStorage fileserver.ObjectStorage,
) (IExperimentApplication, error) {
	wire.Build(
		experimentSet,
	)
	return nil, nil
}

func InitEvaluatorApplication(
	ctx context.Context,
	idgen idgen.IIDGenerator,
	authClient authservice.Client,
	db db.Provider,
	configFactory conf.IConfigLoaderFactory,
	rmqFactory mq.IFactory,
	llmClient llmruntimeservice.Client,
	meter metrics.Meter,
	userClient userservice.Client,
	auditClient audit.IAuditService,
	cmdable redis.Cmdable,
	benefitSvc benefit.IBenefitService,
	limiterFactory limiter.IRateLimiterFactory,
	fileClient fileservice.Client,
) (evaluation.EvaluatorService, error) {
	wire.Build(
		evaluatorSet,
	)
	return nil, nil
}

func InitEvaluationSetApplication(client datasetservice.Client,
	authClient authservice.Client,
	meter metrics.Meter,
	userClient userservice.Client,
) evaluation.EvaluationSetService {
	wire.Build(
		evaluationSetSet,
	)
	return nil
}

func InitEvalTargetApplication(ctx context.Context,
	idgen idgen.IIDGenerator,
	db db.Provider,
	client promptmanageservice.Client,
	executeClient promptexecuteservice.Client,
	authClient authservice.Client,
	cmdable redis.Cmdable,
	meter metrics.Meter,
) evaluation.EvalTargetService {
	wire.Build(
		evalTargetSet,
	)
	return nil
}

func NewEvaluatorSourceServices(llmProvider componentrpc.ILLMProvider, metric mtr.EvaluatorExecMetrics, config evalconf.IConfiger) []domainservice.EvaluatorSourceService {
	return []domainservice.EvaluatorSourceService{
		domainservice.NewEvaluatorSourcePromptServiceImpl(llmProvider, metric, config),
	}
}
