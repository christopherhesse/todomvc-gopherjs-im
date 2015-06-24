package main

import (
	"fmt"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

type Todo struct {
	Id        int
	Text      string
	Completed bool
}

var (
	PreviousRoot *VNode
	todos        = []Todo{{
		Id:        0,
		Text:      "hello0",
		Completed: true,
	}, {
		Id:        1,
		Text:      "hello1",
		Completed: false,
	}, {
		Id:        2,
		Text:      "hello2",
		Completed: false,
	}}
	todoCounter   = 3
	editingTodoId = -1
)

const (
	svgCheckedCheckbox = `<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="-10 -18 100 135"><circle cx="50" cy="50" r="50" fill="none" stroke="#bddad5" stroke-width="3"/><path fill="#5dc2af" d="M72 25L42 71 27 56l-4 4 20 20 34-52z"/></svg>`
	svgEmptyCheckbox   = `<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="-10 -18 100 135"><circle cx="50" cy="50" r="50" fill="none" stroke="#ededed" stroke-width="3"/></svg>`

	keyEnter = 13
	keyEsc   = 27

	filterAll       = "all"
	filterActive    = "active"
	filterCompleted = "completed"
)

func main() {
	Setup()
	Rerender()
}

func render() {
	Init("body")

	Div(func() {
		Style(
			"background", "#fff",
			"margin", "130px 0px 40px 0px",
			"position", "relative",
			"box-shadow", "0 2px 4px 0 rgba(0, 0, 0, 0.2), 0 25px 50px 0 rgba(0, 0, 0, 0.1)",
		)

		Div(func() {
			Tag("h1", func() {
				Style(
					"position", "absolute",
					"top", "-155px",
					"width", "100%",
					"font-size", "100px",
					"font-weight", "100",
					"text-align", "center",
					"color", "rgba(175, 47, 47, 0.15)",
					"text-rendering", "optimizeLegibility",
				)

				Text("todos")
			})

			DrawNewTodo()
		})

		if len(todos) == 0 {
			// shouldn't draw the rest of this stuff in this case
			return
		}

		DrawTodos()

		DrawFooter()
	})

	Div(func() {
		Id("info")

		Style(
			"margin", "65px auto 0",
			"color", "#bfbfbf",
			"font-size", "10px",
			"text-align", "center",
			"text-shadow", "0px 1px 0px rgba(255, 255, 255, 0.5)",
		)

		Raw(`<p>Double-click to edit a todo</p>`)
		Raw(`<p>Written by Christopher Hesse</p>`)
		Raw(`<p>Part of <a href="http://todomvc.com">TodoMVC</a></p>`)
	})

	root := Done()
	patches := DiffNodes(PreviousRoot, root)
	PatchDOM(patches, js.Global.Get("document").Get("body"))
	PreviousRoot = root
}

func getActiveFilter() string {
	filter := js.Global.Get("window").Get("location").Get("hash").String()
	if len(filter) <= 2 {
		return filterAll
	}
	return filter[2:]
}

func setActiveFilter(filter string) {
	js.Global.Get("history").Call("pushState", nil, "", "#/"+filter)
}

func DrawNewTodo() {
	Tag("input", func() {
		newTodo := "new-todo"
		Id(newTodo)

		Style(
			"background", "rgba(0, 0, 0, 0.003)",
			"border", "none",
			"padding", "16px 16px 16px 60px",
			"line-height", "1.4em",
			"font-size", "24px",
			"position", "relative",
			"box-shadow", "inset 0 -2px 1px rgba(0,0,0,0.03)",
		)

		Attr("placeholder", "What needs to be done?")

		if Keyup(newTodo, keyEnter) {
			value := InputValues[newTodo]
			if len(value) > 0 {
				InputValues[newTodo] = ""

				todos = append(todos, Todo{Id: todoCounter, Text: value, Completed: false})
				todoCounter++
				Rerender()
			}
		}
	})
}

func DrawTodos() {
	Div(func() {
		Style(
			"position", "relative",
			"z-index", "2",
			"border-top", "1px solid #e6e6e6",
		)

		Div(func() {
			toggleAll := "toggle-all"
			Id(toggleAll)

			Style(
				"position", "absolute",
				"top", "-65px",
				"width", "34px",
				"height", "65px",
				"cursor", "pointer",
			)

			allCompleted := true
			for _, todo := range todos {
				if !todo.Completed {
					allCompleted = false
					break
				}
			}

			if Clicked(toggleAll) {
				completed := true
				if allCompleted {
					completed = false
				}

				for i := range todos {
					todos[i].Completed = completed
				}

				Rerender()
			}

			Div(func() {
				Style(
					"position", "absolute",
					"top", "21px",
					"left", "18px",
					"font-size", "22px",
					"transform", "rotate(90deg)",
					"color", "#e6e6e6",
					"user-select", "none",
					"-webkit-user-select", "none",
					"-moz-user-select", "none",
					"-ms-user-select", "none",
				)

				if allCompleted {
					Style("color", "#737373")
				}

				Text("❯")
			})
		})

		Div(func() {
			activeFilter := getActiveFilter()
			for i := 0; i < len(todos); i++ {
				todo := &todos[i]
				if activeFilter == filterCompleted && !todo.Completed {
					continue
				}
				if activeFilter == filterActive && todo.Completed {
					continue
				}

				Div(func() {
					DrawTodo(todo)

					if i == len(todos)-1 {
						Style("border-bottom", "none")
					}
				})
			}
		})
	})
}

func DrawTodo(todo *Todo) {
	item := "todo-item-" + strconv.Itoa(todo.Id)
	Id(item)
	editTodo := "edit-" + item

	Style(
		"position", "relative",
		"font-size", "24px",
		"border-bottom", "1px solid #ededed",
	)

	if editingTodoId == todo.Id {
		DrawEditingTodo(editTodo, todo)
	} else {
		DrawNormalTodo(item, editTodo, todo)
	}
}

func DrawEditingTodo(editTodo string, todo *Todo) {
	Style(
		"border-bottom", "none",
		"padding", "0px",
		"margin-bottom", "-1px",
	)

	Tag("input", func() {
		Id(editTodo)

		Style(
			"position", "relative",
			"font-size", "24px",
			"line-height", "1.4em",
			"padding", "6px",
			"border", "1px solid #999",
			"box-shadow", "inset 0 -1px 5px 0 rgba(0, 0, 0, 0.2)",
			"display", "block",
			"width", "471px",
			"padding", "13px 17px 12px 17px",
			"margin", "0px 0px 0px 43px",
		)

		Attr("value", todo.Text)

		if Keyup(editTodo, keyEnter) {
			todo.Text = InputValues[editTodo]
			editingTodoId = -1
			Rerender()
		}

		if Keyup(editTodo, keyEsc) || !Focused(editTodo) {
			editingTodoId = -1
			Rerender()
		}
	})
}

func DrawNormalTodo(item string, editTodo string, todo *Todo) {
	Div(func() {
		checkbox := "checkbox-" + item
		Id(checkbox)

		Style(
			"display", "inline",
			"text-align", "center",
			"width", "40px",
			"height", "40px",
			"position", "absolute",
			"top", "0px",
			"bottom", "0px",
			"margin", "auto 0px",
			"cursor", "pointer",
		)

		if Clicked(checkbox) {
			todo.Completed = !todo.Completed
			Rerender()
		}

		if todo.Completed {
			Raw(svgCheckedCheckbox)
		} else {
			Raw(svgEmptyCheckbox)
		}
	})

	Div(func() {
		textbox := "text-" + item
		Id(textbox)

		Style(
			"display", "inline",
			"white-space", "pre",
			"word-break", "break-word",
			"padding", "15px 60px 15px 15px",
			"margin-left", "45px",
			"display", "block",
			"line-height", "1.2em",
			"transition", "color 0.4s",
		)

		if todo.Completed {
			Style(
				"color", "#d9d9d9",
				"text-decoration", "line-through",
			)
		}

		if DoubleClicked(textbox) {
			editingTodoId = todo.Id
			Focus(editTodo)
			Rerender()
		}

		Text(todo.Text)
	})

	Div(func() {
		destroy := "destroy-" + item
		Id(destroy)

		Style(
			"text-align", "center",
			"cursor", "pointer",
			"position", "absolute",
			"top", "0px",
			"right", "10px",
			"width", "40px",
			"height", "58px",
		)

		if Clicked(destroy) {
			for i := range todos {
				if todos[i].Id == todo.Id {
					todos = append(todos[:i], todos[i+1:]...)
					break
				}
			}
			Rerender()
		}

		Div(func() {
			Style(
				"position", "absolute",
				"font-size", "30px",
				"top", "15px",
				"left", "11px",
				"display", "none",
				"color", "#cc9a9a",
				"transition", "color 0.2s ease-out",
			)

			if Hovering(destroy) {
				Style("color", "#af5b5e")
			}

			if Hovering(item) {
				Style("display", "block")
			}

			Text("×")
		})
	})
}

func DrawFooter() {
	// pattern below footer
	Div(func() {
		Style(
			"position", "absolute",
			"right", "0px",
			"bottom", "0px",
			"left", "0px",
			"height", "50px",
			"overflow", "hidden",
			"box-shadow", "0 1px 1px rgba(0, 0, 0, 0.2), 0 8px 0 -3px #f6f6f6, 0 9px 1px -3px rgba(0, 0, 0, 0.2), 0 16px 0 -6px #f6f6f6, 0 17px 2px -6px rgba(0, 0, 0, 0.2)",
		)
	})

	Div(func() {
		Style(
			"color", "#777",
			"padding", "10px 15px",
			"height", "40px",
			"text-align", "center",
			"border-top", "1px solid #e6e6e6",
		)

		Div(func() {
			Style(
				"display", "inline",
				"float", "left",
				"text-align", "left",
			)

			remaining := 0
			for _, todo := range todos {
				if !todo.Completed {
					remaining++
				}
			}

			if remaining == 1 {
				Text("1 item left")
			} else {
				Text(fmt.Sprintf("%d items left", remaining))
			}
		})

		// filters
		Div(func() {
			Style(
				"position", "absolute",
				"right", "0px",
				"left", "0px",
			)

			createFilterButton := func(name, filter string) {
				Div(func() {
					button := "filter-" + filter
					Id(button)

					Style(
						"display", "inline",
						"margin", "3px",
						"padding", "3px 7px",
						"text-decoration", "none",
						"border", "1px solid transparent",
						"border-radius", "3",
						"cursor", "pointer",
					)

					if Hovering(button) {
						Style("border-color", "rgba(175, 47, 47, 0.1)")
					}

					if getActiveFilter() == filter {
						Style("border-color", "rgba(175, 47, 47, 0.2)")
					}

					if Clicked(button) {
						setActiveFilter(filter)
						Rerender()
					}

					Text(name)
				})
			}

			createFilterButton("All", filterAll)
			createFilterButton("Active", filterActive)
			createFilterButton("Completed", filterCompleted)
		})

		anyCompleted := false
		for _, todo := range todos {
			if todo.Completed {
				anyCompleted = true
				break
			}
		}

		if anyCompleted {
			Div(func() {
				button := "clear-completed"
				Id(button)

				Style(
					"display", "inline",
					"float", "right",
					"position", "relative",
					"line-height", "20px",
					"text-decoration", "none",
					"cursor", "pointer",
				)

				if Hovering(button) {
					Style("text-decoration", "underline")
				}

				if Clicked(button) {
					newTodos := []Todo{}
					for _, todo := range todos {
						if !todo.Completed {
							newTodos = append(newTodos, todo)
						}
					}
					todos = newTodos
					Rerender()
				}

				Text("Clear Completed")
			})
		}
	})
}
