import type {
  AssignStatement,
  BinaryExpression,
  BlockStatement,
  CallExpression,
  CommClause,
  DeferStatement,
  Expression,
  ForStatement,
  FunctionDecl,
  FunctionLiteralExpression,
  GoStatement,
  IfStatement,
  MapLiteralExpression,
  ProgramAst,
  SelectStatement,
  SendStatement,
  ShortVarStatement,
  SliceExpression,
  Statement,
  SwitchStatement,
  UnaryExpression
} from "./ast.js";

export type EffectReasonKind =
  | "go"
  | "channel-send"
  | "channel-receive"
  | "select"
  | "call";

export interface EffectReason {
  kind: EffectReasonKind;
  target?: string;
}

export interface FunctionEffectSummary {
  key: string;
  name: string;
  direct: boolean;
  maySuspend: boolean;
  calls: string[];
  reasons: EffectReason[];
}

export interface ProgramEffectAnalysis {
  topLevel: FunctionEffectSummary;
  functions: Record<string, FunctionEffectSummary>;
}

interface MutableEffectSummary {
  key: string;
  name: string;
  intrinsicReasons: EffectReason[];
  calls: Set<string>;
  maySuspend: boolean;
}

export function analyzeEffects(program: ProgramAst): ProgramEffectAnalysis {
  const functionKeys = new Map<string, string>();
  for (const declaration of program.functions) {
    functionKeys.set(declaration.name, functionEffectKey(declaration));
  }

  const summaries = new Map<string, MutableEffectSummary>();
  for (const declaration of program.functions) {
    const key = functionEffectKey(declaration);
    summaries.set(key, analyzeBlock(key, declaration.name, declaration.body, functionKeys));
  }

  const topLevel = analyzeStatements("<top-level>", "<top-level>", program.body, functionKeys);

  let changed = true;
  while (changed) {
    changed = false;
    for (const summary of summaries.values()) {
      if (summary.maySuspend) continue;
      for (const target of summary.calls) {
        if (summaries.get(target)?.maySuspend) {
          summary.maySuspend = true;
          changed = true;
          break;
        }
      }
    }
    if (!topLevel.maySuspend) {
      for (const target of topLevel.calls) {
        if (summaries.get(target)?.maySuspend) {
          topLevel.maySuspend = true;
          changed = true;
          break;
        }
      }
    }
  }

  const functions: Record<string, FunctionEffectSummary> = {};
  for (const [key, summary] of summaries) {
    functions[key] = freezeSummary(summary, summaries);
  }

  return {
    topLevel: freezeSummary(topLevel, summaries),
    functions
  };
}

export function functionEffectKey(declaration: FunctionDecl): string {
  if (!declaration.receiver) return declaration.name;
  return `${declaration.receiver.type.text}.${declaration.name}`;
}

function analyzeBlock(
  key: string,
  name: string,
  block: BlockStatement,
  functionKeys: Map<string, string>
): MutableEffectSummary {
  return analyzeStatements(key, name, block.statements, functionKeys);
}

function analyzeStatements(
  key: string,
  name: string,
  statements: Statement[],
  functionKeys: Map<string, string>
): MutableEffectSummary {
  const summary: MutableEffectSummary = {
    key,
    name,
    intrinsicReasons: [],
    calls: new Set(),
    maySuspend: false
  };
  for (const statement of statements) analyzeStatement(statement, summary, functionKeys);
  summary.maySuspend = summary.intrinsicReasons.length > 0;
  return summary;
}

