// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"context"
	"time"

	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	obapp "code.byted.org/flowdevops/cozeloop/backend/modules/observability/application"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/component/config"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/json"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/conv"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/logs"
)

type AnnotationConsumer struct {
	handler obapp.IAnnotationQueueConsumer
	conf.IConfigLoader
}

func newAnnotationConsumer(handler obapp.IAnnotationQueueConsumer, loader conf.IConfigLoader) mq.IConsumerWorker {
	return &AnnotationConsumer{
		handler:       handler,
		IConfigLoader: loader,
	}
}

func (e *AnnotationConsumer) ConsumerCfg(ctx context.Context) (*mq.ConsumerConfig, error) {
	const key = "annotation_mq_consumer_config"
	cfg := &config.MqConsumerCfg{}
	if err := e.UnmarshalKey(ctx, key, cfg); err != nil {
		return nil, err
	}
	res := &mq.ConsumerConfig{
		Addr:                 cfg.Addr,
		Topic:                cfg.Topic,
		ConsumerGroup:        cfg.ConsumerGroup,
		ConsumeTimeout:       time.Duration(cfg.Timeout) * time.Millisecond,
		ConsumeGoroutineNums: cfg.WorkerNum,
	}
	return res, nil
}

func (e *AnnotationConsumer) HandleMessage(ctx context.Context, ext *mq.MessageExt) error {
	event := new(entity.AnnotationEvent)
	if err := json.Unmarshal(ext.Body, event); err != nil {
		logs.CtxError(ctx, "annotation msg json unmarshal fail, raw: %v, err: %s", conv.UnsafeBytesToString(ext.Body), err)
		return nil
	}
	logs.CtxInfo(ctx, "Handle annotation message %+v, annotation: %+v", event, event.Annotation)
	return e.handler.Send(ctx, event)
}
