package terraform_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/terraform"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.TerraformReader = terraform.Reader{}

func parse(t *testing.T, src string) *model.TerraformFile {
	t.Helper()
	f, ferr := terraform.Parse("main.tf", []byte(src))
	if ferr != nil {
		t.Fatalf("Parse: %v", ferr)
	}
	return f
}

func TestProviders(t *testing.T) {
	f := parse(t, `
variable "r" { default = "us-east-1" }

provider "aws" {
  region = "eu-central-1"
}

provider "aws" {
  alias  = "dyn"
  region = var.region
}

provider "aws" {
  alias = "norg"
}

provider "google" {
  region = "us-central1"
}
`)
	want := []model.AWSProvider{
		{Region: "eu-central-1", Line: 5},
		{Alias: "dyn", RegionDynamic: true, Line: 10},
		{Alias: "norg", Line: 13},
	}
	if !reflect.DeepEqual(f.Providers, want) {
		t.Errorf("got %+v\nwant %+v", f.Providers, want)
	}
}

func TestPolicyDocument(t *testing.T) {
	f := parse(t, `
data "aws_iam_policy_document" "p" {
  statement {
    effect    = "Allow"
    actions   = ["bedrock:InvokeModel*", "bedrock:ListFoundationModels"]
    resources = ["*"]
  }
  statement {
    actions   = ["bedrock:*"]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "aws:RequestedRegion"
      values   = ["eu-central-1"]
    }
  }
  statement {
    effect    = "Deny"
    actions   = [var.action, "s3:*"]
    resources = local.arns
  }
}
`)
	if len(f.Statements) != 3 {
		t.Fatalf("Statements: %+v", f.Statements)
	}
	s0, s1, s2 := f.Statements[0], f.Statements[1], f.Statements[2]
	if s0.Effect != "Allow" || !reflect.DeepEqual(s0.Actions, []string{"bedrock:InvokeModel*", "bedrock:ListFoundationModels"}) ||
		!reflect.DeepEqual(s0.Resources, []string{"*"}) || s0.HasRegionCondition || s0.Dynamic || s0.Line != 3 {
		t.Errorf("s0: %+v", s0)
	}
	if s1.Effect != "Allow" || !s1.HasRegionCondition || s1.Line != 8 {
		t.Errorf("s1 (Default-Effect Allow, Region-Condition): %+v", s1)
	}
	if s2.Effect != "Deny" || !s2.Dynamic || !reflect.DeepEqual(s2.Actions, []string{"s3:*"}) {
		t.Errorf("s2 (dynamisch, nur Literale): %+v", s2)
	}
}

func TestPolicyResourcesJSONEncodeAndHeredoc(t *testing.T) {
	f := parse(t, `
resource "aws_iam_policy" "a" {
  name = "a"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["bedrock:InvokeModel"]
        Resource = "*"
        Condition = {
          StringEquals = { "aws:RequestedRegion" = "eu-central-1" }
        }
      },
      {
        Effect   = "Allow"
        Action   = "bedrock:*"
        Resource = "arn:aws:bedrock:eu-central-1::foundation-model/*"
      },
    ]
  })
}

resource "aws_iam_role_policy" "b" {
  role   = "r"
  policy = <<EOF
{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"*","Resource":"*"}}
EOF
}

resource "aws_iam_policy" "c" {
  policy = jsonencode({ Statement = [{ Effect = "Allow", Action = "bedrock:*", Resource = var.arn }] })
}

resource "aws_iam_policy" "d" {
  policy = file("policy.json")
}
`)
	if len(f.Statements) != 5 {
		t.Fatalf("Statements: %+v", f.Statements)
	}
	a1, a2, b, c, d := f.Statements[0], f.Statements[1], f.Statements[2], f.Statements[3], f.Statements[4]
	if !a1.HasRegionCondition || a1.Effect != "Allow" || !reflect.DeepEqual(a1.Actions, []string{"bedrock:InvokeModel"}) || !reflect.DeepEqual(a1.Resources, []string{"*"}) {
		t.Errorf("a1: %+v", a1)
	}
	if a2.HasRegionCondition || !reflect.DeepEqual(a2.Actions, []string{"bedrock:*"}) {
		t.Errorf("a2: %+v", a2)
	}
	if b.Effect != "Allow" || !reflect.DeepEqual(b.Actions, []string{"*"}) || b.Dynamic {
		t.Errorf("b (Heredoc, Statement als Objekt): %+v", b)
	}
	if !c.Dynamic || !d.Dynamic {
		t.Errorf("c/d müssen dynamic sein: %+v / %+v", c, d)
	}
}

