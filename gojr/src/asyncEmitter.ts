import type {
  AssignStatement,
  BinaryExpression,
  BlockStatement,
  CallExpression,
  CommClause,
  Expression,
  FunctionDecl,
  FunctionLiteralExpression,
  GoStatement,
  LiteralExpression,
  ProgramAst,
  ReturnStatement,
  SelectStatement,
  SendStatement,
  Statement,
  ShortVarStatement,
  TypeAssertionExpression,
  TypeExpression,
  UnaryExpression
} from "./ast.js";
import { Diagnostic, REPL_FILENAME, SourceSpan } from "./diagnostics.js";
import { analyzeEffects, ProgramEffectAnalysis } from "./effects.js";

export interface AsyncEmitResult {
  diagnostics: Diagnostic[];
  effects: ProgramEffectAnalysis;
  source: string;
}

export interface AsyncEmitOptions {
  readonly sessionState?: boolean;
  readonly returnLastExpression?: boolean;
}

interface EmitFunctionInfo {
  declaration: FunctionDecl;
  jsName: string;
}

interface EmitScope {
  readonly parent?: EmitScope;
  readonly bindings: Map<string, string>;
  readonly topLevel?: boolean;
}

class AsyncEmitter {
  private readonly diagnostics: Diagnostic[] = [];
  private readonly functions = new Map<string, EmitFunctionInfo>();
  private nextLocal = 0;

  public constructor(
    private readonly program: ProgramAst,
    private readonly options: AsyncEmitOptions = {}
  ) {
    for (const declaration of program.functions) {
      if (declaration.receiver) {
        this.unsupported("method emission is not in the async emitter slice yet", declaration.span);
        continue;
      }
      this.functions.set(declaration.name, {
        declaration,
        jsName: generatedFunctionName(declaration.name)
      });
    }
  }

  public emit(): AsyncEmitResult {
    const effects = analyzeEffects(this.program);
    const source = this.diagnostics.length > 0 ? "" : this.emitModule(effects);
    return {
      diagnostics: this.diagnostics,
      effects,
      source: this.diagnostics.length > 0 ? "" : source
    };
  }

  private emitModule(effects: ProgramEffectAnalysis): string {
    const lines: string[] = [];
    lines.push(`"use strict";`);
    lines.push(`const __gojr = ((__runtime = {}) => {`);
    lines.push(`  const __scheduler = __runtime.scheduler;`);
    lines.push(`  const __AsyncGoChannel = __runtime.AsyncGoChannel;`);
    lines.push(`  const __asyncSelect = __runtime.asyncSelect;`);
    lines.push(`  const __state = __runtime.state ?? Object.create(null);`);
    lines.push(`  const __makeChannel = (capacity, zeroValue) => new __AsyncGoChannel(__scheduler, Number(capacity ?? 0), zeroValue);`);
    lines.push(`  const __functions = Object.create(null);`);
    lines.push(`  const __effects = ${JSON.stringify(effects)};`);
    for (const info of this.functions.values()) {
      lines.push(this.indent(this.emitFunction(info), 1));
      lines.push(`  __functions[${JSON.stringify(info.declaration.name)}] = ${info.jsName};`);
      if (this.options.sessionState) {
        lines.push(`  __state[${JSON.stringify(info.declaration.name)}] = ${info.jsName};`);
      }
    }
    lines.push(this.indent(this.emitMain(), 1));
    lines.push(this.indent(this.emitRun(), 1));
    lines.push(`  return { functions: __functions, main: __main, run: __run, effects: __effects };`);
    lines.push(`})(arguments[0] ?? {});`);
    lines.push(`return __gojr;`);
    return lines.join("\n");
  }

  private emitFunction(info: EmitFunctionInfo): string {
    const scope = this.rootScope(false);
    const params: string[] = [];
    for (const [index, parameter] of info.declaration.signature.parameters.entries()) {
      const sourceName = parameter.name ?? `_arg${index}`;
      const jsName = this.declare(scope, sourceName);
      params.push(parameter.variadic ? `...${jsName}` : jsName);
    }
    const body = this.emitBlockStatements(info.declaration.body, scope, false);
    return [
      `async function ${info.jsName}(${params.join(", ")}) {`,
      this.indent(body || `return null;`, 1),
      `}`
    ].join("\n");
  }

