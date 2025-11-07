// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"
	"fmt"

	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth/authservice"
	authentity "code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/domain/auth"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/component/rpc"
	llm_errorx "code.byted.org/flowdevops/cozeloop/backend/modules/llm/pkg/errno"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
)

type AuthRPCAdapter struct {
	client authservice.Client
}

func NewAuthRPCProvider(client authservice.Client) rpc.IAuthProvider {
	return &AuthRPCAdapter{
		client: client,
	}
}

func (a *AuthRPCAdapter) CheckSpacePermission(ctx context.Context, spaceID int64, action string) error {
	authSubject := &authentity.AuthPrincipal{
		AuthPrincipalType:  authentity.AuthPrincipalTypePtr(authentity.AuthPrincipalType_CozeIdentifier),
		AuthCozeIdentifier: &authentity.AuthCozeIdentifier{IdentityTicket: nil},
	}
	authPairs := []*authentity.SubjectActionObjects{
		{
			Subject: authSubject,
			Action:  ptr.Of(action),
			Objects: []*authentity.AuthEntity{
				{
					ID:         ptr.Of(fmt.Sprint(spaceID)),
					EntityType: ptr.Of(authentity.AuthEntityTypeSpace),
				},
			},
		},
	}
	req := &auth.MCheckPermissionRequest{
		Auths:   authPairs,
		SpaceID: ptr.Of(spaceID),
	}
	resp, err := a.client.MCheckPermission(ctx, req)
	if err != nil {
		return err
	}
	for _, authRes := range resp.AuthRes {
		if authRes != nil && !authRes.GetIsAllowed() {
			return errorx.NewByCode(llm_errorx.CommonNoPermissionCode)
		}
	}
	return nil
}
