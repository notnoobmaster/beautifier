package beautifier

import (
	"github.com/yuin/gopher-lua/ast"
	"strings"
)

type someStruct struct {
	str *strings.Builder
	funcs *functions
	tabs int
}

func (s *someStruct) add(strs ...string) {
	for _, str := range strs {
		s.str.WriteString(str)
	}
}

func (s *someStruct) addln(strs ...string) {
	for _, str := range strs {
		s.str.WriteString(str)
	}
	s.str.WriteString("\n")
}


func (s *someStruct) tab() *someStruct {
	s.str.WriteString(strings.Repeat("\t", s.tabs))
	return s
}

func isValid(str string) bool {
	for pos, ch := range str {
        if ch == '_' || 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || ('0' <= ch && ch <= '9') && pos > 0 {
            return true
        }
	}
	return false
}

func (s *someStruct) compileAttrGetExpr(expr *ast.AttrGetExpr) {
	if str, ok := expr.Object.(*ast.StringExpr); ok && str.Value == "" {
		s.add("string")
	} else {
		s.exprToString(&expr.Object)
	}
	if str, ok := expr.Key.(*ast.StringExpr); ok && isValid(str.Value) {
		s.add(".", str.Value)
	} else {
		s.add("[")
		s.exprToString(&expr.Key)
		s.add("]")
	}
}

func (s *someStruct) compileTableExpr(expr *ast.TableExpr) {
	s.add("{")
	for _, field := range expr.Fields {
		if field.Key != nil {
			s.add("[")
			s.exprToString(&field.Key)
			s.add("] = ")
		}
		s.exprToString(&field.Value)
		s.add(",")
	}
	s.add("}")
}

func (s *someStruct) compileUnaryOpExpr(expr *ast.Expr) {
	switch ex := (*expr).(type) {
	case *ast.UnaryMinusOpExpr:
		s.add("(-")
		s.exprToString(&ex.Expr)
	case *ast.UnaryNotOpExpr:
		s.add("not (")
		s.exprToString(&ex.Expr)
	case *ast.UnaryLenOpExpr:
		s.add("#(")
		s.exprToString(&ex.Expr)
	}
	s.add(")")
}

func (s *someStruct) compileRelationalOpExpr(expr *ast.RelationalOpExpr) {
	s.add("(")
	s.exprToString(&expr.Lhs)
	s.add(" ",expr.Operator, " ")
	s.exprToString(&expr.Rhs)
	s.add(")")
}

func (s *someStruct) compileArithmeticOpExpr(expr *ast.ArithmeticOpExpr) {
	s.add("(")
	s.exprToString(&expr.Lhs)
	s.add(" ", expr.Operator, " ")
	s.exprToString(&expr.Rhs)
	s.add(")")
}

func (s *someStruct) compileStringConcatOpExpr(expr *ast.StringConcatOpExpr) {
	s.exprToString(&expr.Lhs)
	s.add(" .. ")
	s.exprToString(&expr.Rhs)
}

func (s *someStruct) compileLogicalOpExpr(expr *ast.LogicalOpExpr) {
	s.add("(")
	s.exprToString(&expr.Lhs)
	s.add(" ", expr.Operator, " ")
	s.exprToString(&expr.Rhs)
	s.add(")")
}

func (s *someStruct) compileFuncCallExpr(expr *ast.FuncCallExpr) {
	if expr.Func != nil { // hoge.func()
		switch expr.Func.(type) {
		case *ast.IdentExpr:
			s.exprToString(&expr.Func)
		case *ast.TableExpr:
			s.exprToString(&expr.Func)
		case *ast.AttrGetExpr:
			s.exprToString(&expr.Func)
		default:
			s.add("(")
			s.exprToString(&expr.Func)
			s.add(")")
		}
	} else { // hoge:method()
		s.exprToString(&expr.Receiver)
		s.add(":", string(expr.Method))
	}

	s.add("(")
	for i := range expr.Args {
		s.exprToString(&expr.Args[i])
		if i < len(expr.Args)-1 {
			s.add(", ")
		}
	}
	s.add(")")
}

