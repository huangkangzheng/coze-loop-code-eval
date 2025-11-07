// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package ctxcache

import (
	"context"

	"github.com/cloudwego/kitex/pkg/endpoint"

	"code.byted.org/flowdevops/cozeloop/backend/pkg/ctxcache"
)

func CtxCacheMW(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp any) (err error) {
		return next(ctxcache.Init(ctx), req, resp)
	}
}
