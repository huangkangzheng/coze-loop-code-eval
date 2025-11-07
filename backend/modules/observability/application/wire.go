// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

//go:build wireinject
// +build wireinject

package application

import (
	"code.byted.org/flowdevops/cozeloop/backend/infra/ck"
	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/infra/external/benefit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/infra/limiter"
	"code.byted.org/flowdevops/cozeloop/backend/infra/metrics"
	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/dataset/datasetservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/tag/tagservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/evaluation/evaluationsetservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/evaluation/evaluatorservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth/authservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file/fileservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/user/userservice"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/component/config"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/collector/exporter"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/collector/processor"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/collector/receiver"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/repo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/service"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/service/collector/exporter/clickhouseexporter"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/service/collector/processor/queueprocessor"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/service/collector/receiver/rmqreceiver"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/service/trace/span_filter"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/service/trace/span_processor"
	obconfig "code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/config"
	obmetrics "code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/metrics"
	mq2 "code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/mq/producer"
	obrepo "code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/repo"
	ckdao "code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/repo/ck"
	mysqldao "code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/repo/mysql"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/rpc/auth"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/rpc/dataset"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/rpc/evaluationset"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/rpc/evaluator"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/rpc/file"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/rpc/tag"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/rpc/user"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/tenant"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/workspace"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
	"github.com/google/wire"
)

var (
	traceDomainSet = wire.NewSet(
		service.NewTraceServiceImpl,
		service.NewTraceExportServiceImpl,
		obrepo.NewTraceCKRepoImpl,
		ckdao.NewSpansCkDaoImpl,
		ckdao.NewAnnotationCkDaoImpl,
		obmetrics.NewTraceMetricsImpl,
		mq2.NewTraceProducerImpl,
		mq2.NewAnnotationProducerImpl,
		file.NewFileRPCProvider,
		NewTraceConfigLoader,
		NewTraceProcessorBuilder,
		obconfig.NewTraceConfigCenter,
		tenant.NewTenantProvider,
		workspace.NewWorkspaceProvider,
		NewDatasetServiceAdapter,
	)
	traceSet = wire.NewSet(
		NewTraceApplication,
		obrepo.NewViewRepoImpl,
		mysqldao.NewViewDaoImpl,
		auth.NewAuthProvider,
		user.NewUserRPCProvider,
		tag.NewTagRPCProvider,
		evaluator.NewEvaluatorRPCProvider,
		traceDomainSet,
	)
	traceIngestionSet = wire.NewSet(
		NewIngestionApplication,
		service.NewIngestionServiceImpl,
		obrepo.NewTraceCKRepoImpl,
		ckdao.NewSpansCkDaoImpl,
		ckdao.NewAnnotationCkDaoImpl,
		obconfig.NewTraceConfigCenter,
		NewTraceConfigLoader,
		NewIngestionCollectorFactory,
	)
	openApiSet = wire.NewSet(
		NewOpenAPIApplication,
		auth.NewAuthProvider,
		traceDomainSet,
	)
)

func NewTraceProcessorBuilder(
	traceConfig config.ITraceConfig,
	fileProvider rpc.IFileProvider,
	benefitSvc benefit.IBenefitService,
) service.TraceFilterProcessorBuilder {
	return service.NewTraceFilterProcessorBuilder(
		span_filter.NewPlatformFilterFactory(
			[]span_filter.Factory{
				span_filter.NewCozeLoopFilterFactory(),
				span_filter.NewPromptFilterFactory(traceConfig),
				span_filter.NewEvaluatorFilterFactory(),
				span_filter.NewEvalTargetFilterFactory(),
			}),
		// get trace processors
		[]span_processor.Factory{
			span_processor.NewPlatformProcessorFactory(traceConfig),
			span_processor.NewCheckProcessorFactory(),
			span_processor.NewAttrTosProcessorFactory(fileProvider),
			span_processor.NewExpireErrorProcessorFactory(benefitSvc),
		},
		// list spans processors
		[]span_processor.Factory{
			span_processor.NewPlatformProcessorFactory(traceConfig),
			span_processor.NewExpireErrorProcessorFactory(benefitSvc),
		},
		// batch get advance info processors
		[]span_processor.Factory{
			span_processor.NewCheckProcessorFactory(),
		},
		// ingest trace processors
		[]span_processor.Factory{},
		// search trace open api processors
		[]span_processor.Factory{
			span_processor.NewPlatformProcessorFactory(traceConfig),
			span_processor.NewCheckProcessorFactory(),
			span_processor.NewAttrTosProcessorFactory(fileProvider),
			span_processor.NewExpireErrorProcessorFactory(benefitSvc),
		},
		// list trace open api processors
		[]span_processor.Factory{
			span_processor.NewPlatformProcessorFactory(traceConfig),
			span_processor.NewExpireErrorProcessorFactory(benefitSvc),
		})
}

func NewIngestionCollectorFactory(mqFactory mq.IFactory, traceRepo repo.ITraceRepo) service.IngestionCollectorFactory {
	return service.NewIngestionCollectorFactory(
		[]receiver.Factory{
			rmqreceiver.NewFactory(mqFactory),
		},
		[]processor.Factory{
			queueprocessor.NewFactory(),
		},
		[]exporter.Factory{
			clickhouseexporter.NewFactory(traceRepo),
		},
	)
}

func NewTraceConfigLoader(confFactory conf.IConfigLoaderFactory) (conf.IConfigLoader, error) {
	return confFactory.NewConfigLoader("observability.yaml")
}

func NewDatasetServiceAdapter(evalSetService evaluationsetservice.Client, datasetService datasetservice.Client) *service.DatasetServiceAdaptor {
	adapter := service.NewDatasetServiceAdaptor()
	datasetProvider := dataset.NewDatasetProvider(datasetService)
	adapter.Register(entity.DatasetCategory_Evaluation, evaluationset.NewEvaluationSetProvider(evalSetService, datasetProvider))
	return adapter
}

func InitTraceApplication(
	db db.Provider,
	ckDb ck.Provider,
	meter metrics.Meter,
	mqFactory mq.IFactory,
	configFactory conf.IConfigLoaderFactory,
	idgen idgen.IIDGenerator,
	fileClient fileservice.Client,
	benefit benefit.IBenefitService,
	authClient authservice.Client,
	userClient userservice.Client,
	evalService evaluatorservice.Client,
	evalSetService evaluationsetservice.Client,
	tagService tagservice.Client,
	datasetService datasetservice.Client,
) (ITraceApplication, error) {
	wire.Build(traceSet)
	return nil, nil
}

func InitOpenAPIApplication(
	mqFactory mq.IFactory,
	configFactory conf.IConfigLoaderFactory,
	fileClient fileservice.Client,
	ckDb ck.Provider,
	benefit benefit.IBenefitService,
	limiterFactory limiter.IRateLimiterFactory,
	authClient authservice.Client,
	meter metrics.Meter,
) (IObservabilityOpenAPIApplication, error) {
	wire.Build(openApiSet)
	return nil, nil
}

func InitTraceIngestionApplication(
	configFactory conf.IConfigLoaderFactory,
	ckDb ck.Provider,
	mqFactory mq.IFactory) (ITraceIngestionApplication, error) {
	wire.Build(traceIngestionSet)
	return nil, nil
}
