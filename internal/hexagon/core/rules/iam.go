package rules

import (
	"fmt"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// Regel-IDs der IAM-Regeln.
const (
	IDIAMRights    = "LH-FA-IAM-001"
	IDIAMBroad     = "LH-FA-IAM-002"
	IDIAMRegionCnd = "LH-FA-IAM-003"
)

func hasWildcardAction(s model.IAMStatement) bool {
	for _, a := range s.Actions {
		l := strings.ToLower(a)
		if l == "*" || l == "bedrock:*" || strings.HasPrefix(l, "bedrock:invokemodel") {
			return true
		}
	}
	return false
}

func touchesBedrock(s model.IAMStatement) bool {
	for _, a := range s.Actions {
		l := strings.ToLower(a)
		if l == "*" || strings.HasPrefix(l, "bedrock:") {
			return true
		}
	}
	return false
}

func hasStarResource(s model.IAMStatement) bool {
	for _, r := range s.Resources {
		if r == "*" {
			return true
		}
	}
	return false
}

// assessBedrockStatements bewertet Bedrock-Allow-Statements (gemeinsam für
// Terraform LH-FA-TF-003 und AWS LH-FA-IAM-002): bedrock:*/InvokeModel*
// (oder *) auf Resource "*" ist FAIL ohne und WARN mit Region-Condition;
// sonst PASS (WARN bei nicht auflösbaren Einträgen). Die Findings tragen
// weder Regel-ID noch Subject/File; das setzt der Aufrufer.
func assessBedrockStatements(stmts []model.IAMStatement) []model.Finding {
	var out []model.Finding
	bedrock, risky, dynamic, dynLine := 0, 0, 0, 0
	for _, s := range stmts {
		if !strings.EqualFold(s.Effect, "Allow") || !touchesBedrock(s) {
			continue
		}
		bedrock++
		if hasWildcardAction(s) && hasStarResource(s) {
			risky++
			fd := model.Finding{Line: s.Line}
			if s.HasRegionCondition {
				fd.Severity, fd.Reason = model.WARN, model.ReasonIAMWildcardCond
				fd.Message = "Bedrock-Rechte auf \"*\" mit Region-Condition; Umfang prüfen"
			} else {
				fd.Severity, fd.Reason = model.FAIL, model.ReasonIAMWildcard
				fd.Message = "Bedrock-Rechte (bedrock:*/InvokeModel) auf \"*\" ohne Region-Condition"
				fd.Recommendation = "Resource einschränken oder aws:RequestedRegion-Condition setzen"
			}
			out = append(out, fd)
		} else if s.Dynamic {
			dynamic++
			if dynLine == 0 {
				dynLine = s.Line
			}
		}
	}
	switch {
	case bedrock == 0 || risky > 0:
		// nichts oder bereits gemeldet
	case dynamic > 0:
		out = append(out, model.Finding{Severity: model.WARN, Reason: model.ReasonIAMDynamic, Line: dynLine,
			Message: "Bedrock-Policy enthält nicht auflösbare Einträge; manuell prüfen"})
	default:
		out = append(out, model.Finding{Severity: model.PASS, Reason: model.ReasonOK,
			Message: "Bedrock-Policies eng eingeschränkt"})
	}
	return out
}

// Aktionen der Rechteprüfung (LH-FA-IAM-001).
var (
	// IAMInvokeActions sind die Aufruf-Rechte (Fehlen = FAIL).
	IAMInvokeActions = []string{"bedrock:InvokeModel", "bedrock:InvokeModelWithResponseStream", "bedrock:Converse", "bedrock:ConverseStream"}
	// IAMDiscoveryActions sind die Erkennungs-Rechte (Fehlen = WARN).
	IAMDiscoveryActions = []string{"bedrock:ListInferenceProfiles", "bedrock:GetInferenceProfile"}
)

// IAMScan bündelt die gelesenen IAM-Daten des aktuellen Principals.
type IAMScan struct {
	Decisions     []model.ActionDecision
	DecisionsErr  *model.AWSError
	Statements    []model.IAMStatement
	StatementsErr *model.AWSError
}

func iamErrorFinding(id string, e *model.AWSError) model.Finding {
	if e.Reason == model.AWSAccessDenied {
		return model.Finding{RuleID: id, Severity: model.WARN, Reason: model.ReasonIAMUnknown, Subject: "iam",
			Message:        "IAM-Rechte des Principals nicht ermittelbar (Zugriff auf " + e.Op + " verweigert)",
			Recommendation: "Dem Prüfwerkzeug lesenden IAM-Zugriff auf den eigenen Principal geben"}
	}
	f := awsErrorFinding(id, "iam", e)
	return f
}

// CheckIAMRights prüft die Bedrock-Rechte des Principals (LH-FA-IAM-001):
// PASS, wenn alle Aufruf- und Erkennungs-Rechte erlaubt sind; FAIL, wenn
// Aufruf-Rechte fehlen; WARN, wenn nur Erkennungs-Rechte fehlen oder die
// Rechte nicht ermittelbar sind.
func CheckIAMRights(s AWSScan, _ model.Config) []model.Finding {
	if s.IAM == nil {
		return nil
	}
	if s.IAM.DecisionsErr != nil {
		return []model.Finding{iamErrorFinding(IDIAMRights, s.IAM.DecisionsErr)}
	}
	allowed := map[string]bool{}
	for _, d := range s.IAM.Decisions {
		allowed[strings.ToLower(d.Action)] = d.Allowed
	}
	missing := func(actions []string) []string {
		var out []string
		for _, a := range actions {
			if !allowed[strings.ToLower(a)] {
				out = append(out, a)
			}
		}
		return out
	}
	f := model.Finding{RuleID: IDIAMRights, Subject: "iam"}
	switch inv, disc := missing(IAMInvokeActions), missing(IAMDiscoveryActions); {
	case len(inv) > 0:
		f.Severity, f.Reason = model.FAIL, model.ReasonIAMInvokeDenied
		f.Message = fmt.Sprintf("Aufruf-Rechte fehlen: %s", strings.Join(inv, ", "))
		f.Recommendation = "bedrock:InvokeModel* bzw. bedrock:Converse* für die erlaubten Ressourcen vergeben"
	case len(disc) > 0:
		f.Severity, f.Reason = model.WARN, model.ReasonIAMDiscoveryDenied
		f.Message = fmt.Sprintf("Erkennungs-Rechte fehlen: %s", strings.Join(disc, ", "))
		f.Recommendation = "bedrock:ListInferenceProfiles und bedrock:GetInferenceProfile ergänzen"
	default:
		f.Severity, f.Reason = model.PASS, model.ReasonOK
		f.Message = "Bedrock-Aufruf- und Erkennungs-Rechte vorhanden"
	}
	return []model.Finding{f}
}

// CheckIAMBroad erkennt zu breite Bedrock-Policies des Principals
// (LH-FA-IAM-002) mit derselben Bewertung wie Terraform (LH-FA-TF-003).
func CheckIAMBroad(s AWSScan, _ model.Config) []model.Finding {
	if s.IAM == nil {
		return nil
	}
	if s.IAM.StatementsErr != nil {
		return []model.Finding{iamErrorFinding(IDIAMBroad, s.IAM.StatementsErr)}
	}
	var out []model.Finding
	for _, fd := range assessBedrockStatements(s.IAM.Statements) {
		fd.RuleID, fd.Subject = IDIAMBroad, "iam"
		out = append(out, fd)
	}
	return out
}

// boundByARN: alle Resources sind Bedrock-ARNs mit erlaubter Region.
func boundByARN(s model.IAMStatement, cfg model.Config) bool {
	if len(s.Resources) == 0 {
		return false
	}
	for _, r := range s.Resources {
		parts := strings.SplitN(r, ":", 6)
		if len(parts) < 6 || parts[0] != "arn" || parts[2] != "bedrock" || !cfg.RegionAllowed(parts[3]) {
			return false
		}
	}
	return true
}

// CheckIAMRegionCondition prüft, ob Bedrock-Aufrufe auf erlaubte Regionen
// eingeschränkt sind (LH-FA-IAM-003): je Bedrock-Allow-Statement muss eine
// aws:RequestedRegion-Condition mit erlaubten Regionen gelten oder die
// Resource an einen ARN einer erlaubten Region gebunden sein. FAIL ohne;
// WARN, wenn die Condition ohne bekannte Regionen vorliegt.
func CheckIAMRegionCondition(s AWSScan, cfg model.Config) []model.Finding {
	if s.IAM == nil {
		return nil
	}
	if s.IAM.StatementsErr != nil {
		return []model.Finding{iamErrorFinding(IDIAMRegionCnd, s.IAM.StatementsErr)}
	}
	var out []model.Finding
	bedrock := 0
	for _, st := range s.IAM.Statements {
		if !strings.EqualFold(st.Effect, "Allow") || !touchesBedrock(st) {
			continue
		}
		bedrock++
		fd := model.Finding{RuleID: IDIAMRegionCnd, Subject: "iam", Line: st.Line}
		switch {
		case st.HasRegionCondition && len(st.ConditionRegions) > 0:
			bad := ""
			for _, r := range st.ConditionRegions {
				if !cfg.RegionAllowed(r) {
					bad = r
					break
				}
			}
			if bad != "" {
				fd.Severity, fd.Reason = model.FAIL, model.ReasonRegionConditionNotAllowed
				fd.Message = "aws:RequestedRegion erlaubt nicht erlaubte Region " + bad
				fd.Recommendation = "Condition auf erlaubte EU-Regionen beschränken"
				out = append(out, fd)
			}
		case st.HasRegionCondition:
			fd.Severity, fd.Reason = model.WARN, model.ReasonRegionConditionUnknown
			fd.Message = "aws:RequestedRegion-Condition vorhanden, Regionen nicht auswertbar; manuell prüfen"
			out = append(out, fd)
		case boundByARN(st, cfg):
			// durch Resource-ARN auf erlaubte Region beschränkt
		default:
			fd.Severity, fd.Reason = model.FAIL, model.ReasonRegionConditionMissing
			fd.Message = "Bedrock-Aufrufe ohne aws:RequestedRegion-Condition erlaubt"
			fd.Recommendation = "aws:RequestedRegion-Condition mit erlaubten EU-Regionen setzen"
			out = append(out, fd)
		}
	}
	if len(out) == 0 && bedrock > 0 {
		out = append(out, model.Finding{RuleID: IDIAMRegionCnd, Severity: model.PASS, Reason: model.ReasonOK, Subject: "iam",
			Message: "Bedrock-Aufrufe sind auf erlaubte Regionen eingeschränkt"})
	}
	return out
}
