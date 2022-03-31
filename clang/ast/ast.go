package ast

// -----------------------------------------------------------------------------

type IncludedFrom struct {
	File string `json:"file"`
}

type Loc struct {
	Offset       int64         `json:"offset,omitempty"` // 432
	File         string        `json:"file,omitempty"`   // "sqlite3.i"
	Line         int           `json:"line,omitempty"`
	PresumedFile string        `json:"presumedFile,omitempty"`
	PresumedLine int           `json:"presumedLine,omitempty"`
	Col          int           `json:"col,omitempty"`
	TokLen       int           `json:"tokLen,omitempty"`
	IncludedFrom *IncludedFrom `json:"includedFrom,omitempty"` // "sqlite3.c"
}

type Pos struct {
	Offset       int64         `json:"offset,omitempty"`
	Col          int           `json:"col,omitempty"`
	TokLen       int           `json:"tokLen,omitempty"`
	IncludedFrom *IncludedFrom `json:"includedFrom,omitempty"` // "sqlite3.c"
	SpellingLoc  *Loc          `json:"spellingLoc,omitempty"`
	ExpansionLoc *Loc          `json:"expansionLoc,omitempty"`
}

type Range struct {
	Begin Pos `json:"begin"`
	End   Pos `json:"end"`
}

// -----------------------------------------------------------------------------

type ID string

type Kind string

const (
	TranslationUnitDecl Kind = "TranslationUnitDecl"
	TypedefType         Kind = "TypedefType"
	TypedefDecl         Kind = "TypedefDecl"
	ElaboratedType      Kind = "ElaboratedType"
	BuiltinType         Kind = "BuiltinType"
	ConstantArrayType   Kind = "ConstantArrayType"
	IncompleteArrayType Kind = "IncompleteArrayType"
	PointerType         Kind = "PointerType"
	RecordType          Kind = "RecordType"
	RecordDecl          Kind = "RecordDecl"
	FieldDecl           Kind = "FieldDecl"
	VarDecl             Kind = "VarDecl"
	EnumDecl            Kind = "EnumDecl"
	EnumConstantDecl    Kind = "EnumConstantDecl"
	AlwaysInlineAttr    Kind = "AlwaysInlineAttr"
	AsmLabelAttr        Kind = "AsmLabelAttr"
	AvailabilityAttr    Kind = "AvailabilityAttr"
	DeprecatedAttr      Kind = "DeprecatedAttr"
	BuiltinAttr         Kind = "BuiltinAttr"
	FormatAttr          Kind = "FormatAttr"
	ColdAttr            Kind = "ColdAttr"
	ConstAttr           Kind = "ConstAttr"
	PackedAttr          Kind = "PackedAttr"
	NoThrowAttr         Kind = "NoThrowAttr"
	MayAliasAttr        Kind = "MayAliasAttr"
	FunctionProtoType   Kind = "FunctionProtoType"
	FunctionDecl        Kind = "FunctionDecl"
	ParmVarDecl         Kind = "ParmVarDecl"
	ParenType           Kind = "ParenType"
	DeclStmt            Kind = "DeclStmt"
	CompoundStmt        Kind = "CompoundStmt"
	NullStmt            Kind = "NullStmt"
	ForStmt             Kind = "ForStmt"
	WhileStmt           Kind = "WhileStmt"
	DoStmt              Kind = "DoStmt"
	GotoStmt            Kind = "GotoStmt"
	BreakStmt           Kind = "BreakStmt"
	ContinueStmt        Kind = "ContinueStmt"
	LabelStmt           Kind = "LabelStmt"
	IfStmt              Kind = "IfStmt"
	SwitchStmt          Kind = "SwitchStmt"
	CaseStmt            Kind = "CaseStmt"
	DefaultStmt         Kind = "DefaultStmt"
	ReturnStmt          Kind = "ReturnStmt"
	ParenExpr           Kind = "ParenExpr"
	CallExpr            Kind = "CallExpr"
	ConstantExpr        Kind = "ConstantExpr"
	CStyleCastExpr      Kind = "CStyleCastExpr"
	DeclRefExpr         Kind = "DeclRefExpr"
	MemberExpr          Kind = "MemberExpr"
	ImplicitCastExpr    Kind = "ImplicitCastExpr"
	BinaryOperator      Kind = "BinaryOperator"
	UnaryOperator       Kind = "UnaryOperator"
	ConditionalOperator Kind = "ConditionalOperator"
	CharacterLiteral    Kind = "CharacterLiteral"
	IntegerLiteral      Kind = "IntegerLiteral"
	StringLiteral       Kind = "StringLiteral"
)

