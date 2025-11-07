// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package evaluator

import (
	"context"

	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/domain/repo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/evaluator/mysql"
	"code.byted.org/flowdevops/cozeloop/backend/modules/evaluation/infra/repo/evaluator/mysql/convertor"
)

type EvaluatorRecordRepoImpl struct {
	idgen              idgen.IIDGenerator
	evaluatorRecordDao mysql.EvaluatorRecordDAO
	dbProvider         db.Provider
}

func NewEvaluatorRecordRepo(idgen idgen.IIDGenerator, provider db.Provider, evaluatorRecordDao mysql.EvaluatorRecordDAO) repo.IEvaluatorRecordRepo {
	singletonEvaluatorRecordRepo := &EvaluatorRecordRepoImpl{
		evaluatorRecordDao: evaluatorRecordDao,
		dbProvider:         provider,
		idgen:              idgen,
	}
	return singletonEvaluatorRecordRepo
}

func (r *EvaluatorRecordRepoImpl) CreateEvaluatorRecord(ctx context.Context, evaluatorRecord *entity.EvaluatorRecord) error {
	po := convertor.ConvertEvaluatorRecordDO2PO(evaluatorRecord)
	return r.evaluatorRecordDao.CreateEvaluatorRecord(ctx, po)
}

func (r *EvaluatorRecordRepoImpl) CorrectEvaluatorRecord(ctx context.Context, evaluatorRecord *entity.EvaluatorRecord) error {
	po := convertor.ConvertEvaluatorRecordDO2PO(evaluatorRecord)
	return r.evaluatorRecordDao.UpdateEvaluatorRecord(ctx, po)
}

func (r *EvaluatorRecordRepoImpl) GetEvaluatorRecord(ctx context.Context, evaluatorRecordID int64, includeDeleted bool) (*entity.EvaluatorRecord, error) {
	po, err := r.evaluatorRecordDao.GetEvaluatorRecord(ctx, evaluatorRecordID, includeDeleted)
	if err != nil {
		return nil, err
	}
	if po == nil {
		return nil, nil
	}
	evaluatorRecord, err := convertor.ConvertEvaluatorRecordPO2DO(po)
	if err != nil {
		return nil, err
	}

	return evaluatorRecord, nil
}

func (r *EvaluatorRecordRepoImpl) BatchGetEvaluatorRecord(ctx context.Context, evaluatorRecordIDs []int64, includeDeleted bool) ([]*entity.EvaluatorRecord, error) {
	pos, err := r.evaluatorRecordDao.BatchGetEvaluatorRecord(ctx, evaluatorRecordIDs, includeDeleted)
	if err != nil {
		return nil, err
	}

	evaluatorRecords := make([]*entity.EvaluatorRecord, 0, len(pos))
	for _, po := range pos {
		evaluatorRecord, err := convertor.ConvertEvaluatorRecordPO2DO(po)
		if err != nil {
			return nil, err
		}
		evaluatorRecords = append(evaluatorRecords, evaluatorRecord)
	}
	return evaluatorRecords, nil
}
