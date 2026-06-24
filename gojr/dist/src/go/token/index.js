// Mechanical TypeScript transliteration of /usr/local/go1.27rc1/src/go/token.
export var Token;
(function (Token) {
    Token[Token["ILLEGAL"] = 0] = "ILLEGAL";
    Token[Token["EOF"] = 1] = "EOF";
    Token[Token["COMMENT"] = 2] = "COMMENT";
    Token[Token["literal_beg"] = 3] = "literal_beg";
    Token[Token["IDENT"] = 4] = "IDENT";
    Token[Token["INT"] = 5] = "INT";
    Token[Token["FLOAT"] = 6] = "FLOAT";
    Token[Token["IMAG"] = 7] = "IMAG";
    Token[Token["CHAR"] = 8] = "CHAR";
    Token[Token["STRING"] = 9] = "STRING";
    Token[Token["literal_end"] = 10] = "literal_end";
    Token[Token["operator_beg"] = 11] = "operator_beg";
    Token[Token["ADD"] = 12] = "ADD";
    Token[Token["SUB"] = 13] = "SUB";
    Token[Token["MUL"] = 14] = "MUL";
    Token[Token["QUO"] = 15] = "QUO";
    Token[Token["REM"] = 16] = "REM";
    Token[Token["AND"] = 17] = "AND";
    Token[Token["OR"] = 18] = "OR";
    Token[Token["XOR"] = 19] = "XOR";
    Token[Token["SHL"] = 20] = "SHL";
    Token[Token["SHR"] = 21] = "SHR";
    Token[Token["AND_NOT"] = 22] = "AND_NOT";
    Token[Token["ADD_ASSIGN"] = 23] = "ADD_ASSIGN";
    Token[Token["SUB_ASSIGN"] = 24] = "SUB_ASSIGN";
    Token[Token["MUL_ASSIGN"] = 25] = "MUL_ASSIGN";
    Token[Token["QUO_ASSIGN"] = 26] = "QUO_ASSIGN";
    Token[Token["REM_ASSIGN"] = 27] = "REM_ASSIGN";
    Token[Token["AND_ASSIGN"] = 28] = "AND_ASSIGN";
    Token[Token["OR_ASSIGN"] = 29] = "OR_ASSIGN";
    Token[Token["XOR_ASSIGN"] = 30] = "XOR_ASSIGN";
    Token[Token["SHL_ASSIGN"] = 31] = "SHL_ASSIGN";
    Token[Token["SHR_ASSIGN"] = 32] = "SHR_ASSIGN";
    Token[Token["AND_NOT_ASSIGN"] = 33] = "AND_NOT_ASSIGN";
    Token[Token["LAND"] = 34] = "LAND";
    Token[Token["LOR"] = 35] = "LOR";
    Token[Token["ARROW"] = 36] = "ARROW";
    Token[Token["INC"] = 37] = "INC";
    Token[Token["DEC"] = 38] = "DEC";
    Token[Token["EQL"] = 39] = "EQL";
    Token[Token["LSS"] = 40] = "LSS";
    Token[Token["GTR"] = 41] = "GTR";
    Token[Token["ASSIGN"] = 42] = "ASSIGN";
    Token[Token["NOT"] = 43] = "NOT";
    Token[Token["NEQ"] = 44] = "NEQ";
    Token[Token["LEQ"] = 45] = "LEQ";
    Token[Token["GEQ"] = 46] = "GEQ";
    Token[Token["DEFINE"] = 47] = "DEFINE";
    Token[Token["ELLIPSIS"] = 48] = "ELLIPSIS";
    Token[Token["LPAREN"] = 49] = "LPAREN";
    Token[Token["LBRACK"] = 50] = "LBRACK";
    Token[Token["LBRACE"] = 51] = "LBRACE";
    Token[Token["COMMA"] = 52] = "COMMA";
    Token[Token["PERIOD"] = 53] = "PERIOD";
    Token[Token["RPAREN"] = 54] = "RPAREN";
    Token[Token["RBRACK"] = 55] = "RBRACK";
    Token[Token["RBRACE"] = 56] = "RBRACE";
    Token[Token["SEMICOLON"] = 57] = "SEMICOLON";
    Token[Token["COLON"] = 58] = "COLON";
    Token[Token["operator_end"] = 59] = "operator_end";
    Token[Token["keyword_beg"] = 60] = "keyword_beg";
    Token[Token["BREAK"] = 61] = "BREAK";
    Token[Token["CASE"] = 62] = "CASE";
    Token[Token["CHAN"] = 63] = "CHAN";
    Token[Token["CONST"] = 64] = "CONST";
    Token[Token["CONTINUE"] = 65] = "CONTINUE";
    Token[Token["DEFAULT"] = 66] = "DEFAULT";
    Token[Token["DEFER"] = 67] = "DEFER";
    Token[Token["ELSE"] = 68] = "ELSE";
    Token[Token["FALLTHROUGH"] = 69] = "FALLTHROUGH";
    Token[Token["FOR"] = 70] = "FOR";
    Token[Token["FUNC"] = 71] = "FUNC";
    Token[Token["GO"] = 72] = "GO";
    Token[Token["GOTO"] = 73] = "GOTO";
    Token[Token["IF"] = 74] = "IF";
    Token[Token["IMPORT"] = 75] = "IMPORT";
    Token[Token["INTERFACE"] = 76] = "INTERFACE";
    Token[Token["MAP"] = 77] = "MAP";
    Token[Token["PACKAGE"] = 78] = "PACKAGE";
    Token[Token["RANGE"] = 79] = "RANGE";
    Token[Token["RETURN"] = 80] = "RETURN";
    Token[Token["SELECT"] = 81] = "SELECT";
    Token[Token["STRUCT"] = 82] = "STRUCT";
    Token[Token["SWITCH"] = 83] = "SWITCH";
    Token[Token["TYPE"] = 84] = "TYPE";
    Token[Token["VAR"] = 85] = "VAR";
    Token[Token["keyword_end"] = 86] = "keyword_end";
    Token[Token["additional_beg"] = 87] = "additional_beg";
    Token[Token["TILDE"] = 88] = "TILDE";
    Token[Token["additional_end"] = 89] = "additional_end";
})(Token || (Token = {}));
const tokens = [];
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
const keywords = new Map();
function init() {
    for (let i = Token.keyword_beg + 1; i < Token.keyword_end; i += 1) {
        const text = tokens[i];
        if (text !== undefined)
            keywords.set(text, i);
    }
}
init();
export function TokenString(tok) {
    let s = "";
    if (0 <= tok && tok < tokens.length) {
        s = tokens[tok] ?? "";
    }
    if (s === "") {
        s = `token(${tok})`;
    }
    return s;
}
export function Precedence(op) {
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
export function Lookup(ident) {
    return keywords.get(ident) ?? Token.IDENT;
}
export function IsLiteral(tok) {
    return Token.literal_beg < tok && tok < Token.literal_end;
}
export function IsOperator(tok) {
    return (Token.operator_beg < tok && tok < Token.operator_end) || tok === Token.TILDE;
}
export function IsKeywordToken(tok) {
    return Token.keyword_beg < tok && tok < Token.keyword_end;
}
export function IsExported(name) {
    const ch = Array.from(name)[0] ?? "";
    return ch.toLocaleUpperCase() === ch && ch.toLocaleLowerCase() !== ch;
}
export function IsKeyword(name) {
    return keywords.has(name);
}
export function IsIdentifier(name) {
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
function isLetter(c) {
    return c.toLocaleLowerCase() !== c.toLocaleUpperCase();
}
function isDigit(c) {
    return "0" <= c && c <= "9";
}
(function (Token) {
    function String(tok) {
        return TokenString(tok);
    }
    Token.String = String;
    function Precedence(tok) {
        return globalThis.Number.isFinite(tok) ? globalPrecedence(tok) : LowestPrec;
    }
    Token.Precedence = Precedence;
    function IsLiteral(tok) {
        return Token.literal_beg < tok && tok < Token.literal_end;
    }
    Token.IsLiteral = IsLiteral;
    function IsOperator(tok) {
        return (Token.operator_beg < tok && tok < Token.operator_end) || tok === Token.TILDE;
    }
    Token.IsOperator = IsOperator;
    function IsKeyword(tok) {
        return Token.keyword_beg < tok && tok < Token.keyword_end;
    }
    Token.IsKeyword = IsKeyword;
})(Token || (Token = {}));
const globalPrecedence = Precedence;
const debug = false;
export class Position {
    Filename;
    Offset;
    Line;
    Column;
    constructor(fields = {}) {
        this.Filename = fields.Filename ?? "";
        this.Offset = fields.Offset ?? 0;
        this.Line = fields.Line ?? 0;
        this.Column = fields.Column ?? 0;
    }
    IsValid() {
        return this.Line > 0;
    }
    String() {
        let s = this.Filename;
        if (this.IsValid()) {
            if (s !== "")
                s += ":";
            s += String(this.Line);
            if (this.Column !== 0)
                s += `:${this.Column}`;
        }
        if (s === "")
            s = "-";
        return s;
    }
    toString() {
        return this.String();
    }
}
export const NoPos = 0;
export function PosIsValid(p) {
    return p !== NoPos;
}
export class File {
    name;
    base;
    size;
    lines;
    infos;
    constructor(name = "", base = 0, size = 0, lines = [0], infos = []) {
        this.name = name;
        this.base = base;
        this.size = size;
        this.lines = lines;
        this.infos = infos;
    }
    String() {
        return `${this.Name()}(${this.Base()}-${this.End()})`;
    }
    toString() {
        return this.String();
    }
    Name() {
        return this.name;
    }
    Base() {
        return this.base;
    }
    Size() {
        return this.size;
    }
    End() {
        return this.base + this.size;
    }
    LineCount() {
        return this.lines.length;
    }
    AddLine(offset) {
        const i = this.lines.length;
        if ((i === 0 || this.lines[i - 1] < offset) && offset < this.size) {
            this.lines.push(offset);
        }
    }
    MergeLine(line) {
        if (line < 1) {
            throw new Error(`invalid line number ${line} (should be >= 1)`);
        }
        if (line >= this.lines.length) {
            throw new Error(`invalid line number ${line} (should be < ${this.lines.length})`);
        }
        this.lines.copyWithin(line, line + 1);
        this.lines.length = this.lines.length - 1;
    }
    Lines() {
        return this.lines;
    }
    SetLines(lines) {
        for (let i = 0; i < lines.length; i += 1) {
            const offset = lines[i];
            if ((i > 0 && offset <= lines[i - 1]) || this.size <= offset) {
                return false;
            }
        }
        this.lines = lines;
        return true;
    }
    SetLinesForContent(content) {
        const text = typeof content === "string" ? content : new TextDecoder().decode(content);
        const lines = [];
        let line = 0;
        for (let offset = 0; offset < text.length; offset += 1) {
            if (line >= 0)
                lines.push(line);
            line = -1;
            if (text[offset] === "\n") {
                line = offset + 1;
            }
        }
        this.lines = lines;
    }
    LineStart(line) {
        if (line < 1) {
            throw new Error(`invalid line number ${line} (should be >= 1)`);
        }
        if (line > this.lines.length) {
            throw new Error(`invalid line number ${line} (should be < ${this.lines.length})`);
        }
        return this.base + this.lines[line - 1];
    }
    AddLineInfo(offset, filename, line) {
        this.AddLineColumnInfo(offset, filename, line, 1);
    }
    AddLineColumnInfo(offset, filename, line, column) {
        const i = this.infos.length;
        if ((i === 0 || this.infos[i - 1].Offset < offset) && offset < this.size) {
            this.infos.push({ Offset: offset, Filename: filename, Line: line, Column: column });
        }
    }
    fixOffset(offset) {
        if (debug && !(0 <= offset && offset <= this.size)) {
            throw new Error(`offset ${offset} out of bounds [0, ${this.size}] (position ${this.base + offset} out of bounds [${this.base}, ${this.base + this.size}])`);
        }
        return Math.max(Math.min(this.size, offset), 0);
    }
    Pos(offset) {
        return this.base + this.fixOffset(offset);
    }
    Offset(p) {
        return this.fixOffset(p - this.base);
    }
    Line(p) {
        return this.Position(p).Line;
    }
    unpack(offset, adjusted) {
        let filename = this.name;
        let line = 0;
        let column = 0;
        const lineIndex = searchInts(this.lines, offset);
        if (lineIndex >= 0) {
            line = lineIndex + 1;
            column = offset - this.lines[lineIndex] + 1;
        }
        if (adjusted && this.infos.length > 0) {
            const i = searchLineInfos(this.infos, offset);
            if (i >= 0) {
                const alt = this.infos[i];
                filename = alt.Filename;
                const altLineIndex = searchInts(this.lines, alt.Offset);
                if (altLineIndex >= 0) {
                    const d = line - (altLineIndex + 1);
                    line = alt.Line + d;
                    if (alt.Column === 0) {
                        column = 0;
                    }
                    else if (d === 0) {
                        column = alt.Column + (offset - alt.Offset);
                    }
                }
            }
        }
        return [filename, line, column];
    }
    position(p, adjusted) {
        const offset = this.fixOffset(p - this.base);
        const [filename, line, column] = this.unpack(offset, adjusted);
        return new Position({ Offset: offset, Filename: filename, Line: line, Column: column });
    }
    PositionFor(p, adjusted) {
        if (p !== NoPos) {
            return this.position(p, adjusted);
        }
        return new Position();
    }
    Position(p) {
        return this.PositionFor(p, true);
    }
    key() {
        return new key(this.base, this.base + this.size);
    }
}
export class FileSet {
    base = 1;
    tree = new tree();
    last;
    Base() {
        return this.base;
    }
    AddFile(filename, base, size) {
        const f = new File(filename, 0, size, [0]);
        if (base < 0)
            base = this.base;
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
    AddExistingFiles(...files) {
        for (const f of files) {
            this.tree.add(f);
            this.base = Math.max(this.base, f.Base() + f.Size() + 1);
        }
    }
    RemoveFile(file) {
        if (this.last === file)
            this.last = undefined;
        this.tree.deleteFile(file);
    }
    Iterate(yieldFn) {
        for (const f of this.tree.all()) {
            if (!yieldFn(f))
                return;
        }
    }
    file(p) {
        const last = this.last;
        if (last && last.base <= p && p <= last.base + last.size) {
            return last;
        }
        const f = this.tree.fileForPos(p);
        if (f)
            this.last = f;
        return f;
    }
    File(p) {
        if (p !== NoPos) {
            return this.file(p);
        }
        return undefined;
    }
    PositionFor(p, adjusted) {
        if (p !== NoPos) {
            const f = this.file(p);
            if (f)
                return f.PositionFor(p, adjusted);
        }
        return new Position();
    }
    Position(p) {
        return this.PositionFor(p, true);
    }
    Read(decode) {
        const ss = { Base: 0, Files: [] };
        const err = decode(ss);
        if (err instanceof Error)
            return err;
        this.base = ss.Base;
        this.tree = new tree();
        for (const f of ss.Files) {
            this.tree.add(new File(f.Name, f.Base, f.Size, [...f.Lines], [...f.Infos]));
        }
        this.last = undefined;
        return undefined;
    }
    Write(encode) {
        const files = [];
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
export function NewFileSet() {
    return new FileSet();
}
function searchLineInfos(a, x) {
    let i = binarySearch(a.length, (index) => a[index].Offset >= x);
    if (i < a.length && a[i].Offset === x) {
        return i;
    }
    i -= 1;
    return i;
}
function searchInts(a, x) {
    let i = 0;
    let j = a.length;
    while (i < j) {
        const h = (i + j) >>> 1;
        if (a[h] <= x) {
            i = h + 1;
        }
        else {
            j = h;
        }
    }
    return i - 1;
}
function binarySearch(n, f) {
    let i = 0;
    let j = n;
    while (i < j) {
        const h = (i + j) >>> 1;
        if (!f(h)) {
            i = h + 1;
        }
        else {
            j = h;
        }
    }
    return i;
}
class key {
    start;
    end;
    constructor(start, end) {
        this.start = start;
        this.end = end;
    }
}
function compareKey(x, y) {
    switch (true) {
        case x.end < y.start:
            return -1;
        case y.end < x.start:
            return +1;
    }
    return 0;
}
class node {
    file;
    key;
    parent;
    left;
    right;
    balance = 0;
    height = 0;
    constructor(file, key) {
        this.file = file;
        this.key = key;
    }
    check(_parent) {
        const debugging = false;
        if (debugging) {
            this.checkBalance();
        }
    }
    checkBalance() {
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
    next() {
        if (!this.right) {
            let x = this;
            while (x.parent && x.parent.right === x) {
                x = x.parent;
            }
            return x.parent;
        }
        let x = this.right;
        while (x.left)
            x = x.left;
        return x;
    }
    setLeft(y) {
        this.left = y;
        if (y)
            y.parent = this;
    }
    setRight(y) {
        this.right = y;
        if (y)
            y.parent = this;
    }
    safeHeight() {
        return this.height;
    }
    update() {
        const lheight = this.left?.safeHeight() ?? -1;
        const rheight = this.right?.safeHeight() ?? -1;
        this.height = Math.max(lheight, rheight) + 1;
        this.balance = rheight - lheight;
    }
}
class tree {
    files = [];
    locate(k) {
        for (let i = 0; i < this.files.length; i += 1) {
            const f = this.files[i];
            const sign = compareKey(k, f.key());
            if (sign === 0)
                return [i, f];
            if (sign < 0)
                return [i, undefined];
        }
        return [this.files.length, undefined];
    }
    *all() {
        for (const f of [...this.files].sort((a, b) => a.Base() - b.Base())) {
            yield f;
        }
    }
    add(file) {
        const [index, prev] = this.locate(file.key());
        if (!prev) {
            this.files.splice(index, 0, file);
            return;
        }
        if (prev !== file) {
            throw new Error(`file ${prev.Name()} (${prev.Base()}-${prev.End()}) overlaps with file ${file.Name()} (${file.Base()}-${file.End()})`);
        }
    }
    set(file, _pos, _parent) {
        const [index, prev] = this.locate(file.key());
        if (prev) {
            throw new Error("unreachable according to current FileSet requirements");
        }
        this.files.splice(index, 0, file);
    }
    deleteFile(file) {
        this.files = this.files.filter((f) => f !== file);
    }
    fileForPos(p) {
        const [, file] = this.locate(new key(p, p));
        return file;
    }
    nextAfter(_pos, _parent) {
        return undefined;
    }
    setRoot(_x) { }
    replaceChild(_parent, _old, _new) { }
    rebalanceUp(_x) { }
    rotateRight(y) { return y; }
    rotateLeft(x) { return x; }
    delete(_pos) { }
    deleteSwap(_pos) { }
    deleteMin(_zpos) { return undefined; }
}
