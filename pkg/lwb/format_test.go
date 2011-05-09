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

package lwb

import (
	"bytes"
	"testing"
)

var spacesTests = []struct {
	in  string
	out string
}{
	{"Hello", "Hello"},
	{"Hello there", "Hello%20there"},
	{"/tag/seattle times", "/tag/seattle%20times"},
}

var convertTests = []struct {
	in  string
	out string
}{
	{"", ""},
	{"Hello", "<p>Hello</p>"},
	{"Hello\nthere", "<p>Hello</p><p>there</p>"},
}

func TestEncodeSpacesFormatter(t *testing.T) {
	for _, lt := range spacesTests {
		var buf bytes.Buffer
		EncodeSpacesFormatter(&buf, "", lt.in)
		bs := buf.String()
		if bs != lt.out {
			t.Errorf("transform(%q) = '%s' want '%s'", lt.in, bs, lt.out)
		}
	}
}

func TestConvertBreaksFormatter(t *testing.T) {
	for _, lt := range convertTests {
		var buf bytes.Buffer
		ConvertBreaksFormatter(&buf, "", lt.in)
		bs := buf.String()
		if bs != lt.out {
			t.Errorf("transform(%q) = '%s' want '%s'", lt.in, bs, lt.out)
		}
	}
}
