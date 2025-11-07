// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package traceutil

import "code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"

func GetTraceStatusCode(err error) int32 {
	if statusErr, ok := errorx.FromStatusError(err); ok {
		return statusErr.Code()
	}
	return -1
}