  private emitMain(): string {
    const scope = this.rootScope(true);
    const body = this.emitStatements(this.program.body, scope, this.options.returnLastExpression === true);
    return [
      `async function __main() {`,
      this.indent(body || `return null;`, 1),
      `}`
    ].join("\n");
  }

  private emitRun(): string {
    return [
      `async function __run() {`,
      `  if (!__scheduler) return __main();`,
      `  let __result = undefined;`,
      `  __scheduler.go(async () => {`,
      `    __result = await __main();`,
      `  });`,
      `  await __scheduler.run();`,
      `  return __result;`,
      `}`
    ].join("\n");
  }

  private emitBlockStatements(block: BlockStatement, scope: EmitScope, createScope: boolean): string {
    return this.emitStatements(block.statements, createScope ? this.childScope(scope) : scope);
  }

  private emitStatements(statements: Statement[], scope: EmitScope, returnLastExpression = false): string {
    const lines: string[] = [];
    for (const [index, statement] of statements.entries()) {
      if (returnLastExpression && index === statements.length - 1 && statement.kind === "ExpressionStatement") {
        lines.push(`return ${this.emitExpression(statement.expression, scope)};`);
        continue;
      }
      const emitted = this.emitStatement(statement, scope);
      if (emitted) lines.push(emitted);
    }
    return lines.join("\n");
  }

  private emitStatement(statement: Statement, scope: EmitScope): string {
    switch (statement.kind) {
      case "BlockStatement":
        return `{\n${this.indent(this.emitBlockStatements(statement, scope, true), 1)}\n}`;
      case "VarDecl":
        return statement.declarations.map((declaration) => {
          const jsName = this.declare(scope, declaration.name);
          const value = declaration.value ? this.emitExpression(declaration.value, scope) : "null";
          return this.emitBindingInitialization(jsName, value);
        }).join("\n");
      case "ShortVarStatement":
        return this.emitShortVar(statement, scope);
      case "AssignStatement":
        return this.emitAssign(statement, scope);
      case "ReturnStatement":
        return this.emitReturn(statement, scope);
      case "ExpressionStatement":
        return `${this.emitExpression(statement.expression, scope)};`;
      case "GoStatement":
        return this.emitGo(statement, scope);
      case "SendStatement":
        return this.emitSend(statement, scope);
      case "SelectStatement":
        return this.emitSelect(statement, scope);
      case "IfStatement": {
        const condition = this.emitExpression(statement.condition, scope);
        const thenBody = this.emitBlockStatements(statement.thenBlock, this.childScope(scope), false);
        const elseBody = statement.elseBranch
          ? ` else ${statement.elseBranch.kind === "IfStatement"
            ? this.emitStatement(statement.elseBranch, scope)
            : `{\n${this.indent(this.emitBlockStatements(statement.elseBranch, this.childScope(scope), false), 1)}\n}`}`
          : "";
        return `if (${condition}) {\n${this.indent(thenBody, 1)}\n}${elseBody}`;
      }
      default:
        this.unsupported(`unsupported statement in async emitter: ${statement.kind}`, statement.span);
        return "";
    }
  }

  private emitAssign(statement: AssignStatement, scope: EmitScope): string {
    if (statement.operator !== "=") {
      this.unsupported(`unsupported assignment operator in async emitter: ${statement.operator}`, statement.span);
      return "";
    }
    const receive = statement.values.length === 1 ? receiveChannelFromExpression(statement.values[0]!) : undefined;
    if (receive && statement.targets.length > 1) {
      return this.emitReceiveAssign(statement, receive, scope);
    }
    return statement.targets.map((target, index) => {
      const value = statement.values[index] ? this.emitExpression(statement.values[index]!, scope) : "null";
      return `${this.emitAssignableExpression(target, scope)} = ${value};`;
    }).join("\n");
  }

