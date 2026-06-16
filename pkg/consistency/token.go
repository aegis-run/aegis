package consistency

import (
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

// Token represents an abstract, database-agnostic causal consistency token.
// Implementations must be comparable within their respective database backends.
type Token interface {
	// Compare returns:
	//   -1 if this token is strictly older than the other token.
	//    0 if this token is equal to the other token.
	//    1 if this token is strictly newer than the other token.
	// An error is returned if the token types are incompatible.
	Compare(other Token) (int, error)

	// String returns the opaque string representation of the token
	// suitable for public API exposure.
	String() string
}

// Encode converts any consistency.Token to its opaque protobuf representation.
func Encode(t Token) *v1.ConsistencyToken {
	if t == nil {
		return nil
	}
	return &v1.ConsistencyToken{
		Token: t.String(),
	}
}

// Strategy defines the database-agnostic causal consistency strategies.
type Strategy int

const (
	// StrategyMinimizeLatency allows read replica stale reads.
	StrategyMinimizeLatency Strategy = iota
	// StrategyFresherThan ensures read is at least as fresh as the provided revision token.
	StrategyFresherThan
	// StrategyFullyConsistent ensures real-time fresh reads from the primary write replica.
	StrategyFullyConsistent
)

// Requirement encapsulates a database-agnostic causal consistency requirement.
type Requirement struct {
	Strategy Strategy
	Token    string
}

// DecodeRequirement maps a Protobuf Consistency message to a database-agnostic Requirement.
func DecodeRequirement(pb *v1.Consistency) Requirement {
	if pb == nil {
		return Requirement{Strategy: StrategyMinimizeLatency}
	}

	switch s := pb.GetStrategy().(type) {
	case *v1.Consistency_FresherThan:
		var token string
		if s.FresherThan != nil {
			token = s.FresherThan.GetToken()
		}
		return Requirement{
			Strategy: StrategyFresherThan,
			Token:    token,
		}
	case *v1.Consistency_FullyConsistent:
		return Requirement{Strategy: StrategyFullyConsistent}
	case *v1.Consistency_MinimizeLatency:
		return Requirement{Strategy: StrategyMinimizeLatency}
	default:
		return Requirement{Strategy: StrategyMinimizeLatency}
	}
}
