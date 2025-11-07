// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/experiment/mysql/gorm_gen/model"
)

func NewExptTurnEvaluatorResultRefConvertor() ExptTurnEvaluatorResultRefConvertor {
	return ExptTurnEvaluatorResultRefConvertor{}
}

type ExptTurnEvaluatorResultRefConvertor struct{}

func (ExptTurnEvaluatorResultRefConvertor) DO2PO(ref *entity.ExptTurnEvaluatorResultRef) *model.ExptTurnEvaluatorResultRef {
	return &model.ExptTurnEvaluatorResultRef{
		ID:                 ref.ID,
		SpaceID:            ref.SpaceID,
		ExptTurnResultID:   ref.ExptTurnResultID,
		EvaluatorVersionID: ref.EvaluatorVersionID,
		EvaluatorResultID:  ref.EvaluatorResultID,
		ExptID:             ref.ExptID,
	}
}

func (ExptTurnEvaluatorResultRefConvertor) PO2DO(ref *model.ExptTurnEvaluatorResultRef) *entity.ExptTurnEvaluatorResultRef {
	return &entity.ExptTurnEvaluatorResultRef{
		ID:                 ref.ID,
		SpaceID:            ref.SpaceID,
		ExptTurnResultID:   ref.ExptTurnResultID,
		EvaluatorVersionID: ref.EvaluatorVersionID,
		EvaluatorResultID:  ref.EvaluatorResultID,
		ExptID:             ref.ExptID,
	}
}
