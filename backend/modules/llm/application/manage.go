// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"strconv"

	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/llm/manage"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/application/convertor"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/service"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
)

type manageApp struct {
	manageSrv service.IManage
	auth      rpc.IAuthProvider
}

func NewManageApplication(
	manageSrv service.IManage,
	auth rpc.IAuthProvider,
) manage.LLMManageService {
	return &manageApp{
		manageSrv: manageSrv,
		auth:      auth,
	}
}

func (m *manageApp) ListModels(ctx context.Context, req *manage.ListModelsRequest) (r *manage.ListModelsResponse, err error) {
	r = manage.NewListModelsResponse()
	if err := m.auth.CheckSpacePermission(ctx, req.GetWorkspaceID(), "listModels"); err != nil {
		return r, err
	}
	var scenario *entity.Scenario
	if req.Scenario != nil {
		scenario = ptr.Of(entity.Scenario(req.GetScenario()))
	} else {
		scenario = ptr.Of(entity.ScenarioDefault)
	}
	pageToken, _ := strconv.ParseInt(req.GetPageToken(), 10, 64)
	models, total, hasMore, nextPageToken, err := m.manageSrv.ListModels(ctx, entity.ListModelReq{
		WorkspaceID: req.WorkspaceID,
		Scenario:    scenario,
		PageSize:    int64(req.GetPageSize()),
		PageToken:   pageToken,
	})
	if err != nil {
		return r, err
	}
	r.SetModels(convertor.ModelsDO2DTO(models, true))
	r.SetTotal(ptr.Of(int32(total)))
	r.SetNextPageToken(ptr.Of(strconv.FormatInt(nextPageToken, 10)))
	r.SetHasMore(ptr.Of(hasMore))
	return r, nil
}

func (m *manageApp) GetModel(ctx context.Context, req *manage.GetModelRequest) (r *manage.GetModelResponse, err error) {
	r = manage.NewGetModelResponse()
	if err := m.auth.CheckSpacePermission(ctx, req.GetWorkspaceID(), "getModel"); err != nil {
		return r, err
	}
	model, err := m.manageSrv.GetModelByID(ctx, req.GetModelID())
	if err != nil {
		return r, err
	}
	r.SetModel(convertor.ModelDO2DTO(model, true))
	return r, nil
}
