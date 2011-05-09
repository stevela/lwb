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
	"fmt"
	"io"
)

var (
	esc_elipses = []byte("&#8230;")
	esc_amp     = []byte("&#38;")
	esc_quote   = []byte("&#34;")
	esc_lquote  = []byte("&#8220;")
	esc_rquote  = []byte("&#8221;")
	esc_lt      = []byte("&#60;")
	esc_gt      = []byte("&#62;")
)

func EntityEscape(w io.Writer, b []byte, encode_all, smart_quotes bool) {
	var esc []byte

	last := 0
	skip := 1
	in_tag := false
	found_closing := -1
	for i := 0; i < len(b); i += skip {
		skip = 1

		if in_tag && b[i] == '>' {
			in_tag = false
		}

		if in_tag {
			continue
		}

		switch b[i] {
		case '<':
			if encode_all {
				esc = esc_lt
			} else {
				in_tag = true
				continue
			}
		case '>':
			if encode_all {
				esc = esc_gt
			} else {
				continue
			}
		case '"':
			if !smart_quotes {
				esc = esc_quote
			} else if found_closing >= 0 {
				if found_closing != i {
					panic("Found closing quote in wrong place.")
				} else {
					esc = esc_rquote
					found_closing = -1
				}
			} else {
				in_tag_in_quote := false
				for j := i + 1; j < len(b) && found_closing < 0; j += 1 {
					if in_tag_in_quote {
						if b[j] == '>' {
							in_tag_in_quote = false
						}

						continue
					}
					switch b[j] {
					case '<':
						in_tag_in_quote = true
					case '"':
						found_closing = j
					}
				}

				if found_closing >= 0 {
					esc = esc_lquote
				} else {
					esc = esc_quote
				}
			}
		case '.':
			if len(b) > i+2 && b[i+1] == '.' && b[i+2] == '.' {
				esc = esc_elipses
				skip = 3
			} else {
				continue
			}
		case '&':
			if len(b) > i+1 && (b[i+1]) == '#' {
				continue
			} else {
				esc = esc_amp
			}
		default:
			continue
		}

		w.Write(b[last:i])
		w.Write(esc)
		last = i + skip
	}

	w.Write(b[last:])
}

func EntityEscapeString(s string, encode_all, smart_quotes bool) string {
	var b = &bytes.Buffer{}
	EntityEscape(b, []byte((s)), encode_all, smart_quotes)
	return b.String()
}

// EncodeEntitiesFormatter replaces various entities. An extension of
// HTMLFormatter.
func EncodeEntitiesFormatter(w io.Writer, format string, value ...interface{}) {
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
	EntityEscape(w, b, false, true)
}
