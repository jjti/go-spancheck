package spanlint

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/cfg"
)

const doc = `check for mistakes with OTEL trace spans

Common mistakes with OTEL trace spans include forgetting to call End:

	func(ctx context.Context) {
		ctx, span := otel.Tracer("app").Start(ctx, "span")
		// defer span.End() should be here

		// do stuff
	}

And forgetting to set an Error status:

	ctx, span := otel.Tracer("app").Start(ctx, "span")
	defer span.End()

	if err := task(); err != nil {
		// span.SetStatus(codes.Error, err.Error()) should be here
		return fmt.Errorf("failed to run task: %w", err)
	}

This checker helps uncover such issues with spans.`

var Analyzer = &analysis.Analyzer{
	Name: "spanlint",
	Doc:  doc,
	Run: func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

		nodeFilter := []ast.Node{
			(*ast.FuncLit)(nil),  // f := func() {}
			(*ast.FuncDecl)(nil), // func foo() {}
		}
		inspect.Preorder(nodeFilter, func(n ast.Node) {
			runFunc(pass, n)
		})

		return nil, nil
	},
	Requires: []*analysis.Analyzer{
		ctrlflow.Analyzer,
		inspect.Analyzer,
	},
}

const (
	msgUnused = "span is unassigned, probable memory leak"
	stackLen  = 32
)

// this approach stolen from errcheck
// https://github.com/kisielk/errcheck/blob/7f94c385d0116ccc421fbb4709e4a484d98325ee/errcheck/errcheck.go#L22
var errorType = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

type spanVar struct {
	stmt ast.Node
	id   *ast.Ident
	vr   *types.Var
}

// runFunc checks if the the node is a function, has a span, and the span never has SetStatus set.
func runFunc(pass *analysis.Pass, node ast.Node) {
	// copying https://cs.opensource.google/go/x/tools/+/master:go/analysis/passes/lostcancel/lostcancel.go

	// Find scope of function node
	var funcScope *types.Scope
	switch v := node.(type) {
	case *ast.FuncLit:
		funcScope = pass.TypesInfo.Scopes[v.Type]
	case *ast.FuncDecl:
		funcScope = pass.TypesInfo.Scopes[v.Type]
	}

	// Maps each span variable to its defining ValueSpec/AssignStmt.
	spanVars := make(map[*ast.Ident]spanVar)

	// Find the set of span vars to analyze.
	stack := make([]ast.Node, 0, stackLen)
	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncLit:
			if len(stack) > 0 {
				return false // don't stray into nested functions
			}
		case nil:
			stack = stack[:len(stack)-1] // pop
			return true
		}
		stack = append(stack, n) // push

		// Look for [{AssignStmt,ValueSpec} CallExpr SelectorExpr]:
		//
		//   ctx, span     := otel.Tracer("app").Start(...)
		//   ctx, span     = otel.Tracer("app").Start(...)
		//   var ctx, span = otel.Tracer("app").Start(...)
		if !isTracerStart(pass.TypesInfo, n) || !isCall(stack[len(stack)-2]) {
			return true
		}

		stmt := stack[len(stack)-3]
		id := getID(stmt)
		if id == nil {
			pass.ReportRangef(n, msgUnused)
			return true
		}

		if id.Name == "_" {
			pass.ReportRangef(id, msgUnused)
		} else if v, ok := pass.TypesInfo.Uses[id].(*types.Var); ok {
			// If the span variable is defined outside function scope,
			// do not analyze it.
			if funcScope.Contains(v.Pos()) {
				spanVars[id] = spanVar{
					vr:   v,
					stmt: stmt,
					id:   id,
				}
			}
		} else if v, ok := pass.TypesInfo.Defs[id].(*types.Var); ok {
			spanVars[id] = spanVar{
				vr:   v,
				stmt: stmt,
				id:   id,
			}
		}

		return true
	})

	if len(spanVars) == 0 {
		return // no need to inspect CFG
	}

	// Obtain the CFG.
	cfgs := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)
	var g *cfg.CFG
	var sig *types.Signature
	switch node := node.(type) {
	case *ast.FuncDecl:
		sig, _ = pass.TypesInfo.Defs[node.Name].Type().(*types.Signature)
		g = cfgs.FuncDecl(node)
	case *ast.FuncLit:
		sig, _ = pass.TypesInfo.Types[node.Type].Type.(*types.Signature)
		g = cfgs.FuncLit(node)
	}
	if sig == nil {
		return // missing type information
	}

	// Check for missing Ends().
	for _, sv := range spanVars {
		// Check if there's no End to the span.
		if ret := missingSpanCalls(pass, g, sv, "End", func(pass *analysis.Pass, ret *ast.ReturnStmt) *ast.ReturnStmt { return ret }); ret != nil {
			pass.ReportRangef(sv.stmt, "%s.End is not called on all paths, possible memory leak", sv.vr.Name())
			pass.ReportRangef(ret, "this return statement may be reached without calling %s.End", sv.vr.Name())
		}

		// Check if there's no SetStatus to the span setting an error.
		if ret := missingSpanCalls(pass, g, sv, "SetStatus", returnsErr); ret != nil {
			pass.ReportRangef(sv.stmt, "%s.SetStatus is not called on all paths", sv.vr.Name())
			pass.ReportRangef(ret, "this return statement may be reached without calling %s.SetStatus", sv.vr.Name())
		}
	}
}

// isTracerStart reports whether n is tracer.Start()
func isTracerStart(info *types.Info, n ast.Node) bool {
	sel, ok := n.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "Start" {
		return false
	}

	obj, ok := info.Uses[sel.Sel]
	if ok {
		return obj.Pkg().Path() == "go.opentelemetry.io/otel/trace"
	}

	return false
}

