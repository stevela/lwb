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
	"fmt"
	"io"
	"regexp"
)

var (
	escSpace = []byte("%20")
)

// EncodeSpacesFormatter replaces spaces with %20.
func EncodeSpacesFormatter(w io.Writer, format string, value ...interface{}) {
	ok := false
	var b []byte
	if len(value) == 1 {
		b, ok = value[0].([]byte)
	}
	if !ok {
		var buf bytes.Buffer
		fmt.Fprint(&buf, value...)
		b = buf.Bytes()
	}

	var esc []byte
	last := 0
	for i, c := range b {
		switch c {
		case ' ':
			esc = escSpace
		default:
			continue
		}
		w.Write(b[last:i])
		w.Write(esc)
		last = i + 1
	}
	w.Write(b[last:])
}

var lineRegexp = regexp.MustCompile("[^\\r\\n]*")

// ConvertBreaksFormatter creates paragraphs split on blank lines.
func ConvertBreaksFormatter(w io.Writer, format string, value ...interface{}) {
	for _, val := range value {
		for _, line := range lineRegexp.FindAllString(val.(string), -1) {
			if len(line) != 0 {
				fmt.Fprintf(w, "<p>%s</p>", line)
			}
		}
	}
}
