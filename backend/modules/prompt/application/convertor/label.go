// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package convertor

import (
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/prompt/domain/prompt"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
)

// BatchLabelDO2DTO converts batch PromptLabel DO to Label DTO
func BatchLabelDO2DTO(dos []*entity.PromptLabel) []*prompt.Label {
	if len(dos) == 0 {
		return nil
	}
	dtos := make([]*prompt.Label, 0, len(dos))
	for _, do := range dos {
		if do == nil {
			continue
		}
		dtos = append(dtos, &prompt.Label{
			Key: ptr.Of(do.LabelKey),
		})
	}
	return dtos
}
