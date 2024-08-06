package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/brunoga/deep"
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
)

type SeqExprVisitor struct {
	ast.NoopVisitor
}

func (v *SeqExprVisitor) VisitStatements(n *ast.Statements) {
	n.VisitChildrenWith(v)

	var newStmts ast.Statements

	for _, stmt := range *n {
		added := false
		if expr, ok := stmt.Stmt.(*ast.ExpressionStatement); ok {
			if seq, ok := expr.Expression.Expr.(*ast.SequenceExpression); ok {
				for _, expr := range seq.Sequence {
					newStmts = append(newStmts, ast.Statement{Stmt: &ast.ExpressionStatement{Expression: DeepCopy(&expr)}})
				}
				added = true
			}
		} else if ret, ok := stmt.Stmt.(*ast.ReturnStatement); ok && ret.Argument != nil {
			if seq, ok := ret.Argument.Expr.(*ast.SequenceExpression); ok {
				last := &seq.Sequence[len(seq.Sequence)-1]
				seq.Sequence = seq.Sequence[:len(seq.Sequence)-1]

				for _, expr := range seq.Sequence {
					newStmts = append(newStmts, ast.Statement{Stmt: &ast.ExpressionStatement{Expression: DeepCopy(&expr)}})
				}
				newStmts = append(newStmts, ast.Statement{Stmt: &ast.ReturnStatement{Argument: last}})
				added = true
			}
		} else if ifStmt, ok := stmt.Stmt.(*ast.IfStatement); ok {
			if seq, ok := ifStmt.Test.Expr.(*ast.SequenceExpression); ok {
				last := &seq.Sequence[len(seq.Sequence)-1]
				seq.Sequence = seq.Sequence[:len(seq.Sequence)-1]

				for _, expr := range seq.Sequence {
					newStmts = append(newStmts, ast.Statement{Stmt: &ast.ExpressionStatement{Expression: DeepCopy(&expr)}})
				}
				newStmts = append(newStmts, ast.Statement{Stmt: &ast.IfStatement{
					Test:       last,
					Consequent: ifStmt.Consequent,
					Alternate:  ifStmt.Alternate,
				}})
				added = true
			}
		} else if forStmt, ok := stmt.Stmt.(*ast.ForStatement); ok && forStmt.Initializer != nil {
			if forLoopInitExpr, ok := (*forStmt.Initializer).(*ast.ForLoopInitializerExpression); ok {
				if seq, ok := forLoopInitExpr.Expression.Expr.(*ast.SequenceExpression); ok {
					last := &seq.Sequence[len(seq.Sequence)-1]
					seq.Sequence = seq.Sequence[:len(seq.Sequence)-1]

					for _, expr := range seq.Sequence {
						newStmts = append(newStmts, ast.Statement{Stmt: &ast.ExpressionStatement{Expression: DeepCopy(&expr)}})
					}
					if last != nil {
						var forLoopInitiator ast.ForLoopInitializer = &ast.ForLoopInitializerExpression{Expression: last}
						newStmts = append(newStmts, ast.Statement{Stmt: &ast.ForStatement{
							Test:        forStmt.Test,
							Initializer: &forLoopInitiator,
							Update:      forStmt.Update,
							Body:        forStmt.Body,
						}})
					}

					added = true
				}
			}
		}

		if !added {
			newStmts = append(newStmts, stmt)
		}
	}
	*n = newStmts
}

func DeepCopy(t *ast.Expression) *ast.Expression {
	return deep.MustCopy(t)
}

func (n *SeqExprVisitor) VisitProgram(program *ast.Program) {
	println("[*] Replacing sequence expressions")
	program.VisitChildrenWith(n)
}

func main() {
	src, _ := os.ReadFile("../input.js")
	f, err := parser.ParseFile(string(src))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(src))
	visitor := SeqExprVisitor{ast.NoopVisitor{}}
	visitor.VisitProgram(f)
	out := generator.Generate(f)
	os.WriteFile("./out/go-fAST.js", []byte(out), fs.ModeAppend)
}
