// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"code.byted.org/flowdevops/cozeloop/backend/infra/external/audit"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/domain/entity"
	prompterr "code.byted.org/flowdevops/cozeloop/backend/modules/prompt/pkg/errno"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/encoding"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/logs"
)

type AuditRPCAdapter struct {
	client audit.IAuditService
}

func NewAuditRPCProvider(client audit.IAuditService) rpc.IAuditProvider {
	return &AuditRPCAdapter{
		client: client,
	}
}

func (a *AuditRPCAdapter) AuditPrompt(ctx context.Context, promptDO *entity.Prompt) error {
	if promptDO == nil {
		return nil
	}

	var auditingTexts []string
	if promptDO.PromptBasic != nil {
		auditingTexts = append(auditingTexts, promptDO.PromptBasic.DisplayName, promptDO.PromptBasic.Description)
	}
	if promptDO.PromptDraft != nil {
		if promptDO.PromptDraft.PromptDetail != nil {
			if promptDO.PromptDraft.PromptDetail.PromptTemplate != nil {
				for _, message := range promptDO.PromptDraft.PromptDetail.PromptTemplate.Messages {
					auditingTexts = append(auditingTexts, ptr.From(message.Content))
				}
			}
		}
	}
	auditingData := map[string]string{
		"texts": strings.Join(auditingTexts, ","),
	}

	auditParam := audit.AuditParam{
		ObjectID: func() int64 {
			if promptDO.ID <= 0 {
				return int64(uuid.New().ID())
			}
			return promptDO.ID
		}(),
		AuditType: audit.AuditType_CozeLoopPEModify,
		AuditData: auditingData,
		ReqID:     encoding.Encode(ctx, auditingData),
	}
	record, err := a.client.Audit(ctx, auditParam)
	// 审核服务不可用，默认通过
	if err != nil {
		logs.CtxError(ctx, "audit: failed to audit, err=%v", err)
		return nil
	}
	if record.AuditStatus != audit.AuditStatus_Approved {
		return errorx.NewByCode(prompterr.RiskContentDetectedCode, errorx.WithExtraMsg(ptr.From(record.FailedReason)))
	}
	return nil
}
