// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"

	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file/fileservice"
	"code.byted.org/flowdevops/cozeloop/backend/modules/prompt/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/errorx"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
)

type FileRPCAdapter struct {
	client fileservice.Client
}

func NewFileRPCProvider(client fileservice.Client) rpc.IFileProvider {
	return &FileRPCAdapter{
		client: client,
	}
}

func (f *FileRPCAdapter) MGetFileURL(ctx context.Context, keys []string) (urls map[string]string, err error) {
	var ttl int64 = 24 * 60 * 60
	req := &file.SignDownloadFileRequest{
		Keys: keys,
		Option: &file.SignFileOption{
			TTL: ptr.Of(ttl),
		},
		BusinessType: ptr.Of(file.BusinessTypePrompt),
	}
	resp, err := f.client.SignDownloadFile(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Uris) != len(keys) {
		return nil, errorx.New("url length mismatch with keys")
	}
	urls = make(map[string]string)
	for idx, key := range keys {
		urls[key] = resp.Uris[idx]
	}
	return urls, nil
}
