import { CstNode, IToken } from "chevrotain";
import {
  AssignStatement,
  BinaryExpression,
  BlockStatement,
  BranchStatement,
  CallExpression,
  ConstDeclStatement,
  DeclarationSpec,
  DeferStatement,
  Expression,
  ExpressionStatement,
  ForStatement,
  FunctionDecl,
  FunctionLiteralExpression,
  IdentifierExpression,
  IfStatement,
  IncDecStatement,
  ImportDecl,
  IndexExpression,
  LabeledStatement,
  LiteralExpression,
  MapEntryExpression,
  MapLiteralExpression,
  ParameterDecl,
  ProgramAst,
  ReceiverDecl,
  ReturnStatement,
  SelectorExpression,
  ShortVarStatement,
  Signature,
  SliceExpression,
  SpreadsheetRangeExpression,
  Statement,
  SwitchClause,
  SwitchStatement,
  TypeDeclStatement,
  TypeNode,
  TypeSpec,
  UnaryExpression,
  VarDeclStatement
} from "./ast.js";
import { Diagnostic, SourceSpan } from "./diagnostics.js";

type BinaryOperator = BinaryExpression["operator"];
type UnaryOperator = UnaryExpression["operator"];

export function cstToAst(cst: CstNode, diagnostics: Diagnostic[] = []): ProgramAst {
  const functions = childNodes(cst, "functionDecl").map(functionDeclToAst);
  const body = childNodes(cst, "statement").map(statementToAst);

  return {
    kind: functions.length > 0 ? "function" : "script",
    imports: collectImports(cst),
    diagnostics,
    body,
    functions
  };
}

function collectImports(program: CstNode): ImportDecl[] {
  const importDecls = childNodes(program, "importDecl");
  return importDecls.flatMap((decl) =>
    childNodes(decl, "importSpec").map(importSpecToAst)
  );
}

function importSpecToAst(spec: CstNode): ImportDecl {
  const pathToken = firstChildToken(spec, "StringLiteral");
  const aliasNode = firstChildNode(spec, "name");
  const aliasToken = aliasNode ? nameToken(aliasNode) : undefined;

  return {
    path: pathToken ? unquote(pathToken.image) : "",
    ...(aliasToken ? { alias: aliasToken.image } : {})
  };
}

function functionDeclToAst(node: CstNode): FunctionDecl {
  const name = firstChildToken(node, "Identifier")?.image ?? "<anonymous>";
  const receiver = firstChildNode(node, "receiver");
  return withSpan(
    {
      kind: "FunctionDecl",
      name,
      ...(receiver ? { receiver: receiverToAst(receiver) } : {}),
      signature: signatureToAst(requiredChildNode(node, "signature")),
      body: blockToAst(requiredChildNode(node, "block"))
    } satisfies FunctionDecl,
    node
  );
}

function receiverToAst(node: CstNode): ReceiverDecl {
  const maybeName = firstChildNode(node, "name");
  const typeNode = requiredChildNode(node, "typeExpression");
  return {
    ...(maybeName ? { name: nameText(maybeName) } : {}),
    type: typeToAst(typeNode)
  };
}

function signatureToAst(node: CstNode): Signature {
  const params = firstChildNode(node, "parameterList");
  const result = firstChildNode(node, "result");
  return {
    parameters: params ? parameterListToAst(params) : [],
    results: result ? resultToAst(result) : []
  };
}

function parameterListToAst(node: CstNode): ParameterDecl[] {
  return childNodes(node, "parameter").flatMap(parameterToAst);
}

function parameterToAst(node: CstNode): ParameterDecl[] {
  const names = childNodes(node, "name");
  const typeNode = firstChildNode(node, "typeExpression");
  const base = {
    type: typeNode ? typeToAst(typeNode) : { text: "<missing>" },
    variadic: childTokens(node, "Ellipsis").length > 0
  };
  if (names.length === 0) return [base];
  return names.map((name) => ({
    name: nameText(name),
    ...base
  }));
}

function resultToAst(node: CstNode): ParameterDecl[] {
  const typeNode = firstChildNode(node, "typeExpression");
  if (typeNode) {
    return [{ type: typeToAst(typeNode), variadic: false }];
  }
  const params = firstChildNode(node, "parameterList");
  return params ? parameterListToAst(params) : [];
}