type ValueCategory string

const (
	RValue ValueCategory = "rvalue"
	LValue ValueCategory = "lvalue"
)

type CC string

const (
	CDecl CC = "cdecl"
)

type StorageClass string

const (
	Static StorageClass = "static"
	Extern StorageClass = "extern"
)

type CastKind string

const (
	LValueToRValue         CastKind = "LValueToRValue"
	IntegralCast           CastKind = "IntegralCast"
	IntegralToPointer      CastKind = "IntegralToPointer"
	PointerToIntegral      CastKind = "PointerToIntegral"
	FunctionToPointerDecay CastKind = "FunctionToPointerDecay"
	ArrayToPointerDecay    CastKind = "ArrayToPointerDecay"
	BuiltinFnToFnPtr       CastKind = "BuiltinFnToFnPtr"
	NoOp                   CastKind = "NoOp"
)

type (
	// OpCode can be:
	//   + - * / || >= -- ++ etc
	OpCode string
)

type Type struct {
	// QualType can be:
	//   unsigned int
	//   struct ConstantString
	//   volatile uint32_t
	//   int (*)(void *, int, char **, char **)
	//   int (*)(const char *, ...)
	//   int (*)(void)
	//   const char *restrict
	//   const char [7]
	//   char *
	//   void
	//   ...
	QualType          string `json:"qualType"`
	DesugaredQualType string `json:"desugaredQualType,omitempty"`
	TypeAliasDeclID   ID     `json:"typeAliasDeclId,omitempty"`
}

type Node struct {
	ID                   ID            `json:"id,omitempty"`
	Kind                 Kind          `json:"kind,omitempty"`
	Loc                  *Loc          `json:"loc,omitempty"`
	Range                *Range        `json:"range,omitempty"`
	ReferencedMemberDecl ID            `json:"referencedMemberDecl,omitempty"`
	PreviousDecl         ID            `json:"previousDecl,omitempty"`
	ParentDeclContextID  ID            `json:"parentDeclContextId,omitempty"`
	IsImplicit           bool          `json:"isImplicit,omitempty"`   // is this type implicit defined
	IsReferenced         bool          `json:"isReferenced,omitempty"` // is this type refered or not
	IsUsed               bool          `json:"isUsed,omitempty"`       // is this variable used or not
	IsArrow              bool          `json:"isArrow,omitempty"`      // is ptr->member not obj.member
	IsPostfix            bool          `json:"isPostfix,omitempty"`
	IsPartOfExplicitCast bool          `json:"isPartOfExplicitCast,omitempty"`
	HasElse              bool          `json:"hasElse,omitempty"`
	Inline               bool          `json:"inline,omitempty"`
	StorageClass         StorageClass  `json:"storageClass,omitempty"`
	TagUsed              string        `json:"tagUsed,omitempty"` // struct | union
	CompleteDefinition   bool          `json:"completeDefinition,omitempty"`
	Name                 string        `json:"name,omitempty"`
	MangledName          string        `json:"mangledName,omitempty"`
	Type                 *Type         `json:"type,omitempty"`
	CC                   CC            `json:"cc,omitempty"`
	Decl                 *Node         `json:"decl,omitempty"`
	OwnedTagDecl         *Node         `json:"ownedTagDecl,omitempty"`
	ReferencedDecl       *Node         `json:"referencedDecl,omitempty"`
	OpCode               OpCode        `json:"opcode,omitempty"`
	Init                 string        `json:"init,omitempty"`
	ValueCategory        ValueCategory `json:"valueCategory,omitempty"`
	Value                interface{}   `json:"value,omitempty"`
	CastKind             CastKind      `json:"castKind,omitempty"`
	Size                 int           `json:"size,omitempty"` // array size
	Inner                []*Node       `json:"inner,omitempty"`
}

// -----------------------------------------------------------------------------