function analyzeStatement(
  statement: Statement | undefined,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  if (!statement) return;
  switch (statement.kind) {
    case "BlockStatement":
      analyzeStatementsInto(statement.statements, summary, functionKeys);
      return;
    case "LabeledStatement":
      analyzeStatement(statement.statement, summary, functionKeys);
      return;
    case "ConstDecl":
    case "VarDecl":
      for (const declaration of statement.declarations) analyzeExpression(declaration.value, summary, functionKeys);
      return;
    case "TypeDecl":
    case "BranchStatement":
    case "IncDecStatement":
      if (statement.kind === "IncDecStatement") analyzeExpression(statement.target, summary, functionKeys);
      return;
    case "ReturnStatement":
      for (const value of statement.values) analyzeExpression(value, summary, functionKeys);
      return;
    case "IfStatement":
      analyzeIf(statement, summary, functionKeys);
      return;
    case "SwitchStatement":
      analyzeSwitch(statement, summary, functionKeys);
      return;
    case "SelectStatement":
      addReason(summary, { kind: "select" });
      analyzeSelect(statement, summary, functionKeys);
      return;
    case "ForStatement":
      analyzeFor(statement, summary, functionKeys);
      return;
    case "DeferStatement":
      analyzeDefer(statement, summary, functionKeys);
      return;
    case "GoStatement":
      analyzeGo(statement, summary, functionKeys);
      return;
    case "SendStatement":
      analyzeSend(statement, summary, functionKeys);
      return;
    case "AssignStatement":
      analyzeAssign(statement, summary, functionKeys);
      return;
    case "ShortVarStatement":
      analyzeShortVar(statement, summary, functionKeys);
      return;
    case "ExpressionStatement":
      analyzeExpression(statement.expression, summary, functionKeys);
      return;
  }
}

function analyzeStatementsInto(
  statements: Statement[],
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  for (const statement of statements) analyzeStatement(statement, summary, functionKeys);
}

function analyzeIf(
  statement: IfStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  analyzeStatement(statement.init, summary, functionKeys);
  analyzeExpression(statement.condition, summary, functionKeys);
  analyzeStatement(statement.thenBlock, summary, functionKeys);
  analyzeStatement(statement.elseBranch, summary, functionKeys);
}

function analyzeSwitch(
  statement: SwitchStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  analyzeStatement(statement.init, summary, functionKeys);
  analyzeExpression(statement.expression, summary, functionKeys);
  analyzeExpression(statement.typeSwitch?.expression, summary, functionKeys);
  for (const clause of statement.clauses) {
    for (const value of clause.values) analyzeExpression(value, summary, functionKeys);
    analyzeStatementsInto(clause.statements, summary, functionKeys);
  }
}

function analyzeSelect(
  statement: SelectStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  for (const clause of statement.clauses) analyzeCommClause(clause, summary, functionKeys);
}

function analyzeCommClause(
  clause: CommClause,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  analyzeStatement(clause.comm, summary, functionKeys);
  analyzeStatementsInto(clause.statements, summary, functionKeys);
}

function analyzeFor(
  statement: ForStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  analyzeStatement(statement.init, summary, functionKeys);
  analyzeExpression(statement.condition, summary, functionKeys);
  analyzeStatement(statement.post, summary, functionKeys);
  analyzeExpression(statement.range?.source, summary, functionKeys);
  analyzeStatement(statement.body, summary, functionKeys);
}

function analyzeDefer(
  statement: DeferStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  analyzeExpression(statement.expression, summary, functionKeys);
}

function analyzeGo(
  statement: GoStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  addReason(summary, { kind: "go" });
  analyzeCall(statement.call, summary, functionKeys);
}

function analyzeSend(
  statement: SendStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  addReason(summary, { kind: "channel-send" });
  analyzeExpression(statement.channel, summary, functionKeys);
  analyzeExpression(statement.value, summary, functionKeys);
}

function analyzeAssign(
  statement: AssignStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  for (const target of statement.targets) analyzeExpression(target, summary, functionKeys);
  for (const value of statement.values) analyzeExpression(value, summary, functionKeys);
}

function analyzeShortVar(
  statement: ShortVarStatement,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  for (const value of statement.values) analyzeExpression(value, summary, functionKeys);
}

