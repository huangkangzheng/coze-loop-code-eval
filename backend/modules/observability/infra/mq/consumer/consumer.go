// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/application"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

func NewConsumerWorkers(
	loader conf.IConfigLoader,
	handler application.IAnnotationQueueConsumer,
) ([]mq.IConsumerWorker, error) {
	return []mq.IConsumerWorker{
		newAnnotationConsumer(handler, loader),
	}, nil
}
