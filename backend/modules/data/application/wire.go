// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

//go:build wireinject
// +build wireinject

package application

import (
	"github.com/google/wire"

	"code.byted.org/flowdevops/cozeloop/backend/infra/db"
	"code.byted.org/flowdevops/cozeloop/backend/infra/external/audit"
	"code.byted.org/flowdevops/cozeloop/backend/infra/fileserver"
	"code.byted.org/flowdevops/cozeloop/backend/infra/idgen"
	"code.byted.org/flowdevops/cozeloop/backend/infra/lock"
	"code.byted.org/flowdevops/cozeloop/backend/infra/mq"
	"code.byted.org/flowdevops/cozeloop/backend/infra/redis"
	tag2 "code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/data/tag"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/auth/authservice"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/user/userservice"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/domain/component/userinfo"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/domain/dataset/service"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/domain/entity"
	service2 "code.byted.org/flowdevops/cozeloop/backend/modules/data/domain/tag/service"
	dataset_config "code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/conf"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/mq/producer"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/repo/dataset"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/repo/dataset/item_dao"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/repo/dataset/mysql"
	oss_dao "code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/repo/dataset/oss"
	redis2 "code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/repo/dataset/redis"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/repo/tag"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/rpc/foundation"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/vfs/oss"
	"code.byted.org/flowdevops/cozeloop/backend/modules/data/infra/vfs/unionfs"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/conf"
)

var (
	datasetSet = wire.NewSet(
		NewDatasetApplicationImpl,
		service.NewDatasetServiceImpl,
		dataset.NewDatasetRepo,
		mysql.NewDatasetDAO,
		mysql.NewDatasetItemDAO,
		mysql.NewDatasetVersionDAO,
		mysql.NewDatasetItemSnapshotDAO,
		mysql.NewDatasetSchemaDAO,
		mysql.NewDatasetIOJobDAO,
		redis2.NewOperationDAO,
		redis2.NewDatasetDAO,
		redis2.NewVersionDAO,
		dataset_config.NewConfiger,
		producer.NewDatasetJobPublisher,
		foundation.NewAuthRPCProvider,
		oss.NewClient,
		unionfs.NewUnionFS,
		lock.NewRedisLocker,
		NewItemProviderDAO,
	)

	tagSet = wire.NewSet(
		NewTagApplicationImpl,
		service2.NewTagServiceImpl,
		tag.NewTagRepoImpl,
		dataset_config.NewConfiger,
		userinfo.NewUserInfoServiceImpl,
		foundation.NewUserRPCProvider,
		lock.NewRedisLocker,
	)
)

func NewItemProviderDAO(batchObjectStorage fileserver.BatchObjectStorage) map[entity.Provider]item_dao.ItemDAO {
	return map[entity.Provider]item_dao.ItemDAO{
		entity.ProviderS3: oss_dao.NewDatasetItemDAO(batchObjectStorage),
	}
}

func InitDatasetApplication(
	idgen idgen.IIDGenerator,
	db db.Provider,
	cmdable redis.Cmdable,
	configFactory conf.IConfigLoaderFactory,
	configLoader conf.IConfigLoader,
	mqFactory mq.IFactory,
	objectStorage fileserver.ObjectStorage,
	batchObjectStorage fileserver.BatchObjectStorage,
	auditClient audit.IAuditService,
	authClient authservice.Client,
) (IDatasetApplication, error) {
	wire.Build(
		datasetSet,
	)
	return nil, nil
}

func InitTagApplication(idgen idgen.IIDGenerator,
	db db.Provider,
	cmdable redis.Cmdable,
	configLoader conf.IConfigLoader,
	userClient userservice.Client,
	authAdapter rpc.IAuthProvider) (tag2.TagService, error) {
	wire.Build(tagSet)
	return nil, nil
}
