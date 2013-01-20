// Copyright 2013 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlfilter

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// Node stores a tree of nodes we are interested in.
type Node struct {
	Token xml.Token
	List  []Node
}

// String returns a simplified string version of the node, just for debugging
// and testing purposes. It doesn't include tag attributes or namespaces.
func (n Node) String() string {
	switch t := n.Token.(type) {
	case xml.StartElement:
		b := new(bytes.Buffer)
		fmt.Fprintf(b, "<%s>", t.Name.Local)
		for _, v := range n.List {
			fmt.Fprintf(b, "%s", v)
		}
		fmt.Fprintf(b, "</%s>", t.Name.Local)
		return b.String()
	case xml.CharData:
		return string(t)
	case string:
		return t
	}
	panic(fmt.Errorf("unexpected %T", n.Token))
}

/*
NextNames looks for the CharData contained in the given tag names. It trims
spaces from the CharData and ignores it if it is empty.

It expects well-formed XML: mismatching closing tags will result in error.

Given the following horrible HTML snippet:

	<?xml version="1.0" encoding="utf-8" standalone="yes"?>
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

Calling the following:

	NextNames(xml.Name{Local:"p"}, xml.Name{Local:"a"}, xml.Name{Local:"sup"})

...results in a Node tree that contains the following HTML structure:

	<p><a></a>Foo<sup>Bar</sup><a>Baz</a></p>

And making the same call again results in a Node with:

	<p>Ding</p>
*/
func NextNames(d *xml.Decoder, names ...xml.Name) (Node, error) {
	t, err := skipUntilStartElement(d, names)
	if err != nil {
		return Node{}, err
	}
	list, err := parseList(d, names, []xml.Name{t.Name})
	if err != nil {
		return Node{}, err
	}
	return Node{Token: t, List: list}, nil
}

func skipUntilStartElement(d *xml.Decoder, names []xml.Name) (xml.StartElement, error) {
	for {
		t, err := d.RawToken()
		if err != nil {
			return xml.StartElement{}, err
		}
		if start, ok := t.(xml.StartElement); ok {
			for _, v := range names {
				if v == start.Name {
					return start, nil
				}
			}
		}
	}
	panic("unreachable")
}

func parseList(d *xml.Decoder, names, stack []xml.Name) ([]Node, error) {
	var c []Node
	for len(stack) > 0 {
		token, err := d.RawToken()
		if err != nil {
			return nil, fmt.Errorf("unclosed tags: %v", stack)
		}
		// A token can be of the following types:
		//
		//   xml.CharData
		//   xml.Comment
		//   xml.Directive
		//   xml.EndElement
		//   xml.ProcInst
		//   xml.StartElement
		switch t := token.(type) {
		case xml.StartElement:
			found := false
			for _, v := range names {
				if v == t.Name {
					found = true
					list, err := parseList(d, names, []xml.Name{t.Name})
					if err != nil {
						return nil, err
					}
					c = append(c, Node{Token: t, List: list})
				}
			}
			if !found {
				stack = append(stack, t.Name)
			}
		case xml.EndElement:
			if stack, err = popName(stack, t.Name); err != nil {
				return nil, err
			}
		case xml.CharData:
			if b := bytes.TrimSpace(t); len(b) > 0 {
				// To pass xml.CharData, we would need a copy of b:
				//
				//     b1 := make(xml.CharData, len(b))
				//     copy(b1, b)
				//
				// Instead of this, we just convert it to string.
				c = append(c, Node{Token: string(b)})
			}
		}
	}
	return c, nil
}

func popName(names []xml.Name, name xml.Name) ([]xml.Name, error) {
	size := len(names)
	if size == 0 {
		return nil, fmt.Errorf("unexpected closing tag %q", name)
	}
	if last := names[size-1]; last != name {
		return nil, fmt.Errorf("expecting closing tag %q, got %q", last, name)
	}
	return names[:size-1], nil
}
