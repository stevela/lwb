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
	"io"
)

type mainIndexHandler struct {
	context *RenderContext
}

func (mih *mainIndexHandler) ServeWeb(req *web.Request) {
	mih.context.Config.Cache.Run(req, func(w io.Writer) bool {
		// Render posts.
		var content bytes.Buffer
		for _, post := range mih.context.Db.GetRecentPosts(mih.context.Config.NumMainIndexPosts) {
			renderPost(&content, mih.context, post, false)
		}

		// Render page.
		templates["main"].Template.Execute(w, makeTemplateParams(mih.context, content.Bytes()))

		return true
	})
}


// MainIndexHandler returns a request handler that serves the main index.
func MainIndexHandler(context *RenderContext) web.Handler {
	return &mainIndexHandler{context}
}
