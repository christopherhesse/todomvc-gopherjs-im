package main

import "github.com/gopherjs/gopherjs/js"

var (
	rendering bool

	clickable       map[string]bool
	doubleClickable map[string]bool
	hoverable       map[string]bool
	keyupable       map[string]bool
	keyupableCodes  map[string][]int

	InputValues = map[string]string{}

	clickId        = ""
	doubleClickId  = ""
	focusId        = ""
	focusSelection = [2]int{}
	keyupId        = ""
	keyupCode      = 0
	hoverIds       = []string{}
)

func Rerender() {
	js.Global.Get("window").Call("requestAnimationFrame", Frame)
}

func Frame() {
	startMs := js.Global.Get("Date").Call("now").Int()

	rendering = true

	// clear monitoring state
	clickable = map[string]bool{}
	doubleClickable = map[string]bool{}
	hoverable = map[string]bool{}
	keyupable = map[string]bool{}
	keyupableCodes = map[string][]int{}

	// store values for inputs
	inputs := js.Global.Get("document").Call("getElementsByTagName", "input")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		id := input.Get("id").String()
		if id != "" {
			InputValues[id] = input.Get("value").String()
		}
	}

	// store selection
	if focusId != "" {
		elem := js.Global.Get("document").Call("getElementById", focusId)
		if elem != nil {
			focusSelection[0] = elem.Get("selectionStart").Int()
			focusSelection[1] = elem.Get("selectionEnd").Int()
		}
	}

	render()

	// restore values for inputs
	for id, value := range InputValues {
		elem := js.Global.Get("document").Call("getElementById", id)
		if elem != nil {
			elem.Set("value", value)
		}
	}

	// set focused element and any selection data
	if focusId != "" {
		elem := js.Global.Get("document").Call("getElementById", focusId)
		if elem != nil {
			// focus will normally cause the page to scroll, which we don't want, so scroll back afterward
			x := js.Global.Get("window").Get("scrollX").Int()
			y := js.Global.Get("window").Get("scrollY").Int()
			elem.Call("focus")
			js.Global.Get("window").Call("scrollTo", x, y)
			if focusSelection == [2]int{-1, -1} {
				elem.Set("selectionStart", elem.Get("value").Get("length"))
				elem.Set("selectionEnd", elem.Get("value").Get("length"))
			} else {
				elem.Set("selectionStart", focusSelection[0])
				elem.Set("selectionEnd", focusSelection[1])
			}
		}
	}

	// clear user input state
	clickId = ""
	doubleClickId = ""
	keyupId = ""
	keyupCode = 0
	InputValues = map[string]string{}
	focusSelection = [2]int{-1, -1}

	rendering = false

	endMs := js.Global.Get("Date").Call("now").Int()
	print("render", endMs-startMs, "ms")
}

func Setup() {
	js.Global.Get("document").Call("addEventListener", "click", func(e *js.Object) {
		ids := findIds(e.Get("target"), clickable)
		if len(ids) > 0 {
			clickId = ids[0]
			Rerender()
		}
	})

	js.Global.Get("document").Call("addEventListener", "dblclick", func(e *js.Object) {
		ids := findIds(e.Get("target"), doubleClickable)
		if len(ids) > 0 {
			doubleClickId = ids[0]
			Rerender()
		}
	})

	js.Global.Get("document").Call("addEventListener", "focus", func(e *js.Object) {
		if rendering {
			return
		}

		newFocusId := e.Get("target").Get("id").String()

		if newFocusId != focusId {
			focusId = newFocusId
			Rerender()
		}
	}, true) // use capture mode because firefox does not support focusin

	js.Global.Get("document").Call("addEventListener", "blur", func(e *js.Object) {
		if rendering {
			return
		}

		if focusId != "" {
			focusId = ""
			Rerender()
		}
	}, true) // use capture mode because firefox does not support focusin

	js.Global.Get("document").Call("addEventListener", "keyup", func(e *js.Object) {
		ids := findIds(e.Get("target"), keyupable)
		if len(ids) > 0 {
			id := ids[0]
			for _, keycode := range keyupableCodes[id] {
				if keycode == e.Get("keyCode").Int() {
					keyupId = id
					keyupCode = keycode
					Rerender()
				}
			}
		}
	})

	js.Global.Get("document").Call("addEventListener", "mouseover", func(e *js.Object) {
		ids := findIds(e.Get("target"), hoverable)

		shouldRender := false
		if len(ids) == len(hoverIds) {
			for i := range ids {
				if ids[i] != hoverIds[i] {
					shouldRender = true
					break
				}
			}
		} else {
			shouldRender = true
		}

		if shouldRender {
			hoverIds = ids
			Rerender()
		}
	})

	js.Global.Get("document").Call("addEventListener", "hashchange", func(e *js.Object) {
		Rerender()
	})
}

func findIds(element *js.Object, set map[string]bool) []string {
	ids := []string{}
	for element != nil {
		id := element.Get("id").String()
		if set[id] {
			ids = append(ids, id)
		}
		element = element.Get("parentNode")
	}
	return ids
}

func Clicked(id string) bool {
	clickable[id] = true
	return clickId == id
}

func DoubleClicked(id string) bool {
	doubleClickable[id] = true
	return doubleClickId == id
}

func Hovering(id string) bool {
	hoverable[id] = true
	for _, hoverId := range hoverIds {
		if hoverId == id {
			return true
		}
	}
	return false
}

func Keyup(id string, keycode int) bool {
	keyupable[id] = true
	keyupableCodes[id] = append(keyupableCodes[id], keycode)

	return keyupId == id && keyupCode == keycode
}

func Focus(id string) {
	focusId = id
	focusSelection = [2]int{-1, -1}
}

func Focused(id string) bool {
	return focusId == id
}
