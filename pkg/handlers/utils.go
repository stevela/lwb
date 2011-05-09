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
	"github.com/stevela/lwb/store"
	"bytes"
	"io"
)

// RenderPost renders a single post.
func renderPost(w io.Writer, context *RenderContext, post *store.Post, withFeedback bool) {
	if context.UseCache {
		if withFeedback {
			if post.CachedPostWithFeedback != nil {
				w.Write(post.CachedPostWithFeedback)
				return
			}
		} else {
			if post.CachedPost != nil {
				w.Write(post.CachedPost)
				return
			}
		}
	}

	data := makeTemplateParams(context, post)

	if withFeedback {
		var feedback bytes.Buffer
		templates["feedback"].Execute(&feedback, data)
		data["feedback"] = feedback.Bytes()
		data["show-footer"] = true
	}

	buf := &bytes.Buffer{}
	templates["post"].Execute(buf, data)
	b := buf.Bytes()

	if context.UseCache {
		if withFeedback {
			post.CachedPostWithFeedback = b
		} else {
			post.CachedPost = b
		}
	}

	w.Write(b)
}


func makeTemplateParams(context *RenderContext, content interface{}) map[string]interface{} {
	return map[string]interface{}{
		"context": context,
		"content": content,
	}
}
