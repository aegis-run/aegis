package gen

type DatasetIndex struct {
	UsersByID    map[UserID]*User
	OrgsByID     map[OrganizationID]*Organization
	GroupsByID   map[GroupID]*Group
	DirsByID     map[DirectoryID]*Directory
	DocsByID     map[DocumentID]*Document
	OrgMembers   map[OrganizationID]map[string]bool
	OrgAdmins    map[OrganizationID]map[string]bool
	GroupMembers map[GroupID]map[string]bool
}

func NewDatasetIndex(ds *Dataset) *DatasetIndex {
	idx := &DatasetIndex{
		UsersByID:    make(map[UserID]*User, len(ds.Users)),
		OrgsByID:     make(map[OrganizationID]*Organization, len(ds.Orgs)),
		GroupsByID:   make(map[GroupID]*Group, len(ds.Groups)),
		DirsByID:     make(map[DirectoryID]*Directory, len(ds.Dirs)),
		DocsByID:     make(map[DocumentID]*Document, len(ds.Docs)),
		OrgMembers:   make(map[OrganizationID]map[string]bool, len(ds.Orgs)),
		OrgAdmins:    make(map[OrganizationID]map[string]bool, len(ds.Orgs)),
		GroupMembers: make(map[GroupID]map[string]bool, len(ds.Groups)),
	}

	for i := range ds.Users {
		idx.UsersByID[ds.Users[i].ID] = &ds.Users[i]
	}

	for i := range ds.Orgs {
		org := &ds.Orgs[i]
		idx.OrgsByID[org.ID] = org

		members := make(map[string]bool, len(org.Members))
		for _, m := range org.Members {
			members[m] = true
		}
		idx.OrgMembers[org.ID] = members

		admins := make(map[string]bool, len(org.Admins))
		for _, a := range org.Admins {
			admins[a] = true
		}
		idx.OrgAdmins[org.ID] = admins
	}

	for i := range ds.Groups {
		group := &ds.Groups[i]
		idx.GroupsByID[group.ID] = group

		members := make(map[string]bool, len(group.Members))
		for _, m := range group.Members {
			members[m] = true
		}
		idx.GroupMembers[group.ID] = members
	}

	for i := range ds.Dirs {
		idx.DirsByID[ds.Dirs[i].ID] = &ds.Dirs[i]
	}

	for i := range ds.Docs {
		idx.DocsByID[ds.Docs[i].ID] = &ds.Docs[i]
	}

	return idx
}
