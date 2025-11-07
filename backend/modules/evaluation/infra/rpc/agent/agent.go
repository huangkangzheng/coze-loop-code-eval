// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"

	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/pkg/errno"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
)

type AgentAdapter struct{}

func NewAgentAdapter() rpc.IAgentAdapter {
	return &AgentAdapter{}
}

func (a AgentAdapter) CallTraceAgent(ctx context.Context, spaceID int64, url string) (int64, error) {
	return 0, errorx.NewByCode(errno.CommonInternalErrorCode, errorx.WithExtraMsg("CallTraceAgent not implement"))
}

func (a AgentAdapter) GetReport(ctx context.Context, spaceID, reportID int64) (report string, status entity.ReportStatus, err error) {
	return "", 0, errorx.NewByCode(errno.CommonInternalErrorCode, errorx.WithExtraMsg("GetReport not implement"))
}