function typeToAst(node: CstNode): TypeNode {
  const text = allTokensIn(node)
    .sort(byOffset)
    .map((token) => token.image)
    .join("");
  const span = spanFromNode(node);
  return {
    text,
    ...(span ? { span } : {})
  };
}

function blockToAst(node: CstNode): BlockStatement {
  return withSpan(
    {
      kind: "BlockStatement",
      statements: childNodes(node, "statement").map(statementToAst)
    } satisfies BlockStatement,
    node
  );
}

function statementToAst(node: CstNode): Statement {
  const labeledStmt = firstChildNode(node, "labeledStmt");
  if (labeledStmt) return labeledStmtToAst(labeledStmt);

  const constDecl = firstChildNode(node, "constDecl");
  if (constDecl) return constDeclToAst(constDecl);

  const varDecl = firstChildNode(node, "varDecl");
  if (varDecl) return varDeclToAst(varDecl);

  const typeDecl = firstChildNode(node, "typeDecl");
  if (typeDecl) return typeDeclToAst(typeDecl);

  const returnStmt = firstChildNode(node, "returnStmt");
  if (returnStmt) return returnStmtToAst(returnStmt);

  const ifStmt = firstChildNode(node, "ifStmt");
  if (ifStmt) return ifStmtToAst(ifStmt);

  const switchStmt = firstChildNode(node, "switchStmt");
  if (switchStmt) return switchStmtToAst(switchStmt);

  const forStmt = firstChildNode(node, "forStmt");
  if (forStmt) return forStmtToAst(forStmt);

  const deferStmt = firstChildNode(node, "deferStmt");
  if (deferStmt) return deferStmtToAst(deferStmt);

  const branch = firstChildNode(node, "branchStmt");
  if (branch) return branchToAst(branch);

  const simpleStmt = firstChildNode(node, "simpleStmt");
  if (simpleStmt) return simpleStmtToAst(simpleStmt);

  return withSpan(
    {
      kind: "ExpressionStatement",
      expression: missingExpression()
    } satisfies ExpressionStatement,
    node
  );
}

function labeledStmtToAst(node: CstNode): LabeledStatement {
  const label = firstChildToken(node, "Identifier")?.image ?? "<missing>";
  const statement = firstChildNode(node, "statement");
  return withSpan(
    {
      kind: "LabeledStatement",
      label,
      ...(statement ? { statement: statementToAst(statement) } : {})
    } satisfies LabeledStatement,
    node
  );
}

function constDeclToAst(node: CstNode): ConstDeclStatement {
  return withSpan(
    {
      kind: "ConstDecl",
      declarations: childNodes(node, "declarationSpec").map(declarationSpecToAst)
    } satisfies ConstDeclStatement,
    node
  );
}

function varDeclToAst(node: CstNode): VarDeclStatement {
  return withSpan(
    {
      kind: "VarDecl",
      declarations: childNodes(node, "declarationSpec").map(declarationSpecToAst)
    } satisfies VarDeclStatement,
    node
  );
}

function declarationSpecToAst(node: CstNode): DeclarationSpec {
  const nameNode = firstChildNode(node, "name");
  const typeNode = firstChildNode(node, "typeExpression");
  const valueNode = firstChildNode(node, "expression");
  return {
    name: nameNode ? nameText(nameNode) : "<missing>",
    ...(typeNode ? { type: typeToAst(typeNode) } : {}),
    ...(valueNode ? { value: expressionToAst(valueNode) } : {})
  };
}

function typeDeclToAst(node: CstNode): TypeDeclStatement {
  const names = childNodes(node, "name");
  const types = childNodes(node, "typeExpression");
  const declarations: TypeSpec[] = names.map((nameNode, index) => ({
    name: nameText(nameNode),
    type: types[index] ? typeToAst(types[index]) : { text: "<missing>" }
  }));

  return withSpan(
    {
      kind: "TypeDecl",
      declarations
    } satisfies TypeDeclStatement,
    node
  );
}

function returnStmtToAst(node: CstNode): ReturnStatement {
  const expressionList = firstChildNode(node, "expressionList");
  return withSpan(
    {
      kind: "ReturnStatement",
      values: expressionList ? expressionListToAst(expressionList) : []
    } satisfies ReturnStatement,
    node
  );
}

