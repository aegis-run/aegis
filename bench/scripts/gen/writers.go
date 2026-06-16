package gen

import (
	"encoding/csv"
	"encoding/json"
	"os"
)

func WriteTuplesCSV(path string, tuples []TupleRow) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	w.Write([]string{"resource.type", "resource.id", "relation", "subject.instance.type", "subject.instance.id", "subject.permission"})

	for _, t := range tuples {
		w.Write([]string{t.ResourceType, t.ResourceID, t.Relation, t.SubjectType, t.SubjectID, t.SubjectRel})
	}

	return w.Error()
}

func WriteChecksJSONL(path string, checks []Check) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	for _, c := range checks {
		if err := enc.Encode(c); err != nil {
			return err
		}
	}

	return nil
}

func WriteJSON(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
