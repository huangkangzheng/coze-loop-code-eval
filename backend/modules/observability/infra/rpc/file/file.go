// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"context"
	"fmt"

	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file"
	"code.byted.org/flowdevops/cozeloop/backend/kitex_gen/coze/loop/foundation/file/fileservice"
	"code.byted.org/flowdevops/cozeloop/backend/modules/observability/domain/component/rpc"
	"code.byted.org/flowdevops/cozeloop/backend/pkg/lang/ptr"
)

type FileProvider struct {
	client fileservice.Client
}

func NewFileRPCProvider(client fileservice.Client) rpc.IFileProvider {
	return &FileProvider{
		client: client,
	}
}

func (f *FileProvider) GetDownloadUrls(ctx context.Context, spaceId string, keys []string) (map[string]string, error) {
	var ttl int64 = 5 * 60 * 60
	keyList := make([]string, 0)
	for _, key := range keys {
		keyList = append(keyList, fmt.Sprintf("%s/%s", spaceId, key))
	}
	req := &file.SignDownloadFileRequest{
		Keys: keyList,
		Option: &file.SignFileOption{
			TTL: ptr.Of(ttl),
		},
		BusinessType: ptr.Of(file.BusinessTypeObservability),
	}
	resp, err := f.client.SignDownloadFile(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Uris) != len(keys) {
		return nil, fmt.Errorf("url length mismatch with keys")
	}
	urlMap := make(map[string]string)
	for idx, key := range keys {
		urlMap[key] = resp.Uris[idx]
	}
	return urlMap, nil
}
