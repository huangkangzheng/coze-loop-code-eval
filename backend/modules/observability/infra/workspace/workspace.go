// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"strconv"

	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/observability/domain/span"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/component/workspace"
)

func NewWorkspaceProvider() workspace.IWorkSpaceProvider {
	return &WorkspaceProviderImpl{}
}

type WorkspaceProviderImpl struct{}

func (t *WorkspaceProviderImpl) GetIngestWorkSpaceID(ctx context.Context, spans []*span.InputSpan) string {
	if len(spans) == 0 {
		return ""
	}
	return spans[0].WorkspaceID
}

func (t *WorkspaceProviderImpl) GetThirdPartyQueryWorkSpaceID(ctx context.Context, requestWorkspaceID int64) string {
	return strconv.FormatInt(requestWorkspaceID, 10)
}
