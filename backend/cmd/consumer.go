// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	dataapp "code.byted.org/flowdevops/cozeloop/backend/modules/data/application"
	dataconsumer "code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/mq/consumer"
	exptapp "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/application"
	evalconsumer "code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/mq/rocket/consumer"
	obapp "code.byted.org/flowdevops/cozeloop/backend/modules/observability/application"
	obconsumer "code.byted.org/flowdevops/cozeloop/backend/modules/observability/infra/mq/consumer"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

func mustInitConsumerWorkers(
	cfactory conf.IConfigLoaderFactory,
	experimentApplication exptapp.IExperimentApplication,
	datasetApplication dataapp.IJobRunMsgHandler,
	obApplication obapp.IObservabilityOpenAPIApplication,
) []mq.IConsumerWorker {
	var res []mq.IConsumerWorker

	workers, err := evalconsumer.NewConsumerWorkers(cfactory, experimentApplication)
	if err != nil {
		panic(err)
	}
	res = append(res, workers...)

	workers, err = dataconsumer.NewConsumerWorkers(cfactory, datasetApplication)
	if err != nil {
		panic(err)
	}
	res = append(res, workers...)

	loader, err := cfactory.NewConfigLoader("observability.yaml")
	if err != nil {
		panic(err)
	}
	workers, err = obconsumer.NewConsumerWorkers(loader, obApplication)
	if err != nil {
		panic(err)
	}
	res = append(res, workers...)

	return res
}
