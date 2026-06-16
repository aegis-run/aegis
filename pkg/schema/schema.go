package schema

// Type represents an entity type in the authorization model.
type Type struct {
	Name        string
	Relations   map[string]*Relation
	Permissions map[string]*Permission
}

func (t Type) Permission(name string) *Permission {
	return t.Permissions[name]
}

func (t Type) Relation(name string) *Relation {
	return t.Relations[name]
}

// Relation represents a direct connection between an instance and actors.
type Relation struct {
	Name       string
	ActorTypes []ActorType
}

// ActorType specifies what types of subjects can be bound to a relation.
type ActorType interface {
	isActorType()
}

// DirectActorType represents a direct connection to a specific type (e.g., "user").
type DirectActorType struct {
	Type string
}

func (DirectActorType) isActorType() {}

// UsersetActorType represents a connection to actors that have a specific relation
// to another instance (e.g., "group#member").
type UsersetActorType struct {
	Type     string
	Relation string
}

func (UsersetActorType) isActorType() {}

// Permission represents a computed relationship based on other relations/permissions.
type Permission struct {
	Name string
	Expr Expr
}

// Expr represents the AST for a computed permission.
type Expr interface {
	isExpression()
}

// ExprUnion represents an OR operation.
type ExprUnion struct {
	Terms []Expr
}

func (ExprUnion) isExpression() {}

// ExprIntersection represents an AND operation.
type ExprIntersection struct {
	Terms []Expr
}

func (ExprIntersection) isExpression() {}

// ExprDifference represents a subtract operation (Base - Subtract).
type ExprDifference struct {
	LHS Expr
	RHS Expr
}

func (ExprDifference) isExpression() {}

// ExprSelfRef evaluates a local relation on the current instance.
type ExprSelfRef struct {
	Relation string
}

func (ExprSelfRef) isExpression() {}

// ExprTraversal evaluates a permission on instances found via a local relation.
type ExprTraversal struct {
	Relation   string
	Permission string
}

func (ExprTraversal) isExpression() {}
