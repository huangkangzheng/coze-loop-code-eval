// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/component/conf"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/repo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/service/llmfactory"
)

func NewRuntime(
	llmFact llmfactory.IFactory,
	idGen idgen.IIDGenerator,
	runtimeRepo repo.IRuntimeRepo,
	cfg conf.IConfigRuntime,
) IRuntime {
	return &RuntimeImpl{
		llmFact:     llmFact,
		idGen:       idGen,
		runtimeRepo: runtimeRepo,
		runtimeCfg:  cfg,
	}
}

func NewManage(cfg conf.IConfigManage) IManage {
	return &ManageImpl{
		conf: cfg,
	}
}
