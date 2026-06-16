package gen

import (
	"strings"
)

type Evaluator struct {
	idx *DatasetIndex
}

func NewEvaluator(idx *DatasetIndex) *Evaluator {
	return &Evaluator{idx: idx}
}

func (e *Evaluator) Check(c Check) bool {
	return e.Evaluate(c.Resource.Type, c.Resource.ID, c.Permission, c.Actor.Type, c.Actor.ID)
}

func (e *Evaluator) Evaluate(rType, rID, permission, aType, aID string) bool {
	switch rType {
	case "group":
		return e.evalGroup(rID, permission, aType, aID)
	case "organization":
		return e.evalOrg(rID, permission, aType, aID)
	case "directory":
		return e.evalDir(rID, permission, aType, aID)
	case "document":
		return e.evalDoc(rID, permission, aType, aID)
	}
	return false
}

func (e *Evaluator) evalGroup(rID, permission, aType, aID string) bool {
	if permission != "member" || aType != "user" {
		return false
	}
	if m, ok := e.idx.GroupMembers[GroupID(rID)]; ok {
		return m[aID]
	}
	return false
}

func (e *Evaluator) evalOrg(rID, permission, aType, aID string) bool {
	id := OrganizationID(rID)
	switch permission {
	case "admin":
		if aType != "user" {
			return false
		}
		if a, ok := e.idx.OrgAdmins[id]; ok {
			return a[aID]
		}
	case "member":
		if aType == "user" {
			if m, ok := e.idx.OrgMembers[id]; ok {
				return m[aID]
			}
		}
		// In our current generation, we don't have group members in orgs yet,
		// but if we did, we'd check them here.
	}
	return false
}

func (e *Evaluator) evalDir(rID, permission, aType, aID string) bool {
	id := DirectoryID(rID)
	dir, ok := e.idx.DirsByID[id]
	if !ok {
		return false
	}

	switch permission {
	case "view":
		// def view = .viewer | .editor | .parent.view;
		if e.isDirViewer(dir, aType, aID) {
			return true
		}
		if e.isDirEditor(dir, aType, aID) {
			return true
		}
		if dir.ParentID != nil {
			return e.evalDir(dir.ParentID.String(), "view", aType, aID)
		}
	case "edit":
		// def edit = .editor & .organization.member;
		if e.isDirEditor(dir, aType, aID) {
			return e.evalOrg(dir.OrgID.String(), "member", aType, aID)
		}
	case "delete":
		// def delete = .organization.admin - .viewer;
		if e.evalOrg(dir.OrgID.String(), "admin", aType, aID) {
			return !e.isDirViewer(dir, aType, aID)
		}
	}
	return false
}

func (e *Evaluator) evalDoc(rID, permission, aType, aID string) bool {
	id := DocumentID(rID)
	doc, ok := e.idx.DocsByID[id]
	if !ok {
		return false
	}

	switch permission {
	case "view":
		// def view = .owner | .commenter | .directory.view;
		if aType == "user" && doc.OwnerID == aID {
			return true
		}
		if e.isDocCommenter(doc, aType, aID) {
			return true
		}
		return e.evalDir(doc.DirID.String(), "view", aType, aID)
	case "edit":
		// def edit = .owner | .directory.edit;
		if aType == "user" && doc.OwnerID == aID {
			return true
		}
		return e.evalDir(doc.DirID.String(), "edit", aType, aID)
	case "comment":
		// def comment = .commenter | .owner;
		if aType == "user" && doc.OwnerID == aID {
			return true
		}
		return e.isDocCommenter(doc, aType, aID)
	case "delete":
		// def delete = .owner & .directory.delete;
		if aType == "user" && doc.OwnerID == aID {
			return e.evalDir(doc.DirID.String(), "delete", aType, aID)
		}
	}
	return false
}

// --- Helpers ---

func (e *Evaluator) isDirViewer(dir *Directory, aType, aID string) bool {
	for _, v := range dir.Viewers {
		if e.matches(v, aType, aID) {
			return true
		}
	}
	return false
}

func (e *Evaluator) isDirEditor(dir *Directory, aType, aID string) bool {
	for _, ed := range dir.Editors {
		if aType == "user" && ed == aID {
			return true
		}
	}
	return false
}

func (e *Evaluator) isDocCommenter(doc *Document, aType, aID string) bool {
	for _, c := range doc.Commenters {
		if e.matches(c, aType, aID) {
			return true
		}
	}
	return false
}

func (e *Evaluator) matches(entry, aType, aID string) bool {
	if aType == "user" && entry == aID {
		return true
	}
	if strings.HasPrefix(entry, "group_") {
		// Check if user is in group
		if aType == "user" {
			if m, ok := e.idx.GroupMembers[GroupID(entry)]; ok {
				return m[aID]
			}
		}
	}
	return false
}
