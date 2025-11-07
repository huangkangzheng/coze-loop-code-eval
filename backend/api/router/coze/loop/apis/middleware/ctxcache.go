// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"code.byted.org/flowdevops/cozeloop/backend/pkg/ctxcache"
)

func CtxCacheMW() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctxcache.Init(ctx))
	}
}
