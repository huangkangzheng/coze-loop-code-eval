// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package vfs

import (
	"context"
	"io/fs"

	"code.byted.org/flowdevops/cozeloop/backend/modules/data/domain/entity"
)

//go:generate mockgen -destination=mocks/union.go -package=mocks . IUnionFS
type IUnionFS interface {
	StatFile(ctx context.Context, provider entity.Provider, path string) (fs.FileInfo, error)
	GetROFileSystem(provider entity.Provider) (ROFileSystem, error)
}
