// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package convertor

import (
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/llm/domain/common"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
)

func ScenarioDO2DTO(s entity.Scenario) common.Scenario {
	return common.Scenario(s)
}

func ScenarioPtrDTO2DTO(s *common.Scenario) *entity.Scenario {
	if s == nil {
		return nil
	}
	return ptr.Of(entity.Scenario(*s))
}
