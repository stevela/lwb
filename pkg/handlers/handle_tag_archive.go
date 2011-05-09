/*
Copyright 2011 Steve Lacey

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"bytes"
	"github.com/garyburd/twister/web"
	"github.com/stevela/lwb/store"
	"io"
)

type TagLookupFunc func(string) ([]*store.Post, bool)

type tagArchiveHandler struct {
	context  *RenderContext
	fnLookup TagLookupFunc
}

func (tah *tagArchiveHandler) ServeWeb(req *web.Request) {
	tah.context.Config.Cache.Run(req, func(w io.Writer) bool {
		var posts []*store.Post

		found := false
		if tag := req.Param.Get("tag"); tag != "" {
			if posts, found = tah.fnLookup(tag); !found {
				return false
			}
		}

		// Render posts.
		var content bytes.Buffer
		for _, post := range posts {
			renderPost(&content, tah.context, post, false)
		}

		// Render page.
		templates["main"].Template.Execute(w, makeTemplateParams(tah.context, content.Bytes()))

		return true
	})
}

// TagArchiveHandler returns a request handler that serves tags or categories.
func TagArchiveHandler(context *RenderContext, fn TagLookupFunc) web.Handler {
	return &tagArchiveHandler{context, fn}
}
