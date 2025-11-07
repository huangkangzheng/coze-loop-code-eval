// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package rmqreceiver

import (
	"context"
	"fmt"

	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/collector/component"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/collector/consumer"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/collector/receiver"
)

const (
	TypeStr = "rmq"
)

func createDefaultConfig() component.Config {
	return &Config{}
}

func NewFactory(rmqFactory mq.IFactory) receiver.Factory {
	return receiver.NewFactory(
		TypeStr,
		createDefaultConfig,
		func(ctx context.Context, params receiver.CreateSettings, baseCfg component.Config, c consumer.Consumer) (receiver.Receiver, error) {
			if c == nil {
				return nil, fmt.Errorf("no next consumer")
			}
			return &rmqReceiver{
				componentID:  params.ID,
				nextConsumer: c,
				config:       baseCfg.(*Config),
				mqFactory:    rmqFactory,
			}, nil
		},
	)
}
