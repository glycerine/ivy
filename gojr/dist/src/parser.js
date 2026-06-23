import { CstParser } from "chevrotain";
import { spanFromToken } from "./diagnostics.js";
import { allTokens, Amp, AndAnd, Assign, Bang, Break, Case, CellAddress, Colon, Comma, Const, Continue, Default, Defer, Define, Dot, Ellipsis, Else, Equal, Fallthrough, False, FloatLiteral, For, Func, Goto, Greater, GreaterEqual, Identifier, If, Import, IntLiteral, Interface, LBrace, LBracket, Less, LessEqual, LParen, MapTok, Minus, MinusMinus, Nil, NotEqual, OrOr, Percent, Plus, PlusPlus, Range, Return, RBrace, RBracket, RParen, Semicolon, Slash, Star, StringLiteral, Struct, Switch, True, Type, Var, goJuniorLexer } from "./tokens.js";
export class GoJuniorParser extends CstParser {
    program;
    importDecl;
    importSpec;
    functionDecl;
    receiver;
    signature;
    parameterList;
    parameter;
    result;
    typeExpression;
    interfaceType;
    structType;
    fieldDecl;
    block;
    statement;
    labeledStmt;
    constDecl;
    varDecl;
    typeDecl;
    returnStmt;
    ifStmt;
    switchStmt;
    switchClause;
    forStmt;
    forClause;
    forInitClause;
    forPostClause;
    rangeClause;
    deferStmt;
    branchStmt;
    declarationSpec;
    simpleStmt;
    expressionStatement;
    expression;
    orExpr;
    andExpr;
    equalityExpr;
    compareExpr;
    addExpr;
    mulExpr;
    unaryExpr;
    primaryExpr;
    atom;
    mapLiteral;
    mapElement;
    functionLiteral;
    arguments;
    expressionList;
    qualifiedName;
    selectorName;
    name;
    literal;
    constructor() {
        super(allTokens, {
            recoveryEnabled: true,
            maxLookahead: 4
        });
        const $ = this;
        this.program = $.RULE("program", () => {
            $.MANY(() => {
                $.SUBRULE(this.importDecl);
                $.OPTION(() => $.CONSUME(Semicolon));
            });
            $.OPTION2(() => {
                $.OR([
                    { GATE: () => this.nextTokensStartFunctionDecl(), ALT: () => $.SUBRULE(this.functionDecl) },
                    {
                        ALT: () => {
                            $.MANY1(() => {
                                $.SUBRULE(this.statement);
                                $.OPTION3(() => $.CONSUME2(Semicolon));
                            });
                        }
                    }
                ]);
            });
        });
        this.importDecl = $.RULE("importDecl", () => {
            $.CONSUME(Import);
            $.OR([
                { ALT: () => $.SUBRULE(this.importSpec) },
                {
                    ALT: () => {
                        $.CONSUME(LParen);
                        $.MANY(() => {
                            $.SUBRULE2(this.importSpec);
                            $.OPTION(() => $.CONSUME(Semicolon));
                        });
                        $.CONSUME(RParen);
                    }
                }
            ]);
        });
        this.importSpec = $.RULE("importSpec", () => {
            $.OPTION(() => $.SUBRULE(this.name));
            $.CONSUME(StringLiteral);
        });
        this.functionDecl = $.RULE("functionDecl", () => {
            $.CONSUME(Func);
            $.OPTION(() => $.SUBRULE(this.receiver));
            $.CONSUME(Identifier);
            $.SUBRULE(this.signature);
            $.SUBRULE(this.block);
        });
        this.receiver = $.RULE("receiver", () => {
            $.CONSUME(LParen);
            $.OPTION(() => $.SUBRULE(this.name));
            $.SUBRULE(this.typeExpression);
            $.CONSUME(RParen);
        });
        this.signature = $.RULE("signature", () => {
            $.SUBRULE(this.parameterList);
            $.OPTION(() => $.SUBRULE(this.result));
        });
        this.parameterList = $.RULE("parameterList", () => {
            $.CONSUME(LParen);
            $.OPTION(() => {
                $.SUBRULE(this.parameter);
                $.MANY(() => {
                    $.CONSUME(Comma);
                    $.SUBRULE2(this.parameter);
                });
                $.OPTION2(() => $.CONSUME2(Comma));
            });
            $.CONSUME(RParen);
        });
        this.parameter = $.RULE("parameter", () => {
            $.OR([
                {
                    GATE: () => this.nextTokensAreNamedParameter(),
                    ALT: () => {
                        $.SUBRULE(this.name);
                        $.MANY(() => {
                            $.CONSUME(Comma);
                            $.SUBRULE2(this.name);
                        });
                        $.OPTION(() => $.CONSUME(Ellipsis));
                        $.SUBRULE(this.typeExpression);
                    }
                },
                {
                    ALT: () => {
                        $.OPTION2(() => $.CONSUME2(Ellipsis));
                        $.SUBRULE2(this.typeExpression);
                    }
                }
            ]);
        });
        this.result = $.RULE("result", () => {
            $.OR([
                { ALT: () => $.SUBRULE(this.typeExpression) },
                { ALT: () => $.SUBRULE(this.parameterList) }
            ]);
        });
        this.typeExpression = $.RULE("typeExpression", () => {
            $.OR([
                {
                    ALT: () => {
                        $.CONSUME(Star);
                        $.SUBRULE(this.typeExpression);
                    }
                },
                {
                    ALT: () => {
                        $.CONSUME(LBracket);
                        $.OPTION(() => $.CONSUME(IntLiteral));
                        $.CONSUME(RBracket);
                        $.SUBRULE2(this.typeExpression);
                    }
                },
                {
                    ALT: () => {
                        $.CONSUME(MapTok);
                        $.CONSUME2(LBracket);
                        $.SUBRULE3(this.typeExpression);
                        $.CONSUME2(RBracket);
                        $.SUBRULE4(this.typeExpression);
                    }
                },
                { ALT: () => $.SUBRULE(this.structType) },
                { ALT: () => $.SUBRULE(this.interfaceType) },
                {
                    ALT: () => {
                        $.CONSUME(Func);
                        $.SUBRULE(this.signature);
                    }
                },
                { ALT: () => $.SUBRULE(this.qualifiedName) }
            ]);
        });
        this.structType = $.RULE("structType", () => {
            $.CONSUME(Struct);
            $.CONSUME(LBrace);
            $.MANY(() => $.SUBRULE(this.fieldDecl));
            $.CONSUME(RBrace);
        });
        this.fieldDecl = $.RULE("fieldDecl", () => {
            $.SUBRULE(this.name);
            $.SUBRULE(this.typeExpression);
            $.OPTION(() => $.CONSUME(Semicolon));
        });
        this.interfaceType = $.RULE("interfaceType", () => {
            $.CONSUME(Interface);
            $.CONSUME(LBrace);
            $.MANY(() => {
                $.SUBRULE(this.name);
                $.SUBRULE(this.signature);
                $.OPTION(() => $.CONSUME(Semicolon));
            });
            $.CONSUME(RBrace);
        });
        this.block = $.RULE("block", () => {
            $.CONSUME(LBrace);
            $.MANY(() => {
                $.SUBRULE(this.statement);
                $.OPTION(() => $.CONSUME(Semicolon));
            });
            $.CONSUME(RBrace);
        });
        this.statement = $.RULE("statement", () => {
            $.OR([
                { GATE: () => this.nextTokensAreLabel(), ALT: () => $.SUBRULE(this.labeledStmt) },
                { ALT: () => $.SUBRULE(this.constDecl) },
                { ALT: () => $.SUBRULE(this.varDecl) },
                { ALT: () => $.SUBRULE(this.typeDecl) },
                { ALT: () => $.SUBRULE(this.returnStmt) },
                { ALT: () => $.SUBRULE(this.ifStmt) },
                { ALT: () => $.SUBRULE(this.switchStmt) },
                { ALT: () => $.SUBRULE(this.forStmt) },
                { ALT: () => $.SUBRULE(this.deferStmt) },
                { ALT: () => $.SUBRULE(this.branchStmt) },
                { ALT: () => $.SUBRULE(this.simpleStmt) }
            ]);
        });
        this.labeledStmt = $.RULE("labeledStmt", () => {
            $.CONSUME(Identifier);
            $.CONSUME(Colon);
            $.OPTION(() => $.SUBRULE(this.statement));
        });
        this.constDecl = $.RULE("constDecl", () => {
            $.CONSUME(Const);
            $.OR([
                { ALT: () => $.SUBRULE(this.declarationSpec) },
                {
                    ALT: () => {
                        $.CONSUME(LParen);
                        $.MANY(() => {
                            $.SUBRULE2(this.declarationSpec);
                            $.OPTION(() => $.CONSUME(Semicolon));
                        });
                        $.CONSUME(RParen);
                    }
                }
            ]);
        });
        this.varDecl = $.RULE("varDecl", () => {
            $.CONSUME(Var);
            $.OR([
                { ALT: () => $.SUBRULE(this.declarationSpec) },
                {
                    ALT: () => {
                        $.CONSUME(LParen);
                        $.MANY(() => {
                            $.SUBRULE2(this.declarationSpec);
                            $.OPTION(() => $.CONSUME(Semicolon));
                        });
                        $.CONSUME(RParen);
                    }
                }
            ]);
        });
        this.typeDecl = $.RULE("typeDecl", () => {
            $.CONSUME(Type);
            $.OR([
                {
                    ALT: () => {
                        $.SUBRULE(this.name);
                        $.SUBRULE(this.typeExpression);
                    }
                },
                {
                    ALT: () => {
                        $.CONSUME(LParen);
                        $.MANY(() => {
                            $.SUBRULE2(this.name);
                            $.SUBRULE2(this.typeExpression);
                            $.OPTION(() => $.CONSUME(Semicolon));
                        });
                        $.CONSUME(RParen);
                    }
                }
            ]);
        });
        this.returnStmt = $.RULE("returnStmt", () => {
            $.CONSUME(Return);
            $.OPTION(() => $.SUBRULE(this.expressionList));
        });
        this.ifStmt = $.RULE("ifStmt", () => {
            $.CONSUME(If);
            $.SUBRULE(this.expression);
            $.SUBRULE(this.block);
            $.OPTION(() => {
                $.CONSUME(Else);
                $.OR([
                    { ALT: () => $.SUBRULE(this.ifStmt) },
                    { ALT: () => $.SUBRULE2(this.block) }
                ]);
            });
        });
        this.switchStmt = $.RULE("switchStmt", () => {
            $.CONSUME(Switch);
            $.OPTION(() => $.SUBRULE(this.expression));
            $.CONSUME(LBrace);
            $.MANY(() => $.SUBRULE(this.switchClause));
            $.CONSUME(RBrace);
        });
        this.switchClause = $.RULE("switchClause", () => {
            $.OR([
                {
                    ALT: () => {
                        $.CONSUME(Case);
                        $.SUBRULE(this.expressionList);
                    }
                },
                { ALT: () => $.CONSUME(Default) }
            ]);
            $.CONSUME(Colon);
            $.MANY(() => {
                $.SUBRULE(this.statement);
                $.OPTION(() => $.CONSUME(Semicolon));
            });
        });
        this.forStmt = $.RULE("forStmt", () => {
            $.CONSUME(For);
            $.OR([
                { GATE: $.BACKTRACK(this.rangeClause), ALT: () => $.SUBRULE(this.rangeClause) },
                { GATE: $.BACKTRACK(this.forClause), ALT: () => $.SUBRULE(this.forClause) },
                {
                    ALT: () => {
                        $.OPTION(() => $.SUBRULE(this.expression));
                    }
                }
            ]);
            $.SUBRULE(this.block);
        });
        this.forClause = $.RULE("forClause", () => {
            $.OPTION(() => $.SUBRULE(this.forInitClause));
            $.CONSUME(Semicolon);
            $.OPTION2(() => $.SUBRULE(this.expression));
            $.CONSUME2(Semicolon);
            $.OPTION3(() => $.SUBRULE(this.forPostClause));
        });
        this.forInitClause = $.RULE("forInitClause", () => {
            $.SUBRULE(this.name);
            $.OR([
                { ALT: () => $.CONSUME(Define) },
                { ALT: () => $.CONSUME(Assign) }
            ]);
            $.SUBRULE(this.expression);
        });
        this.forPostClause = $.RULE("forPostClause", () => {
            $.SUBRULE(this.expression);
            $.OR([
                { ALT: () => $.CONSUME(PlusPlus) },
                { ALT: () => $.CONSUME(MinusMinus) }
            ]);
        });
        this.rangeClause = $.RULE("rangeClause", () => {
            $.OPTION(() => {
                $.SUBRULE(this.name);
                $.OPTION2(() => {
                    $.CONSUME(Comma);
                    $.SUBRULE2(this.name);
                });
                $.OR([
                    { ALT: () => $.CONSUME(Define) },
                    { ALT: () => $.CONSUME(Assign) }
                ]);
            });
            $.CONSUME(Range);
            $.SUBRULE(this.expression);
        });
        this.deferStmt = $.RULE("deferStmt", () => {
            $.CONSUME(Defer);
            $.SUBRULE(this.expression);
        });
        this.branchStmt = $.RULE("branchStmt", () => {
            $.OR([
                {
                    ALT: () => {
                        $.CONSUME(Break);
                        $.OPTION(() => $.CONSUME(Identifier));
                    }
                },
                {
                    ALT: () => {
                        $.CONSUME(Continue);
                        $.OPTION2(() => $.CONSUME2(Identifier));
                    }
                },
                { ALT: () => $.CONSUME(Fallthrough) },
                {
                    ALT: () => {
                        $.CONSUME(Goto);
                        $.CONSUME3(Identifier);
                    }
                }
            ]);
        });
        this.declarationSpec = $.RULE("declarationSpec", () => {
            $.SUBRULE(this.name);
            $.OPTION(() => $.SUBRULE(this.typeExpression));
            $.OPTION2(() => {
                $.CONSUME(Assign);
                $.SUBRULE(this.expression);
            });
        });
        this.simpleStmt = $.RULE("simpleStmt", () => {
            $.SUBRULE(this.expression);
            $.OPTION(() => {
                $.OR([
                    {
                        ALT: () => {
                            $.OR2([
                                { ALT: () => $.CONSUME(Define) },
                                { ALT: () => $.CONSUME(Assign) }
                            ]);
                            $.SUBRULE2(this.expression);
                        }
                    },
                    { ALT: () => $.CONSUME(PlusPlus) },
                    { ALT: () => $.CONSUME(MinusMinus) }
                ]);
            });
        });
        this.expressionStatement = $.RULE("expressionStatement", () => {
            $.SUBRULE(this.expression);
        });
        this.expressionList = $.RULE("expressionList", () => {
            $.SUBRULE(this.expression);
            $.MANY(() => {
                $.CONSUME(Comma);
                $.SUBRULE2(this.expression);
            });
        });
        this.expression = $.RULE("expression", () => $.SUBRULE(this.orExpr));
        this.orExpr = $.RULE("orExpr", () => {
            $.SUBRULE(this.andExpr);
            $.MANY(() => {
                $.CONSUME(OrOr);
                $.SUBRULE2(this.andExpr);
            });
        });
        this.andExpr = $.RULE("andExpr", () => {
            $.SUBRULE(this.equalityExpr);
            $.MANY(() => {
                $.CONSUME(AndAnd);
                $.SUBRULE2(this.equalityExpr);
            });
        });
        this.equalityExpr = $.RULE("equalityExpr", () => {
            $.SUBRULE(this.compareExpr);
            $.MANY(() => {
                $.OR([
                    { ALT: () => $.CONSUME(Equal) },
                    { ALT: () => $.CONSUME(NotEqual) }
                ]);
                $.SUBRULE2(this.compareExpr);
            });
        });
        this.compareExpr = $.RULE("compareExpr", () => {
            $.SUBRULE(this.addExpr);
            $.MANY(() => {
                $.OR([
                    { ALT: () => $.CONSUME(Less) },
                    { ALT: () => $.CONSUME(LessEqual) },
                    { ALT: () => $.CONSUME(Greater) },
                    { ALT: () => $.CONSUME(GreaterEqual) }
                ]);
                $.SUBRULE2(this.addExpr);
            });
        });
        this.addExpr = $.RULE("addExpr", () => {
            $.SUBRULE(this.mulExpr);
            $.MANY(() => {
                $.OR([
                    { ALT: () => $.CONSUME(Plus) },
                    { ALT: () => $.CONSUME(Minus) }
                ]);
                $.SUBRULE2(this.mulExpr);
            });
        });
        this.mulExpr = $.RULE("mulExpr", () => {
            $.SUBRULE(this.unaryExpr);
            $.MANY(() => {
                $.OR([
                    { ALT: () => $.CONSUME(Star) },
                    { ALT: () => $.CONSUME(Slash) },
                    { ALT: () => $.CONSUME(Percent) }
                ]);
                $.SUBRULE2(this.unaryExpr);
            });
        });
        this.unaryExpr = $.RULE("unaryExpr", () => {
            $.OR([
                {
                    ALT: () => {
                        $.OR2([
                            { ALT: () => $.CONSUME(Plus) },
                            { ALT: () => $.CONSUME(Minus) },
                            { ALT: () => $.CONSUME(Bang) },
                            { ALT: () => $.CONSUME(Amp) },
                            { ALT: () => $.CONSUME(Star) }
                        ]);
                        $.SUBRULE(this.unaryExpr);
                    }
                },
                { ALT: () => $.SUBRULE(this.primaryExpr) }
            ]);
        });
        this.primaryExpr = $.RULE("primaryExpr", () => {
            $.SUBRULE(this.atom);
            $.MANY(() => {
                $.OR([
                    {
                        ALT: () => {
                            $.CONSUME(Dot);
                            $.SUBRULE(this.selectorName);
                        }
                    },
                    {
                        ALT: () => {
                            $.CONSUME(LBracket);
                            $.OPTION(() => $.SUBRULE(this.expression));
                            $.OPTION2(() => {
                                $.CONSUME(Colon);
                                $.OPTION3(() => $.SUBRULE2(this.expression));
                            });
                            $.CONSUME(RBracket);
                        }
                    },
                    { ALT: () => $.SUBRULE(this.arguments) }
                ]);
            });
            $.OPTION4({ GATE: () => this.nextTokensAreSpreadsheetRangeSuffix(), DEF: () => {
                    $.CONSUME2(Colon);
                    $.SUBRULE3(this.selectorName);
                } });
        });
        this.atom = $.RULE("atom", () => {
            $.OR([
                { ALT: () => $.SUBRULE(this.functionLiteral) },
                { ALT: () => $.SUBRULE(this.mapLiteral) },
                { ALT: () => $.SUBRULE(this.literal) },
                { ALT: () => $.SUBRULE(this.qualifiedName) },
                {
                    ALT: () => {
                        $.CONSUME(LParen);
                        $.SUBRULE(this.expression);
                        $.CONSUME(RParen);
                    }
                }
            ]);
        });
        this.mapLiteral = $.RULE("mapLiteral", () => {
            $.CONSUME(MapTok);
            $.CONSUME(LBracket);
            $.SUBRULE(this.typeExpression);
            $.CONSUME(RBracket);
            $.SUBRULE2(this.typeExpression);
            $.CONSUME(LBrace);
            $.OPTION(() => {
                $.SUBRULE(this.mapElement);
                $.MANY(() => {
                    $.CONSUME(Comma);
                    $.SUBRULE2(this.mapElement);
                });
                $.OPTION2(() => $.CONSUME2(Comma));
            });
            $.CONSUME(RBrace);
        });
        this.mapElement = $.RULE("mapElement", () => {
            $.SUBRULE(this.expression);
            $.CONSUME(Colon);
            $.SUBRULE2(this.expression);
        });
        this.functionLiteral = $.RULE("functionLiteral", () => {
            $.CONSUME(Func);
            $.SUBRULE(this.signature);
            $.SUBRULE(this.block);
        });
        this.arguments = $.RULE("arguments", () => {
            $.CONSUME(LParen);
            $.OPTION(() => {
                $.SUBRULE(this.expression);
                $.OPTION2(() => $.CONSUME(Ellipsis));
                $.MANY(() => {
                    $.CONSUME(Comma);
                    $.SUBRULE2(this.expression);
                    $.OPTION3(() => $.CONSUME2(Ellipsis));
                });
                $.OPTION4(() => $.CONSUME2(Comma));
            });
            $.CONSUME(RParen);
        });
        this.qualifiedName = $.RULE("qualifiedName", () => {
            $.SUBRULE(this.name);
            $.MANY(() => {
                $.CONSUME(Dot);
                $.SUBRULE2(this.selectorName);
            });
        });
        this.selectorName = $.RULE("selectorName", () => {
            $.OR([
                { ALT: () => $.CONSUME(Identifier) },
                { ALT: () => $.CONSUME(CellAddress) }
            ]);
        });
        this.name = $.RULE("name", () => {
            $.OR([
                { ALT: () => $.CONSUME(Identifier) },
                { ALT: () => $.CONSUME(CellAddress) }
            ]);
        });
        this.literal = $.RULE("literal", () => {
            $.OR([
                { ALT: () => $.CONSUME(StringLiteral) },
                { ALT: () => $.CONSUME(FloatLiteral) },
                { ALT: () => $.CONSUME(IntLiteral) },
                { ALT: () => $.CONSUME(True) },
                { ALT: () => $.CONSUME(False) },
                { ALT: () => $.CONSUME(Nil) }
            ]);
        });
        this.performSelfAnalysis();
    }
    nextTokensAreLabel() {
        return this.LA(1).tokenType === Identifier && this.LA(2).tokenType === Colon;
    }
    nextTokensStartFunctionDecl() {
        if (this.LA(1).tokenType !== Func)
            return false;
        if (this.LA(2).tokenType === Identifier)
            return true;
        if (this.LA(2).tokenType !== LParen)
            return false;
        let depth = 0;
        for (let offset = 2; offset < 64; offset += 1) {
            const tokenType = this.LA(offset).tokenType;
            if (tokenType === LParen) {
                depth += 1;
                continue;
            }
            if (tokenType === RParen) {
                depth -= 1;
                if (depth === 0) {
                    return this.LA(offset + 1).tokenType === Identifier;
                }
            }
            if (tokenType.name === "EOF")
                return false;
        }
        return false;
    }
    nextTokensAreSpreadsheetRangeSuffix() {
        return this.LA(1).tokenType === Colon &&
            (this.LA(2).tokenType === Identifier || this.LA(2).tokenType === CellAddress);
    }
    nextTokensAreNamedParameter() {
        const first = this.LA(1).tokenType;
        const second = this.LA(2).tokenType;
        if (!this.isNameToken(first))
            return false;
        if (second === Ellipsis || this.isTypeStartToken(second))
            return true;
        if (second !== Comma)
            return false;
        let offset = 2;
        while (this.LA(offset).tokenType === Comma && this.isNameToken(this.LA(offset + 1).tokenType)) {
            offset += 2;
        }
        const afterNames = this.LA(offset).tokenType;
        return afterNames === Ellipsis || this.isTypeStartToken(afterNames);
    }
    isNameToken(tokenType) {
        return tokenType === Identifier || tokenType === CellAddress;
    }
    isTypeStartToken(tokenType) {
        return tokenType === Star ||
            tokenType === LBracket ||
            tokenType === MapTok ||
            tokenType === Struct ||
            tokenType === Interface ||
            tokenType === Func ||
            tokenType === Identifier ||
            tokenType === CellAddress;
    }
}
export const goJuniorParser = new GoJuniorParser();
export function parseGoJunior(source) {
    const lexResult = goJuniorLexer.tokenize(source);
    goJuniorParser.input = lexResult.tokens;
    const cst = goJuniorParser.program();
    const diagnostics = [
        ...lexResult.errors.map((error) => lexDiagnostic(error)),
        ...goJuniorParser.errors.map(parserDiagnostic)
    ];
    return { cst, diagnostics, tokens: lexResult.tokens };
}
function lexDiagnostic(error) {
    return {
        code: "GJLEX001",
        severity: "error",
        message: error.message,
        span: {
            offset: error.offset,
            length: error.length,
            line: error.line ?? 1,
            column: error.column ?? 1
        }
    };
}
function parserDiagnostic(error) {
    const token = error.token;
    const diagnostic = {
        code: "GJPARSE001",
        severity: "error",
        message: error.message
    };
    if (token) {
        diagnostic.span = spanFromToken(token);
    }
    return diagnostic;
}