function ifStmtToAst(node: CstNode): IfStatement {
  const condition = firstChildNode(node, "expression");
  const blocks = childNodes(node, "block");
  const nestedIf = firstChildNode(node, "ifStmt");
  const elseBlock = blocks[1];

  return withSpan(
    {
      kind: "IfStatement",
      condition: condition ? expressionToAst(condition) : missingExpression(),
      thenBlock: blocks[0] ? blockToAst(blocks[0]) : emptyBlock(),
      ...(nestedIf
        ? { elseBranch: ifStmtToAst(nestedIf) }
        : elseBlock
          ? { elseBranch: blockToAst(elseBlock) }
          : {})
    } satisfies IfStatement,
    node
  );
}

function switchStmtToAst(node: CstNode): SwitchStatement {
  const expression = firstChildNode(node, "expression");
  return withSpan(
    {
      kind: "SwitchStatement",
      ...(expression ? { expression: expressionToAst(expression) } : {}),
      clauses: childNodes(node, "switchClause").map(switchClauseToAst)
    } satisfies SwitchStatement,
    node
  );
}

function switchClauseToAst(node: CstNode): SwitchClause {
  const expressionList = firstChildNode(node, "expressionList");
  return {
    kind: "SwitchClause",
    values: expressionList ? expressionListToAst(expressionList) : [],
    default: childTokens(node, "Default").length > 0,
    statements: childNodes(node, "statement").map(statementToAst)
  };
}

function forStmtToAst(node: CstNode): ForStatement {
  const range = firstChildNode(node, "rangeClause");
  const forClause = firstChildNode(node, "forClause");
  const condition = firstChildNode(node, "expression");
  const body = firstChildNode(node, "block");
  return withSpan(
    {
      kind: "ForStatement",
      ...(range ? { range: rangeClauseToAst(range) } : {}),
      ...(forClause ? forClauseToAst(forClause) : {}),
      ...(!range && condition ? { condition: expressionToAst(condition) } : {}),
      body: body ? blockToAst(body) : emptyBlock()
    } satisfies ForStatement,
    node
  );
}

function forClauseToAst(node: CstNode): Pick<ForStatement, "init" | "condition" | "post"> {
  const init = firstChildNode(node, "forInitClause");
  const post = firstChildNode(node, "forPostClause");
  const condition = firstChildNode(node, "expression");
  const result: Pick<ForStatement, "init" | "condition" | "post"> = {};

  if (init) result.init = forInitClauseToAst(init);
  if (condition) {
    result.condition = expressionToAst(condition);
  }
  if (post) result.post = forPostClauseToAst(post);

  return result;
}

function forInitClauseToAst(node: CstNode): ShortVarStatement | AssignStatement {
  const name = firstChildNode(node, "name");
  const value = firstChildNode(node, "expression");
  const target: IdentifierExpression = {
    kind: "Identifier",
    name: name ? nameText(name) : "<missing>"
  };

  if (childTokens(node, "Define").length > 0) {
    return withSpan(
      {
        kind: "ShortVarStatement",
        name: target.name,
        value: value ? expressionToAst(value) : missingExpression()
      } satisfies ShortVarStatement,
      node
    );
  }

  return withSpan(
    {
      kind: "AssignStatement",
      target,
      value: value ? expressionToAst(value) : missingExpression()
    } satisfies AssignStatement,
    node
  );
}

function forPostClauseToAst(node: CstNode): IncDecStatement {
  const target = firstChildNode(node, "expression");
  const operator = childTokens(node, "PlusPlus").length > 0 ? "++" : "--";
  return withSpan(
    {
      kind: "IncDecStatement",
      target: target ? expressionToAst(target) : missingExpression(),
      operator
    } satisfies IncDecStatement,
    node
  );
}

function rangeClauseToAst(node: CstNode) {
  const names = childNodes(node, "name");
  const source = firstChildNode(node, "expression");
  return {
    ...(names[0] ? { keyName: nameText(names[0]) } : {}),
    ...(names[1] ? { valueName: nameText(names[1]) } : {}),
    define: childTokens(node, "Define").length > 0,
    source: source ? expressionToAst(source) : missingExpression()
  };
}

function deferStmtToAst(node: CstNode): DeferStatement {
  const expression = firstChildNode(node, "expression");
  return withSpan(
    {
      kind: "DeferStatement",
      expression: expression ? expressionToAst(expression) : missingExpression()
    } satisfies DeferStatement,
    node
  );
}

