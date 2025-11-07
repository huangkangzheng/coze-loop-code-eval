// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"context"

	"gorm.io/gorm"

	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/entity"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/domain/repo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/infra/repo/convertor"
	"code.byted.org/flowdevops/cozeloop/backend/modules/llm/infra/repo/dao"
)

type RuntimeRepoImpl struct {
	db                db.Provider
	modelReqRecordDao dao.IModelRequestRecordDao
}

func NewRuntimeRepo(db db.Provider, modelReqRecordDao dao.IModelRequestRecordDao) repo.IRuntimeRepo {
	return &RuntimeRepoImpl{
		db:                db,
		modelReqRecordDao: modelReqRecordDao,
	}
}

func (r *RuntimeRepoImpl) CreateModelRequestRecord(ctx context.Context, record *entity.ModelRequestRecord) (err error) {
	return r.db.Transaction(ctx, func(tx *gorm.DB) error {
		opt := db.WithTransaction(tx)
		return r.modelReqRecordDao.Create(ctx, convertor.ModelReqRecordDO2PO(record), opt)
	})
}