  private emitShortVar(statement: ShortVarStatement, scope: EmitScope): string {
    const receive = statement.values.length === 1 ? receiveChannelFromExpression(statement.values[0]!) : undefined;
    if (receive && statement.names.length > 1) {
      return this.emitReceiveShortVar(statement, receive, scope);
    }
    return statement.names.map((name, index) => {
      const jsName = this.declare(scope, name);
      const value = statement.values[index] ? this.emitExpression(statement.values[index]!, scope) : "null";
      return this.emitBindingInitialization(jsName, value);
    }).join("\n");
  }

  private emitReceiveShortVar(statement: ShortVarStatement, channel: Expression, scope: EmitScope): string {
    if (statement.names.length !== 2) {
      this.unsupported("channel receive short declaration must bind one or two names", statement.span);
      return "";
    }
    const resultName = `_recv${this.nextLocal++}`;
    const lines = [`const ${resultName} = await ${this.emitExpression(channel, scope)}.receive();`];
    for (const [index, name] of statement.names.entries()) {
      const jsName = this.declare(scope, name);
      lines.push(this.emitBindingInitialization(jsName, `${resultName}[${index}]`));
    }
    return lines.join("\n");
  }

  private emitReceiveAssign(statement: AssignStatement, channel: Expression, scope: EmitScope): string {
    if (statement.targets.length !== 2) {
      this.unsupported("channel receive assignment must bind one or two targets", statement.span);
      return "";
    }
    const resultName = `_recv${this.nextLocal++}`;
    const lines = [`const ${resultName} = await ${this.emitExpression(channel, scope)}.receive();`];
    for (const [index, target] of statement.targets.entries()) {
      lines.push(`${this.emitAssignableExpression(target, scope)} = ${resultName}[${index}];`);
    }
    return lines.join("\n");
  }

  private emitReturn(statement: ReturnStatement, scope: EmitScope): string {
    const values = statement.values.map((value) => this.emitExpression(value, scope));
    if (values.length === 0) return `return null;`;
    if (values.length === 1) return `return ${values[0]};`;
    return `return [${values.join(", ")}];`;
  }

  private emitGo(statement: GoStatement, scope: EmitScope): string {
    return `__scheduler.go(async () => { await ${this.emitInvocation(statement.call, scope)}; });`;
  }

  private emitSend(statement: SendStatement, scope: EmitScope): string {
    return `await ${this.emitExpression(statement.channel, scope)}.send(${this.emitExpression(statement.value, scope)});`;
  }

  private emitSelect(statement: SelectStatement, scope: EmitScope): string {
    const resultName = `_sel${this.nextLocal++}`;
    const cases = statement.clauses.map((clause) => this.emitSelectCaseObject(clause, scope));
    const lines: string[] = [];
    lines.push(`{`);
    lines.push(`  const ${resultName} = await __asyncSelect(__scheduler, [${cases.join(", ")}]);`);
    lines.push(`  switch (${resultName}.index) {`);
    for (const [index, clause] of statement.clauses.entries()) {
      const caseScope = this.childScope(scope);
      const prelude = this.emitSelectCasePrelude(clause, resultName, caseScope);
      const body = this.emitStatements(clause.statements, caseScope);
      lines.push(`    case ${index}: {`);
      const caseLines = [prelude, body, `break;`].filter(Boolean).join("\n");
      lines.push(this.indent(caseLines, 3));
      lines.push(`    }`);
    }
    lines.push(`  }`);
    lines.push(`}`);
    return lines.join("\n");
  }

