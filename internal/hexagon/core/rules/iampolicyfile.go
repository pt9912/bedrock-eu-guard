package rules

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// IDIAMPolicyFile ist die Regel-ID der Prüfung eigenständiger IAM-Policy-Dateien.
const IDIAMPolicyFile = "LH-FA-IAM-004"

// IAMPolicyFileState ist ein geladener Dateizustand.
type IAMPolicyFileState struct {
	Path string
	F    *model.IAMPolicyFile
	Err  *model.FileError
}

// IAMPolicyScan bündelt die Suche nach IAM-Policy-Dateien. Searched=false
// heißt: nicht aktiv, keine Findings.
type IAMPolicyScan struct {
	Searched bool
	FindErr  *model.FileError
	Files    []IAMPolicyFileState
}

// CheckIAMPolicyFiles bewertet eigenständige IAM-Policy-Dateien
// (LH-FA-IAM-004) mit der gemeinsamen Bewertung assessBedrockStatements.
// Eine nicht lesbare oder ungültige Datei ist bei eindeutigem Namen
// (iam-policy*.json) ein FAIL, beim breiten Muster *-policy.json nur ein
// WARN. Gültiges JSON ohne Policy-Statement oder ohne Bedrock-Aktion ergibt
// kein Finding.
func CheckIAMPolicyFiles(sc ScanContext, _ model.Config) []model.Finding {
	p := sc.IAMPolicies
	if !p.Searched {
		return nil
	}
	if p.FindErr != nil {
		return []model.Finding{fileErrorFinding(IDIAMPolicyFile, SettingsFile{Path: ".", Err: p.FindErr})}
	}
	var out []model.Finding
	for _, f := range p.Files {
		if f.Err != nil {
			fd := fileErrorFinding(IDIAMPolicyFile, SettingsFile{Path: f.Path, Err: f.Err})
			if !model.StrictIAMPolicyName(f.Path) {
				fd.Severity = model.WARN
			}
			out = append(out, fd)
			continue
		}
		if f.F == nil {
			continue
		}
		for _, fd := range assessBedrockStatements(f.F.Statements) {
			fd.RuleID, fd.Subject, fd.File = IDIAMPolicyFile, f.Path, f.Path
			out = append(out, fd)
		}
	}
	return out
}
