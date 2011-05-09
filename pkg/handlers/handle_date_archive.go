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
	"strconv"
)

type dateArchiveHandler struct {
	context *RenderContext
}

func (dah *dateArchiveHandler) ServeWeb(req *web.Request) {
	dah.context.Config.Cache.Run(req, func(w io.Writer) bool {
		yearStr := req.Param.Get("year")
		monthStr := req.Param.Get("month")

		var posts []*store.Post
		var found = false
		if yearStr != "" {
			year, _ := strconv.Atoi(yearStr)
			if monthStr != "" {
				month, _ := strconv.Atoi(monthStr)
				posts, found = dah.context.Db.GetPostsByYearMonth(year, month)
			} else {
				posts, found = dah.context.Db.GetPostsByYear(year)
			}
		}

		if !found {
			return false
		}

		// Render posts.
		var content bytes.Buffer
		for _, post := range posts {
			renderPost(&content, dah.context, post, false)
		}

		// Render page.
		templates["main"].Template.Execute(w, makeTemplateParams(dah.context, content.Bytes()))

		return true
	})
}


// DataArchiveHandler returns a request handler that serves date archives.
func DateArchiveHandler(context *RenderContext) web.Handler {
	return &dateArchiveHandler{context}
}
