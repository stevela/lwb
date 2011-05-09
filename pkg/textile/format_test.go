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

package textile

import (
	"bytes"
	"testing"
)

var enttests = []struct {
	in  string
	out string
}{
	{"Hello", "Hello"},
	{"...", "&#8230;"},
	{"....", "&#8230;."},
	{"foo....", "foo&#8230;."},
	{"Hello...", "Hello&#8230;"},
	{"Hello...there", "Hello&#8230;there"},
	{"&", "&#38;"},
	{"&#", "&#"},
	{"& ", "&#38; "},
	{"&#123;", "&#123;"},
	{"&&", "&#38;&#38;"},
	{"\"", "&#34;"},
}

func TestEncodeEntitiesFormatter(t *testing.T) {
	for _, lt := range enttests {
		var buf bytes.Buffer
		EncodeEntitiesFormatter(&buf, "", lt.in)
		bs := buf.String()
		if bs != lt.out {
			t.Errorf("transform(%q) = '%s' want '%s'", lt.in, bs, lt.out)
		}
	}
}