function branchToAst(node: CstNode): BranchStatement {
  const token = firstChildTokenAny(node, ["Break", "Continue", "Fallthrough", "Goto"]);
  const image = token?.image as "break" | "continue" | "fallthrough" | "goto";
  const label = firstChildToken(node, "Identifier")?.image;
  return withSpan(
    {
      kind: "BranchStatement",
      branch: image,
      ...(label ? { label } : {})
    } satisfies BranchStatement,
    node
  );
}

function simpleStmtToAst(node: CstNode): AssignStatement | ShortVarStatement | IncDecStatement | ExpressionStatement {
  const expressions = childNodes(node, "expression").map(expressionToAst);
  const target = expressions[0] ?? missingExpression();
  const value = expressions[1];

  if (value && childTokens(node, "Define").length > 0) {
    return withSpan(
      {
        kind: "ShortVarStatement",
        name: target.kind === "Identifier" ? target.name : "<invalid>",
        value
      } satisfies ShortVarStatement,
      node
    );
  }

  if (value && childTokens(node, "Assign").length > 0) {
    return withSpan(
      {
        kind: "AssignStatement",
        target,
        value
      } satisfies AssignStatement,
      node
    );
  }

  const incDecToken = firstChildTokenAny(node, ["PlusPlus", "MinusMinus"]);
  if (incDecToken) {
    return withSpan(
      {
        kind: "IncDecStatement",
        target,
        operator: incDecToken.tokenType.name === "PlusPlus" ? "++" : "--"
      } satisfies IncDecStatement,
      node
    );
  }

  return withSpan(
    {
      kind: "ExpressionStatement",
      expression: target
    } satisfies ExpressionStatement,
    node
  );
}

function expressionListToAst(node: CstNode): Expression[] {
  return childNodes(node, "expression").map(expressionToAst);
}

function expressionToAst(node: CstNode): Expression {
  return orExprToAst(requiredChildNode(node, "orExpr"));
}

function orExprToAst(node: CstNode): Expression {
  return binaryRuleToAst(node, "andExpr", ["OrOr"], andExprToAst);
}

function andExprToAst(node: CstNode): Expression {
  return binaryRuleToAst(node, "equalityExpr", ["AndAnd"], equalityExprToAst);
}

function equalityExprToAst(node: CstNode): Expression {
  return binaryRuleToAst(node, "compareExpr", ["Equal", "NotEqual"], compareExprToAst);
}

function compareExprToAst(node: CstNode): Expression {
  return binaryRuleToAst(
    node,
    "addExpr",
    ["Less", "LessEqual", "Greater", "GreaterEqual"],
    addExprToAst
  );
}

function addExprToAst(node: CstNode): Expression {
  return binaryRuleToAst(node, "mulExpr", ["Plus", "Minus"], mulExprToAst);
}

function mulExprToAst(node: CstNode): Expression {
  return binaryRuleToAst(node, "unaryExpr", ["Star", "Slash", "Percent"], unaryExprToAst);
}

function binaryRuleToAst(
  node: CstNode,
  operandRuleName: string,
  operatorTokenNames: string[],
  operandMapper: (node: CstNode) => Expression
): Expression {
  const operands = childNodes(node, operandRuleName)
    .sort(byNodeOffset)
    .map(operandMapper);
  const operators = childTokensAny(node, operatorTokenNames).sort(byOffset);

  let expression = operands[0] ?? missingExpression();
  for (let index = 0; index < operators.length; index += 1) {
    const operator = operators[index]?.image as BinaryOperator | undefined;
    expression = withSpan(
      {
        kind: "BinaryExpression",
        operator: operator ?? "+",
        left: expression,
        right: operands[index + 1] ?? missingExpression()
      } satisfies BinaryExpression,
      node
    );
  }
  return expression;
}

function unaryExprToAst(node: CstNode): Expression {
  const primary = firstChildNode(node, "primaryExpr");
  if (primary) return primaryExprToAst(primary);

  const operator = firstChildTokenAny(node, ["Plus", "Minus", "Bang", "Amp", "Star"]);
  const operand = firstChildNode(node, "unaryExpr");
  return withSpan(
    {
      kind: "UnaryExpression",
      operator: unaryOperatorImage(operator),
      operand: operand ? unaryExprToAst(operand) : missingExpression()
    } satisfies UnaryExpression,
    node
  );
}

