package gen

import (
	"fmt"
)

func GenerateWorkload(g *Generator, ds *Dataset, idx *DatasetIndex, eval *Evaluator) ([]Check, *Check) {
	total := g.cfg.Workload.Checks
	allowedRatio := g.cfg.Workload.AllowedRatio
	if allowedRatio == 0 {
		allowedRatio = 0.7 // Default
	}

	allowedN := int(float64(total) * allowedRatio)
	// deniedN := total - allowedN

	var checks []Check

	// 1. Generate Allowed Checks
	for len(checks) < allowedN {
		// Randomly pick a document
		doc := ds.Docs[g.r.IntN(len(ds.Docs))]

		var check Check
		check.Resource = ResourceRef{Type: "document", ID: doc.String()}

		// Pick a class of allowed access
		switch g.r.IntN(4) {
		case 0: // Owner
			check.Permission = "view"
			check.Actor = ResourceRef{Type: "user", ID: doc.OwnerID}
			check.Class = "doc_owner_view"
		case 1: // Commenter
			if len(doc.Commenters) > 0 {
				c := doc.Commenters[g.r.IntN(len(doc.Commenters))]
				check.Permission = "comment"
				if len(c) > 5 && c[:5] == "group" {
					// We need a user from this group
					gID := GroupID(c)
					// I'll use the index
					group := idx.GroupsByID[gID]
					if len(group.Members) > 0 {
						check.Actor = ResourceRef{Type: "user", ID: group.Members[g.r.IntN(len(group.Members))]}
					} else {
						continue
					}
				} else {
					check.Actor = ResourceRef{Type: "user", ID: c}
				}
				check.Class = "doc_commenter_comment"
			} else {
				continue
			}
		case 2: // Directory Viewer
			dir := idx.DirsByID[doc.DirID]
			if len(dir.Viewers) > 0 {
				v := dir.Viewers[g.r.IntN(len(dir.Viewers))]
				check.Permission = "view"
				if len(v) > 5 && v[:5] == "group" {
					group := idx.GroupsByID[GroupID(v)]
					if len(group.Members) > 0 {
						check.Actor = ResourceRef{Type: "user", ID: group.Members[g.r.IntN(len(group.Members))]}
					} else {
						continue
					}
				} else {
					check.Actor = ResourceRef{Type: "user", ID: v}
				}
				check.Class = "doc_dir_viewer_view"
			} else {
				continue
			}
		case 3: // Directory Editor
			dir := idx.DirsByID[doc.DirID]
			if len(dir.Editors) > 0 {
				e := dir.Editors[g.r.IntN(len(dir.Editors))]
				check.Permission = "edit"
				// Check if they are also org member
				if eval.Evaluate("organization", dir.OrgID.String(), "member", "user", e) {
					check.Actor = ResourceRef{Type: "user", ID: e}
					check.Class = "doc_dir_editor_edit"
				} else {
					continue
				}
			} else {
				continue
			}
		}

		check.Expected = true
		if eval.Check(check) {
			checks = append(checks, check)
		}
	}

	// 2. Generate Denied Checks
	for len(checks) < total {
		doc := ds.Docs[g.r.IntN(len(ds.Docs))]
		user := ds.Users[g.r.IntN(len(ds.Users))]

		check := Check{
			Resource:   ResourceRef{Type: "document", ID: doc.String()},
			Permission: "edit", // Most are denied edit
			Actor:      ResourceRef{Type: "user", ID: user.String()},
			Expected:   false,
			Class:      "doc_random_user_denied_edit",
		}

		// Maybe try some near misses
		if g.r.Float64() < 0.2 {
			check.Permission = "delete" // Harder to get delete
			check.Class = "doc_near_miss_denied_delete"
		}

		if !eval.Check(check) {
			checks = append(checks, check)
		}
	}

	// 3. Generate Herd Check
	var herd *Check
	if ds.DeepChainFixture != nil {
		herd = &Check{
			Resource:   ResourceRef{Type: "document", ID: ds.DeepChainFixture.DocumentID.String()},
			Permission: "view",
			Actor:      ResourceRef{Type: "user", ID: ds.DeepChainFixture.UserID},
			Expected:   true,
			Class:      "deep_chain_herd",
		}
		if !eval.Check(*herd) {
			fmt.Printf("WARNING: Herd check failed evaluation! User=%s Doc=%s\n", herd.Actor.ID, herd.Resource.ID)
		}
	}

	// 4. Shuffle
	g.r.Shuffle(len(checks), func(i, j int) {
		checks[i], checks[j] = checks[j], checks[i]
	})

	return checks, herd
}
