import {
  createTokenInstance,
  CstNode,
  CstParser,
  IRecognitionException,
  IToken
} from "chevrotain";
import { Diagnostic, spanFromToken } from "./diagnostics.js";
import {
  allTokens,
  Amp,
  AndAnd,
  Assign,
  Bang,
  Break,
  Case,
  CellAddress,
  Colon,
  Comma,
  Const,
  Continue,
  Default,
  Defer,
  Define,
  Dot,
  Ellipsis,
  Else,
  Equal,
  Fallthrough,
  False,
  FloatLiteral,
  For,
  Func,
  Goto,
  Greater,
  GreaterEqual,
  Identifier,
  If,
  Import,
  IntLiteral,
  Interface,
  LBrace,
  LBracket,
  Less,
  LessEqual,
  LParen,
  MapTok,
  Minus,
  MinusMinus,
  Nil,
  NotEqual,
  OrOr,
  Percent,
  Plus,
  PlusPlus,
  Range,
  Return,
  RBrace,
  RBracket,
  RParen,
  Semicolon,
  Slash,
  Star,
  StringLiteral,
  Struct,
  Switch,
  True,
  Type,
  Var,
  goJuniorLexer
} from "./tokens.js";

type Rule = () => CstNode;

export class GoJuniorParser extends CstParser {
  public program!: Rule;
  public importDecl!: Rule;
  private importSpec!: Rule;
  private functionDecl!: Rule;
  private receiver!: Rule;
  private signature!: Rule;
  private parameterList!: Rule;
  private parameter!: Rule;
  private result!: Rule;
  private typeExpression!: Rule;
  private interfaceType!: Rule;
  private structType!: Rule;
  private fieldDecl!: Rule;
  private block!: Rule;
  private statement!: Rule;
  private labeledStmt!: Rule;
  private constDecl!: Rule;
  private varDecl!: Rule;
  private typeDecl!: Rule;
  private returnStmt!: Rule;
  private ifStmt!: Rule;
  private switchStmt!: Rule;
  private typeSwitchHeader!: Rule;
  private switchClause!: Rule;
  private typeSwitchClause!: Rule;
  private typeSwitchType!: Rule;
  private forStmt!: Rule;
  private forClause!: Rule;
  private forInitClause!: Rule;
  private forPostClause!: Rule;
  private rangeClause!: Rule;
  private deferStmt!: Rule;
  private branchStmt!: Rule;
  private declarationSpec!: Rule;
  private simpleStmt!: Rule;
  private expressionStatement!: Rule;
  private expression!: Rule;
  private orExpr!: Rule;
  private andExpr!: Rule;
  private equalityExpr!: Rule;
  private compareExpr!: Rule;
  private addExpr!: Rule;
  private mulExpr!: Rule;
  private unaryExpr!: Rule;
  private primaryExpr!: Rule;
  private atom!: Rule;
  private arrayLiteral!: Rule;
  private structLiteral!: Rule;
  private structLiteralField!: Rule;
  private mapLiteral!: Rule;
  private mapElement!: Rule;
  private functionLiteral!: Rule;
  private arguments!: Rule;
  private expressionList!: Rule;
  private qualifiedName!: Rule;
  private selectorName!: Rule;
  private name!: Rule;
  private literal!: Rule;

