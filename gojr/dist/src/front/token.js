export var TokenKind;
(function (TokenKind) {
    TokenKind["Illegal"] = "Illegal";
    TokenKind["EOF"] = "EOF";
    TokenKind["Identifier"] = "Identifier";
    TokenKind["CellAddress"] = "CellAddress";
    TokenKind["IntLiteral"] = "IntLiteral";
    TokenKind["FloatLiteral"] = "FloatLiteral";
    TokenKind["ImagLiteral"] = "ImagLiteral";
    TokenKind["RuneLiteral"] = "RuneLiteral";
    TokenKind["StringLiteral"] = "StringLiteral";
    TokenKind["Import"] = "Import";
    TokenKind["Package"] = "Package";
    TokenKind["Func"] = "Func";
    TokenKind["Return"] = "Return";
    TokenKind["If"] = "If";
    TokenKind["Else"] = "Else";
    TokenKind["Switch"] = "Switch";
    TokenKind["Case"] = "Case";
    TokenKind["Default"] = "Default";
    TokenKind["Fallthrough"] = "Fallthrough";
    TokenKind["Goto"] = "Goto";
    TokenKind["For"] = "For";
    TokenKind["Range"] = "Range";
    TokenKind["Break"] = "Break";
    TokenKind["Continue"] = "Continue";
    TokenKind["Defer"] = "Defer";
    TokenKind["Var"] = "Var";
    TokenKind["Const"] = "Const";
    TokenKind["Type"] = "Type";
    TokenKind["Struct"] = "Struct";
    TokenKind["Interface"] = "Interface";
    TokenKind["Map"] = "Map";
    TokenKind["Chan"] = "Chan";
    TokenKind["Go"] = "Go";
    TokenKind["Select"] = "Select";
    TokenKind["True"] = "True";
    TokenKind["False"] = "False";
    TokenKind["Nil"] = "Nil";
    TokenKind["Ellipsis"] = "Ellipsis";
    TokenKind["Define"] = "Define";
    TokenKind["Assign"] = "Assign";
    TokenKind["PlusAssign"] = "PlusAssign";
    TokenKind["MinusAssign"] = "MinusAssign";
    TokenKind["StarAssign"] = "StarAssign";
    TokenKind["SlashAssign"] = "SlashAssign";
    TokenKind["PercentAssign"] = "PercentAssign";
    TokenKind["AmpAssign"] = "AmpAssign";
    TokenKind["OrAssign"] = "OrAssign";
    TokenKind["CaretAssign"] = "CaretAssign";
    TokenKind["ShlAssign"] = "ShlAssign";
    TokenKind["ShrAssign"] = "ShrAssign";
    TokenKind["BitClearAssign"] = "BitClearAssign";
    TokenKind["Equal"] = "Equal";
    TokenKind["NotEqual"] = "NotEqual";
    TokenKind["Less"] = "Less";
    TokenKind["LessEqual"] = "LessEqual";
    TokenKind["Greater"] = "Greater";
    TokenKind["GreaterEqual"] = "GreaterEqual";
    TokenKind["AndAnd"] = "AndAnd";
    TokenKind["OrOr"] = "OrOr";
    TokenKind["Arrow"] = "Arrow";
    TokenKind["Plus"] = "Plus";
    TokenKind["PlusPlus"] = "PlusPlus";
    TokenKind["Minus"] = "Minus";
    TokenKind["MinusMinus"] = "MinusMinus";
    TokenKind["Star"] = "Star";
    TokenKind["Slash"] = "Slash";
    TokenKind["Percent"] = "Percent";
    TokenKind["Or"] = "Or";
    TokenKind["Caret"] = "Caret";
    TokenKind["Shl"] = "Shl";
    TokenKind["Shr"] = "Shr";
    TokenKind["BitClear"] = "BitClear";
    TokenKind["Bang"] = "Bang";
    TokenKind["Amp"] = "Amp";
    TokenKind["Dot"] = "Dot";
    TokenKind["Comma"] = "Comma";
    TokenKind["Colon"] = "Colon";
    TokenKind["Semicolon"] = "Semicolon";
    TokenKind["LParen"] = "LParen";
    TokenKind["RParen"] = "RParen";
    TokenKind["LBrace"] = "LBrace";
    TokenKind["RBrace"] = "RBrace";
    TokenKind["LBracket"] = "LBracket";
    TokenKind["RBracket"] = "RBracket";
})(TokenKind || (TokenKind = {}));
const keywords = new Map([
    ["import", TokenKind.Import],
    ["package", TokenKind.Package],
    ["func", TokenKind.Func],
    ["return", TokenKind.Return],
    ["if", TokenKind.If],
    ["else", TokenKind.Else],
    ["switch", TokenKind.Switch],
    ["case", TokenKind.Case],
    ["default", TokenKind.Default],
    ["fallthrough", TokenKind.Fallthrough],
    ["goto", TokenKind.Goto],
    ["for", TokenKind.For],
    ["range", TokenKind.Range],
    ["break", TokenKind.Break],
    ["continue", TokenKind.Continue],
    ["defer", TokenKind.Defer],
    ["var", TokenKind.Var],
    ["const", TokenKind.Const],
    ["type", TokenKind.Type],
    ["struct", TokenKind.Struct],
    ["interface", TokenKind.Interface],
    ["map", TokenKind.Map],
    ["chan", TokenKind.Chan],
    ["go", TokenKind.Go],
    ["select", TokenKind.Select],
    ["true", TokenKind.True],
    ["false", TokenKind.False],
    ["nil", TokenKind.Nil]
]);
export function keywordKind(text) {
    return keywords.get(text);
}
export function isIdentifierLike(kind) {
    return kind === TokenKind.Identifier || kind === TokenKind.CellAddress;
}
export function tokenCanEndStatement(kind) {
    return kind === TokenKind.Identifier ||
        kind === TokenKind.CellAddress ||
        kind === TokenKind.IntLiteral ||
        kind === TokenKind.FloatLiteral ||
        kind === TokenKind.ImagLiteral ||
        kind === TokenKind.RuneLiteral ||
        kind === TokenKind.StringLiteral ||
        kind === TokenKind.True ||
        kind === TokenKind.False ||
        kind === TokenKind.Nil ||
        kind === TokenKind.Return ||
        kind === TokenKind.Break ||
        kind === TokenKind.Continue ||
        kind === TokenKind.Fallthrough ||
        kind === TokenKind.PlusPlus ||
        kind === TokenKind.MinusMinus ||
        kind === TokenKind.RParen ||
        kind === TokenKind.RBracket ||
        kind === TokenKind.RBrace;
}
export function isAssignmentToken(kind) {
    return kind === TokenKind.Assign ||
        kind === TokenKind.Define ||
        kind === TokenKind.PlusAssign ||
        kind === TokenKind.MinusAssign ||
        kind === TokenKind.StarAssign ||
        kind === TokenKind.SlashAssign ||
        kind === TokenKind.PercentAssign ||
        kind === TokenKind.AmpAssign ||
        kind === TokenKind.OrAssign ||
        kind === TokenKind.CaretAssign ||
        kind === TokenKind.ShlAssign ||
        kind === TokenKind.ShrAssign ||
        kind === TokenKind.BitClearAssign;
}
