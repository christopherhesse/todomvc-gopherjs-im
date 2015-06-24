package main

import "github.com/gopherjs/gopherjs/js"

const (
	tagText = "_TEXT_"
	tagRaw  = "_RAW_"

	patchReplace         = "replace"
	patchUpdate          = "update"
	patchRemoveLastChild = "remove-last-child"
	patchAppendChild     = "append-child"
)

var (
	Root   *VNode
	Active *VNode
)

type Patch struct {
	Type       string
	Location   []int
	VNode      *VNode
	Attributes map[string]string
}

type VNode struct {
	Tag      string
	Parent   *VNode
	Children []*VNode

	// used by text and raw
	Data string

	// used by normal elements
	Attributes map[string]string
	Styles     map[string]string
}

func NewVNode(tag string) *VNode {
	return &VNode{Tag: tag, Attributes: map[string]string{}, Styles: map[string]string{}}
}

func Init(tag string) {
	Root = NewVNode(tag)
	Active = Root
}

func Done() *VNode {
	if Root != Active {
		panic("not at root element")
	}

	root := Root
	Root = nil
	Active = nil
	return root
}

func RenderNode(vnode *VNode) *js.Object {
	var dnode *js.Object
	switch vnode.Tag {
	case tagText:
		dnode = js.Global.Get("document").Call("createTextNode", vnode.Data)
	case tagRaw:
		dnode = js.Global.Get("document").Call("createRange").Call("createContextualFragment", vnode.Data)
	default:
		dnode = js.Global.Get("document").Call("createElement", vnode.Tag)

		for k, v := range vnode.Attributes {
			dnode.Call("setAttribute", k, v)
		}

		for _, child := range vnode.Children {
			dnode.Call("appendChild", RenderNode(child))
		}
	}

	return dnode
}

func Begin(tag string) {
	vnode := NewVNode(tag)
	Active.Children = append(Active.Children, vnode)
	vnode.Parent = Active
	Active = vnode
}

func sortStrings(a []string, left int, right int) {
	if left >= right {
		return
	}

	// partition
	pivotIndex := (left + right) / 2
	pivot := a[pivotIndex]
	a[pivotIndex], a[right] = a[right], a[pivotIndex]
	storeIndex := left

	for i := left; i < right; i++ {
		if a[i] < pivot {
			a[storeIndex], a[i] = a[i], a[storeIndex]
			storeIndex++
		}
	}
	a[storeIndex], a[right] = a[right], a[storeIndex]

	sortStrings(a, left, storeIndex-1)
	sortStrings(a, storeIndex+1, right)
}

func End(tag string) {
	vnode := Active
	if vnode.Tag != tag {
		panic("attempted to end non-active tag tag=" + tag + " active=" + vnode.Tag)
	}

	// browser implementations of node.style differ, so convert it into an attribute on the node
	if len(vnode.Styles) > 0 {
		keys := []string{}
		for k := range vnode.Styles {
			keys = append(keys, k)
		}
		sortStrings(keys, 0, len(keys)-1)
		style := ""
		for _, k := range keys {
			style += k + ":" + vnode.Styles[k] + ";"
		}
		vnode.Attributes["style"] = style
	}

	Active = Active.Parent
}

func Text(data string) {
	vnode := NewVNode(tagText)
	vnode.Data = data
	Active.Children = append(Active.Children, vnode)
}

func Raw(data string) {
	vnode := NewVNode(tagRaw)
	vnode.Data = data
	Active.Children = append(Active.Children, vnode)
}

func Attr(args ...string) {
	for i := 0; i < len(args); i += 2 {
		Active.Attributes[args[i]] = args[i+1]
	}
}

func Style(args ...string) {
	for i := 0; i < len(args); i += 2 {
		Active.Styles[args[i]] = args[i+1]
	}
}

func Tag(tag string, arg interface{}) {
	Begin(tag)
	switch v := arg.(type) {
	case func():
		v()
	case string:
		Text(v)
	default:
		panic("invalid parameter to Tag")
	}
	End(tag)
}

func DiffNodes(o, n *VNode) []Patch {
	return diffHelper(o, n, []int{})
}

func diffHelper(o, n *VNode, loc []int) []Patch {
	if o == nil || o.Tag != n.Tag || o.Attributes["id"] != n.Attributes["id"] {
		return []Patch{{Type: patchReplace, VNode: n, Location: loc}}
	}

	if (o.Tag == tagText && n.Tag == tagText) || (o.Tag == tagRaw && n.Tag == tagRaw) {
		// these nodes cannot have children
		if o.Data == n.Data {
			return nil
		}
		return []Patch{{Type: patchReplace, VNode: n, Location: loc}}
	}

	updated := false
	attributes := map[string]string{}

	for k := range n.Attributes {
		// added/changed attributes
		if o.Attributes[k] != n.Attributes[k] {
			attributes[k] = n.Attributes[k]
			updated = true
		}
	}

	for k := range o.Attributes {
		// removed attributes
		_, ok := n.Attributes[k]
		if !ok {
			attributes[k] = ""
			updated = true
		}
	}

	patches := []Patch{}

	if updated {
		patches = append(patches, Patch{Type: patchUpdate, Attributes: attributes, Location: loc})
	}

	i := 0
	for i < len(o.Children) && i < len(n.Children) {
		subloc := make([]int, len(loc)+1)
		copy(subloc, loc)
		subloc[len(subloc)-1] = i
		patches = append(patches, diffHelper(o.Children[i], n.Children[i], subloc)...)
		i++
	}

	for i < len(o.Children) {
		patches = append(patches, Patch{Type: patchRemoveLastChild, Location: loc})
		i++
	}

	for i < len(n.Children) {
		patches = append(patches, Patch{Type: patchAppendChild, VNode: n.Children[i], Location: loc})
		i++
	}

	return patches
}

func PatchDOM(patches []Patch, root *js.Object) {
	for _, patch := range patches {
		dnode := root
		for _, i := range patch.Location {
			dnode = dnode.Get("childNodes").Index(i)
		}

		switch patch.Type {
		case patchReplace:
			dnode.Get("parentNode").Call("replaceChild", RenderNode(patch.VNode), dnode)
		case patchRemoveLastChild:
			dnode.Call("removeChild", dnode.Get("childNodes").Index(dnode.Get("childNodes").Get("length").Int()-1))
		case patchAppendChild:
			dnode.Call("appendChild", RenderNode(patch.VNode))
		case patchUpdate:
			for k, v := range patch.Attributes {
				if v == "" {
					dnode.Call("removeAttribute", k)
				} else {
					dnode.Call("setAttribute", k, v)
				}
			}
		}
	}
}

func Div(arg interface{}) {
	Tag("div", arg)
}

func Id(v string) {
	Attr("id", v)
}

func Debug(s string) {
	Attr("data-debug", s)
}
