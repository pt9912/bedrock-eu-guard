// Package fakeiampolicy ist ein Test-Double für driven.IAMPolicyFileReader.
package fakeiampolicy

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// Reader liefert vorbereitete IAM-Policy-Dateien.
type Reader struct {
	Files   map[string]*model.IAMPolicyFile
	Errs    map[string]error
	FindErr error
}

// Find liefert die Pfade aus Files und Errs (unsortiert; Scan sortiert).
func (r Reader) Find() ([]string, error) {
	if r.FindErr != nil {
		return nil, r.FindErr
	}
	var out []string
	for p := range r.Files {
		out = append(out, p)
	}
	for p := range r.Errs {
		out = append(out, p)
	}
	return out, nil
}

// Read implementiert driven.IAMPolicyFileReader.
func (r Reader) Read(rel string) (*model.IAMPolicyFile, error) {
	if err := r.Errs[rel]; err != nil {
		return nil, err
	}
	return r.Files[rel], nil
}
