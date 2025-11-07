// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package convertor

import (
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/domain/tag"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
)

func TagValueDTO2DO(tagValue *tag.TagValue) *rpc.TagValue {
	return &rpc.TagValue{
		TagValueId:   ptr.From(tagValue.TagValueID),
		TagValueName: ptr.From(tagValue.TagValueName),
		TagValues:    TagValueListDTO2DO(tagValue.GetChildren()),
	}
}

func TagValueListDTO2DO(tagValues []*tag.TagValue) []*rpc.TagValue {
	ret := make([]*rpc.TagValue, 0, len(tagValues))
	for _, tagValue := range tagValues {
		ret = append(ret, TagValueDTO2DO(tagValue))
	}
	return ret
}

func TagDTO2DO(tagInfo *tag.TagInfo) *rpc.TagInfo {
	return &rpc.TagInfo{
		TagKeyId:       ptr.From(tagInfo.TagKeyID),
		TagKeyName:     ptr.From(tagInfo.TagKeyName),
		InActive:       ptr.From(tagInfo.Status) != "active",
		TagValues:      TagValueListDTO2DO(tagInfo.TagValues),
		TagContentType: rpc.TagContentType(ptr.From(tagInfo.ContentType)),
	}
}

func TagListDTO2DO(tagInfos []*tag.TagInfo) []*rpc.TagInfo {
	ret := make([]*rpc.TagInfo, 0, len(tagInfos))
	for _, tagInfo := range tagInfos {
		ret = append(ret, TagDTO2DO(tagInfo))
	}
	return ret
}
