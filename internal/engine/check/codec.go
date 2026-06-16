package check

import (
	"github.com/aegis-run/aegis/pkg/tuple"
	v1 "github.com/aegis-run/aegis/proto/aegis/authz/v1"
	corev1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

// Request represents the internal domain model for an evaluation operation.
type Request struct {
	Resource   tuple.Instance
	Permission string
	Actor      tuple.Instance
}

func NewRequest(resource tuple.Instance, permission string, actor tuple.Instance) *Request {
	return &Request{
		Resource:   resource,
		Permission: permission,
		Actor:      actor,
	}
}

type Decision int

const (
	decisionUnknown Decision = iota
	DecisionAllowed
	DecisionDenied
)

// Response represents the internal domain model for an evaluation result.
type Response struct {
	Decision Decision
}

func NewResponse(decision Decision) *Response {
	return &Response{Decision: decision}
}

func Allowed() *Response {
	return &Response{Decision: DecisionAllowed}
}

func Denied() *Response {
	return &Response{Decision: DecisionDenied}
}

func (r *Response) IsAllowed() bool {
	return r != nil && r.Decision == DecisionAllowed
}

func DecodeRequest(req *v1.AuthorizeRequest) *Request {
	if req == nil {
		return nil
	}

	var resource tuple.Instance
	if req.Resource != nil {
		resource = tuple.Instance{
			Type: req.Resource.Type,
			ID:   req.Resource.Id,
		}
	}

	var actor tuple.Instance
	if req.Actor != nil {
		actor = tuple.Instance{
			Type: req.Actor.Type,
			ID:   req.Actor.Id,
		}
	}

	return NewRequest(resource, req.Permission, actor)
}

func (r *Response) Encode(token *corev1.ConsistencyToken) *v1.AuthorizeResponse {
	if r == nil {
		return nil
	}

	var d v1.AuthorizeResponse_Decision
	switch r.Decision {
	case decisionUnknown:
		d = v1.AuthorizeResponse_DECISION_UNSPECIFIED
	case DecisionAllowed:
		d = v1.AuthorizeResponse_DECISION_ALLOWED
	case DecisionDenied:
		d = v1.AuthorizeResponse_DECISION_DENIED
	default:
		d = v1.AuthorizeResponse_DECISION_UNSPECIFIED
	}

	return &v1.AuthorizeResponse{
		Decision:  d,
		CheckedAt: token,
	}
}
