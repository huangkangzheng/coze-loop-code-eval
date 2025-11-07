// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"context"

	"github.com/bytedance/sonic"

	"code.byted.org/flowdevops/cozeloop/backend/infra/middleware/session"
	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/service"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/conv"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/logs"
)

func NewExptRecordEvalConsumer(recordEval service.ExptItemEvalEvent) mq.IConsumerHandler {
	return &ExptRecordEvalConsumer{
		ExptItemEvalEvent: recordEval,
	}
}

type ExptRecordEvalConsumer struct {
	service.ExptItemEvalEvent
}

func (e *ExptRecordEvalConsumer) HandleMessage(ctx context.Context, ext *mq.MessageExt) error {
	rawLogID := logs.GetLogID(ctx)
	ctx = logs.SetLogID(ctx, logs.NewLogID())

	body := ext.Body
	event := &entity.ExptItemEvalEvent{}
	if err := sonic.Unmarshal(body, event); err != nil {
		logs.CtxError(ctx, "ExptItemEvalEvent json unmarshal fail, raw: %v, err: %s", conv.UnsafeBytesToString(body), err)
		return nil
	}

	if event.Session != nil {
		ctx = session.WithCtxUser(ctx, &session.User{
			ID: event.Session.UserID,
		})
	}

	logs.CtxInfo(ctx, "ExptRecordEvalConsumer consume message, event: %v, msg_id: %v, rawlogid: %v", conv.UnsafeBytesToString(body), ext.MsgID, rawLogID)

	return e.Eval(ctx, event)
}
