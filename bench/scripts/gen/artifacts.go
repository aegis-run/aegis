package gen

type Dataset struct {
	Users  []User
	Orgs   []Organization
	Groups []Group
	Dirs   []Directory
	Docs   []Document

	*DeepChainFixture
}

type DeepChainFixture struct {
	UserID     string
	DocumentID DocumentID
	Depth      int
}

type TupleRow struct {
	ResourceType string
	ResourceID   string
	Relation     string
	SubjectType  string
	SubjectID    string
	SubjectRel   string
}

type Metadata struct {
	Config    Config `json:"config"`
	Artifacts struct {
		TupleCount   int `json:"tuple_count"`
		CheckCount   int `json:"check_count"`
		AllowedCount int `json:"allowed_count"`
		DeniedCount  int `json:"denied_count"`
	} `json:"artifacts"`
	DeepChain *DeepChainFixture `json:"deep_chain"`
}

type GeneratedArtifacts struct {
	Dataset *Dataset
	Tuples  []TupleRow
	Checks  []Check
	Herd    *Check
	Meta    Metadata
}

// type Check struct {
// 	Resource   Instance `json:"resource"`
// 	Permission string   `json:"permission"`
// 	Actor      Instance `json:"actor"`
// 	Expected   bool     `json:"expected"`
// 	Class      string   `json:"class"`
// }

// type Instance struct {
// 	Type string `json:"type"`
// 	ID   string `json:"id"`
// }
