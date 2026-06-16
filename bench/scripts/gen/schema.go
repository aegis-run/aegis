package gen

type UserID string

func (id UserID) String() string { return string(id) }

type OrganizationID string

func (id OrganizationID) String() string { return string(id) }

type GroupID string

func (id GroupID) String() string { return string(id) }

type DirectoryID string

func (id DirectoryID) String() string { return string(id) }

type DocumentID string

func (id DocumentID) String() string { return string(id) }

type User struct {
	ID UserID
}

func (u User) String() string { return u.ID.String() }

type Organization struct {
	ID      OrganizationID
	Members []string
	Admins  []string
}

func (o Organization) String() string { return o.ID.String() }

type Group struct {
	ID      GroupID
	Members []string
}

func (g Group) String() string { return g.ID.String() }

type Directory struct {
	ID       DirectoryID
	OrgID    OrganizationID
	ParentID *DirectoryID
	Viewers  []string
	Editors  []string
}

func (d Directory) String() string { return d.ID.String() }

type Document struct {
	ID         DocumentID
	DirID      DirectoryID
	OwnerID    string
	Commenters []string
}

func (d Document) String() string { return d.ID.String() }
