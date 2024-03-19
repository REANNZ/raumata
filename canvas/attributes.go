package canvas

// Attributes for canvas objects
type Attributes struct {
	Id      string
	Style   *Style
	Classes []string
	Extra   map[string]any
}

// EnsureStyle ensures that a.Style is not
// nil
func (a *Attributes) EnsureStyle() {
	if a.Style == nil {
		a.Style = NewStyle()
	}
}

// AddClass adds the given class to Classes
func (a *Attributes) AddClass(class string) {
	for _, cls := range a.Classes {
		if cls == class {
			return
		}
	}

	a.Classes = append(a.Classes, class)
}

// RemoveClass removes the given class from Classes.
//
// If the class isn't in Classes, it does nothing
func (a *Attributes) RemoveClass(class string) {
	p := -1
	for i, cls := range a.Classes {
		if cls == class {
			p = i
			break
		}
	}

	if p >= 0 {
		// Since order doesn't matter, remove the class
		// by moving the last element into it's position
		// and shrinking the list
		a.Classes[p] = a.Classes[len(a.Classes)-1]
		a.Classes = a.Classes[:len(a.Classes)-1]
	}
}

func (a *Attributes) SetExtra(name string, val any) {
	if a.Extra == nil {
		a.Extra = make(map[string]any)
	}

	a.Extra[name] = val
}
