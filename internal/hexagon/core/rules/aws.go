package rules

import (
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

// Regel-IDs der AWS-Regeln.
const (
	IDAWSIdentity = "LH-FA-AWS-001"
	IDAWSProfile  = "LH-FA-AWS-002"
	IDAWSAccount  = "LH-FA-AWS-003"
	IDAuth        = "LH-FA-AUTH-001"
)

// AWSScan bündelt die einmal ermittelten AWS-Fakten für alle AWS-Regeln.
// Genau eines von Identity und IdentityErr ist gesetzt.
type AWSScan struct {
	Env         driven.Environment
	Identity    *model.AWSIdentity
	IdentityErr *model.AWSError

	// Bedrock ist nil, wenn kein BedrockReader konfiguriert ist (Regeln
	// LH-FA-BEDROCK-* laufen dann nicht).
	Bedrock *BedrockScan

	// IAM ist nil, wenn kein IAMReader konfiguriert ist (Regeln LH-FA-IAM-*
	// laufen dann nicht).
	IAM *IAMScan
}

// AWSRule wertet AWS-Fakten gegen die Config aus.
type AWSRule func(s AWSScan, cfg model.Config) []model.Finding

// AWSRules liefert die AWS-Regeln in stabiler Reihenfolge.
func AWSRules() []AWSRule {
	return []AWSRule{CheckAWSIdentity, CheckAWSProfile, CheckAWSAccount, CheckAuthMethod,
		CheckBedrockProfiles, CheckBedrockBlocked, CheckBedrockAllowlist, CheckBedrockAvailability,
		CheckIAMRights, CheckIAMBroad, CheckIAMRegionCondition}
}

func awsErrorFinding(id, subject string, e *model.AWSError) model.Finding {
	f := model.Finding{RuleID: id, Severity: model.FAIL, Reason: e.Reason, Subject: subject}
	switch e.Reason {
	case model.AWSNoCredentials:
		f.Message = "Keine gültigen AWS-Credentials gefunden"
		f.Recommendation = "AWS_PROFILE setzen und `aws sso login` bzw. Credentials einrichten"
	case model.AWSAccessDenied:
		f.Message = "Zugriff auf " + e.Op + " verweigert"
		f.Recommendation = "Minimale IAM-Rechte für bedrock-eu-check vergeben"
	default:
		f.Message = "AWS nicht erreichbar (" + e.Op + ")"
		f.Recommendation = "Netzwerk, Proxy und Region prüfen"
	}
	return f
}

// CheckAWSIdentity prüft die STS-Identität (LH-FA-AWS-001): PASS mit Account
// und Principal-ARN, FAIL ohne gültige Credentials oder ohne Erreichbarkeit.
func CheckAWSIdentity(s AWSScan, _ model.Config) []model.Finding {
	if s.IdentityErr != nil {
		return []model.Finding{awsErrorFinding(IDAWSIdentity, "sts", s.IdentityErr)}
	}
	if s.Identity == nil {
		return nil
	}
	return []model.Finding{{RuleID: IDAWSIdentity, Severity: model.PASS, Reason: model.ReasonOK, Subject: "sts",
		Message: "Account " + s.Identity.Account + ", Principal " + s.Identity.ARN}}
}

// CheckAWSProfile zeigt das aktive Profil (LH-FA-AWS-002): PASS bei gesetztem
// und verwendbarem AWS_PROFILE, WARN ohne AWS_PROFILE oder wenn die Identität
// damit nicht ermittelbar ist.
func CheckAWSProfile(s AWSScan, _ model.Config) []model.Finding {
	name, ok := s.Env.Lookup("AWS_PROFILE")
	f := model.Finding{RuleID: IDAWSProfile, Subject: "AWS_PROFILE"}
	switch {
	case !ok || name == "":
		f.Severity, f.Reason = model.WARN, model.ReasonProfileNotSet
		f.Message = "AWS_PROFILE ist nicht gesetzt (Default-Credential-Chain)"
		f.Recommendation = "AWS_PROFILE explizit setzen"
	case s.IdentityErr != nil:
		f.Severity, f.Reason = model.WARN, model.ReasonProfileUnusable
		f.Message = "AWS_PROFILE=" + name + " ist gesetzt, aber nicht verwendbar"
	default:
		f.Severity, f.Reason = model.PASS, model.ReasonProfileSet
		f.Message = "AWS_PROFILE=" + name
	}
	return []model.Finding{f}
}

// CheckAWSAccount prüft die Account-ID gegen allowedAccounts
// (LH-FA-AWS-003). Leere Liste = nicht geprüft (kein Finding); ist die
// Identität nicht ermittelbar, bleibt die Prüfung WARN.
func CheckAWSAccount(s AWSScan, cfg model.Config) []model.Finding {
	if len(cfg.AllowedAccounts) == 0 {
		return nil
	}
	f := model.Finding{RuleID: IDAWSAccount, Subject: "account"}
	if s.Identity == nil {
		f.Severity, f.Reason = model.WARN, model.ReasonAccountUnchecked
		f.Message = "Account-ID nicht prüfbar (Identität nicht ermittelbar)"
		return []model.Finding{f}
	}
	for _, a := range cfg.AllowedAccounts {
		if a == s.Identity.Account {
			f.Severity, f.Reason = model.PASS, model.ReasonAccountAllowed
			f.Message = "Account " + s.Identity.Account + " ist erlaubt"
			return []model.Finding{f}
		}
	}
	f.Severity, f.Reason = model.FAIL, model.ReasonAccountNotAllowed
	f.Message = "Account " + s.Identity.Account + " ist nicht erlaubt"
	f.Recommendation = "Mit einem erlaubten Account arbeiten oder allowedAccounts anpassen"
	return []model.Finding{f}
}

func has(env driven.Environment, key string) bool {
	v, ok := env.Lookup(key)
	return ok && v != ""
}

// CheckAuthMethod klassifiziert die Authentifizierungsart (LH-FA-AUTH-001)
// allein aus der Existenz von Umgebungsvariablen; Werte werden nie gelesen
// oder ausgegeben. Pro erkannter Art ein Finding.
func CheckAuthMethod(s AWSScan, _ model.Config) []model.Finding {
	var out []model.Finding
	add := func(sev model.Severity, reason, subject, msg, rec string) {
		out = append(out, model.Finding{RuleID: IDAuth, Severity: sev, Reason: reason, Subject: subject, Message: msg, Recommendation: rec})
	}
	if has(s.Env, "ANTHROPIC_API_KEY") {
		add(model.FAIL, model.ReasonAuthAnthropicKey, "ANTHROPIC_API_KEY",
			"Authentifizierung über ANTHROPIC_API_KEY (direkte Anthropic-API)", "ANTHROPIC_API_KEY entfernen und Bedrock mit AWS-Credentials nutzen")
	}
	if has(s.Env, "AWS_BEARER_TOKEN_BEDROCK") {
		add(model.WARN, model.ReasonAuthBedrockAPIKey, "AWS_BEARER_TOKEN_BEDROCK",
			"Authentifizierung über Bedrock API Key", "Kurzlebige Credentials (SSO/AWS_PROFILE) bevorzugen")
	}
	if has(s.Env, "AWS_ACCESS_KEY_ID") {
		if has(s.Env, "AWS_SESSION_TOKEN") {
			add(model.PASS, model.ReasonAuthTemporary, "AWS_ACCESS_KEY_ID", "Temporäre Credentials (Session-Token) in der Umgebung", "")
		} else {
			add(model.WARN, model.ReasonAuthStaticKeys, "AWS_ACCESS_KEY_ID",
				"Statische Access Keys in der Umgebung", "SSO bzw. temporäre Credentials statt statischer Keys verwenden")
		}
	}
	if has(s.Env, "AWS_PROFILE") && !has(s.Env, "AWS_ACCESS_KEY_ID") {
		add(model.PASS, model.ReasonAuthProfile, "AWS_PROFILE", "Authentifizierung über AWS_PROFILE", "")
	}
	if len(out) == 0 {
		add(model.WARN, model.ReasonAuthNoneDetected, "auth",
			"Keine Authentifizierungsart in der Umgebung erkannt (Default-Credential-Chain)", "AWS_PROFILE setzen")
	}
	return out
}