function primaryExprToAst(node: CstNode): Expression {
  let expression = atomToAst(requiredChildNode(node, "atom"));
  const events = primaryEvents(node);

  for (const event of events) {
    if (event.kind === "selector") {
      expression = withSpan(
        {
          kind: "SelectorExpression",
          object: expression,
          field: event.field
        } satisfies SelectorExpression,
        event.token
      );
      continue;
    }

    if (event.kind === "call") {
      expression = withSpan(
        {
          kind: "CallExpression",
          callee: expression,
          args: argumentExpressions(event.node),
          spreadLast: childTokens(event.node, "Ellipsis").length > 0
        } satisfies CallExpression,
        event.node
      );
      continue;
    }

    if (event.kind === "index") {
      expression = withSpan(
        {
          kind: "IndexExpression",
          object: expression,
          index: event.index
        } satisfies IndexExpression,
        event.token
      );
      continue;
    }

    if (event.kind === "slice") {
      expression = withSpan(
        {
          kind: "SliceExpression",
          object: expression,
          ...(event.start ? { start: event.start } : {}),
          ...(event.end ? { end: event.end } : {})
        } satisfies SliceExpression,
        event.token
      );
      continue;
    }

    if (event.kind === "range") {
      expression = withSpan(
        {
          kind: "SpreadsheetRangeExpression",
          start: selectorStart(expression),
          endCell: event.endCell
        } satisfies SpreadsheetRangeExpression,
        event.token
      );
    }
  }

  return expression;
}

type PrimaryEvent =
  | { kind: "selector"; offset: number; token: IToken; field: string }
  | { kind: "call"; offset: number; node: CstNode }
  | { kind: "index"; offset: number; token: IToken; index: Expression }
  | { kind: "slice"; offset: number; token: IToken; start?: Expression; end?: Expression }
  | { kind: "range"; offset: number; token: IToken; endCell: string };

function primaryEvents(node: CstNode): PrimaryEvent[] {
  const events: PrimaryEvent[] = [];
  const selectorNodes = childNodes(node, "selectorName").sort(byNodeOffset);
  const usedSelectors = new Set<CstNode>();
  const brackets = bracketPairs(node);

  for (const dot of childTokens(node, "Dot")) {
    const selector = selectorNodes.find(
      (candidate) => !usedSelectors.has(candidate) && nodeOffset(candidate) > tokenEnd(dot)
    );
    if (!selector) continue;
    usedSelectors.add(selector);
    events.push({
      kind: "selector",
      offset: dot.startOffset,
      token: dot,
      field: nameText(selector)
    });
  }

  for (const argsNode of childNodes(node, "arguments")) {
    events.push({
      kind: "call",
      offset: nodeOffset(argsNode),
      node: argsNode
    });
  }

  for (const bracket of brackets) {
    const expressions = childNodes(node, "expression")
      .filter((expr) => nodeOffset(expr) > bracket.left.startOffset && nodeOffset(expr) < bracket.right.startOffset)
      .sort(byNodeOffset)
      .map(expressionToAst);
    const colon = childTokens(node, "Colon").find(
      (token) => token.startOffset > bracket.left.startOffset && token.startOffset < bracket.right.startOffset
    );
    if (colon) {
      events.push({
        kind: "slice",
        offset: bracket.left.startOffset,
        token: bracket.left,
        ...(expressions[0] ? { start: expressions[0] } : {}),
        ...(expressions[1] ? { end: expressions[1] } : {})
      });
    } else if (expressions[0]) {
      events.push({
        kind: "index",
        offset: bracket.left.startOffset,
        token: bracket.left,
        index: expressions[0]
      });
    }
  }

  const bracketRanges = brackets.map((bracket) => [bracket.left.startOffset, bracket.right.startOffset] as const);
  for (const colon of childTokens(node, "Colon")) {
    const insideBracket = bracketRanges.some(([start, end]) => colon.startOffset > start && colon.startOffset < end);
    if (insideBracket) continue;

    const selector = selectorNodes.find(
      (candidate) => !usedSelectors.has(candidate) && nodeOffset(candidate) > colon.startOffset
    );
    if (!selector) continue;
    usedSelectors.add(selector);
    events.push({
      kind: "range",
      offset: colon.startOffset,
      token: colon,
      endCell: nameText(selector)
    });
  }

  return events.sort((left, right) => left.offset - right.offset);
}

