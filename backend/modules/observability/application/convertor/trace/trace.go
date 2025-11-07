// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	traced "code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/observability/domain/trace"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/observability/trace"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/loop_span"
)

func AdvanceInfoDO2DTO(info *loop_span.TraceAdvanceInfo) *trace.TraceAdvanceInfo {
	return &trace.TraceAdvanceInfo{
		TraceID: info.TraceId,
		Tokens: &trace.TokenCost{
			Input:  info.InputCost,
			Output: info.OutputCost,
		},
	}
}

func BatchAdvanceInfoDO2DTO(infos []*loop_span.TraceAdvanceInfo) []*trace.TraceAdvanceInfo {
	ret := make([]*trace.TraceAdvanceInfo, len(infos))
	for i, info := range infos {
		ret[i] = AdvanceInfoDO2DTO(info)
	}
	return ret
}

func FileMetaDO2DTO() {
}

func AdvanceInfoDO2TraceDTO(info *loop_span.TraceAdvanceInfo) *traced.Trace {
	return &traced.Trace{
		TraceID: &info.TraceId,
		Tokens: &traced.TokenCost{
			InputToken:  info.InputCost,
			OutputToken: info.OutputCost,
		},
	}
}

func BatchAdvanceInfoDO2TraceDTO(infos []*loop_span.TraceAdvanceInfo) []*traced.Trace {
	ret := make([]*traced.Trace, len(infos))
	for i, info := range infos {
		ret[i] = AdvanceInfoDO2TraceDTO(info)
	}
	return ret
}