  private emitExpression(expression: Expression, scope: EmitScope): string {
    switch (expression.kind) {
      case "Identifier":
        return this.lookup(scope, expression.name, expression.span);
      case "Literal":
        return this.emitLiteral(expression);
      case "FunctionLiteralExpression":
        return this.emitFunctionLiteral(expression, scope);
      case "UnaryExpression":
        return this.emitUnary(expression, scope);
      case "BinaryExpression":
        return this.emitBinary(expression, scope);
      case "CallExpression":
        return this.emitCall(expression, scope);
      case "TypeAssertionExpression":
        return this.emitTypeAssertion(expression, scope);
      default:
        this.unsupported(`unsupported expression in async emitter: ${expression.kind}`, expression.span);
        return "undefined";
    }
  }

  private emitAssignableExpression(expression: Expression, scope: EmitScope): string {
    if (expression.kind === "Identifier") return this.lookup(scope, expression.name, expression.span);
    this.unsupported(`unsupported assignment target in async emitter: ${expression.kind}`, expression.span);
    return "undefined";
  }

  private emitLiteral(expression: LiteralExpression): string {
    switch (expression.literalKind) {
      case "int":
      case "rune":
        return `${String(expression.value)}n`;
      case "float":
        return JSON.stringify(expression.value);
      case "string":
        return JSON.stringify(expression.value);
      case "bool":
        return expression.value ? "true" : "false";
      case "nil":
        return "null";
      case "imag":
        return `{ real: 0, imag: ${JSON.stringify((expression.value as { imag: number }).imag)} }`;
    }
  }

  private emitUnary(expression: UnaryExpression, scope: EmitScope): string {
    const operand = this.emitExpression(expression.operand, scope);
    if (expression.operator === "<-") {
      return `(await ${operand}.receive())[0]`;
    }
    return `(${expression.operator}${operand})`;
  }

  private emitBinary(expression: BinaryExpression, scope: EmitScope): string {
    const left = this.emitExpression(expression.left, scope);
    const right = this.emitExpression(expression.right, scope);
    return `(${left} ${expression.operator} ${right})`;
  }

  private emitCall(expression: CallExpression, scope: EmitScope): string {
    if (expression.callee.kind === "Identifier" && expression.callee.name === "make") {
      return this.emitMake(expression, scope);
    }
    const builtin = this.emitBuiltinCall(expression, scope);
    if (builtin !== undefined) return builtin;
    return `(await ${this.emitInvocation(expression, scope)})`;
  }

  private emitBuiltinCall(expression: CallExpression, scope: EmitScope): string | undefined {
    if (expression.callee.kind !== "Identifier") return undefined;
    if (expression.callee.name === "close") {
      if (expression.args.length !== 1) {
        this.unsupported("close expects exactly one argument", expression.span);
        return "undefined";
      }
      return `(${this.emitExpression(expression.args[0]!, scope)}.close(), null)`;
    }
    if (expression.callee.name === "len") {
      if (expression.args.length !== 1) {
        this.unsupported("len expects exactly one argument", expression.span);
        return "undefined";
      }
      return `BigInt(${this.emitExpression(expression.args[0]!, scope)}.len())`;
    }
    if (expression.callee.name === "cap") {
      if (expression.args.length !== 1) {
        this.unsupported("cap expects exactly one argument", expression.span);
        return "undefined";
      }
      return `BigInt(${this.emitExpression(expression.args[0]!, scope)}.cap())`;
    }
    return undefined;
  }

  private emitInvocation(expression: CallExpression, scope: EmitScope): string {
    if (expression.callee.kind !== "Identifier") {
      if (expression.callee.kind === "FunctionLiteralExpression") {
        const callee = this.emitFunctionLiteral(expression.callee, scope);
        const args = expression.args.map((arg) => this.emitExpression(arg, scope)).join(", ");
        return `${callee}(${args})`;
      }
      this.unsupported("only direct function calls are supported in this async emitter slice", expression.span);
      return "undefined";
    }
    const target = this.functions.get(expression.callee.name);
    if (!target) {
      if (this.options.sessionState) {
        const callee = this.lookup(scope, expression.callee.name, expression.callee.span);
        const args = expression.args.map((arg) => this.emitExpression(arg, scope)).join(", ");
        return `${callee}(${args})`;
      }
      this.unsupported(`unknown function call in async emitter: ${expression.callee.name}`, expression.callee.span);
      return "undefined";
    }
    const args = expression.args.map((arg) => this.emitExpression(arg, scope)).join(", ");
    return `${target.jsName}(${args})`;
  }