func TestModelRefsOnlyTokens(t *testing.T) {
	f := parse(t, `
resource "aws_ssm_parameter" "m" {
  name  = "model"
  value = "us.anthropic.claude-3-5-sonnet-20241022-v2:0"
}
locals {
  desc = "Visit Amazon.com. Mehr Text GEHEIM-NOTIZ, anthropic.claude-sonnet-4-5-20250929-v1:0 und danach Text"
  arn  = "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-haiku-20240307-v1:0"
  ok   = "eu.anthropic.claude-sonnet-4-6"
  none = "amazon.com"
}
`)
	got := map[string]int{}
	for _, r := range f.ModelRefs {
		got[r.Value] = r.Line
	}
	want := map[string]int{
		"us.anthropic.claude-3-5-sonnet-20241022-v2:0":                                       4,
		"anthropic.claude-sonnet-4-5-20250929-v1:0":                                          7,
		"arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-haiku-20240307-v1:0": 8,
		"eu.anthropic.claude-sonnet-4-6":                                                     9,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
	for _, r := range f.ModelRefs {
		if strings.Contains(r.Value, "GEHEIM") || strings.Contains(r.Value, " ") {
			t.Errorf("Treffer enthält mehr als das Token: %q", r.Value)
		}
	}
}

func TestModelRefInHeredocHasRightLine(t *testing.T) {
	f := parse(t, "resource \"x\" \"y\" {\n  v = <<EOT\nzeile eins\n{\"m\": \"eu.anthropic.claude-haiku-4-5\"}\nEOT\n}\n")
	if len(f.ModelRefs) != 1 || f.ModelRefs[0].Line != 4 {
		t.Errorf("got %+v", f.ModelRefs)
	}
}

func TestNoVariableValuesInResult(t *testing.T) {
	f := parse(t, `
variable "api_key" { default = "GEHEIMER-DEFAULT" }
locals { token = "GEHEIMER-LOCAL" }
output "x" { value = "GEHEIMER-OUTPUT" }
`)
	if len(f.Providers) != 0 || len(f.Statements) != 0 || len(f.ModelRefs) != 0 {
		t.Errorf("nur Fakten erwartet, got %+v", f)
	}
}

func TestInvalidFile(t *testing.T) {
	const secret = "GEHEIM"
	f, ferr := terraform.Parse("main.tf", []byte("provider \"aws\" {\n  region = \"eu-central-1\"\n  x = ["+secret+"\n}\n"))
	if f != nil || ferr == nil || ferr.Reason != model.ReasonFileInvalid || ferr.Path != "main.tf" || ferr.Line == 0 {
		t.Fatalf("want FILE_INVALID with line, got %+v, %+v", f, ferr)
	}
	if strings.Contains(ferr.Error(), secret) {
		t.Errorf("Fehlertext zitiert Inhalt: %q", ferr.Error())
	}
}

func TestFindAndRead(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("main.tf", `provider "aws" { region = "eu-central-1" }`)
	write("modules/net/net.tf", `provider "aws" { region = "us-east-1" }`)
	write("a.tf.json", `{}`)
	write("notes.txt", "x")
	for _, skip := range []string{".terraform", ".git", "vendor", "node_modules", ".tmp"} {
		write(skip+"/hidden.tf", `provider "aws" { region = "us-west-2" }`)
	}
	if err := os.Symlink(filepath.Join(root, "modules"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	write("kaputt.tf", `provider "aws" {`)

	r := terraform.New(root)
	got, err := r.Find()
	want := []string{"kaputt.tf", "main.tf", "modules/net/net.tf"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Find: got %v, %v; want %v", got, err, want)
	}

	f, err := r.Read("main.tf")
	if err != nil || f == nil || f.Path != "main.tf" || len(f.Providers) != 1 {
		t.Errorf("Read: %+v, %v", f, err)
	}
	if f, err := r.Read("fehlt.tf"); f != nil || err != nil {
		t.Errorf("fehlend: %+v, %v", f, err)
	}
	var fe *model.FileError
	if _, err := r.Read("kaputt.tf"); !errors.As(err, &fe) || fe.Reason != model.ReasonFileInvalid {
		t.Errorf("kaputt: %v", err)
	}
	for _, p := range []string{"../x.tf", "/etc/x.tf"} {
		if _, err := r.Read(p); !errors.As(err, &fe) {
			t.Errorf("%s muss abgelehnt werden: %v", p, err)
		}
	}
}
