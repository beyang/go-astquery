package astquery

import (
	"fmt"
	"go/ast"
	"reflect"
	"regexp"
)

type Filter interface {
	Filter(node ast.Node) bool
}

// SetFilter matches nodes whose names are in the specified set of names.
type SetFilter struct {
	// Names is a set of names that match the filter
	Names []string

	// Type is the type of AST node to filter for
	Type reflect.Type
}

func (f SetFilter) Filter(node ast.Node) bool {
	nodeName, exists := GetName(node)
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

// RegexpFilter matches nodes whose names match a regular expression.
type RegexpFilter struct {
	// Pattern is a regular expression matching AST node names
	Pattern *regexp.Regexp

	// Type is the type of AST node to filter for
	Type reflect.Type
}

func (s RegexpFilter) Filter(node ast.Node) bool {
	nodeName, exists := GetName(node)
	if !exists {
		return false
	}
	return reflect.TypeOf(node) == s.Type && s.Pattern.MatchString(nodeName)
}

// MethodFilter matches method declaration nodes that have the specified receiver type.
type MethodFilter struct {
	// ReceiverType is the name of the receiver's type (without the '*' if a pointer).
	ReceiverType string

	// ExportedOnly is if the filter should select only exported methods.
	ExportedOnly bool
}

func (f MethodFilter) Filter(node ast.Node) bool {
	switch node := node.(type) {
	case *ast.FuncDecl:
		recv := node.Recv
		if recv == nil || len(recv.List) != 1 {
			return false // not a method
		}
		recvType, _ := typeName(recv.List[0].Type)
		if recvType != f.ReceiverType {
			return false // receiver doesn't match
		}
		if f.ExportedOnly && !node.Name.IsExported() {
			return false // not exported
		}
		return true
	default:
		return false
	}
}

// FilterFunc lets you specify a function for custom filtering logic.
type FilterFunc func(node ast.Node) bool

func (f FilterFunc) Filter(node ast.Node) bool { return f(node) }

// Find recursively searches the AST nodes passed as the first argument and returns all
// AST nodes that match the filter. It does not descend into matching nodes for additional
// matching nodes.
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

// GetName gets the name of a node's identifier. For TypeSpecs and FuncDecls, it looks at the .Name field. For
// SelectorExpr's, it looks at the Sel field.
func GetName(n ast.Node) (name string, exists bool) {
	var ident_ interface{}
	if idt, exists := getStructField(n, "Name"); exists {
		ident_ = idt
	} else if idt, exists := getStructField(n, "Sel"); exists {
		ident_ = idt
	}
	if ident_ == nil {
		return "", false
	}

	nodeName, isIdent := ident_.(*ast.Ident)
	if !isIdent {
		return "", false
	}
	return nodeName.Name, true
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

// typeName returns the name of the type referenced by typeExpr.
func typeName(typeExpr ast.Expr) (string, error) {
	switch typeExpr := typeExpr.(type) {
	case *ast.StarExpr:
		return typeName(typeExpr.X)
	case *ast.Ident:
		return typeExpr.Name, nil
	default:
		return "", fmt.Errorf("expr %+v is not a type expression", typeExpr)
	}
}
