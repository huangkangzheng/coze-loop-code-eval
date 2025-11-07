// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package notify

import (
	"context"

	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/component/rpc"
)

type NotifyRPCAdapter struct{}

func NewNotifyRPCAdapter() rpc.INotifyRPCAdapter {
	return NotifyRPCAdapter{}
}

func (n NotifyRPCAdapter) SendMessageCard(ctx context.Context, userID, cardID string, param map[string]string) error {
	return nil
}