function bracketPairs(node: CstNode): Array<{ left: IToken; right: IToken }> {
  const lefts = childTokens(node, "LBracket").sort(byOffset);
  const rights = childTokens(node, "RBracket").sort(byOffset);
  return lefts.flatMap((left, index) => {
    const right = rights[index];
    return right ? [{ left, right }] : [];
  });
}

function argumentExpressions(node: CstNode): Expression[] {
  return childNodes(node, "expression").map(expressionToAst);
}

function selectorStart(expression: Expression): SelectorExpression {
  if (expression.kind === "SelectorExpression") return expression;
  return {
    kind: "SelectorExpression",
    object: expression,
    field: "<invalid>"
  };
}

function atomToAst(node: CstNode): Expression {
  const functionLiteral = firstChildNode(node, "functionLiteral");
  if (functionLiteral) return functionLiteralToAst(functionLiteral);

  const mapLiteral = firstChildNode(node, "mapLiteral");
  if (mapLiteral) return mapLiteralToAst(mapLiteral);

  const literal = firstChildNode(node, "literal");
  if (literal) return literalToAst(literal);

  const qualified = firstChildNode(node, "qualifiedName");
  if (qualified) return qualifiedNameToAst(qualified);

  const expression = firstChildNode(node, "expression");
  return expression ? expressionToAst(expression) : missingExpression();
}

function functionLiteralToAst(node: CstNode): FunctionLiteralExpression {
  return withSpan(
    {
      kind: "FunctionLiteralExpression",
      signature: signatureToAst(requiredChildNode(node, "signature")),
      body: blockToAst(requiredChildNode(node, "block"))
    } satisfies FunctionLiteralExpression,
    node
  );
}

function mapLiteralToAst(node: CstNode): MapLiteralExpression {
  const types = childNodes(node, "typeExpression");
  return withSpan(
    {
      kind: "MapLiteralExpression",
      keyType: types[0] ? typeToAst(types[0]) : { text: "<missing>" },
      valueType: types[1] ? typeToAst(types[1]) : { text: "<missing>" },
      entries: childNodes(node, "mapElement").map(mapElementToAst)
    } satisfies MapLiteralExpression,
    node
  );
}

function mapElementToAst(node: CstNode): MapEntryExpression {
  const expressions = childNodes(node, "expression");
  return {
    key: expressions[0] ? expressionToAst(expressions[0]) : missingExpression(),
    value: expressions[1] ? expressionToAst(expressions[1]) : missingExpression()
  };
}

function qualifiedNameToAst(node: CstNode): Expression {
  const baseName = firstChildNode(node, "name");
  let expression: Expression = withSpan(
    {
      kind: "Identifier",
      name: baseName ? nameText(baseName) : "<missing>"
    } satisfies IdentifierExpression,
    baseName ?? node
  );

  for (const selector of childNodes(node, "selectorName").sort(byNodeOffset)) {
    expression = withSpan(
      {
        kind: "SelectorExpression",
        object: expression,
        field: nameText(selector)
      } satisfies SelectorExpression,
      selector
    );
  }

  return expression;
}

function literalToAst(node: CstNode): LiteralExpression {
  const token = firstChildTokenAny(node, ["StringLiteral", "FloatLiteral", "IntLiteral", "True", "False", "Nil"]);
  if (!token) {
    return { kind: "Literal", literalKind: "nil", value: null, raw: "nil" };
  }

  if (token.tokenType.name === "StringLiteral") {
    return withSpan(
      {
        kind: "Literal",
        literalKind: "string",
        value: unquote(token.image),
        raw: token.image
      } satisfies LiteralExpression,
      token
    );
  }

  if (token.tokenType.name === "FloatLiteral") {
    return withSpan(
      {
        kind: "Literal",
        literalKind: "float",
        value: Number(token.image),
        raw: token.image
      } satisfies LiteralExpression,
      token
    );
  }

  if (token.tokenType.name === "IntLiteral") {
    return withSpan(
      {
        kind: "Literal",
        literalKind: "int",
        value: BigInt(token.image),
        raw: token.image
      } satisfies LiteralExpression,
      token
    );
  }

  if (token.tokenType.name === "True" || token.tokenType.name === "False") {
    return withSpan(
      {
        kind: "Literal",
        literalKind: "bool",
        value: token.tokenType.name === "True",
        raw: token.image
      } satisfies LiteralExpression,
      token
    );
  }

  return withSpan(
    {
      kind: "Literal",
      literalKind: "nil",
      value: null,
      raw: token.image
    } satisfies LiteralExpression,
    token
  );
}