func (s *someStruct) compileFunctionExpr(expr *ast.FunctionExpr) {
	s.add("function(")
	for i, name := range expr.ParList.Names {
		s.add(name)
		s.addComma(i, len(expr.ParList.Names))
	}
	if expr.ParList.HasVargs {
		if len(expr.ParList.Names) > 0 {
			s.add(", ")
		}
		s.add("...")
	}
	s.addln(")")
	s.tabs++
	s.funcs.Traverse(expr.Stmts)
	s.tabs--
	s.tab().add("end")
}

func (s *someStruct) exprToString(expr *ast.Expr) {
	switch ex := (*expr).(type) {
	case *ast.StringExpr:
		s.add("\"",ex.Value,"\"")
	case *ast.NumberExpr:
		s.add(ex.Value)
	case *constLValueExpr:
		s.add("PANIC")
	case *ast.NilExpr:
		s.add("nil")
	case *ast.FalseExpr:
		s.add("false")
	case *ast.TrueExpr:
		s.add("true")
	case *ast.IdentExpr:
		s.add(ex.Value)
	case *ast.Comma3Expr:
		s.add("...")
	case *ast.AttrGetExpr:
		s.compileAttrGetExpr(ex)
	case *ast.TableExpr:
		s.compileTableExpr(ex)
	case *ast.ArithmeticOpExpr:
		s.compileArithmeticOpExpr(ex)
	case *ast.StringConcatOpExpr:
		s.compileStringConcatOpExpr(ex)
	case *ast.UnaryMinusOpExpr, *ast.UnaryNotOpExpr, *ast.UnaryLenOpExpr:
		s.compileUnaryOpExpr(expr)
	case *ast.RelationalOpExpr:
		s.compileRelationalOpExpr(ex)
	case *ast.LogicalOpExpr:
		s.compileLogicalOpExpr(ex)
	case *ast.FuncCallExpr:
		s.compileFuncCallExpr(ex)
	case *ast.FunctionExpr:
		s.compileFunctionExpr(ex)
	}
}

func (s *someStruct) whileStmt(stmt *ast.WhileStmt) {
	s.tab().add("while ")
	s.exprToString(&stmt.Condition)
	s.addln(" do")
	s.tabs++
	s.funcs.Traverse(stmt.Stmts)
	s.tabs--
	s.tab().addln("end")
}

func (s *someStruct) repeatStmt(stmt *ast.RepeatStmt) {
	s.tab().addln("repeat")
	s.tabs++
	s.funcs.Traverse(stmt.Stmts)
	s.tabs--
	s.tab().add("until ")
	s.exprToString(&stmt.Condition)
	s.addln()
}

func (s *someStruct) doBlockStmt(stmt *ast.DoBlockStmt) {
	s.tab().addln("do")
	s.tabs++
	s.funcs.Traverse(stmt.Stmts)
	s.tabs--
	s.tab().addln("end")
}

func (s *someStruct) addComma(idx int, length int) {
	if idx < length-1 {
		s.add(", ")
	}
}

func (s *someStruct) assignStmt(stmt *ast.AssignStmt) {
	s.tab()
	for i, ex := range stmt.Lhs {
		s.exprToString(&ex)
		s.addComma(i, len(stmt.Lhs))
	}
	s.add(" = ")
	for i, ex := range stmt.Rhs {
		s.exprToString(&ex)
		s.addComma(i, len(stmt.Rhs))
	}
	s.addln()
}

func (s *someStruct) localAssignStmt(stmt *ast.LocalAssignStmt) {
	s.tab().add("local ")
	for i, name := range stmt.Names {
		s.add(name)
		s.addComma(i, len(stmt.Names))
	}
	if len(stmt.Exprs) > 0 {
		s.add(" = ")
		for i, ex := range stmt.Exprs {
			s.exprToString(&ex)
			s.addComma(i, len(stmt.Exprs))
		}
	}
	s.addln()
}

