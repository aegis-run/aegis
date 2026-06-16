package tuple

import (
	datav1 "github.com/aegis-run/aegis/proto/aegis/data/v1"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

// DecodeInstance decodes a protobuf Instance into a tuple.Instance.
func DecodeInstance(pb *v1.Instance) (Instance, error) {
	if pb == nil {
		return Instance{}, ErrInstanceRequired
	}
	return Instance{
		Type: pb.Type,
		ID:   pb.Id,
	}, nil
}

// Encode encodes a tuple.Instance into a protobuf Instance.
func (i Instance) Encode() *v1.Instance {
	return &v1.Instance{
		Type: i.Type,
		Id:   i.ID,
	}
}

// DecodeSubject decodes a protobuf Subject into a tuple.Subject.
func DecodeSubject(pb *v1.Subject) (Subject, error) {
	if pb == nil {
		return Subject{}, ErrSubjectRequired
	}

	instance, err := DecodeInstance(pb.Instance)
	if err != nil {
		return Subject{}, err
	}

	return Subject{
		Instance:   instance,
		Permission: pb.Permission,
	}, nil
}

// Encode encodes a tuple.Subject into a protobuf Subject.
func (s Subject) Encode() *v1.Subject {
	return &v1.Subject{
		Instance:   s.Instance.Encode(),
		Permission: s.Permission,
	}
}

// DecodeTuple decodes a protobuf Tuple into a tuple.Tuple.
func DecodeTuple(pb *v1.Tuple) (Tuple, error) {
	if pb == nil {
		return Tuple{}, ErrTupleRequired
	}

	resource, err := DecodeInstance(pb.Resource)
	if err != nil {
		return Tuple{}, err
	}

	subject, err := DecodeSubject(pb.Subject)
	if err != nil {
		return Tuple{}, err
	}

	return Tuple{
		Resource: resource,
		Relation: pb.Relation,
		Subject:  subject,
	}, nil
}

// Encode encodes a tuple.Tuple into a protobuf Tuple.
func (t Tuple) Encode() *v1.Tuple {
	return &v1.Tuple{
		Resource: t.Resource.Encode(),
		Relation: t.Relation,
		Subject:  t.Subject.Encode(),
	}
}

// DecodeMutation decodes a protobuf TupleMutation into a tuple.TupleMutation.
func DecodeMutation(pb *datav1.TupleMutation) (TupleMutation, error) {
	if pb == nil {
		return TupleMutation{}, ErrMutationRequired
	}

	var op MutationOp
	switch pb.Operation {
	case datav1.TupleMutation_OPERATION_UNSPECIFIED:
		return TupleMutation{}, ErrUnsupportedMutationOp{
			Op: pb.Operation.String(),
		}
	case datav1.TupleMutation_OPERATION_WRITE:
		op = OpWrite
	case datav1.TupleMutation_OPERATION_DELETE:
		op = OpDelete
	default:
		return TupleMutation{}, ErrUnsupportedMutationOp{
			Op: pb.Operation.String(),
		}
	}

	tuple, err := DecodeTuple(pb.Tuple)
	if err != nil {
		return TupleMutation{}, err
	}

	return TupleMutation{
		Op:    op,
		Tuple: tuple,
	}, nil
}

// DecodeFilter decodes a protobuf TupleFilter into a tuple.TupleFilter.
func DecodeFilter(pb *datav1.TupleFilter) (TupleFilter, error) {
	if pb == nil {
		return TupleFilter{}, ErrFilterRequired
	}

	var target QueryTarget
	var rType, rID, sType, sID, sPerm string

	switch t := pb.QueryTarget.(type) {
	case *datav1.TupleFilter_Resource:
		target = TargetResource
		if t.Resource != nil {
			rType = t.Resource.Type
			rID = t.Resource.Id
		}
	case *datav1.TupleFilter_Subject:
		target = TargetSubject
		if t.Subject != nil {
			sPerm = t.Subject.Permission
			if t.Subject.Instance != nil {
				sType = t.Subject.Instance.Type
				sID = t.Subject.Instance.Id
			}
		}
	default:
		return TupleFilter{}, ErrInvalidTarget
	}

	return TupleFilter{
		Target:            target,
		ResourceType:      rType,
		ResourceID:        rID,
		Relation:          pb.Relation,
		SubjectType:       sType,
		SubjectID:         sID,
		SubjectPermission: sPerm,
	}, nil
}
