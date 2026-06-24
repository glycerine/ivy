// Mechanical TypeScript transliteration of /usr/local/go1.27rc1/src/go/token.

export enum Token {
  ILLEGAL,
  EOF,
  COMMENT,
  literal_beg,
  IDENT,
  INT,
  FLOAT,
  IMAG,
  CHAR,
  STRING,
  literal_end,
  operator_beg,
  ADD,
  SUB,
  MUL,
  QUO,
  REM,
  AND,
  OR,
  XOR,
  SHL,
  SHR,
  AND_NOT,
  ADD_ASSIGN,
  SUB_ASSIGN,
  MUL_ASSIGN,
  QUO_ASSIGN,
  REM_ASSIGN,
  AND_ASSIGN,
  OR_ASSIGN,
  XOR_ASSIGN,
  SHL_ASSIGN,
  SHR_ASSIGN,
  AND_NOT_ASSIGN,
  LAND,
  LOR,
  ARROW,
  INC,
  DEC,
  EQL,
  LSS,
  GTR,
  ASSIGN,
  NOT,
  NEQ,
  LEQ,
  GEQ,
  DEFINE,
  ELLIPSIS,
  LPAREN,
  LBRACK,
  LBRACE,
  COMMA,
  PERIOD,
  RPAREN,
  RBRACK,
  RBRACE,
  SEMICOLON,
  COLON,
  operator_end,
  keyword_beg,
  BREAK,
  CASE,
  CHAN,
  CONST,
  CONTINUE,
  DEFAULT,
  DEFER,
  ELSE,
  FALLTHROUGH,
  FOR,
  FUNC,
  GO,
  GOTO,
  IF,
  IMPORT,
  INTERFACE,
  MAP,
  PACKAGE,
  RANGE,
  RETURN,
  SELECT,
  STRUCT,
  SWITCH,
  TYPE,
  VAR,
  keyword_end,
  additional_beg,
  TILDE,
  additional_end
}

const tokens: string[] = [];

tokens[Token.ILLEGAL] = "ILLEGAL";
tokens[Token.EOF] = "EOF";
tokens[Token.COMMENT] = "COMMENT";
tokens[Token.IDENT] = "IDENT";
tokens[Token.INT] = "INT";
tokens[Token.FLOAT] = "FLOAT";
tokens[Token.IMAG] = "IMAG";
tokens[Token.CHAR] = "CHAR";
tokens[Token.STRING] = "STRING";
tokens[Token.ADD] = "+";
tokens[Token.SUB] = "-";
tokens[Token.MUL] = "*";
tokens[Token.QUO] = "/";
tokens[Token.REM] = "%";
tokens[Token.AND] = "&";
tokens[Token.OR] = "|";
tokens[Token.XOR] = "^";
tokens[Token.SHL] = "<<";
tokens[Token.SHR] = ">>";
tokens[Token.AND_NOT] = "&^";
tokens[Token.ADD_ASSIGN] = "+=";
tokens[Token.SUB_ASSIGN] = "-=";
tokens[Token.MUL_ASSIGN] = "*=";
tokens[Token.QUO_ASSIGN] = "/=";
tokens[Token.REM_ASSIGN] = "%=";
tokens[Token.AND_ASSIGN] = "&=";
tokens[Token.OR_ASSIGN] = "|=";
tokens[Token.XOR_ASSIGN] = "^=";
tokens[Token.SHL_ASSIGN] = "<<=";
tokens[Token.SHR_ASSIGN] = ">>=";
tokens[Token.AND_NOT_ASSIGN] = "&^=";
tokens[Token.LAND] = "&&";
tokens[Token.LOR] = "||";
tokens[Token.ARROW] = "<-";
tokens[Token.INC] = "++";
tokens[Token.DEC] = "--";
tokens[Token.EQL] = "==";
tokens[Token.LSS] = "<";
tokens[Token.GTR] = ">";
tokens[Token.ASSIGN] = "=";
tokens[Token.NOT] = "!";
tokens[Token.NEQ] = "!=";
tokens[Token.LEQ] = "<=";
tokens[Token.GEQ] = ">=";
tokens[Token.DEFINE] = ":=";
tokens[Token.ELLIPSIS] = "...";
tokens[Token.LPAREN] = "(";
tokens[Token.LBRACK] = "[";
tokens[Token.LBRACE] = "{";
tokens[Token.COMMA] = ",";
tokens[Token.PERIOD] = ".";
tokens[Token.RPAREN] = ")";
tokens[Token.RBRACK] = "]";
tokens[Token.RBRACE] = "}";
tokens[Token.SEMICOLON] = ";";
tokens[Token.COLON] = ":";
tokens[Token.BREAK] = "break";
tokens[Token.CASE] = "case";
tokens[Token.CHAN] = "chan";
tokens[Token.CONST] = "const";
tokens[Token.CONTINUE] = "continue";
tokens[Token.DEFAULT] = "default";
tokens[Token.DEFER] = "defer";
tokens[Token.ELSE] = "else";
tokens[Token.FALLTHROUGH] = "fallthrough";
tokens[Token.FOR] = "for";
tokens[Token.FUNC] = "func";
tokens[Token.GO] = "go";
tokens[Token.GOTO] = "goto";
tokens[Token.IF] = "if";
tokens[Token.IMPORT] = "import";
tokens[Token.INTERFACE] = "interface";
tokens[Token.MAP] = "map";
tokens[Token.PACKAGE] = "package";
tokens[Token.RANGE] = "range";
tokens[Token.RETURN] = "return";
tokens[Token.SELECT] = "select";
tokens[Token.STRUCT] = "struct";
tokens[Token.SWITCH] = "switch";
tokens[Token.TYPE] = "type";
tokens[Token.VAR] = "var";
tokens[Token.TILDE] = "~";

