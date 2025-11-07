// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"context"

	"github.com/bytedance/sonic"

	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/evaluation"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/evaluation/expt"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/logs"
)

type ExptTurnResultFilterConsumer struct {
	exptManager evaluation.ExperimentService
}

func NewExptTurnResultFilterConsumer(exptManager evaluation.ExperimentService) mq.IConsumerHandler {
	return &ExptTurnResultFilterConsumer{
		exptManager: exptManager,
	}
}

func (c *ExptTurnResultFilterConsumer) HandleMessage(ctx context.Context, ext *mq.MessageExt) (err error) {
	defer func() {
		if err != nil {
			logs.CtxError(ctx, "ExptTurnResultFilterConsumer HandleMessage fail, err: %v", err)
		}
	}()

	event := &entity.ExptTurnResultFilterEvent{}
	body := ext.Body
	if err := sonic.Unmarshal(body, event); err != nil {
		logs.CtxError(ctx, "ExptTurnResultFilterEvent json unmarshal fail, raw: %v, err: %s", string(body), err)
		return nil
	}

	logs.CtxInfo(ctx, "ExptTurnResultFilterConsumer consume message, event: %v, msg_id: %v", string(body), ext.MsgID)

	upsertExptTurnResultFilterRequest := &expt.UpsertExptTurnResultFilterRequest{
		WorkspaceID:  ptr.Of(event.SpaceID),
		ExperimentID: ptr.Of(event.ExperimentID),
		ItemIds:      event.ItemID,
	}
	if event.FilterType != nil {
		upsertExptTurnResultFilterRequest.FilterType = (*expt.UpsertExptTurnResultFilterType)(event.FilterType)
	}
	if event.RetryTimes != nil {
		upsertExptTurnResultFilterRequest.RetryTimes = ptr.Of(*event.RetryTimes)
	}
	// 调用 ExptMangerImpl.UpsertExptTurnResultFilter
	resp, err := c.exptManager.UpsertExptTurnResultFilter(ctx, upsertExptTurnResultFilterRequest)
	if err != nil {
		return err
	}
	if resp.BaseResp != nil && resp.BaseResp.StatusCode != 0 {
		return errorx.NewByCode(resp.BaseResp.StatusCode, errorx.WithExtraMsg(resp.BaseResp.StatusMessage))
	}
	return nil
}