function analyzeExpression(
  expression: Expression | undefined,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  if (!expression) return;
  switch (expression.kind) {
    case "Identifier":
    case "TypeExpression":
    case "Literal":
      return;
    case "FunctionLiteralExpression":
      analyzeFunctionLiteral(expression, summary, functionKeys);
      return;
    case "ArrayLiteralExpression":
      for (const element of expression.elements) analyzeExpression(element, summary, functionKeys);
      return;
    case "StructLiteralExpression":
      for (const field of expression.fields) analyzeExpression(field.value, summary, functionKeys);
      return;
    case "MapLiteralExpression":
      analyzeMapLiteral(expression, summary, functionKeys);
      return;
    case "UnaryExpression":
      analyzeUnary(expression, summary, functionKeys);
      return;
    case "BinaryExpression":
      analyzeBinary(expression, summary, functionKeys);
      return;
    case "TypeAssertionExpression":
      analyzeExpression(expression.expression, summary, functionKeys);
      return;
    case "SelectorExpression":
      analyzeExpression(expression.object, summary, functionKeys);
      return;
    case "CallExpression":
      analyzeCall(expression, summary, functionKeys);
      return;
    case "IndexExpression":
      analyzeExpression(expression.object, summary, functionKeys);
      analyzeExpression(expression.index, summary, functionKeys);
      return;
    case "SliceExpression":
      analyzeSlice(expression, summary, functionKeys);
      return;
    case "SpreadsheetRangeExpression":
      analyzeExpression(expression.start, summary, functionKeys);
      return;
  }
}

function analyzeFunctionLiteral(
  expression: FunctionLiteralExpression,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  // A function literal value does not execute where it appears. Its body is
  // analyzed when the literal is immediately called or launched.
  void expression;
  void summary;
  void functionKeys;
}

function analyzeMapLiteral(
  expression: MapLiteralExpression,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  for (const entry of expression.entries) {
    analyzeExpression(entry.key, summary, functionKeys);
    analyzeExpression(entry.value, summary, functionKeys);
  }
}

function analyzeUnary(
  expression: UnaryExpression,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  if (expression.operator === "<-") addReason(summary, { kind: "channel-receive" });
  analyzeExpression(expression.operand, summary, functionKeys);
}

function analyzeBinary(
  expression: BinaryExpression,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  analyzeExpression(expression.left, summary, functionKeys);
  analyzeExpression(expression.right, summary, functionKeys);
}

function analyzeSlice(
  expression: SliceExpression,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  analyzeExpression(expression.object, summary, functionKeys);
  analyzeExpression(expression.start, summary, functionKeys);
  analyzeExpression(expression.end, summary, functionKeys);
  analyzeExpression(expression.max, summary, functionKeys);
}

function analyzeCall(
  expression: CallExpression,
  summary: MutableEffectSummary,
  functionKeys: Map<string, string>
): void {
  if (expression.callee.kind === "Identifier") {
    const target = functionKeys.get(expression.callee.name);
    if (target) summary.calls.add(target);
  } else if (expression.callee.kind === "FunctionLiteralExpression") {
    const literalSummary = analyzeBlock("<function-literal>", "<function-literal>", expression.callee.body, functionKeys);
    for (const reason of literalSummary.intrinsicReasons) addReason(summary, reason);
    for (const call of literalSummary.calls) summary.calls.add(call);
  } else {
    analyzeExpression(expression.callee, summary, functionKeys);
  }
  for (const arg of expression.args) analyzeExpression(arg, summary, functionKeys);
}

function addReason(summary: MutableEffectSummary, reason: EffectReason): void {
  if (summary.intrinsicReasons.some((item) => item.kind === reason.kind && item.target === reason.target)) return;
  summary.intrinsicReasons.push(reason);
}

function freezeSummary(
  summary: MutableEffectSummary,
  summaries: Map<string, MutableEffectSummary>
): FunctionEffectSummary {
  const reasons = [...summary.intrinsicReasons];
  for (const target of summary.calls) {
    if (summaries.get(target)?.maySuspend && !reasons.some((item) => item.kind === "call" && item.target === target)) {
      reasons.push({ kind: "call", target });
    }
  }
  const maySuspend = summary.maySuspend || reasons.length > 0;
  return {
    key: summary.key,
    name: summary.name,
    direct: !maySuspend,
    maySuspend,
    calls: [...summary.calls].sort(),
    reasons
  };
}