func (s *someStruct) returnStmt(stmt *ast.ReturnStmt) {
	s.tab().add("return ")
	for i, ex := range stmt.Exprs {
		s.exprToString(&ex)
		s.addComma(i, len(stmt.Exprs))
	}
	s.addln()
}

func (s *someStruct) funcDefStmt(stmt *ast.FuncDefStmt) {
	
}


func (s *someStruct) compileBranchCondition(expr ast.Expr) {
	switch ex := expr.(type) {
	case *ast.UnaryNotOpExpr:
		s.add("not (")
		s.compileBranchCondition(ex.Expr)
		s.add(")")
	case *ast.LogicalOpExpr:
		s.add("(")
		s.compileBranchCondition(ex.Lhs)
		s.add(" ", ex.Operator, " ")
		s.compileBranchCondition(ex.Rhs)
		s.add(")")
	case *ast.RelationalOpExpr:
		s.add("(")
		s.exprToString(&ex.Lhs)
		s.add(" ", ex.Operator, " ")
		s.exprToString(&ex.Rhs)
		s.add(")")
	default:
		s.exprToString(&expr)
	}
}

func (s *someStruct) elseBody(elseStmt []ast.Stmt) {
	if len(elseStmt) > 0 {
		if elseif, ok := elseStmt[0].(*ast.IfStmt); ok && len(elseStmt) == 1 {
			s.tab().add("elseif ")
			s.compileBranchCondition(elseif.Condition)
			s.addln(" then")
			s.tabs++
			s.funcs.Traverse(elseif.Then)
			s.tabs--
			s.elseBody(elseif.Else)
		} else {
			s.tab().addln("else")
			s.tabs++
			s.funcs.Traverse(elseStmt)
			s.tabs--
		}
	}
}

func (s *someStruct) ifStmt(stmt *ast.IfStmt) {
	s.tab().add("if ")
	s.compileBranchCondition(stmt.Condition)
	s.addln(" then")
	s.tabs++
	s.funcs.Traverse(stmt.Then)
	s.tabs--
	s.elseBody(stmt.Else)
	s.tab().addln("end")
}

func (s *someStruct) breakStmt() {
	s.tab().addln("break")
}

func (s *someStruct) numberForStmt(stmt *ast.NumberForStmt) {
	s.tab().add("for ", stmt.Name, " = ")
	s.exprToString(&stmt.Init)
	s.add(", ")
	s.exprToString(&stmt.Limit)
	if stmt.Step != nil {
		s.add(", ")
		s.exprToString(&stmt.Step)
	}
	s.addln(" do")
	s.tabs++
	s.funcs.Traverse(stmt.Stmts)
	s.tabs--
	s.tab().addln("end")
}

func (s *someStruct) genericForStmt(stmt *ast.GenericForStmt) {
	s.tab().add("for ")
	for i, name := range stmt.Names {
		s.add(name)
		s.addComma(i, len(stmt.Names))
	}
	s.add(" in ")
	for _, ex := range stmt.Exprs {
		s.exprToString(&ex)
	}
	s.addln(" do")
	s.tabs++
	s.funcs.Traverse(stmt.Stmts)
	s.tabs--
	s.tab().addln("end")
}

func (s *someStruct) funcCallStmt(expr *ast.FuncCallExpr) {
	s.tab()
	s.compileFuncCallExpr(expr)
	s.addln()
}

// Beautify the Abstract Syntax Tree
func Beautify(ast []ast.Stmt) string {
	s := &someStruct{str: &strings.Builder{}}
	f := &functions{
		WhileStmt: s.whileStmt,
		RepeatStmt: s.repeatStmt,
		DoBlockStmt: s.doBlockStmt,
		LocalAssignStmt: s.localAssignStmt,
		FuncDefStmt: s.funcDefStmt,
		AssignStmt: s.assignStmt,
		ReturnStmt: s.returnStmt,
		IfStmt: s.ifStmt,
		NumberForStmt: s.numberForStmt,
		GenericForStmt: s.genericForStmt,
		BreakStmt: s.breakStmt,
		FuncCallStmt: s.funcCallStmt,
	}
	s.funcs = f
	f.Traverse(ast)
	return s.str.String()
}