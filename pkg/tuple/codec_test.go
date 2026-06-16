package tuple

import (
	"reflect"
	"testing"

	"github.com/aegis-run/aegis/pkg/assert"
	datav1 "github.com/aegis-run/aegis/proto/aegis/data/v1"
	v1 "github.com/aegis-run/aegis/proto/aegis/v1"
)

func TestInstanceCodec(t *testing.T) {
	pb := &v1.Instance{Type: "document", Id: "doc-1"}
	inst, err := DecodeInstance(pb)
	assert.Err(t, err, nil)
	assert.Equal(t, inst.Type, "document")
	assert.Equal(t, inst.ID, "doc-1")

	encoded := inst.Encode()
	assert.Equal(t, encoded.Type, "document")
	assert.Equal(t, encoded.Id, "doc-1")

	// Test nil Decode
	_, err = DecodeInstance(nil)
	assert.Err(t, err, ErrInstanceRequired)
}

func TestSubjectCodec(t *testing.T) {
	pb := &v1.Subject{
		Instance:   &v1.Instance{Type: "team", Id: "t1"},
		Permission: "member",
	}

	sub, err := DecodeSubject(pb)
	assert.Err(t, err, nil)
	assert.Equal(t, sub.Instance.Type, "team")
	assert.Equal(t, sub.Instance.ID, "t1")
	assert.Equal(t, sub.Permission, "member")

	encoded := sub.Encode()
	assert.Equal(t, encoded.Instance.Type, "team")
	assert.Equal(t, encoded.Instance.Id, "t1")
	assert.Equal(t, encoded.Permission, "member")

	// Test nil Decode
	_, err = DecodeSubject(nil)
	assert.Err(t, err, ErrSubjectRequired)
}

func TestTupleCodec(t *testing.T) {
	pb := &v1.Tuple{
		Resource: &v1.Instance{Type: "folder", Id: "f1"},
		Relation: "viewer",
		Subject: &v1.Subject{
			Instance:   &v1.Instance{Type: "team", Id: "t2"},
			Permission: "editor",
		},
	}

	tpl, err := DecodeTuple(pb)
	assert.Err(t, err, nil)
	assert.Equal(t, tpl.Resource.Type, "folder")
	assert.Equal(t, tpl.Resource.ID, "f1")
	assert.Equal(t, tpl.Relation, "viewer")
	assert.Equal(t, tpl.Subject.Instance.Type, "team")
	assert.Equal(t, tpl.Subject.Instance.ID, "t2")
	assert.Equal(t, tpl.Subject.Permission, "editor")

	encoded := tpl.Encode()
	assert.Equal(t, encoded.Resource.Type, "folder")
	assert.Equal(t, encoded.Resource.Id, "f1")
	assert.Equal(t, encoded.Relation, "viewer")
	assert.Equal(t, encoded.Subject.Instance.Type, "team")
	assert.Equal(t, encoded.Subject.Instance.Id, "t2")
	assert.Equal(t, encoded.Subject.Permission, "editor")

	// Test nil Decode
	_, err = DecodeTuple(nil)
	assert.Err(t, err, ErrTupleRequired)
}

func TestMutationCodec(t *testing.T) {
	pb := &datav1.TupleMutation{
		Operation: datav1.TupleMutation_OPERATION_WRITE,
		Tuple: &v1.Tuple{
			Resource: &v1.Instance{Type: "doc", Id: "1"},
			Relation: "viewer",
			Subject: &v1.Subject{
				Instance: &v1.Instance{Type: "user", Id: "alice"},
			},
		},
	}

	mut, err := DecodeMutation(pb)
	assert.Err(t, err, nil)
	assert.Equal(t, mut.Op, OpWrite)
	assert.Equal(t, mut.Tuple.Resource.Type, "doc")
	assert.Equal(t, mut.Tuple.Resource.ID, "1")
	assert.Equal(t, mut.Tuple.Relation, "viewer")
	assert.Equal(t, mut.Tuple.Subject.Instance.Type, "user")
	assert.Equal(t, mut.Tuple.Subject.Instance.ID, "alice")
	assert.Equal(t, mut.Tuple.Subject.Permission, "")

	// Test nil Decode
	_, err = DecodeMutation(nil)
	assert.Err(t, err, ErrMutationRequired)

	// Test unspecified op
	pbInvalid := &datav1.TupleMutation{Operation: datav1.TupleMutation_OPERATION_UNSPECIFIED}
	_, err = DecodeMutation(pbInvalid)
	assert.Err(t, err, reflect.TypeOf(ErrUnsupportedMutationOp{}))
}

func TestFilterCodec(t *testing.T) {
	// 1. Test Resource filter decoding
	pbResource := &datav1.TupleFilter{
		QueryTarget: &datav1.TupleFilter_Resource{
			Resource: &datav1.TupleFilter_InstanceFilter{
				Type: "doc",
				Id:   "1",
			},
		},
		Relation: "viewer",
	}

	filterRes, err := DecodeFilter(pbResource)
	assert.Err(t, err, nil)
	assert.Equal(t, filterRes.Target, TargetResource)
	assert.Equal(t, filterRes.ResourceType, "doc")
	assert.Equal(t, filterRes.ResourceID, "1")
	assert.Equal(t, filterRes.Relation, "viewer")
	assert.Equal(t, filterRes.SubjectType, "")
	assert.Equal(t, filterRes.SubjectID, "")
	assert.Equal(t, filterRes.SubjectPermission, "")

	// 2. Test Subject filter decoding
	pbSubject := &datav1.TupleFilter{
		QueryTarget: &datav1.TupleFilter_Subject{
			Subject: &datav1.TupleFilter_SubjectFilter{
				Instance: &datav1.TupleFilter_InstanceFilter{
					Type: "team",
					Id:   "t1",
				},
				Permission: "member",
			},
		},
		Relation: "viewer",
	}

	filterSub, err := DecodeFilter(pbSubject)
	assert.Err(t, err, nil)
	assert.Equal(t, filterSub.Target, TargetSubject)
	assert.Equal(t, filterSub.ResourceType, "")
	assert.Equal(t, filterSub.ResourceID, "")
	assert.Equal(t, filterSub.Relation, "viewer")
	assert.Equal(t, filterSub.SubjectType, "team")
	assert.Equal(t, filterSub.SubjectID, "t1")
	assert.Equal(t, filterSub.SubjectPermission, "member")

	// Test nil Decode
	_, err = DecodeFilter(nil)
	assert.Err(t, err, ErrFilterRequired)

	// Test missing query target
	pbInvalid := &datav1.TupleFilter{Relation: "viewer"}
	_, err = DecodeFilter(pbInvalid)
	assert.Err(t, err, ErrInvalidTarget)
}