func isCall(n ast.Node) bool {
	_, ok := n.(*ast.CallExpr)
	return ok
}

func getID(node ast.Node) *ast.Ident {
	switch stmt := node.(type) {
	case *ast.ValueSpec:
		if len(stmt.Names) > 1 {
			return stmt.Names[1]
		}
	case *ast.AssignStmt:
		if len(stmt.Lhs) > 1 {
			id, _ := stmt.Lhs[1].(*ast.Ident)
			return id
		}
	}
	return nil
}

// missingSpanCalls finds a path through the CFG, from stmt (which defines
// the 'span' variable v) to a return statement, that doesn't call End() on the span.
//
// Additionally, if the CFG returns an error without calling SetStatus, an error is returned.
func missingSpanCalls(
	pass *analysis.Pass,
	g *cfg.CFG,
	sv spanVar,
	selName string,
	checkErr func(pass *analysis.Pass, ret *ast.ReturnStmt) *ast.ReturnStmt,
) *ast.ReturnStmt {
	// usesCall reports whether stmts contain a use of the selName call on variable v.
	usesCall := func(pass *analysis.Pass, stmts []ast.Node) bool {
		found, reAssigned := false, false
		for _, subStmt := range stmts {
			stack := []ast.Node{}
			ast.Inspect(subStmt, func(n ast.Node) bool {
				switch n.(type) {
				case *ast.FuncLit:
					if len(stack) > 0 {
						// return false // don't stray into nested functions
						return true // don't stray into nested functions
					}
				case nil:
					stack = stack[:len(stack)-1] // pop
					return true
				}
				stack = append(stack, n) // push

				if isTracerStart(pass.TypesInfo, n) {
					if id := getID(stack[len(stack)-3]); id != nil && id.Obj.Decl == sv.id.Obj.Decl {
						reAssigned = true
						return false
					}
				}

				if n, ok := n.(*ast.SelectorExpr); ok && n.Sel.Name == selName {
					id, ok := n.X.(*ast.Ident)
					found = ok && id.Obj.Decl == sv.id.Obj.Decl
				}

				return !found
			})
		}
		return found && !reAssigned
	}

	// blockUses computes "uses" for each block, caching the result.
	memo := make(map[*cfg.Block]bool)
	blockUses := func(pass *analysis.Pass, b *cfg.Block) bool {
		res, ok := memo[b]
		if !ok {
			res = usesCall(pass, b.Nodes)
			memo[b] = res
		}
		return res
	}

	// Find the var's defining block in the CFG,
	// plus the rest of the statements of that block.
	var defBlock *cfg.Block
	var rest []ast.Node
outer:
	for _, b := range g.Blocks {
		for i, n := range b.Nodes {
			if n == sv.stmt {
				defBlock = b
				rest = b.Nodes[i+1:]
				break outer
			}
		}
	}
	if defBlock == nil {
		panic("internal error: can't find defining block for span var")
	}

	// Is the call "used" in the remainder of its defining block?
	if usesCall(pass, rest) {
		return nil
	}

	// Does the defining block return without using v.End()?
	if ret := defBlock.Return(); ret != nil {
		return checkErr(pass, ret)
	}

	// Search the CFG depth-first for a path, from defblock to a
	// return block, in which v is never "used".
	seen := make(map[*cfg.Block]bool)
	var search func(blocks []*cfg.Block) *ast.ReturnStmt
	search = func(blocks []*cfg.Block) *ast.ReturnStmt {
		for _, b := range blocks {
			if seen[b] {
				continue
			}
			seen[b] = true

			// Prune the search if the block uses v.
			if blockUses(pass, b) {
				continue
			}

			// Found path to return statement?
			if ret := returnsErr(pass, b.Return()); ret != nil {
				return ret // found
			}

			// Recur
			if ret := returnsErr(pass, search(b.Succs)); ret != nil {
				return ret
			}
		}
		return nil
	}

	return search(defBlock.Succs)
}

func returnsErr(pass *analysis.Pass, ret *ast.ReturnStmt) *ast.ReturnStmt {
	if ret == nil {
		return nil
	}

	for _, r := range ret.Results {
		if isErrorType(pass.TypesInfo.TypeOf(r)) {
			return ret
		}

		if r, ok := r.(*ast.CallExpr); ok {
			for _, err := range errorsByArg(pass, r) {
				if err {
					return ret
				}
			}
		}
	}

	return nil
}

// errorsByArg returns a slice s such that
// len(s) == number of return types of call
// s[i] == true iff return type at position i from left is an error type
//
// copied from https://github.com/kisielk/errcheck/blob/master/errcheck/errcheck.go
func errorsByArg(pass *analysis.Pass, call *ast.CallExpr) []bool {
	switch t := pass.TypesInfo.Types[call].Type.(type) {
	case *types.Named:
		// Single return
		return []bool{isErrorType(t)}
	case *types.Pointer:
		// Single return via pointer
		return []bool{isErrorType(t)}
	case *types.Tuple:
		// Multiple returns
		s := make([]bool, t.Len())
		for i := 0; i < t.Len(); i++ {
			switch et := t.At(i).Type().(type) {
			case *types.Named:
				// Single return
				s[i] = isErrorType(et)
			case *types.Pointer:
				// Single return via pointer
				s[i] = isErrorType(et)
			default:
				s[i] = false
			}
		}
		return s
	}
	return []bool{false}
}

func isErrorType(t types.Type) bool {
	return types.Implements(t, errorType)
}
