const parser = require("@babel/parser");
const generate = require("@babel/generator").default;
const traverse = require("@babel/traverse").default;
const t = require("@babel/types");
const fs = require("node:fs");

const file = fs.readFileSync("./input.js", "utf-8");

// Taken from https://github.com/wwhtrbbtt/deob-transformations/blob/3b6baac12f992cba5e816e41cf3ed08e655c07ea/transformations/general.js#L266
const traversal = {
  ReturnStatement(path) {
    const { node } = path;

    if (t.isSequenceExpression(node.argument)) {
      let expressionArr = [];
      const { expressions } = node.argument;
      expressions.forEach((node, indx) => {
        if (indx !== expressions.length - 1) {
          if (
            node.type === "CallExpression" &&
            node.callee.type === "FunctionExpression"
          )
            expressionArr.push(
              t.emptyStatement(),
              t.unaryExpression("!", node),
              t.emptyStatement()
            );
          else expressionArr.push(node);
        } else expressionArr.push(t.returnStatement(node));
      });
      path.replaceWithMultiple(expressionArr);
    }
  },

  IfStatement(path) {
    const { node } = path;
    const { test } = node;
    if (!t.isSequenceExpression(test)) return;
    const { expressions } = test;

    expressions.forEach((expression, index) => {
      if (index !== expressions.length - 1) {
        path.insertBefore(t.expressionStatement(expression));
      } else {
        node.test = expression;
      }
    });
  },

  AssignmentExpression(path) {
    const { node } = path;
    if (!t.isSequenceExpression(node.right)) return;
    const { expressions } = node.right;
    expressions.forEach((expression, index) => {
      if (index !== expressions.length - 1) {
        path.insertBefore(t.expressionStatement(expression));
      } else {
        path.replaceWith(t.AssignmentExpression("=", node.left, expression));
      }
    });
  },

  FunctionDeclaration(path) {
    if (!t.isBlockStatement(path.node.body)) return;

    const { body } = path.node.body;
    if (body.length !== 1) return;

    const [statement] = body;
    if (!t.isExpressionStatement(statement)) return;
    const { expression } = statement;
    if (!t.isSequenceExpression(expression)) return;

    newExpressions = [];
    expression.expressions.forEach((e) => {
      if (t.isAssignmentExpression(e)) {
        newExpressions.push(t.expressionStatement(e));
      } else {
        newExpressions.push(e);
      }
    });

    path.node.body.body = newExpressions;
  },

  ExpressionStatement(path) {
    if (!t.isSequenceExpression(path.node.expression)) return;
    expressions = path.node.expression.expressions;
    expressionsAndStatements = [];
    expressions.forEach((ex) =>
      expressionsAndStatements.push(ex, t.emptyStatement())
    );
    path.replaceWithMultiple(expressionsAndStatements);
  },

  ForStatement(path) {
    if (!t.isSequenceExpression(path.node.init)) return;
    if (path.node.init.expressions.length < 1) return;
    const expressions = path.node.init.expressions;

    expressions.forEach((e, i) => {
      if (i === expressions.length - 1) return;
      path.insertBefore(t.expressionStatement(e));
    });
    if (!path.node.init) return;
    if (!path.node.init.expressions) return;
    path.node.init.expressions = [expressions[expressions.length - 1]];
  },
};

console.time("Total");
console.time("parsing");
const ast = parser.parse(file);
console.timeEnd("parsing");
console.time("traversal");
traverse(ast, traversal);
console.timeEnd("traversal");
console.time("generating");
const out = generate(ast).code;
console.timeEnd("generating");
console.timeEnd("Total");

fs.writeFileSync("./output/js-babel.js", out);
