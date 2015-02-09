package astquery

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"testing"
)

type nodeInfo struct {
	Name string
	Type reflect.Type
}

func TestSetFilter(t *testing.T) {
	servicePkg := getTestPkg(t)

	serviceTypes := Find([]ast.Node{servicePkg}, SetFilter{
		Names: []string{"ServiceOne", "ServiceTwo"},
		Type:  reflect.TypeOf((*ast.TypeSpec)(nil)),
	})
	expServiceTypes := []nodeInfo{
		{Name: "ServiceOne", Type: reflect.TypeOf((*ast.TypeSpec)(nil))},
		{Name: "ServiceTwo", Type: reflect.TypeOf((*ast.TypeSpec)(nil))},
	}
	checkNodesExpected(t, expServiceTypes, serviceTypes)
}

func TestRegexpFilter(t *testing.T) {
	servicePkg := getTestPkg(t)

	serviceTypes := Find([]ast.Node{servicePkg}, RegexpFilter{
		Pattern: regexp.MustCompile(`^Service[A-Za-z]*$`),
		Type:    reflect.TypeOf((*ast.TypeSpec)(nil)),
	})
	expServiceTypes := []nodeInfo{
		{Name: "ServiceOne", Type: reflect.TypeOf((*ast.TypeSpec)(nil))},
		{Name: "ServiceTwo", Type: reflect.TypeOf((*ast.TypeSpec)(nil))},
	}
	checkNodesExpected(t, expServiceTypes, serviceTypes)
}

func TestNestedFilters(t *testing.T) {
	servicePkg := getTestPkg(t)

	type expMethodInfo struct {
		method nodeInfo
		calls  []nodeInfo
	}
	type expServiceInfo struct {
		service nodeInfo
		methods []expMethodInfo
	}
	testcases := []expServiceInfo{{
		service: nodeInfo{Name: "ServiceOne", Type: reflect.TypeOf((*ast.TypeSpec)(nil))},
		methods: []expMethodInfo{{
			method: nodeInfo{Name: "Get", Type: reflect.TypeOf((*ast.FuncDecl)(nil))},
			calls:  []nodeInfo{{Name: "Check", Type: reflect.TypeOf((*ast.SelectorExpr)(nil))}},
		}, {
			method: nodeInfo{Name: "List", Type: reflect.TypeOf((*ast.FuncDecl)(nil))},
			calls:  []nodeInfo{{Name: "Check", Type: reflect.TypeOf((*ast.SelectorExpr)(nil))}},
		}},
	}, {
		service: nodeInfo{Name: "ServiceTwo", Type: reflect.TypeOf((*ast.TypeSpec)(nil))},
		methods: []expMethodInfo{{
			method: nodeInfo{Name: "Get", Type: reflect.TypeOf((*ast.FuncDecl)(nil))},
			calls:  []nodeInfo{{Name: "Check", Type: reflect.TypeOf((*ast.SelectorExpr)(nil))}},
		}, {
			method: nodeInfo{Name: "List", Type: reflect.TypeOf((*ast.FuncDecl)(nil))},
			calls:  []nodeInfo{{Name: "Check", Type: reflect.TypeOf((*ast.SelectorExpr)(nil))}},
		}, {
			method: nodeInfo{Name: "UncheckedMeth", Type: reflect.TypeOf((*ast.FuncDecl)(nil))},
			calls:  []nodeInfo{},
		}},
	}}

	for _, test := range testcases {
		service_ := Find([]ast.Node{servicePkg},
			SetFilter{Names: []string{test.service.Name}, Type: reflect.TypeOf((*ast.TypeSpec)(nil))})
		if len(service_) != 1 {
			t.Fatalf("expected to get 1 AST node back, but got %d: %v", len(service_), service_)
		}
		service := service_[0]
		serviceName, _ := getName(service)

		expMethods := make([]nodeInfo, len(test.methods))
		for i, m := range test.methods {
			expMethods[i] = m.method
		}
		actMethods := Find([]ast.Node{servicePkg}, MethodFilter{
			ReceiverType: serviceName,
			ExportedOnly: true,
		})
		checkNodesExpected(t, expMethods, actMethods)

		for _, method := range actMethods {
			methodInfo := nodeInfoFromNode(method)
			var expCalls []nodeInfo
			for _, expMethodInfo := range test.methods {
				if expMethodInfo.method == methodInfo {
					expCalls = expMethodInfo.calls
				}
			}

			calls := Find([]ast.Node{method}, RegexpFilter{Pattern: regexp.MustCompile(`.*`), Type: reflect.TypeOf((*ast.SelectorExpr)(nil))})
			checkNodesExpected(t, expCalls, calls)
		}
	}
}

//
// Helpers
//

func nodeInfoFromNode(node ast.Node) nodeInfo {
	var info nodeInfo
	if name, nameExists := getName(node); nameExists {
		info.Name = name
	}
	info.Type = reflect.TypeOf(node)
	return info
}

func checkNodesExpected(t *testing.T, exp []nodeInfo, actual []ast.Node) {
	exp_ := make(map[nodeInfo]bool)
	for _, e := range exp {
		exp_[e] = true
	}

	actual_ := make(map[nodeInfo]bool)
	for _, node := range actual {
		actual_[nodeInfoFromNode(node)] = true
	}
	if !reflect.DeepEqual(exp_, actual_) {
		t.Errorf("expected nodes %+v, but got %+v", exp_, actual_)
	}
}

func getTestPkg(t *testing.T) *ast.Package {
	pkg, err := build.Import("github.com/beyang/go-astquery/testpkg", "", build.FindOnly)
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := parser.ParseDir(token.NewFileSet(), pkg.Dir, nil, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	servicePkg, in := pkgs["service"]
	if !in {
		t.Fatal("service package not found")
	}
	return servicePkg
}
