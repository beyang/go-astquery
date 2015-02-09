package astquery

import (
	"go/ast"
	"regexp"

	"code.google.com/p/go/src/pkg/reflect"
)

type Filter interface {
	Filter(node ast.Node) bool
}

type SetFilter struct {
	// Names is a set of names that match the filter
	Names []string

	// Type is the type of AST node to filter for
	Type reflect.Type
}

func (f SetFilter) Filter(node ast.Node) bool {
	nodeName, exists := getName(node)
	if !exists {
		return false
	}

	matched := false
	for _, name := range f.Names {
		if name == nodeName {
			matched = true
			break
		}
	}
	return reflect.TypeOf(node) == f.Type && matched
}

type RegexpFilter struct {
	// Pattern is a regular expression matching AST node names
	Pattern regexp.Regexp

	// Type is the type of AST node to filter for
	Type reflect.Type
}

func (s RegexpFilter) Filter(node ast.Node) bool {
	nodeName, exists := getName(node)
	if !exists {
		return false
	}
	return reflect.TypeOf(node) == s.Type && s.Pattern.MatchString(nodeName)
}

func Find(nodes []ast.Node, filter Filter) []ast.Node {
	var found []ast.Node
	for _, node := range nodes {
		found = append(found, find(node, filter)...)
	}
	return found
}

func find(node ast.Node, filter Filter) []ast.Node {
	var found []ast.Node
	ast.Walk(visitFunc(func(node ast.Node) bool {
		if filter.Filter(node) {
			found = append(found, node)
			return false
		}
		return true
	}), node)
	return found
}

// visitFunc is a wrapper for traversing nodes in the AST
type visitFunc func(node ast.Node) (descend bool)

func (v visitFunc) Visit(node ast.Node) ast.Visitor {
	descend := v(node)
	if descend {
		return v
	} else {
		return nil
	}
}

func getName(n ast.Node) (name string, exists bool) {
	nameValue, exists := getStructField(n, "Name")
	if !exists {
		return "", false
	}
	nodeName, isStr := nameValue.(string)
	if !isStr {
		return "", false
	}
	return nodeName, true
}

// getStructField returns the value of v's field with the given name
// if it exists. v must be a struct or a pointer to a struct.
func getStructField(v interface{}, field string) (fieldVal interface{}, exists bool) {
	vv := reflect.ValueOf(v)
	if !vv.IsValid() {
		return nil, false
	}
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	fv := vv.FieldByName(field)
	if !fv.IsValid() {
		return nil, false
	}
	return fv.Interface(), true
}
