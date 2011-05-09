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

type singlePostHandler struct {
	context *RenderContext
}

func (sph *singlePostHandler) ServeWeb(req *web.Request) {
	sph.context.Config.Cache.Run(req, func(w io.Writer) bool {
		// Render post.
		post, found := sph.context.Db.GetPostByPath(req.URL.Path)
		if !found {
			return false
		}

		local_context := *sph.context
		local_context.Title = post.Title
		local_context.Path = post.CanonicalBlogUrl.String() + post.CanonicalPath

		var content bytes.Buffer
		renderPost(&content, &local_context, post, true)

		// Render page.
		templates["main"].Execute(w, makeTemplateParams(&local_context, content.Bytes()))

		return true
	})
}

func SinglePostHandler(context *RenderContext) web.Handler {
	return &singlePostHandler{context}
}
