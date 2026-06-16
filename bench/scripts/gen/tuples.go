package gen

func GenerateTuples(ds *Dataset) []TupleRow {
	var rows []TupleRow
	seen := make(map[TupleRow]struct{})

	add := func(rt, ri, rel, st, si, sp string) {
		row := TupleRow{rt, ri, rel, st, si, sp}
		if _, ok := seen[row]; ok {
			return
		}
		seen[row] = struct{}{}
		rows = append(rows, row)
	}

	for _, o := range ds.Orgs {
		for _, a := range o.Admins {
			add("organization", o.String(), "admin", "user", a, "")
		}
		for _, m := range o.Members {
			add("organization", o.String(), "member", "user", m, "")
		}
	}
	for _, g := range ds.Groups {
		for _, m := range g.Members {
			add("group", g.String(), "member", "user", m, "")
		}
	}
	for _, d := range ds.Dirs {
		add("directory", d.String(), "organization", "organization", d.OrgID.String(), "")
		if d.ParentID != nil {
			add("directory", d.String(), "parent", "directory", d.ParentID.String(), "")
		}
		for _, e := range d.Editors {
			add("directory", d.String(), "editor", "user", e, "")
		}
		for _, v := range d.Viewers {
			if len(v) > 5 && v[:5] == "group" {
				add("directory", d.String(), "viewer", "group", v, "member")
			} else {
				add("directory", d.String(), "viewer", "user", v, "")
			}
		}
	}
	for _, d := range ds.Docs {
		add("document", d.String(), "directory", "directory", d.DirID.String(), "")
		add("document", d.String(), "owner", "user", d.OwnerID, "")
		for _, c := range d.Commenters {
			if len(c) > 5 && c[:5] == "group" {
				add("document", d.String(), "commenter", "group", c, "member")
			} else {
				add("document", d.String(), "commenter", "user", c, "")
			}
		}
	}
	return rows
}
