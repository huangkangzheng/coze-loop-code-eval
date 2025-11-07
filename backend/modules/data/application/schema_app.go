// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"strings"

	"github.com/bytedance/gg/gptr"
	"github.com/bytedance/gg/gslice"

	"code.byted.org/flowdevops/cozeloop/backend/infra/external/audit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/middleware/session"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/dataset"
	idl "code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/domain/dataset"
	convertor "code.byted.org/flowdevops/cozeloop/backend/modules/data/application/convertor/dataset"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/pkg/errno"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/encoding"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/logs"
)

func (h *DatasetApplicationImpl) GetDatasetSchema(ctx context.Context, req *dataset.GetDatasetSchemaRequest) (resp *dataset.GetDatasetSchemaResponse, err error) {
	// 鉴权
	err = h.authByDatasetID(ctx, req.GetWorkspaceID(), req.GetDatasetID(), rpc.CommonActionRead)
	if err != nil {
		return nil, err
	}
	ds, err := h.repo.GetDataset(ctx, req.GetWorkspaceID(), req.GetDatasetID())
	if err != nil {
		return nil, err
	}
	if ds == nil {
		return nil, errno.NotFoundErrorf("dataset=%d is not found", req.GetDatasetID())
	}
	schemaID := ds.SchemaID

	schema, err := h.repo.GetSchema(ctx, req.GetWorkspaceID(), schemaID)
	if err != nil {
		return nil, err
	}
	if !req.GetWithDeleted() {
		schema.Fields = schema.AvailableFields()
	}
	dtos, err := gslice.TryMap(schema.AvailableFields(), convertor.FieldSchemaDO2DTO).Get()
	if err != nil {
		return nil, err
	}
	return &dataset.GetDatasetSchemaResponse{Fields: dtos}, nil
}

func (h *DatasetApplicationImpl) UpdateDatasetSchema(ctx context.Context, req *dataset.UpdateDatasetSchemaRequest) (resp *dataset.UpdateDatasetSchemaResponse, err error) {
	// 鉴权
	err = h.authByDatasetID(ctx, req.GetWorkspaceID(), req.GetDatasetID(), rpc.CommonActionEdit)
	if err != nil {
		return nil, err
	}
	fields, err := gslice.TryMap(req.Fields, convertor.FieldSchemaDTO2DO).Get()
	if err != nil {
		return nil, errno.BadReqErr(err)
	}
	ds, err := h.repo.GetDataset(ctx, req.GetWorkspaceID(), req.GetDatasetID())
	if err != nil {
		return nil, err
	}

	userID := session.UserIDInCtxOrEmpty(ctx)
	fieldNames := gslice.FilterMap(req.Fields, func(f *idl.FieldSchema) (string, bool) { return f.GetName(), f.GetName() != "" })
	fieldDescs := gslice.FilterMap(req.Fields, func(f *idl.FieldSchema) (string, bool) { return f.GetDescription(), f.GetDescription() != "" })
	data := map[string]string{
		"texts": strings.Join(gslice.Union(fieldNames, fieldDescs), ","),
	}
	record, err := h.auditClient.Audit(ctx, audit.AuditParam{
		ObjectID:  0,
		AuditType: audit.AuditType_CozeLoopDatasetModify,
		AuditData: data,
		ReqID:     encoding.Encode(ctx, data),
	})
	if err != nil {
		logs.CtxError(ctx, "call BatchAuditDatas failed, err=%v", err)
	}
	if record.AuditStatus != audit.AuditStatus_Approved {
		return nil, errno.GetContentAuditFailedErrorf("reason=%s", gptr.Indirect(record.FailedReason))
	}
	if err := h.svc.UpdateSchema(ctx, ds, fields, userID); err != nil {
		return nil, err
	}
	return &dataset.UpdateDatasetSchemaResponse{}, nil
}
