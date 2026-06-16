package tuple

// Instance represents an atomic, uniquely identifiable object of a specific type.
// Used for resources, subjects, and actors.
type Instance struct {
	Type string
	ID   string
}

// Subject represents the target of a relationship assignment.
// Can be a direct Instance, or a Userset reference containing a nested permission.
type Subject struct {
	Instance   Instance
	Permission string // Empty string represents a direct subject instance
}

func (s *Subject) IsDirect() bool {
	return s.Permission == ""
}

func (s *Subject) IsUserset() bool {
	return s.Permission != ""
}

// Tuple is the fundamental relationship fact populating the ReBAC graph.
// Maps a Resource Instance via a direct Relation to a Subject.
type Tuple struct {
	Resource Instance
	Relation string
	Subject  Subject
}

// MutationOp defines the operation type: WRITE or DELETE.
type MutationOp int

const (
	OpWrite = iota + 1
	OpDelete
)

// TupleMutation represents an atomic change operation applied to a Tuple.
type TupleMutation struct {
	Op    MutationOp
	Tuple Tuple
}

// QueryTarget defines whether the query filters by Resource properties
// (forward lookup) or Subject properties (reverse lookup).
type QueryTarget int

const (
	TargetResource QueryTarget = iota + 1
	TargetSubject
)

// TupleFilter represents a query pattern to find stored relationship tuples.
type TupleFilter struct {
	Target            QueryTarget
	ResourceType      string
	ResourceID        string
	Relation          string
	SubjectType       string
	SubjectID         string
	SubjectPermission string
}

func ResourceFilter(res Instance, relation string) TupleFilter {
	return TupleFilter{
		Target:       TargetResource,
		ResourceType: res.Type,
		ResourceID:   res.ID,
		Relation:     relation,
	}
}

func SubjectFilter(subject Subject, relation string) TupleFilter {
	return TupleFilter{
		Target:            TargetSubject,
		SubjectType:       subject.Instance.Type,
		SubjectID:         subject.Instance.ID,
		SubjectPermission: subject.Permission,
		Relation:          relation,
	}
}
