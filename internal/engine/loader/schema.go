package loader

import (
	"context"
	"fmt"
	"unsafe"

	"golang.org/x/sync/singleflight"

	"github.com/aegis-run/aegis/internal/datalayer"
	"github.com/aegis-run/aegis/pkg/cache"
	cacheCore "github.com/aegis-run/aegis/pkg/cache/core"
	"github.com/aegis-run/aegis/pkg/schema"
	"github.com/aegis-run/aegis/pkg/telemetry"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

func cacheKey(hash schema.Hash, name string) string {
	return hash.Hex() + "|" + name
}

// Schema provides in-memory caching and O(1) lookups for the engine's AST evaluation.
// It translates the raw protobuf schemas into internal domain models and guarantees
// that evaluations are pinned to a specific schema hash.
type Schema interface {
	GetType(ctx context.Context, typeName string, hash schema.Hash) (*schema.Type, error)
}

type manager struct {
	store datalayer.Schema
	cache cacheCore.Cache[*schema.Type]
	sg    singleflight.Group
}

// NewSchema creates a new hash-pinned schema manager that wraps the datalayer and cache.
func NewSchema(store datalayer.Schema, cache cacheCore.Cache[*schema.Type]) Schema {
	return &manager{
		store: store,
		cache: cache,
	}
}

func NewCache(cfg *cacheCore.Config) (cacheCore.Cache[*schema.Type], error) {
	return cache.New[*schema.Type](cfg)
}

// GetType retrieves a TypeDefinition by name for a specific frozen schema hash.
func (m *manager) GetType(
	ctx context.Context,
	name string,
	hash schema.Hash,
) (_ *schema.Type, err error) {
	ctx, span := telemetry.Start(ctx, "engine.schema.get")
	defer telemetry.End(span, &err)

	key := cacheKey(hash, name)

	if val, ok := m.cache.Get(key); ok {
		return val, nil
	}

	res, err, _ := m.sg.Do(hash.Hex(), func() (any, error) {
		if val, ok := m.cache.Get(key); ok {
			return val, nil
		}

		version, err := m.store.ReadByHash(ctx, hash)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema by hash %s: %w", hash.Hex(), err)
		}

		pbSchema, err := schema.Decode(version.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode schema payload for hash %s: %w", hash.Hex(), err)
		}

		var target *schema.Type
		for _, def := range pbSchema.GetTypes() {
			mappedDef := decodeTypeDef(def)
			m.cache.Set(cacheKey(hash, def.Name), mappedDef, int64(unsafe.Sizeof(*mappedDef)), 0)

			if def.Name == name {
				target = mappedDef
			}
		}
		m.cache.Wait()

		if target != nil {
			return target, nil
		}

		return nil, fmt.Errorf("type definition '%s' not found in schema %s", name, hash.Hex())
	})

	if err != nil {
		return nil, err
	}

	return res.(*schema.Type), nil
}

func decodeTypeDef(pb *v1.TypeDefinition) *schema.Type {
	def := &schema.Type{
		Name:        pb.Name,
		Relations:   make(map[string]*schema.Relation),
		Permissions: make(map[string]*schema.Permission),
	}

	for _, rel := range pb.GetRelations() {
		def.Relations[rel.Name] = decodeRelation(rel)
	}

	for _, perm := range pb.GetPermissions() {
		def.Permissions[perm.Name] = decodePermission(perm)
	}

	return def
}

func decodeRelation(pb *v1.Relation) *schema.Relation {
	var actors []schema.ActorType
	for _, a := range pb.GetActors() {
		switch act := a.GetActor().(type) {
		case *v1.ActorType_Direct:
			actors = append(actors, schema.DirectActorType{Type: act.Direct})
		case *v1.ActorType_Userset:
			actors = append(actors, schema.UsersetActorType{
				Type:     act.Userset.Type,
				Relation: act.Userset.Member,
			})
		}
	}
	return &schema.Relation{
		Name:       pb.Name,
		ActorTypes: actors,
	}
}

func decodePermission(pb *v1.Permission) *schema.Permission {
	return &schema.Permission{
		Name: pb.Name,
		Expr: decodeExpression(pb.GetExpr()),
	}
}

func decodeExpression(pb *v1.Expression) schema.Expr {
	if pb == nil {
		return nil
	}

	switch kind := pb.GetKind().(type) {
	case *v1.Expression_Union:
		var terms []schema.Expr
		for _, e := range kind.Union.GetTerms() {
			terms = append(terms, decodeExpression(e))
		}
		return schema.ExprUnion{Terms: terms}

	case *v1.Expression_Intersection:
		var terms []schema.Expr
		for _, e := range kind.Intersection.GetTerms() {
			terms = append(terms, decodeExpression(e))
		}
		return schema.ExprIntersection{Terms: terms}

	case *v1.Expression_Difference:
		return schema.ExprDifference{
			LHS: decodeExpression(kind.Difference.GetLhs()),
			RHS: decodeExpression(kind.Difference.GetRhs()),
		}

	case *v1.Expression_Term:
		switch term := kind.Term.GetTerm().(type) {
		case *v1.TermExpr_SelfRef:
			return schema.ExprSelfRef{Relation: term.SelfRef.GetRelation()}
		case *v1.TermExpr_Traversal:
			return schema.ExprTraversal{
				Relation:   term.Traversal.GetRelation(),
				Permission: term.Traversal.GetPermission(),
			}
		}
	}
	return nil
}
