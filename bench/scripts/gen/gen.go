package gen

import (
	"fmt"
	"log"
	"math/rand/v2"
)

type Generator struct {
	cfg    *Config
	r      *rand.Rand
	zUser  *rand.Zipf
	zOrg   *rand.Zipf
	zGroup *rand.Zipf
	zDir   *rand.Zipf
}

func NewGenerator(cfg *Config) *Generator {
	if cfg.Model.Users.Dimension <= 0 ||
		cfg.Model.Organizations.Dimension <= 0 ||
		cfg.Model.Groups.Dimension <= 0 ||
		cfg.Model.Directories.Dimension <= 0 ||
		cfg.Model.Documents.Dimension <= 0 {
		log.Fatal("all model dimensions (users, organizations, groups, directories, documents) must be positive")
	}

	s1, s2 := uint64(42), uint64(0)
	if len(cfg.Seed) > 0 {
		s1 = cfg.Seed[0]
	}
	if len(cfg.Seed) > 1 {
		s2 = cfg.Seed[1]
	}

	pcg := rand.NewPCG(s1, s2)
	r := rand.New(pcg)

	s, v := cfg.Distribution.ZipfS, cfg.Distribution.ZipfV

	return &Generator{
		cfg:    cfg,
		r:      r,
		zUser:  rand.NewZipf(r, s, v, uint64(cfg.Model.Users.Dimension)),
		zOrg:   rand.NewZipf(r, s, v, uint64(cfg.Model.Organizations.Dimension)),
		zGroup: rand.NewZipf(r, s, v, uint64(cfg.Model.Groups.Dimension)),
		zDir:   rand.NewZipf(r, s, v, uint64(cfg.Model.Directories.Dimension)),
	}
}

func (g *Generator) Generate() *GeneratedArtifacts {
	d := &Dataset{}
	d.genUsers(g)
	d.genOrgs(g)
	d.genGroups(g)
	d.genDirs(g)
	d.genDocs(g)
	d.genDeepChain(g)

	return &GeneratedArtifacts{
		Dataset: d,
	}
}

// --- Constructors ---

func (g *Generator) NewUser(id int) User {
	return User{ID: UserID(fmt.Sprintf("user_%d", id))}
}

func (g *Generator) NewOrganization(id int) Organization {
	org := Organization{ID: OrganizationID(fmt.Sprintf("org_%d", id))}

	for range g.cfg.Model.Organizations.Admins.Roll(g.r) {
		org.Admins = append(org.Admins, g.RandUserID())
	}

	for range g.cfg.Model.Organizations.Members.Roll(g.r) {
		org.Members = append(org.Members, g.RandUserID())
	}

	return org
}

func (g *Generator) NewGroup(id int) Group {
	group := Group{ID: GroupID(fmt.Sprintf("group_%d", id))}

	for range g.cfg.Model.Groups.Members.Roll(g.r) {
		group.Members = append(group.Members, g.RandUserID())
	}

	return group
}

func (g *Generator) NewDirectory(id int, org Organization, parentID *DirectoryID) Directory {
	dir := Directory{
		ID:       DirectoryID(fmt.Sprintf("dir_%d", id)),
		OrgID:    org.ID,
		ParentID: parentID,
	}

	// Editors
	for range g.cfg.Model.Directories.Editors.Roll(g.r) {
		if len(org.Members) > 0 {
			dir.Editors = append(dir.Editors, org.Members[g.r.IntN(len(org.Members))])
		}
	}

	// Viewers
	for range g.cfg.Model.Directories.Viewers.Roll(g.r) {
		if g.r.Float64() < g.cfg.Model.Directories.Viewers.IsUserProb {
			dir.Viewers = append(dir.Viewers, g.RandUserID())
		} else {
			dir.Viewers = append(dir.Viewers, g.RandGroupID())
		}
	}

	return dir
}

func (g *Generator) NewDocument(id int, dir Directory, org *Organization) Document {
	doc := Document{
		ID:    DocumentID(fmt.Sprintf("doc_%d", id)),
		DirID: dir.ID,
	}

	// Owner (from Org)
	if org != nil {
		if len(org.Members) > 0 {
			doc.OwnerID = org.Members[g.r.IntN(len(org.Members))]
		} else if len(org.Admins) > 0 {
			doc.OwnerID = org.Admins[g.r.IntN(len(org.Admins))]
		}
	}

	// Commenters
	for range g.cfg.Model.Documents.Commenters.Roll(g.r) {
		if g.r.Float64() < g.cfg.Model.Documents.Commenters.IsUserProb {
			doc.Commenters = append(doc.Commenters, g.RandUserID())
		} else {
			doc.Commenters = append(doc.Commenters, g.RandGroupID())
		}
	}

	return doc
}

// --- Random IDs (Raw) ---

func (g *Generator) RandUserID() string {
	return fmt.Sprintf("user_%d", g.zipfID(g.zUser, g.cfg.Model.Users.Dimension))
}

func (g *Generator) RandOrgID() int {
	return g.zipfID(g.zOrg, g.cfg.Model.Organizations.Dimension)
}

func (g *Generator) RandGroupID() string {
	return fmt.Sprintf("group_%d", g.zipfID(g.zGroup, g.cfg.Model.Groups.Dimension))
}

func (g *Generator) RandDirID() int {
	return g.zipfID(g.zDir, g.cfg.Model.Directories.Dimension)
}

// --- Internal ---

func (g *Generator) zipfID(z *rand.Zipf, bound int) int {
	if bound <= 0 {
		return 0
	}
	for {
		if v := int(z.Uint64()); v < bound {
			return v
		}
	}
}