  private emitMake(expression: CallExpression, scope: EmitScope): string {
    const typeArg = expression.args[0];
    if (!typeArg || typeArg.kind !== "TypeExpression" || !isChannelType(typeArg)) {
      this.unsupported("only make(chan T[, n]) is supported in this async emitter slice", expression.span);
      return "undefined";
    }
    const capacity = expression.args[1] ? this.emitExpression(expression.args[1], scope) : "0";
    return `__makeChannel(${capacity}, ${zeroValueFactoryForType(channelElementType(typeArg.type.text))})`;
  }

  private emitFunctionLiteral(expression: FunctionLiteralExpression, scope: EmitScope): string {
    const literalScope = this.childScope(scope);
    const params: string[] = [];
    for (const [index, parameter] of expression.signature.parameters.entries()) {
      const sourceName = parameter.name ?? `_arg${index}`;
      const jsName = this.declare(literalScope, sourceName);
      params.push(parameter.variadic ? `...${jsName}` : jsName);
    }
    const body = this.emitBlockStatements(expression.body, literalScope, false);
    return `(async function(${params.join(", ")}) {\n${this.indent(body || `return null;`, 1)}\n})`;
  }

  private emitSelectCaseObject(clause: CommClause, scope: EmitScope): string {
    if (clause.default || !clause.comm) return `{ op: "default" }`;
    if (clause.comm.kind === "SendStatement") {
      return `{ op: "send", channel: ${this.emitExpression(clause.comm.channel, scope)}, value: ${this.emitExpression(clause.comm.value, scope)} }`;
    }
    const receive = receiveChannelFromStatement(clause.comm);
    if (receive) return `{ op: "receive", channel: ${this.emitExpression(receive, scope)} }`;
    this.unsupported(`unsupported select communication clause in async emitter: ${clause.comm.kind}`, clause.comm.span);
    return `{ op: "default" }`;
  }

  private emitSelectCasePrelude(clause: CommClause, resultName: string, scope: EmitScope): string {
    if (!clause.comm) return "";
    if (clause.comm.kind === "ShortVarStatement") {
      return this.emitReceiveShortVarPrelude(clause.comm, resultName, scope);
    }
    if (clause.comm.kind === "AssignStatement") {
      return this.emitReceiveAssignPrelude(clause.comm, resultName, scope);
    }
    return "";
  }

  private emitReceiveShortVarPrelude(statement: ShortVarStatement, resultName: string, scope: EmitScope): string {
    if (!receiveChannelFromStatement(statement)) return "";
    return statement.names.map((name, index) => {
      const jsName = this.declare(scope, name);
      const property = index === 1 ? "ok" : "value";
      return `let ${jsName} = ${resultName}.${property};`;
    }).join("\n");
  }

  private emitReceiveAssignPrelude(statement: AssignStatement, resultName: string, scope: EmitScope): string {
    if (!receiveChannelFromStatement(statement)) return "";
    return statement.targets.map((target, index) => {
      const property = index === 1 ? "ok" : "value";
      return `${this.emitAssignableExpression(target, scope)} = ${resultName}.${property};`;
    }).join("\n");
  }

  private emitTypeAssertion(expression: TypeAssertionExpression, scope: EmitScope): string {
    this.unsupported("type assertion emission is not in the first async emitter slice", expression.span);
    return this.emitExpression(expression.expression, scope);
  }

  private rootScope(topLevel: boolean): EmitScope {
    return { bindings: new Map(), topLevel };
  }

  private childScope(parent: EmitScope): EmitScope {
    return { parent, bindings: new Map() };
  }

  private declare(scope: EmitScope, name: string): string {
    if (this.options.sessionState && scope.topLevel) {
      const jsName = `__state[${JSON.stringify(name)}]`;
      scope.bindings.set(name, jsName);
      return jsName;
    }
    const jsName = `_v${this.nextLocal++}_${sanitizeIdentifierPart(name)}`;
    scope.bindings.set(name, jsName);
    return jsName;
  }

