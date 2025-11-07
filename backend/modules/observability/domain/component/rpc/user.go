// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"

	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/common"
)

//go:generate mockgen -destination=mocks/user.go -package=mocks . IUserProvider
type IUserProvider interface {
	GetUserInfo(ctx context.Context, userIDs []string) ([]*common.UserInfo, map[string]*common.UserInfo, error)
}
