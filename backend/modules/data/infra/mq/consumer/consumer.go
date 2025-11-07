// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/application"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/pkg/consts"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

func NewConsumerWorkers(
	cfactory conf.IConfigLoaderFactory,
	handler application.IJobRunMsgHandler,
) ([]mq.IConsumerWorker, error) {
	loader, err := cfactory.NewConfigLoader(consts.DataConfigFileName)
	if err != nil {
		return nil, err
	}
	return []mq.IConsumerWorker{
		newDatasetJobConsumer(handler, loader),
	}, nil
}
