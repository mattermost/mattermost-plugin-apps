// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package httputils

import (
	"io"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type LimitReadCloser struct {
	ReadCloser io.ReadCloser
	TotalRead  utils.ByteSize
	Limit      utils.ByteSize
	OnClose    func(*LimitReadCloser) error
}

func (r *LimitReadCloser) Read(data []byte) (int, error) {
	if r.Limit >= 0 {
		remain := r.Limit - r.TotalRead
		if remain <= 0 {
			return 0, io.EOF
		}
		if len(data) > int(remain) {
			data = data[0:remain]
		}
	}
	n, err := r.ReadCloser.Read(data)
	r.TotalRead += utils.ByteSize(n)
	return n, err
}

func (r *LimitReadCloser) Close() error {
	if r.OnClose != nil {
		err := r.OnClose(r)
		if err != nil {
			return err
		}
	}
	return r.ReadCloser.Close()
}