export const LowestPrec = 0;
export const UnaryPrec = 6;
export const HighestPrec = 7;

const keywords = new Map<string, Token>();

function init(): void {
  for (let i = Token.keyword_beg + 1; i < Token.keyword_end; i += 1) {
    const text = tokens[i];
    if (text !== undefined) keywords.set(text, i);
  }
}

init();

export function TokenString(tok: Token): string {
  let s = "";
  if (0 <= tok && tok < tokens.length) {
    s = tokens[tok] ?? "";
  }
  if (s === "") {
    s = `token(${tok})`;
  }
  return s;
}

export function Precedence(op: Token): number {
  switch (op) {
    case Token.LOR:
      return 1;
    case Token.LAND:
      return 2;
    case Token.EQL:
    case Token.NEQ:
    case Token.LSS:
    case Token.LEQ:
    case Token.GTR:
    case Token.GEQ:
      return 3;
    case Token.ADD:
    case Token.SUB:
    case Token.OR:
    case Token.XOR:
      return 4;
    case Token.MUL:
    case Token.QUO:
    case Token.REM:
    case Token.SHL:
    case Token.SHR:
    case Token.AND:
    case Token.AND_NOT:
      return 5;
  }
  return LowestPrec;
}

export function Lookup(ident: string): Token {
  return keywords.get(ident) ?? Token.IDENT;
}

export function IsLiteral(tok: Token): boolean {
  return Token.literal_beg < tok && tok < Token.literal_end;
}

export function IsOperator(tok: Token): boolean {
  return (Token.operator_beg < tok && tok < Token.operator_end) || tok === Token.TILDE;
}

export function IsKeywordToken(tok: Token): boolean {
  return Token.keyword_beg < tok && tok < Token.keyword_end;
}

export function IsExported(name: string): boolean {
  const ch = Array.from(name)[0] ?? "";
  return ch.toLocaleUpperCase() === ch && ch.toLocaleLowerCase() !== ch;
}

export function IsKeyword(name: string): boolean {
  return keywords.has(name);
}

export function IsIdentifier(name: string): boolean {
  if (name === "" || IsKeyword(name)) {
    return false;
  }
  let i = 0;
  for (const c of name) {
    if (!isLetter(c) && c !== "_" && (i === 0 || !isDigit(c))) {
      return false;
    }
    i += 1;
  }
  return true;
}

function isLetter(c: string): boolean {
  return c.toLocaleLowerCase() !== c.toLocaleUpperCase();
}

function isDigit(c: string): boolean {
  return "0" <= c && c <= "9";
}

export namespace Token {
  export function String(tok: Token): string {
    return TokenString(tok);
  }

  export function Precedence(tok: Token): number {
    return globalThis.Number.isFinite(tok) ? globalPrecedence(tok) : LowestPrec;
  }

  export function IsLiteral(tok: Token): boolean {
    return Token.literal_beg < tok && tok < Token.literal_end;
  }

  export function IsOperator(tok: Token): boolean {
    return (Token.operator_beg < tok && tok < Token.operator_end) || tok === Token.TILDE;
  }

  export function IsKeyword(tok: Token): boolean {
    return Token.keyword_beg < tok && tok < Token.keyword_end;
  }
}

const globalPrecedence = Precedence;

const debug = false;

export interface PositionFields {
  Filename?: string;
  Offset?: number;
  Line?: number;
  Column?: number;
}

