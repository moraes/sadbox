// Copyright 2013 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmlfilter

import (
	"bytes"
	"exp/html"
	"testing"
)

func TestNextTextFilter(t *testing.T) {
	src := `<html>
	<p>
		<a name="foo"/>
		<small>
			<font face="Arial">
				Foo
				<sup>
					<u>
						<b>
							Bar
						</b>
					</u>
				</sup>
			</font>
		</small>
		<a href="/path/to/somewhere">
			<i>
				Baz
			</i>
		</a>
	</p>

	<p>
		<span>
			Ding
		</span>
	</p>

</html>`

	expected := []string{
		"<p><a/>Foo<sup>Bar</sup><a>Baz</a></p>",
		"<p>Ding</p>",
	}

	r := bytes.NewBufferString(src)
	d := html.NewTokenizer(r)

	for _, v := range expected {
		node, err := NextTextFilter(d, "p", "a", "sup")
		if err != nil {
			t.Fatal(err)
		}
		if node.String() != v {
			t.Errorf("expected %q, got %q", v, node.String())
		}
	}
}

func TestMalformed(t *testing.T) {
	// some mal-formed html
	type test struct {
		src string
		exp string
	}
	tests := []test{
		{`<p><span>Foo</i></p>`, `<html><head></head><body><p><span>Foo</span></p></body></html>`},
	}

	for _, test := range tests {
		r := bytes.NewBufferString(test.src)
		nodes, err := html.ParseFragment(r, nil)
		if err != nil {
			t.Fatal(err)
		}

		b := new(bytes.Buffer)
		for _, v := range nodes {
			err := html.Render(b, v)
			if err != nil {
				t.Error(err)
			}
		}
		if b.String() != test.exp {
			t.Errorf("expected %q, got %q", test.exp, b)
		}
	}
}
