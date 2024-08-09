use std::fs;
use std::time::Instant;
use swc_common::input::StringInput;
use swc_common::util::take::Take;
use swc_common::{SourceMap, Span};
use swc_core::common::sync::Lrc;
use swc_core::common::FileName;
use swc_core::ecma::codegen::text_writer::JsWriter;
use swc_core::ecma::codegen::Emitter;
use swc_core::ecma::utils::ExprFactory;
use swc_core::ecma::visit::{noop_visit_mut_type, VisitMut, VisitMutWith};
use swc_ecma_ast::{ExprStmt, ForStmt, IfStmt, ReturnStmt, Stmt};
use swc_ecma_parser::{EsSyntax, Parser, Syntax};

pub struct Visitor;

impl VisitMut for Visitor {
    noop_visit_mut_type!();
    fn visit_mut_stmts(&mut self, path: &mut Vec<Stmt>) {
        path.visit_mut_children_with(self);

        // Reserve capacity to avoid reallocations
        let mut new_stmts: Vec<Stmt> = Vec::with_capacity(path.len());

        for stmt in path {
            match stmt {
                Stmt::Expr(ExprStmt { expr, .. }) if expr.is_seq() => {
                    let seq = expr.as_seq().unwrap();

                    new_stmts.reserve(seq.exprs.len());

                    for expr in &seq.exprs {
                        new_stmts.push(expr.clone().into_stmt());
                    }
                }
                Stmt::Return(ReturnStmt { arg, .. })
                    if arg.is_some() && arg.as_ref().unwrap().is_seq() =>
                {
                    let seq = arg.as_mut().unwrap().as_mut_seq().unwrap();
                    let last = seq.exprs.pop().unwrap();

                    new_stmts.reserve(seq.exprs.len() + 1);

                    for expr in &seq.exprs {
                        new_stmts.push(expr.clone().into_stmt());
                    }

                    new_stmts.push(Stmt::Return(ReturnStmt {
                        span: Span::dummy(),
                        arg: Some(last),
                    }));
                }
                Stmt::If(IfStmt {
                    ref mut test,
                    cons,
                    alt,
                    ..
                }) if test.is_seq() => {
                    let seq = test.as_mut_seq().unwrap();
                    let last = seq.exprs.pop().unwrap();

                    new_stmts.reserve(seq.exprs.len() + 1);

                    for expr in &seq.exprs {
                        new_stmts.push(expr.clone().into_stmt());
                    }

                    new_stmts.push(Stmt::If(IfStmt {
                        span: Span::dummy(),
                        test: last,
                        cons: cons.clone(),
                        alt: alt.clone(),
                    }));
                }
                Stmt::For(ForStmt {
                    init,
                    test,
                    update,
                    body,
                    ..
                }) if init.is_some()
                    && init.as_ref().unwrap().is_expr()
                    && init.as_ref().unwrap().as_expr().unwrap().is_seq() =>
                {
                    let seq = init
                        .as_mut()
                        .unwrap()
                        .as_mut_expr()
                        .unwrap()
                        .as_mut_seq()
                        .unwrap();
                    let last = seq.exprs.pop().unwrap();

                    new_stmts.reserve(seq.exprs.len() + 1);

                    for expr in &seq.exprs {
                        new_stmts.push(expr.clone().into_stmt());
                    }

                    new_stmts.push(Stmt::For(ForStmt {
                        span: Span::dummy(),
                        init: Some(swc_ecma_ast::VarDeclOrExpr::Expr(last)),
                        test: test.clone(),
                        update: update.clone(),
                        body: body.clone(),
                    }));
                }
                _ => new_stmts.push(stmt.clone()),
            }
        }
    }
}

fn main() {
    let cm: Lrc<SourceMap> = std::default::Default::default();
    let fm = cm.new_source_file(
        FileName::Custom("input.js".into()).into(),
        fs::read_to_string("../input.js").unwrap(),
    );

    let start_parse = Instant::now();
    let mut parser = Parser::new(
        Syntax::Es(EsSyntax::default()),
        StringInput::from(&*fm),
        None,
    );
    let script = &mut parser.parse_script().expect("");
    let end_parse = Instant::now();

    let start_visit = Instant::now();
    script.visit_mut_with(&mut Visitor {});
    let end_visit = Instant::now();

    let start_gen = Instant::now();
    let mut buf = Vec::new();
    let mut emitter = Emitter {
        cfg: Default::default(),
        cm: cm.clone(),
        comments: None,
        wr: JsWriter::new(cm, "\n", &mut buf, None),
    };
    emitter.emit_script(script).unwrap();
    let code = String::from_utf8_lossy(&buf).to_string();
    let end_gen = Instant::now();

    println!(
        "Parsing: {:?}\nTraversal: {:?}\nGenerating: {:?}\nTotal: {:?}",
        end_parse - start_parse,
        end_visit - start_visit,
        end_gen - start_gen,
        end_gen - start_parse,
    );
    let _ = fs::write("../output/rust-swc.js", code);
}