function nameText(node: CstNode): string {
  return nameToken(node)?.image ?? "<missing>";
}

function nameToken(node: CstNode): IToken | undefined {
  return firstChildToken(node, "Identifier") ?? firstChildToken(node, "CellAddress");
}

function unaryOperatorImage(token: IToken | undefined): UnaryOperator {
  if (!token) return "+";
  if (token.tokenType.name === "Bang") return "!";
  return token.image as UnaryOperator;
}

function missingExpression(): IdentifierExpression {
  return {
    kind: "Identifier",
    name: "<missing>"
  };
}

function emptyBlock(): BlockStatement {
  return {
    kind: "BlockStatement",
    statements: []
  };
}

function firstChildNode(node: CstNode, name: string): CstNode | undefined {
  return childNodes(node, name)[0];
}

function requiredChildNode(node: CstNode, name: string): CstNode {
  return firstChildNode(node, name) ?? {
    name,
    children: {}
  };
}

function childNodes(node: CstNode, name: string): CstNode[] {
  const values = node.children[name] ?? [];
  return values.filter(isCstNode);
}

function firstChildToken(node: CstNode, name: string): IToken | undefined {
  return childTokens(node, name)[0];
}

function firstChildTokenAny(node: CstNode, names: string[]): IToken | undefined {
  return childTokensAny(node, names).sort(byOffset)[0];
}

function childTokens(node: CstNode, name: string): IToken[] {
  const values = node.children[name] ?? [];
  return values.filter(isToken);
}

function childTokensAny(node: CstNode, names: string[]): IToken[] {
  return names.flatMap((name) => childTokens(node, name));
}

function allTokensIn(node: CstNode): IToken[] {
  const tokens: IToken[] = [];
  for (const values of Object.values(node.children)) {
    for (const value of values) {
      if (isToken(value)) {
        tokens.push(value);
      } else if (isCstNode(value)) {
        tokens.push(...allTokensIn(value));
      }
    }
  }
  return tokens;
}

function hasTokens(node: CstNode): boolean {
  return allTokensIn(node).length > 0;
}

function spanFromNode(node: CstNode): SourceSpan | undefined {
  if (!hasTokens(node)) return undefined;
  const tokens = allTokensIn(node).sort(byOffset);
  const first = tokens[0];
  const last = tokens[tokens.length - 1];
  if (!first || !last) return undefined;
  const endOffset = tokenEnd(last);
  return {
    offset: first.startOffset,
    length: Math.max(0, endOffset - first.startOffset + 1),
    line: first.startLine ?? 1,
    column: first.startColumn ?? 1
  };
}

function spanFromToken(token: IToken): SourceSpan {
  const startOffset = finiteOr(token.startOffset, 0);
  const endOffset = finiteOr(tokenEnd(token), startOffset);
  return {
    offset: startOffset,
    length: Math.max(0, endOffset - startOffset + 1),
    line: finiteOr(token.startLine, 1),
    column: finiteOr(token.startColumn, 1)
  };
}

function withSpan<T extends object>(value: T, source: CstNode | IToken | undefined): T {
  const span = source && isToken(source) ? spanFromToken(source) : source ? spanFromNode(source) : undefined;
  if (span) {
    (value as T & { span: SourceSpan }).span = span;
  }
  return value;
}

function nodeOffset(node: CstNode): number {
  return allTokensIn(node).sort(byOffset)[0]?.startOffset ?? Number.MAX_SAFE_INTEGER;
}

function byNodeOffset(left: CstNode, right: CstNode): number {
  return nodeOffset(left) - nodeOffset(right);
}

function byOffset(left: IToken, right: IToken): number {
  return left.startOffset - right.startOffset;
}

function tokenEnd(token: IToken): number {
  return token.endOffset ?? token.startOffset + token.image.length - 1;
}

function finiteOr(value: number | undefined, fallback: number): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}

function isCstNode(value: unknown): value is CstNode {
  return Boolean(value && typeof value === "object" && "children" in value);
}

function isToken(value: unknown): value is IToken {
  return Boolean(value && typeof value === "object" && "image" in value && "tokenType" in value);
}

function unquote(image: string): string {
  try {
    return JSON.parse(image) as string;
  } catch {
    return image.slice(1, -1);
  }
}
