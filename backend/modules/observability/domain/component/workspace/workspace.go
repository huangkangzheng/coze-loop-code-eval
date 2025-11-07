// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/observability/domain/span"
)

//go:generate mockgen -destination=mocks/workspace_provider.go -package=mocks . IWorkSpaceProvider
type IWorkSpaceProvider interface {
	GetIngestWorkSpaceID(ctx context.Context, spans []*span.InputSpan) string
	GetThirdPartyQueryWorkSpaceID(ctx context.Context, requestWorkspaceID int64) string
}
