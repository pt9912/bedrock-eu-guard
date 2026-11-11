package model

import "strings"

// Reason-Codes der Terraform-Regeln.
const (
	ReasonNoTerraform     = "NO_TERRAFORM_FILES"
	ReasonRegionDynamic   = "REGION_DYNAMIC"
	ReasonIAMWildcard     = "IAM_WILDCARD"
	ReasonIAMWildcardCond = "IAM_WILDCARD_REGION_CONDITION"
	ReasonIAMDynamic      = "IAM_DYNAMIC"
	ReasonTerraformFound  = "TERRAFORM_FOUND"
)

// AWSProvider ist ein `provider "aws"`-Block. Region ist nur gesetzt, wenn
// ein String-Literal vorliegt; RegionDynamic, wenn die Region ein nicht
// auflösbarer Ausdruck ist (var.x, local.x, ...). Sind beide leer, fehlt
// das Attribut.
type AWSProvider struct {
	Alias         string
	Region        string
	RegionDynamic bool
	Line          int
}

// IAMStatement ist ein IAM-Statement aus einem Policy-Dokument
// (aws_iam_policy_document oder jsonencode-Policy). Actions und Resources
// enthalten nur Literale; Dynamic=true, wenn Einträge nicht auflösbar
// waren. HasRegionCondition ist wahr, wenn eine Condition auf
// aws:RequestedRegion (o. Ä.) einschränkt.
type IAMStatement struct {
	Effect             string // "Allow" | "Deny"
	Actions            []string
	Resources          []string
	HasRegionCondition bool
	// ConditionRegions sind die Werte der aws:RequestedRegion-Condition,
	// soweit auswertbar (leer bei HasRegionCondition = Regionen unbekannt).
	ConditionRegions []string
	Dynamic          bool
	Line             int
}

// ModelRef ist ein Fund einer Modell-/Profil-ID oder Bedrock-ARN in einem
// String-Literal (nur das gefundene Token, nie das ganze Literal).
type ModelRef struct {
	Value string
	Line  int
}

// TerraformFile sind die geparsten Fakten einer .tf-Datei. Der Kern sieht
// keine übrigen Literale, Variablenwerte oder .tfvars (LH-NF-004).
type TerraformFile struct {
	Path       string
	Providers  []AWSProvider
	Statements []IAMStatement
	ModelRefs  []ModelRef
}

// IAMPolicyFile sind die Fakten einer eigenständigen IAM-Policy-Datei
// (LH-FA-IAM-004): die flach gelegten Statements. Gültiges JSON ohne
// erkennbares Policy-Dokument (Fremdformat) ergibt eine leere Liste.
type IAMPolicyFile struct {
	Statements []IAMStatement
}

// StrictIAMPolicyName meldet, ob der Dateiname dem eindeutigen Muster
// iam-policy*.json entspricht (LH-FA-IAM-004): Dort ist eine kaputte Datei
// ein FAIL, beim breiten Muster *-policy.json nur ein WARN.
func StrictIAMPolicyName(rel string) bool {
	base := rel
	if i := strings.LastIndexByte(rel, '/'); i >= 0 {
		base = rel[i+1:]
	}
	return strings.HasPrefix(base, "iam-policy") && strings.HasSuffix(base, ".json")
}