  private emitBindingInitialization(target: string, value: string): string {
    return target.startsWith("__state[") ? `${target} = ${value};` : `let ${target} = ${value};`;
  }

  private lookup(scope: EmitScope, name: string, span?: SourceSpan): string {
    let cursor: EmitScope | undefined = scope;
    while (cursor) {
      const found = cursor.bindings.get(name);
      if (found) return found;
      cursor = cursor.parent;
    }
    const functionInfo = this.functions.get(name);
    if (functionInfo) return functionInfo.jsName;
    if (this.options.sessionState) return `__state[${JSON.stringify(name)}]`;
    this.unsupported(`unknown identifier in async emitter: ${name}`, span);
    return "undefined";
  }

  private unsupported(message: string, span?: SourceSpan): void {
    this.diagnostics.push({
      filename: span?.filename ?? REPL_FILENAME,
      code: "GOJR_EMIT001",
      severity: "error",
      message,
      ...(span ? { span } : {})
    });
  }

  private indent(text: string, level: number): string {
    if (!text) return "";
    const prefix = "  ".repeat(level);
    return text.split("\n").map((line) => line ? `${prefix}${line}` : line).join("\n");
  }
}

export function emitAsyncJavaScript(program: ProgramAst, options: AsyncEmitOptions = {}): AsyncEmitResult {
  return new AsyncEmitter(program, options).emit();
}

function generatedFunctionName(name: string): string {
  return `_fn_${sanitizeIdentifierPart(name)}`;
}

function sanitizeIdentifierPart(name: string): string {
  const sanitized = name.replace(/[^A-Za-z0-9_$]/g, "_");
  if (/^[A-Za-z_$]/.test(sanitized)) return sanitized || "_";
  return `_${sanitized}`;
}

function receiveChannelFromStatement(statement: Statement): Expression | undefined {
  if (statement.kind === "ExpressionStatement") {
    return receiveChannelFromExpression(statement.expression);
  }
  if (statement.kind === "ShortVarStatement" && statement.values.length === 1) {
    return receiveChannelFromExpression(statement.values[0]!);
  }
  if (statement.kind === "AssignStatement" && statement.values.length === 1) {
    return receiveChannelFromExpression(statement.values[0]!);
  }
  return undefined;
}

function receiveChannelFromExpression(expression: Expression): Expression | undefined {
  if (expression.kind === "UnaryExpression" && expression.operator === "<-") {
    return expression.operand;
  }
  return undefined;
}

function isChannelType(expression: TypeExpression): boolean {
  return normalizeTypeText(expression.type.text).startsWith("chan ")
    || normalizeTypeText(expression.type.text).startsWith("<-chan ")
    || normalizeTypeText(expression.type.text).startsWith("chan<- ");
}

function channelElementType(typeText: string): string {
  const normalized = normalizeTypeText(typeText);
  if (normalized.startsWith("<-chan ")) return normalized.slice("<-chan ".length).trim();
  if (normalized.startsWith("chan<- ")) return normalized.slice("chan<- ".length).trim();
  if (normalized.startsWith("chan ")) return normalized.slice("chan ".length).trim();
  return "any";
}

function zeroValueFactoryForType(typeText: string): string {
  const normalized = normalizeTypeText(typeText);
  if (/^(?:u?int(?:8|16|32|64)?|uintptr|byte|rune)$/.test(normalized)) return `() => 0n`;
  if (/^(?:float32|float64)$/.test(normalized)) return `() => 0`;
  if (/^(?:complex64|complex128)$/.test(normalized)) return `() => ({ real: 0, imag: 0 })`;
  if (normalized === "string") return `() => ""`;
  if (normalized === "bool") return `() => false`;
  return `() => null`;
}

function normalizeTypeText(typeText: string): string {
  return typeText.replace(/\s+/g, " ").trim();
}
