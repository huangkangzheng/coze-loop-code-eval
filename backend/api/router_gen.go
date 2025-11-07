// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"code.byted.org/flowdevops/cozeloop/backend/api/handler/coze/loop/apis"
	router "code.byted.org/flowdevops/cozeloop/backend/api/router"
)

// register registers all routers.
func register(r *server.Hertz, handler *apis.APIHandler) {
	router.GeneratedRegister(r, handler)

	customizedRegister(r)
}
