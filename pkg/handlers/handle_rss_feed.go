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
	"time"
)

type rssFeedHandler struct {
	context *RenderContext
}

func (rfh *rssFeedHandler) ServeWeb(req *web.Request) {
	rfh.context.Config.Cache.Run(req, func(w io.Writer) bool {
		// Render posts.
		var content bytes.Buffer
		posts := rfh.context.Db.GetRecentPosts(rfh.context.Config.NumRssFeedPosts)
		for _, post := range posts {
			data := makeTemplateParams(rfh.context, post)
			templates["rss_item"].Template.Execute(&content, data)
		}

		// Render page.
		data := makeTemplateParams(rfh.context, content.Bytes())
		data["lastBuildDate"] = posts[0].Published.Format(time.RFC1123)

		templates["rss"].Template.Execute(w, data)

		return true
	})
}

func RssFeedHandler(context *RenderContext) web.Handler {
	return &rssFeedHandler{context}
}
