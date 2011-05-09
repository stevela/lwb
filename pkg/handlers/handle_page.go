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
	"fmt"
	"github.com/garyburd/twister/web"
	"os"
)

type pageHandler struct {
	context *RenderContext
}

func (ph *pageHandler) ServeWeb(req *web.Request) {
	// Render page.
	fmt.Println(req.URL.Path)
	post, found := ph.context.Db.GetPage(req.URL.Path)
	if !found {
		req.Error(web.StatusNotFound, os.NewError("Not Found."))
		return
	}

	local_context := *ph.context
	local_context.Title = post.Title
	local_context.Path = post.CanonicalBlogUrl.String() + post.CanonicalPath

	var content bytes.Buffer
	renderPost(&content, &local_context, post, post.CommentOnPage)

	// Render page.
	templates["main"].Execute(
		req.Respond(web.StatusOK, web.HeaderContentType, "text/html"),
		makeTemplateParams(&local_context, content.Bytes()))
}

func PageHandler(context *RenderContext) web.Handler {
	return &pageHandler{context}
}
