// Package awscli implementiert den AWSReader-Port (ARC-006) über das
// installierte `aws`-Binary. Es wird ohne Shell mit festen Argumenten
// gestartet, die Umgebung wird per Allowlist durchgereicht und nie geloggt.
// stderr der CLI wird ausschließlich zur Fehlerklassifikation gelesen und nie
// in Meldungen übernommen (LH-SEC-001, LH-NF-004). Rein lesend (LH-SEC-003).
package awscli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// DefaultBinary ist der Name des AWS-CLI-Binarys im PATH.
const DefaultBinary = "aws"

const (
	opIdentity = "sts:GetCallerIdentity"
	opProfiles = "bedrock:ListInferenceProfiles"
	opModels   = "bedrock:ListFoundationModels"
)

// timeout begrenzt einen CLI-Aufruf (Variable nur für Tests).
var timeout = 30 * time.Second

// Reader ruft die AWS-CLI auf.
type Reader struct {
	bin    string
	region string
}

// New liefert einen Reader. bin ist Name oder Pfad des aws-Binarys (leer =
// DefaultBinary), region die Region für den Aufruf (leer = keine Angabe).
func New(bin, region string) Reader {
	if bin == "" {
		bin = DefaultBinary
	}
	return Reader{bin: bin, region: region}
}

// Identity ruft `aws sts get-caller-identity` auf. Fehler sind immer
// *model.AWSError ohne CLI-Ausgabe.
func (r Reader) Identity() (*model.AWSIdentity, error) {
	args := []string{"sts", "get-caller-identity", "--output", "json"}
	if r.region != "" {
		args = append(args, "--region", r.region)
	}
	var out struct {
		UserID  string `json:"UserId"`
		Account string `json:"Account"`
		ARN     string `json:"Arn"`
	}
	if err := r.call(opIdentity, args, &out); err != nil {
		return nil, err
	}
	if out.Account == "" {
		return nil, &model.AWSError{Op: opIdentity, Reason: model.AWSUnreachable}
	}
	return &model.AWSIdentity{Account: out.Account, ARN: out.ARN, UserID: out.UserID}, nil
}

// maxPages begrenzt die Paginierung (Schutz vor Endlosschleifen).
const maxPages = 100

// ListInferenceProfiles ruft `aws bedrock list-inference-profiles` für eine
// Region auf. Die AWS-CLI paginiert selbst; liefert sie dennoch ein
// nextToken, wird bis zum Ende nachgeladen.
func (r Reader) ListInferenceProfiles(region string) ([]model.InferenceProfile, error) {
	var all []model.InferenceProfile
	token := ""
	for page := 0; page < maxPages; page++ {
		var out struct {
			Summaries []struct {
				ID   string `json:"inferenceProfileId"`
				Name string `json:"inferenceProfileName"`
				Type string `json:"type"`
			} `json:"inferenceProfileSummaries"`
			NextToken string `json:"nextToken"`
		}
		if err := r.call(opProfiles, bedrockArgs("list-inference-profiles", region, token), &out); err != nil {
			return nil, err
		}
		for _, p := range out.Summaries {
			all = append(all, model.InferenceProfile{ID: p.ID, Name: p.Name, Type: p.Type})
		}
		if out.NextToken == "" || out.NextToken == token {
			return all, nil
		}
		token = out.NextToken
	}
	return nil, &model.AWSError{Op: opProfiles, Reason: model.AWSUnreachable}
}

// ListFoundationModels ruft `aws bedrock list-foundation-models` auf.
func (r Reader) ListFoundationModels(region string) ([]model.FoundationModel, error) {
	var out struct {
		Summaries []struct {
			ID string `json:"modelId"`
		} `json:"modelSummaries"`
	}
	if err := r.call(opModels, bedrockArgs("list-foundation-models", region, ""), &out); err != nil {
		return nil, err
	}
	all := make([]model.FoundationModel, 0, len(out.Summaries))
	for _, m := range out.Summaries {
		all = append(all, model.FoundationModel{ID: m.ID})
	}
	return all, nil
}

func bedrockArgs(sub, region, token string) []string {
	args := []string{"bedrock", sub, "--output", "json"}
	if region != "" {
		args = append(args, "--region", region)
	}
	if token != "" {
		args = append(args, "--next-token", token)
	}
	return args
}

// call führt einen CLI-Aufruf aus und liest die JSON-Antwort in out. Fehler
// sind immer *model.AWSError ohne CLI-Ausgabe.
func (r Reader) call(op string, args []string, out any) error {
	stdout, stderr, err := r.run(args)
	if err != nil {
		return classify(op, err, stderr)
	}
	if jerr := json.Unmarshal(stdout, out); jerr != nil {
		return &model.AWSError{Op: op, Reason: model.AWSUnreachable}
	}
	return nil
}

// run startet das Binary ohne Shell mit Allowlist-Umgebung und Timeout.
func (r Reader) run(args []string) (stdout, stderr []byte, err error) {
	path, err := exec.LookPath(r.bin)
	if err != nil {
		return nil, nil, errCLIMissing
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = allowedEnv(os.Environ())
	cmd.WaitDelay = 2 * time.Second // Kindprozesse dürfen die Pipes nicht ewig offen halten
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err = cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, nil, errTimeout
	}
	return so.Bytes(), se.Bytes(), err
}

var (
	errCLIMissing = errors.New("aws cli missing")
	errTimeout    = errors.New("aws cli timeout")
)

// allowedEnv reicht nur PATH, HOME, TMPDIR und AWS_* weiter und schaltet den
// Pager der CLI ab. Andere Variablen (auch Secrets) erreichen die CLI nicht.
func allowedEnv(environ []string) []string {
	var out []string
	for _, kv := range environ {
		key, _, _ := strings.Cut(kv, "=")
		if key == "PATH" || key == "HOME" || key == "TMPDIR" || strings.HasPrefix(key, "AWS_") {
			if key == "AWS_PAGER" {
				continue
			}
			out = append(out, kv)
		}
	}
	return append(out, "AWS_PAGER=")
}

// Muster in stderr der AWS-CLI (kleingeschrieben), nur zur Klassifikation.
var (
	noCredentialsPatterns = []string{
		"unable to locate credentials", "nocredentialserror", "expiredtoken",
		"token has expired", "security token included in the request is expired",
		"invalidclienttokenid", "signaturedoesnotmatch", "could not be found",
		"sso session", "error loading sso token", "to refresh this sso session",
	}
	accessDeniedPatterns = []string{"accessdenied", "not authorized", "unauthorizedoperation"}
)

// classify bildet einen Fehler des CLI-Aufrufs auf einen *model.AWSError ab.
func classify(op string, err error, stderr []byte) *model.AWSError {
	switch {
	case errors.Is(err, errCLIMissing):
		return &model.AWSError{Op: op, Reason: model.AWSCLIMissing}
	case errors.Is(err, errTimeout):
		return &model.AWSError{Op: op, Reason: model.AWSUnreachable}
	}
	msg := strings.ToLower(string(stderr))
	for _, p := range noCredentialsPatterns {
		if strings.Contains(msg, p) {
			return &model.AWSError{Op: op, Reason: model.AWSNoCredentials}
		}
	}
	for _, p := range accessDeniedPatterns {
		if strings.Contains(msg, p) {
			return &model.AWSError{Op: op, Reason: model.AWSAccessDenied}
		}
	}
	return &model.AWSError{Op: op, Reason: model.AWSUnreachable}
}
