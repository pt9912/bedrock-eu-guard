package rules

import (
	"fmt"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Regel-IDs der Terraform-Regeln.
const (
	IDTerraformFiles  = "LH-FA-TF-001"
	IDTerraformRegion = "LH-FA-TF-002"
	IDTerraformIAM    = "LH-FA-TF-003"
	IDTerraformModels = "LH-FA-TF-004"
)

// TerraformFileState ist ein geladener Dateizustand.
type TerraformFileState struct {
	Path string
	F    *model.TerraformFile
	Err  *model.FileError
}

// TerraformScan bündelt die Terraform-Suche. Searched=false heißt: Modus
// nicht aktiv, keine Terraform-Findings.
type TerraformScan struct {
	Searched bool
	FindErr  *model.FileError
	Files    []TerraformFileState
}

// CheckTerraformFiles meldet, ob Terraform-Dateien gefunden wurden
// (LH-FA-TF-001): PASS bei Fund, WARN ohne Fund; kaputte Dateien sind FAIL.
func CheckTerraformFiles(sc ScanContext, _ model.Config) []model.Finding {
	t := sc.Terraform
	if !t.Searched {
		return nil
	}
	if t.FindErr != nil {
		return []model.Finding{fileErrorFinding(IDTerraformFiles, SettingsFile{Path: ".", Err: t.FindErr})}
	}
	var out []model.Finding
	ok := 0
	for _, f := range t.Files {
		if f.Err != nil {
			out = append(out, fileErrorFinding(IDTerraformFiles, SettingsFile{Path: f.Path, Err: f.Err}))
			continue
		}
		if f.F != nil {
			ok++
		}
	}
	switch {
	case ok > 0:
		out = append([]model.Finding{{RuleID: IDTerraformFiles, Severity: model.PASS, Reason: model.ReasonTerraformFound,
			Subject: "terraform", Message: fmt.Sprintf("%d Terraform-Datei(en) gefunden und gelesen", ok)}}, out...)
	case len(t.Files) == 0:
		out = append(out, model.Finding{RuleID: IDTerraformFiles, Severity: model.WARN, Reason: model.ReasonNoTerraform,
			Subject: "terraform", Message: "Keine Terraform-Dateien gefunden",
			Recommendation: "Scan-Root prüfen"})
	}
	return out
}

// tfFiles liefert die gelesenen Dateien (ohne fehlende/kaputte).
func tfFiles(sc ScanContext) []TerraformFileState {
	var out []TerraformFileState
	for _, f := range sc.Terraform.Files {
		if f.Err == nil && f.F != nil {
			out = append(out, f)
		}
	}
	return out
}

// CheckTerraformRegion prüft AWS-Provider-Regionen je Datei
// (LH-FA-TF-002, SPEC-011): FAIL bei nicht erlaubter Region, WARN bei
// dynamischer oder fehlender Region, PASS sonst.
func CheckTerraformRegion(sc ScanContext, cfg model.Config) []model.Finding {
	var out []model.Finding
	for _, f := range tfFiles(sc) {
		if len(f.F.Providers) == 0 {
			continue
		}
		fd := model.Finding{RuleID: IDTerraformRegion, Subject: f.Path, File: f.Path}
		var bad, soft *model.AWSProvider
		for i := range f.F.Providers {
			p := &f.F.Providers[i]
			switch {
			case p.Region != "" && !cfg.RegionAllowed(p.Region):
				if bad == nil {
					bad = p
				}
			case p.Region == "":
				if soft == nil {
					soft = p
				}
			}
		}
		switch {
		case bad != nil:
			fd.Severity, fd.Reason, fd.Line = model.FAIL, model.ReasonRegionNotAllowed, bad.Line
			fd.Message = "AWS-Provider-Region " + bad.Region + " ist nicht erlaubt"
			fd.Recommendation = "Provider-Region auf eine erlaubte EU-Region setzen"
		case soft != nil && soft.RegionDynamic:
			fd.Severity, fd.Reason, fd.Line = model.WARN, model.ReasonRegionDynamic, soft.Line
			fd.Message = "AWS-Provider-Region ist dynamisch und nicht auflösbar; manuell prüfen"
		case soft != nil:
			fd.Severity, fd.Reason, fd.Line = model.WARN, model.ReasonRegionMissing, soft.Line
			fd.Message = "AWS-Provider ohne region-Attribut; Region stammt aus der Umgebung"
			fd.Recommendation = "region explizit setzen"
		default:
			fd.Severity, fd.Reason = model.PASS, model.ReasonOK
			fd.Message = "Nur erlaubte Provider-Regionen"
		}
		out = append(out, fd)
	}
	return out
}

// CheckTerraformIAM prüft Bedrock-IAM-Statements (LH-FA-TF-003) mit der
// gemeinsamen Bewertung assessBedrockStatements.
func CheckTerraformIAM(sc ScanContext, _ model.Config) []model.Finding {
	var out []model.Finding
	for _, f := range tfFiles(sc) {
		for _, fd := range assessBedrockStatements(f.F.Statements) {
			fd.RuleID, fd.Subject, fd.File = IDTerraformIAM, f.Path, f.Path
			out = append(out, fd)
		}
	}
	return out
}

// CheckTerraformModels klassifiziert gefundene Modell-/Profil-IDs je Datei
// (LH-FA-TF-004, SPEC-010): FAIL bei blockierten, WARN bei
// unklassifizierbaren, PASS bei nur erlaubten IDs; ohne Funde kein Finding.
func CheckTerraformModels(sc ScanContext, cfg model.Config) []model.Finding {
	var out []model.Finding
	for _, f := range tfFiles(sc) {
		if len(f.F.ModelRefs) == 0 {
			continue
		}
		var blocked, unclass []model.ModelRef
		for _, r := range f.F.ModelRefs {
			switch Classify(r.Value, cfg) {
			case model.ModelBlocked:
				blocked = append(blocked, r)
			case model.ModelUnclassified:
				unclass = append(unclass, r)
			}
		}
		fd := model.Finding{RuleID: IDTerraformModels, Subject: f.Path, File: f.Path}
		switch {
		case len(blocked) > 0:
			fd.Severity, fd.Reason, fd.Line = model.FAIL, model.ReasonModelBlocked, blocked[0].Line
			fd.Message = fmt.Sprintf("%d nicht erlaubte Modell-/Profil-ID(s), z. B. %s", len(blocked), blocked[0].Value)
			fd.Recommendation = "EU-Profil-IDs (eu.*) verwenden"
		case len(unclass) > 0:
			fd.Severity, fd.Reason, fd.Line = model.WARN, model.ReasonModelUnclassified, unclass[0].Line
			fd.Message = fmt.Sprintf("%d nicht klassifizierbare Modell-ID(s), z. B. %s", len(unclass), unclass[0].Value)
		default:
			fd.Severity, fd.Reason = model.PASS, model.ReasonOK
			fd.Message = "Nur erlaubte Modell-/Profil-IDs"
		}
		out = append(out, fd)
	}
	return out
}