  public constructor() {
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
        $.MANY1(() => {
          $.OR([
            { GATE: () => this.nextTokensStartFunctionDecl(), ALT: () => $.SUBRULE(this.functionDecl) },
            { ALT: () => $.SUBRULE(this.statement) }
          ]);
          $.OPTION3(() => $.CONSUME2(Semicolon));
        });
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
            $.OPTION(() => {
              $.OR2([
                { ALT: () => $.CONSUME(IntLiteral) },
                { ALT: () => $.CONSUME(Ellipsis) }
              ]);
            });
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
      $.MANY(() => {
        $.CONSUME(Comma);
        $.SUBRULE2(this.name);
      });
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
      $.MANY(() => {
        $.CONSUME(Identifier);
        $.CONSUME(Colon);
      });
      $.OR([
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
      $.SUBRULE(this.statement);
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
      $.OR([
        {
          GATE: $.BACKTRACK(this.typeSwitchHeader),
          ALT: () => {
            $.SUBRULE(this.typeSwitchHeader);
            $.CONSUME(LBrace);
            $.MANY(() => $.SUBRULE(this.typeSwitchClause));
            $.CONSUME(RBrace);
          }
        },
        {
          ALT: () => {
            $.OPTION(() => $.SUBRULE(this.expression));
            $.CONSUME2(LBrace);
            $.MANY2(() => $.SUBRULE(this.switchClause));
            $.CONSUME2(RBrace);
          }
        }
      ]);
    });

    this.typeSwitchHeader = $.RULE("typeSwitchHeader", () => {
      $.OPTION({ GATE: () => this.isNameToken(this.LA(1).tokenType) &&
        (this.LA(2).tokenType === Define || this.LA(2).tokenType === Assign), DEF: () => {
        $.SUBRULE(this.name);
        $.OR([
          { ALT: () => $.CONSUME(Define) },
          { ALT: () => $.CONSUME(Assign) }
        ]);
      } });
      $.SUBRULE(this.qualifiedName);
      $.CONSUME(Dot);
      $.CONSUME(LParen);
      $.CONSUME(Type);
      $.CONSUME(RParen);
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

    this.typeSwitchClause = $.RULE("typeSwitchClause", () => {
      $.OR([
        {
          ALT: () => {
            $.CONSUME(Case);
            $.SUBRULE(this.typeSwitchType);
            $.MANY(() => {
              $.CONSUME(Comma);
              $.SUBRULE2(this.typeSwitchType);
            });
          }
        },
        { ALT: () => $.CONSUME(Default) }
      ]);
      $.CONSUME(Colon);
      $.MANY2(() => {
        $.SUBRULE(this.statement);
        $.OPTION(() => $.CONSUME(Semicolon));
      });
    });

    this.typeSwitchType = $.RULE("typeSwitchType", () => {
      $.OR([
        { ALT: () => $.CONSUME(Nil) },
        { ALT: () => $.SUBRULE(this.typeExpression) }
      ]);
    });

    this.forStmt = $.RULE("forStmt", () => {
      $.CONSUME(For);
      $.OR([
        { GATE: () => this.nextTokensStartRangeClause(), ALT: () => $.SUBRULE(this.rangeClause) },
        { GATE: () => this.nextTokensStartForClause(), ALT: () => $.SUBRULE(this.forClause) },
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
      $.SUBRULE(this.expressionList);
      $.OPTION(() => {
        $.OR([
          {
            ALT: () => {
              $.OR2([
                { ALT: () => $.CONSUME(Define) },
                { ALT: () => $.CONSUME(Assign) }
              ]);
              $.SUBRULE2(this.expressionList);
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
              $.CONSUME2(Dot);
              $.CONSUME(LParen);
              $.SUBRULE(this.typeExpression);
              $.CONSUME(RParen);
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
        { GATE: () => this.LA(1).tokenType === LBracket, ALT: () => $.SUBRULE(this.arrayLiteral) },
        { ALT: () => $.SUBRULE(this.mapLiteral) },
        { GATE: $.BACKTRACK(this.structLiteral), ALT: () => $.SUBRULE(this.structLiteral) },
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

    this.arrayLiteral = $.RULE("arrayLiteral", () => {
      $.CONSUME(LBracket);
      $.OPTION(() => {
        $.OR2([
          { ALT: () => $.CONSUME(IntLiteral) },
          { ALT: () => $.CONSUME(Ellipsis) }
        ]);
      });
      $.CONSUME(RBracket);
      $.SUBRULE(this.typeExpression);
      $.CONSUME(LBrace);
      $.OPTION2(() => {
        $.SUBRULE(this.expression);
        $.MANY(() => {
          $.CONSUME(Comma);
          $.SUBRULE2(this.expression);
        });
        $.OPTION3(() => $.CONSUME2(Comma));
      });
      $.CONSUME(RBrace);
    });

    this.structLiteral = $.RULE("structLiteral", () => {
      $.SUBRULE(this.qualifiedName);
      $.CONSUME(LBrace);
      $.OPTION(() => {
        $.SUBRULE(this.structLiteralField);
        $.MANY(() => {
          $.CONSUME(Comma);
          $.SUBRULE2(this.structLiteralField);
        });
        $.OPTION2(() => $.CONSUME2(Comma));
      });
      $.CONSUME(RBrace);
    });

    this.structLiteralField = $.RULE("structLiteralField", () => {
      $.OR([
        {
          GATE: () => this.nextTokensAreStructLiteralKey(),
          ALT: () => {
            $.SUBRULE(this.selectorName);
            $.CONSUME(Colon);
            $.SUBRULE(this.expression);
          }
        },
        { ALT: () => $.SUBRULE2(this.expression) }
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
      $.MANY({ GATE: () => this.LA(1).tokenType === Dot && this.isNameToken(this.LA(2).tokenType), DEF: () => {
        $.CONSUME(Dot);
        $.SUBRULE2(this.selectorName);
      } });
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

  private nextTokensAreLabel(): boolean {
    return this.LA(1).tokenType === Identifier && this.LA(2).tokenType === Colon;
  }

  private nextTokensStartFunctionDecl(): boolean {
    if (this.LA(1).tokenType !== Func) return false;
    if (this.LA(2).tokenType === Identifier) return true;
    if (this.LA(2).tokenType !== LParen) return false;

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
      if (tokenType.name === "EOF") return false;
    }
    return false;
  }

  private nextTokensAreSpreadsheetRangeSuffix(): boolean {
    return this.LA(1).tokenType === Colon &&
      (this.LA(2).tokenType === Identifier || this.LA(2).tokenType === CellAddress);
  }

  private nextTokensAreStructLiteralKey(): boolean {
    return this.isNameToken(this.LA(1).tokenType) && this.LA(2).tokenType === Colon;
  }

  private nextTokensStartForClause(): boolean {
    return this.nextTopLevelTokenBeforeForBody(Semicolon);
  }

  private nextTokensStartRangeClause(): boolean {
    return this.nextTopLevelTokenBeforeForBody(Range);
  }

  private nextTopLevelTokenBeforeForBody(target: IToken["tokenType"]): boolean {
    let parenDepth = 0;
    let bracketDepth = 0;
    for (let offset = 1; offset < 128; offset += 1) {
      const tokenType = this.LA(offset).tokenType;
      if (tokenType.name === "EOF") return false;
      if (tokenType === LParen) parenDepth += 1;
      if (tokenType === RParen) parenDepth = Math.max(0, parenDepth - 1);
      if (tokenType === LBracket) bracketDepth += 1;
      if (tokenType === RBracket) bracketDepth = Math.max(0, bracketDepth - 1);
      if (parenDepth === 0 && bracketDepth === 0) {
        if (tokenType === target) return true;
        if (tokenType === LBrace) return false;
      }
    }
    return false;
  }

  private nextTokensAreNamedParameter(): boolean {
    const first = this.LA(1).tokenType;
    const second = this.LA(2).tokenType;
    if (!this.isNameToken(first)) return false;
    if (second === Ellipsis || this.isTypeStartToken(second)) return true;
    if (second !== Comma) return false;

    let offset = 2;
    while (this.LA(offset).tokenType === Comma && this.isNameToken(this.LA(offset + 1).tokenType)) {
      offset += 2;
    }

    const afterNames = this.LA(offset).tokenType;
    return afterNames === Ellipsis || this.isTypeStartToken(afterNames);
  }

  private isNameToken(tokenType: IToken["tokenType"]): boolean {
    return tokenType === Identifier || tokenType === CellAddress;
  }

  private isTypeStartToken(tokenType: IToken["tokenType"]): boolean {
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

export interface ParseResult {
  cst?: CstNode;
  diagnostics: Diagnostic[];
  tokens: IToken[];
}

export function parseGoJunior(source: string): ParseResult {
  const lexResult = goJuniorLexer.tokenize(source);
  const tokens = insertImplicitSemicolons(lexResult.tokens);
  goJuniorParser.input = tokens;
  const cst = goJuniorParser.program();
  const diagnostics = [
    ...lexResult.errors.map((error) => lexDiagnostic(error)),
    ...goJuniorParser.errors.map(parserDiagnostic)
  ];

  return { cst, diagnostics, tokens };
}

function insertImplicitSemicolons(tokens: IToken[]): IToken[] {
  if (tokens.length === 0) return tokens;

  const result: IToken[] = [];
  for (let index = 0; index < tokens.length; index += 1) {
    const token = tokens[index]!;
    const next = tokens[index + 1];
    result.push(token);
    if (!next) continue;
    if (!shouldInsertSemicolon(token, next)) continue;

    const endOffset = token.endOffset ?? token.startOffset;
    const line = token.endLine ?? token.startLine ?? 1;
    const column = token.endColumn ?? token.startColumn ?? 1;
    result.push(createTokenInstance(
      Semicolon,
      ";",
      endOffset + 1,
      endOffset + 1,
      line,
      line,
      column + 1,
      column + 1
    ));
  }
  return result;
}

function shouldInsertSemicolon(token: IToken, next: IToken): boolean {
  if ((next.startLine ?? token.endLine ?? 1) <= (token.endLine ?? token.startLine ?? 1)) return false;
  return canEndStatement(token.tokenType) && canStartImplicitlySeparatedStatement(next.tokenType);
}

function canEndStatement(tokenType: IToken["tokenType"]): boolean {
  return tokenType === Identifier ||
    tokenType === CellAddress ||
    tokenType === IntLiteral ||
    tokenType === FloatLiteral ||
    tokenType === StringLiteral ||
    tokenType === True ||
    tokenType === False ||
    tokenType === Nil ||
    tokenType === Break ||
    tokenType === Continue ||
    tokenType === Fallthrough ||
    tokenType === Return ||
    tokenType === PlusPlus ||
    tokenType === MinusMinus ||
    tokenType === RParen ||
    tokenType === RBracket ||
    tokenType === RBrace;
}

function canStartImplicitlySeparatedStatement(tokenType: IToken["tokenType"]): boolean {
  return tokenType === Identifier ||
    tokenType === CellAddress ||
    tokenType === StringLiteral ||
    tokenType === FloatLiteral ||
    tokenType === IntLiteral ||
    tokenType === True ||
    tokenType === False ||
    tokenType === Nil ||
    tokenType === Func ||
    tokenType === MapTok ||
    tokenType === LBracket ||
    tokenType === LParen ||
    tokenType === Plus ||
    tokenType === Minus ||
    tokenType === Bang ||
    tokenType === Amp ||
    tokenType === Star ||
    tokenType === Return ||
    tokenType === If ||
    tokenType === Switch ||
    tokenType === For ||
    tokenType === Defer ||
    tokenType === Break ||
    tokenType === Continue ||
    tokenType === Fallthrough ||
    tokenType === Goto ||
    tokenType === Var ||
    tokenType === Const ||
    tokenType === Type ||
    tokenType === Case ||
    tokenType === Default ||
    tokenType === RBrace;
}

function lexDiagnostic(error: {
  message: string;
  offset: number;
  length: number;
  line: number | undefined;
  column: number | undefined;
}): Diagnostic {
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

function parserDiagnostic(error: IRecognitionException): Diagnostic {
  const token = error.token;
  const diagnostic: Diagnostic = {
    code: "GJPARSE001",
    severity: "error",
    message: error.message
  };
  if (token) {
    diagnostic.span = spanFromToken(token);
  }
  return diagnostic;
}
