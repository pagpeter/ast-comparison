package main

import (
	"fmt"
	"os"
	"time"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/generator"
	"github.com/t14raptor/go-fast/parser"
)

type SeqExprVisitor struct {
	ast.NoopVisitor
}

func (v *SeqExprVisitor) VisitStatements(n *ast.Statements) {
	n.VisitChildrenWith(v)

	newStmts := make(ast.Statements, 0, len(*n))

	for _, stmt := range *n {
		added := false
		if expr, ok := stmt.Stmt.(*ast.ExpressionStatement); ok {
			if seq, ok := expr.Expression.Expr.(*ast.SequenceExpression); ok {
				for i := range seq.Sequence {
					newStmts = append(newStmts, ast.Statement{Stmt: &ast.ExpressionStatement{Expression: &seq.Sequence[i]}})
				}
				added = true
			}
		} else if ret, ok := stmt.Stmt.(*ast.ReturnStatement); ok && ret.Argument != nil {
			if seq, ok := ret.Argument.Expr.(*ast.SequenceExpression); ok {
				last := &seq.Sequence[len(seq.Sequence)-1]
				seq.Sequence = seq.Sequence[:len(seq.Sequence)-1]

				for i := range seq.Sequence {
					newStmts = append(newStmts, ast.Statement{Stmt: &ast.ExpressionStatement{Expression: &seq.Sequence[i]}})
				}
				newStmts = append(newStmts, ast.Statement{Stmt: &ast.ReturnStatement{Argument: last}})
				added = true
			}
		} else if ifStmt, ok := stmt.Stmt.(*ast.IfStatement); ok {
			if seq, ok := ifStmt.Test.Expr.(*ast.SequenceExpression); ok {
				last := &seq.Sequence[len(seq.Sequence)-1]
				seq.Sequence = seq.Sequence[:len(seq.Sequence)-1]

				for i := range seq.Sequence {
					newStmts = append(newStmts, ast.Statement{Stmt: &ast.ExpressionStatement{Expression: &seq.Sequence[i]}})
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

					for i := range seq.Sequence {
						newStmts = append(newStmts, ast.Statement{Stmt: &ast.ExpressionStatement{Expression: &seq.Sequence[i]}})
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

func (n *SeqExprVisitor) VisitProgram(program *ast.Program) {
	program.VisitChildrenWith(n)
}

func main() {
	src, _ := os.ReadFile("../input.js")
	parseStart := time.Now()
	f, err := parser.ParseFile(string(src))
	if err != nil {
		panic(err)
	}
	parseEnd := time.Now()
	visitor := SeqExprVisitor{ast.NoopVisitor{}}
	visitor.V = &visitor
	traversalStart := time.Now()
	visitor.VisitProgram(f)
	traversalEnd := time.Now()
	genStart := time.Now()
	out := generator.Generate(f)
	genEnd := time.Now()
	os.WriteFile("../output/go-fAST.js", []byte(out), 0644)
	fmt.Printf("Total: %v\nparsing: %v\ntraversal: %v\ngenerating: %v\n", genEnd.Sub(parseStart), parseEnd.Sub(parseStart), traversalEnd.Sub(traversalStart), genEnd.Sub(genStart))
}