export class Position {
  public Filename: string;
  public Offset: number;
  public Line: number;
  public Column: number;

  public constructor(fields: PositionFields = {}) {
    this.Filename = fields.Filename ?? "";
    this.Offset = fields.Offset ?? 0;
    this.Line = fields.Line ?? 0;
    this.Column = fields.Column ?? 0;
  }

  public IsValid(): boolean {
    return this.Line > 0;
  }

  public String(): string {
    let s = this.Filename;
    if (this.IsValid()) {
      if (s !== "") s += ":";
      s += String(this.Line);
      if (this.Column !== 0) s += `:${this.Column}`;
    }
    if (s === "") s = "-";
    return s;
  }

  public toString(): string {
    return this.String();
  }
}

export type Pos = number;

export const NoPos: Pos = 0;

export function PosIsValid(p: Pos): boolean {
  return p !== NoPos;
}

export interface lineInfo {
  Offset: number;
  Filename: string;
  Line: number;
  Column: number;
}

export class File {
  public name: string;
  public base: number;
  public size: number;
  public lines: number[];
  public infos: lineInfo[];

  public constructor(name = "", base = 0, size = 0, lines: number[] = [0], infos: lineInfo[] = []) {
    this.name = name;
    this.base = base;
    this.size = size;
    this.lines = lines;
    this.infos = infos;
  }

  public String(): string {
    return `${this.Name()}(${this.Base()}-${this.End()})`;
  }

  public toString(): string {
    return this.String();
  }

  public Name(): string {
    return this.name;
  }

  public Base(): number {
    return this.base;
  }

  public Size(): number {
    return this.size;
  }

  public End(): Pos {
    return this.base + this.size;
  }

  public LineCount(): number {
    return this.lines.length;
  }

  public AddLine(offset: number): void {
    const i = this.lines.length;
    if ((i === 0 || this.lines[i - 1]! < offset) && offset < this.size) {
      this.lines.push(offset);
    }
  }

  public MergeLine(line: number): void {
    if (line < 1) {
      throw new Error(`invalid line number ${line} (should be >= 1)`);
    }
    if (line >= this.lines.length) {
      throw new Error(`invalid line number ${line} (should be < ${this.lines.length})`);
    }
    this.lines.copyWithin(line, line + 1);
    this.lines.length = this.lines.length - 1;
  }

  public Lines(): number[] {
    return this.lines;
  }

  public SetLines(lines: number[]): boolean {
    for (let i = 0; i < lines.length; i += 1) {
      const offset = lines[i]!;
      if ((i > 0 && offset <= lines[i - 1]!) || this.size <= offset) {
        return false;
      }
    }
    this.lines = lines;
    return true;
  }

  public SetLinesForContent(content: Uint8Array | string): void {
    const text = typeof content === "string" ? content : new TextDecoder().decode(content);
    const lines: number[] = [];
    let line = 0;
    for (let offset = 0; offset < text.length; offset += 1) {
      if (line >= 0) lines.push(line);
      line = -1;
      if (text[offset] === "\n") {
        line = offset + 1;
      }
    }
    this.lines = lines;
  }

  public LineStart(line: number): Pos {
    if (line < 1) {
      throw new Error(`invalid line number ${line} (should be >= 1)`);
    }
    if (line > this.lines.length) {
      throw new Error(`invalid line number ${line} (should be < ${this.lines.length})`);
    }
    return this.base + this.lines[line - 1]!;
  }

  public AddLineInfo(offset: number, filename: string, line: number): void {
    this.AddLineColumnInfo(offset, filename, line, 1);
  }

  public AddLineColumnInfo(offset: number, filename: string, line: number, column: number): void {
    const i = this.infos.length;
    if ((i === 0 || this.infos[i - 1]!.Offset < offset) && offset < this.size) {
      this.infos.push({ Offset: offset, Filename: filename, Line: line, Column: column });
    }
  }

  public fixOffset(offset: number): number {
    if (debug && !(0 <= offset && offset <= this.size)) {
      throw new Error(`offset ${offset} out of bounds [0, ${this.size}] (position ${this.base + offset} out of bounds [${this.base}, ${this.base + this.size}])`);
    }
    return Math.max(Math.min(this.size, offset), 0);
  }

  public Pos(offset: number): Pos {
    return this.base + this.fixOffset(offset);
  }

  public Offset(p: Pos): number {
    return this.fixOffset(p - this.base);
  }

  public Line(p: Pos): number {
    return this.Position(p).Line;
  }

