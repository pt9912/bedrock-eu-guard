# Mindestrechte für `bedrock-eu-check aws`

`bedrock-eu-check aws` (und `all`) arbeitet ausschließlich **lesend**
([`LH-SEC-003`](../../spec/lastenheft.md#lh-sec-003--read-only-aws-zugriff)) und
speichert keine Credentials
([`LH-SEC-001`](../../spec/lastenheft.md#lh-sec-001--keine-speicherung-von-aws-credentials)).
Diese Seite listet die Rechte, die das Werkzeug für seine Prüfungen braucht
([`LH-SEC-004`](../../spec/lastenheft.md#lh-sec-004--least-privilege)). Fehlt ein Recht,
meldet das Werkzeug `WARN` ("nicht ermittelbar"), nie ein stilles `PASS`.

Der Modus `local` und der Modus `terraform` brauchen **keinen** AWS-Zugang.

## Aufrufe und benötigte Rechte

| Prüfung | AWS-CLI-Aufruf | IAM-Aktion |
|---|---|---|
| Identität, Konto, Profil | `sts get-caller-identity` | `sts:GetCallerIdentity` (braucht in der Praxis keine Berechtigung) |
| Inference-Profile je erlaubter Region | `bedrock list-inference-profiles` | `bedrock:ListInferenceProfiles` |
| Modellverfügbarkeit | `bedrock list-foundation-models` | `bedrock:ListFoundationModels` |
| Rechte des eigenen Principals | `iam simulate-principal-policy` | `iam:SimulatePrincipalPolicy` |
| Auflösung einer Sitzung (`assumed-role`) zur IAM-Rolle | `iam get-role` | `iam:GetRole` |
| Policies des eigenen Principals | `iam list-attached-*-policies`, `get-policy`, `get-policy-version`, `list-*-policies`, `get-*-policy`, `list-groups-for-user` | `iam:ListAttachedRolePolicies`, `iam:ListRolePolicies`, `iam:GetRolePolicy`, `iam:GetPolicy`, `iam:GetPolicyVersion` (bei Benutzern zusätzlich `…User…`/`…Group…`-Varianten und `iam:ListGroupsForUser`) |

Das Werkzeug liest nur den **eigenen** Principal; andere Benutzer, Rollen
oder Konten werden nicht untersucht.

## Beispiel-Policy (für eine Rolle)

`ACCOUNT_ID` und `ROLE_NAME` sind durch die eigenen Werte zu ersetzen (bei einer
Rolle mit Pfad den Pfad im ARN ergänzen). Die Bedrock-Listen-Aktionen
unterstützen keine Einschränkung auf Ressourcen und stehen daher mit `*`.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Identity",
      "Effect": "Allow",
      "Action": "sts:GetCallerIdentity",
      "Resource": "*"
    },
    {
      "Sid": "BedrockDiscovery",
      "Effect": "Allow",
      "Action": [
        "bedrock:ListInferenceProfiles",
        "bedrock:ListFoundationModels"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": { "aws:RequestedRegion": ["eu-central-1"] }
      }
    },
    {
      "Sid": "OwnPrincipalPolicies",
      "Effect": "Allow",
      "Action": [
        "iam:GetRole",
        "iam:ListAttachedRolePolicies",
        "iam:ListRolePolicies",
        "iam:GetRolePolicy",
        "iam:SimulatePrincipalPolicy"
      ],
      "Resource": "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
    },
    {
      "Sid": "ReadAttachedPolicyDocuments",
      "Effect": "Allow",
      "Action": [
        "iam:GetPolicy",
        "iam:GetPolicyVersion"
      ],
      "Resource": [
        "arn:aws:iam::ACCOUNT_ID:policy/*",
        "arn:aws:iam::aws:policy/*"
      ]
    }
  ]
}
```

Die `aws:RequestedRegion`-Bedingung ist an die erlaubten Regionen
anzupassen (`allowedRegions` in `bedrock-eu-check.yml`).

## Hinweise

- **Näherung:** Die Rechtesimulation berücksichtigt Organisations-SCPs und Permission
  Boundaries nur eingeschränkt; das Ergebnis ist eine Näherung.
- **Abgelaufene Sitzungen** (z. B. SSO) werden als fehlende Credentials gemeldet
  (`AWS_NO_CREDENTIALS`, Exit `1`); ein fehlendes `aws`-Binary ist ein Tool-Fehler
  (`AWS_CLI_MISSING`, Exit `2`).
- **Keine Ausgabe von Zugangsdaten:** Weder Credentials noch Ausgaben der AWS-CLI
  werden in Meldungen übernommen ([`LH-NF-004`](../../spec/lastenheft.md#lh-nf-004--keine-secret-ausgabe)).
