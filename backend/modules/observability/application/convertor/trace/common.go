// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	commondto "code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/observability/domain/common"
	commonentity "code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/trace/entity/common"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
)

func UserInfoDO2DTO(info *commonentity.UserInfo) *commondto.UserInfo {
	if info == nil {
		return nil
	}
	ret := &commondto.UserInfo{}
	if info.Name != "" {
		ret.Name = ptr.Of(info.Name)
	}
	if info.EnName != "" {
		ret.EnName = ptr.Of(info.EnName)
	}
	if info.AvatarURL != "" {
		ret.AvatarURL = ptr.Of(info.AvatarURL)
	}
	if info.AvatarThumb != "" {
		ret.AvatarThumb = ptr.Of(info.AvatarThumb)
	}
	if info.OpenID != "" {
		ret.OpenID = ptr.Of(info.OpenID)
	}
	if info.UnionID != "" {
		ret.UnionID = ptr.Of(info.UnionID)
	}
	if info.Email != "" {
		ret.Email = ptr.Of(info.Email)
	}
	if info.UserID != "" {
		ret.UserID = ptr.Of(info.UserID)
	}
	return ret
}
