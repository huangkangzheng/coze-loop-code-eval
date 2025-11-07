// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package foundation

import (
	"context"

	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/user"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/user/userservice"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/pkg/errno"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
)

type UserRPCAdapter struct {
	client userservice.Client
}

func NewUserRPCProvider(client userservice.Client) rpc.IUserProvider {
	return &UserRPCAdapter{
		client: client,
	}
}

func (u UserRPCAdapter) MGetUserInfo(ctx context.Context, userIDs []string) ([]*entity.UserInfo, error) {
	resp, err := u.client.MGetUserInfo(ctx, &user.MGetUserInfoRequest{
		UserIds: userIDs,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errorx.NewByCode(errno.CommonRPCErrorCode)
	}
	if resp.BaseResp != nil && resp.BaseResp.StatusCode != 0 {
		return nil, errorx.NewByCode(resp.BaseResp.StatusCode, errorx.WithExtraMsg(resp.BaseResp.StatusMessage))
	}
	res := make([]*entity.UserInfo, 0)
	for _, userInfo := range resp.UserInfos {
		if userInfo == nil {
			continue
		}
		res = append(res, &entity.UserInfo{
			Name:      userInfo.NickName,
			AvatarURL: userInfo.AvatarURL,
			// AvatarThumb: userInfo.AvatarThumb,
			Email:  userInfo.Email,
			UserID: userInfo.UserID,
		})
	}
	return res, nil
}
