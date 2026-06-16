package gen

import "fmt"

func (ds *Dataset) genUsers(g *Generator) {
	ds.Users = make([]User, 0, g.cfg.Model.Users.Dimension)
	for i := range g.cfg.Model.Users.Dimension {
		ds.Users = append(ds.Users, g.NewUser(i))
	}
}

func (ds *Dataset) genOrgs(g *Generator) {
	ds.Orgs = make([]Organization, 0, g.cfg.Model.Organizations.Dimension)
	for i := range g.cfg.Model.Organizations.Dimension {
		ds.Orgs = append(ds.Orgs, g.NewOrganization(i))
	}
}

func (ds *Dataset) genGroups(g *Generator) {
	ds.Groups = make([]Group, 0, g.cfg.Model.Groups.Dimension)
	for i := range g.cfg.Model.Groups.Dimension {
		ds.Groups = append(ds.Groups, g.NewGroup(i))
	}
}

func (ds *Dataset) genDirs(g *Generator) {
	ds.Dirs = make([]Directory, 0, g.cfg.Model.Directories.Dimension)
	dirsByOrg := make(map[OrganizationID][]DirectoryID)

	for i := range g.cfg.Model.Directories.Dimension {
		org := ds.Orgs[g.RandOrgID()]

		var parentID *DirectoryID
		if i > 0 && g.r.Float64() < g.cfg.Model.Directories.ParentDirProb {
			parents := dirsByOrg[org.ID]
			if len(parents) > 0 {
				id := parents[g.r.IntN(len(parents))]
				parentID = &id
			}
		}

		dir := g.NewDirectory(i, org, parentID)
		dirsByOrg[org.ID] = append(dirsByOrg[org.ID], dir.ID)
		ds.Dirs = append(ds.Dirs, dir)
	}
}

func (ds *Dataset) genDocs(g *Generator) {
	orgByID := make(map[OrganizationID]*Organization, len(ds.Orgs))
	for i := range ds.Orgs {
		orgByID[ds.Orgs[i].ID] = &ds.Orgs[i]
	}

	ds.Docs = make([]Document, 0, g.cfg.Model.Documents.Dimension)
	for i := range g.cfg.Model.Documents.Dimension {
		dir := ds.Dirs[g.RandDirID()]
		doc := g.NewDocument(i, dir, orgByID[dir.OrgID])
		ds.Docs = append(ds.Docs, doc)
	}
}

func (ds *Dataset) genDeepChain(g *Generator) {
	if g.cfg.Model.DeepChainDepth <= 0 {
		return
	}

	var org *Organization
	for i := range ds.Orgs {
		if len(ds.Orgs[i].Members) > 0 {
			org = &ds.Orgs[i]
			break
		}
	}
	if org == nil {
		return
	}

	chainUser := org.Members[0]
	startDirID := g.cfg.Model.Directories.Dimension
	var parentID *DirectoryID

	for i := range g.cfg.Model.DeepChainDepth {
		id := DirectoryID(fmt.Sprintf("dir_%d", startDirID+i))

		dir := Directory{
			ID:       id,
			OrgID:    org.ID,
			ParentID: parentID,
		}

		if i == 0 {
			dir.Viewers = []string{chainUser}
		}

		ds.Dirs = append(ds.Dirs, dir)
		nextParent := id
		parentID = &nextParent
	}

	targetDocID := DocumentID(fmt.Sprintf("doc_%d", g.cfg.Model.Documents.Dimension))

	ds.Docs = append(ds.Docs, Document{
		ID:      targetDocID,
		DirID:   *parentID,
		OwnerID: ds.pickDifferentUser(chainUser),
	})

	ds.DeepChainFixture = &DeepChainFixture{
		UserID:     chainUser,
		DocumentID: targetDocID,
		Depth:      g.cfg.Model.DeepChainDepth,
	}
}

func (ds *Dataset) pickDifferentUser(exclude string) string {
	for _, u := range ds.Users {
		id := u.String()
		if id != exclude {
			return id
		}
	}
	return exclude
}
