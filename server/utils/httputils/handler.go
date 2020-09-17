// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package httputils

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	*mux.Router
}

func NewHandler() *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}
	h.Router.Handle("{anything:.*}", http.NotFoundHandler())
	return h
}