  private unpack(offset: number, adjusted: boolean): [string, number, number] {
    let filename = this.name;
    let line = 0;
    let column = 0;
    const lineIndex = searchInts(this.lines, offset);
    if (lineIndex >= 0) {
      line = lineIndex + 1;
      column = offset - this.lines[lineIndex]! + 1;
    }
    if (adjusted && this.infos.length > 0) {
      const i = searchLineInfos(this.infos, offset);
      if (i >= 0) {
        const alt = this.infos[i]!;
        filename = alt.Filename;
        const altLineIndex = searchInts(this.lines, alt.Offset);
        if (altLineIndex >= 0) {
          const d = line - (altLineIndex + 1);
          line = alt.Line + d;
          if (alt.Column === 0) {
            column = 0;
          } else if (d === 0) {
            column = alt.Column + (offset - alt.Offset);
          }
        }
      }
    }
    return [filename, line, column];
  }

  private position(p: Pos, adjusted: boolean): Position {
    const offset = this.fixOffset(p - this.base);
    const [filename, line, column] = this.unpack(offset, adjusted);
    return new Position({ Offset: offset, Filename: filename, Line: line, Column: column });
  }

  public PositionFor(p: Pos, adjusted: boolean): Position {
    if (p !== NoPos) {
      return this.position(p, adjusted);
    }
    return new Position();
  }

  public Position(p: Pos): Position {
    return this.PositionFor(p, true);
  }

  public key(): key {
    return new key(this.base, this.base + this.size);
  }
}

export class FileSet {
  private base = 1;
  private tree = new tree();
  private last: File | undefined;

  public Base(): number {
    return this.base;
  }

  public AddFile(filename: string, base: number, size: number): File {
    const f = new File(filename, 0, size, [0]);
    if (base < 0) base = this.base;
    if (base < this.base) {
      throw new Error(`invalid base ${base} (should be >= ${this.base})`);
    }
    f.base = base;
    if (size < 0) {
      throw new Error(`invalid size ${size} (should be >= 0)`);
    }
    base += size + 1;
    if (base < 0) {
      throw new Error("token.Pos offset overflow (> 2G of source code in file set)");
    }
    this.base = base;
    this.tree.add(f);
    this.last = f;
    return f;
  }

  public AddExistingFiles(...files: File[]): void {
    for (const f of files) {
      this.tree.add(f);
      this.base = Math.max(this.base, f.Base() + f.Size() + 1);
    }
  }

  public RemoveFile(file: File): void {
    if (this.last === file) this.last = undefined;
    this.tree.deleteFile(file);
  }

  public Iterate(yieldFn: (file: File) => boolean): void {
    for (const f of this.tree.all()) {
      if (!yieldFn(f)) return;
    }
  }

  private file(p: Pos): File | undefined {
    const last = this.last;
    if (last && last.base <= p && p <= last.base + last.size) {
      return last;
    }
    const f = this.tree.fileForPos(p);
    if (f) this.last = f;
    return f;
  }

  public File(p: Pos): File | undefined {
    if (p !== NoPos) {
      return this.file(p);
    }
    return undefined;
  }

  public PositionFor(p: Pos, adjusted: boolean): Position {
    if (p !== NoPos) {
      const f = this.file(p);
      if (f) return f.PositionFor(p, adjusted);
    }
    return new Position();
  }

  public Position(p: Pos): Position {
    return this.PositionFor(p, true);
  }

  public Read(decode: (value: unknown) => void | Error): Error | undefined {
    const ss: serializedFileSet = { Base: 0, Files: [] };
    const err = decode(ss);
    if (err instanceof Error) return err;
    this.base = ss.Base;
    this.tree = new tree();
    for (const f of ss.Files) {
      this.tree.add(new File(f.Name, f.Base, f.Size, [...f.Lines], [...f.Infos]));
    }
    this.last = undefined;
    return undefined;
  }

  public Write(encode: (value: serializedFileSet) => void | Error): Error | undefined {
    const files: serializedFile[] = [];
    for (const f of this.tree.all()) {
      files.push({
        Name: f.name,
        Base: f.base,
        Size: f.size,
        Lines: [...f.lines],
        Infos: [...f.infos]
      });
    }
    const err = encode({ Base: this.base, Files: files });
    return err instanceof Error ? err : undefined;
  }
}

export function NewFileSet(): FileSet {
  return new FileSet();
}

function searchLineInfos(a: lineInfo[], x: number): number {
  let i = binarySearch(a.length, (index) => a[index]!.Offset >= x);
  if (i < a.length && a[i]!.Offset === x) {
    return i;
  }
  i -= 1;
  return i;
}

function searchInts(a: number[], x: number): number {
  let i = 0;
  let j = a.length;
  while (i < j) {
    const h = (i + j) >>> 1;
    if (a[h]! <= x) {
      i = h + 1;
    } else {
      j = h;
    }
  }
  return i - 1;
}

function binarySearch(n: number, f: (i: number) => boolean): number {
  let i = 0;
  let j = n;
  while (i < j) {
    const h = (i + j) >>> 1;
    if (!f(h)) {
      i = h + 1;
    } else {
      j = h;
    }
  }
  return i;
}

class key {
  public constructor(
    public start: number,
    public end: number
  ) {}
}

function compareKey(x: key, y: key): number {
  switch (true) {
    case x.end < y.start:
      return -1;
    case y.end < x.start:
      return +1;
  }
  return 0;
}

class node {
  public parent: node | undefined;
  public left: node | undefined;
  public right: node | undefined;
  public balance = 0;
  public height = 0;

  public constructor(
    public file: File,
    public key: key
  ) {}

  public check(_parent: node | undefined): void {
    const debugging = false;
    if (debugging) {
      this.checkBalance();
    }
  }

  public checkBalance(): void {
    const lheight = this.left?.safeHeight() ?? -1;
    const rheight = this.right?.safeHeight() ?? -1;
    const balance = rheight - lheight;
    if (balance !== this.balance) {
      throw new Error("bad node.balance");
    }
    if (!(-2 <= balance && balance <= +2)) {
      throw new Error(`node.balance out of range: ${balance}`);
    }
    const h = 1 + Math.max(lheight, rheight);
    if (h !== this.height) {
      throw new Error("bad node.height");
    }
  }

  public next(): node | undefined {
    if (!this.right) {
      let x: node | undefined = this;
      while (x.parent && x.parent.right === x) {
        x = x.parent;
      }
      return x.parent;
    }
    let x = this.right;
    while (x.left) x = x.left;
    return x;
  }

  public setLeft(y: node | undefined): void {
    this.left = y;
    if (y) y.parent = this;
  }

  public setRight(y: node | undefined): void {
    this.right = y;
    if (y) y.parent = this;
  }

  public safeHeight(): number {
    return this.height;
  }

  public update(): void {
    const lheight = this.left?.safeHeight() ?? -1;
    const rheight = this.right?.safeHeight() ?? -1;
    this.height = Math.max(lheight, rheight) + 1;
    this.balance = rheight - lheight;
  }
}

class tree {
  private files: File[] = [];

  public locate(k: key): [number, File | undefined] {
    for (let i = 0; i < this.files.length; i += 1) {
      const f = this.files[i]!;
      const sign = compareKey(k, f.key());
      if (sign === 0) return [i, f];
      if (sign < 0) return [i, undefined];
    }
    return [this.files.length, undefined];
  }

  public *all(): IterableIterator<File> {
    for (const f of [...this.files].sort((a, b) => a.Base() - b.Base())) {
      yield f;
    }
  }

  public add(file: File): void {
    const [index, prev] = this.locate(file.key());
    if (!prev) {
      this.files.splice(index, 0, file);
      return;
    }
    if (prev !== file) {
      throw new Error(`file ${prev.Name()} (${prev.Base()}-${prev.End()}) overlaps with file ${file.Name()} (${file.Base()}-${file.End()})`);
    }
  }

  public set(file: File, _pos?: unknown, _parent?: unknown): void {
    const [index, prev] = this.locate(file.key());
    if (prev) {
      throw new Error("unreachable according to current FileSet requirements");
    }
    this.files.splice(index, 0, file);
  }

  public deleteFile(file: File): void {
    this.files = this.files.filter((f) => f !== file);
  }

  public fileForPos(p: Pos): File | undefined {
    const [, file] = this.locate(new key(p, p));
    return file;
  }

  public nextAfter(_pos: unknown, _parent: unknown): node | undefined {
    return undefined;
  }

  public setRoot(_x: node | undefined): void {}
  public replaceChild(_parent: node | undefined, _old: node, _new: node | undefined): void {}
  public rebalanceUp(_x: node | undefined): void {}
  public rotateRight(y: node): node { return y; }
  public rotateLeft(x: node): node { return x; }
  public delete(_pos: unknown): void {}
  public deleteSwap(_pos: unknown): void {}
  public deleteMin(_zpos: unknown): node | undefined { return undefined; }
}

export interface serializedFile {
  Name: string;
  Base: number;
  Size: number;
  Lines: number[];
  Infos: lineInfo[];
}

export interface serializedFileSet {
  Base: number;
  Files: serializedFile[];
}

