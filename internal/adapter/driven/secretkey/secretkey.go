// Package secretkey entscheidet anhand des Schlüsselnamens, ob ein Wert ein
// Secret sein könnte (LH-NF-004). Die Adapter schwärzen solche Werte, der Kern
// sieht nie einen Secret-Wert.
package secretkey

import "strings"

// needles sind Teilstrings (Großschreibung egal), bei denen ein Schlüssel als
// Secret gilt. Bewusst großzügig: Fehlalarme (zu viel geschwärzt) sind
// akzeptabel, Lücken nicht.
var needles = []string{
	"TOKEN", "SECRET", "PASSWORD", "CREDENTIAL", "API_KEY", "ACCESS_KEY",
	"PRIVATE_KEY", "AWS_ACCESS_KEY_ID", "APIKEYHELPER",
}

// IsSecret meldet, ob der Schlüsselname auf ein Secret hindeutet.
func IsSecret(key string) bool {
	up := strings.ToUpper(key)
	if strings.HasSuffix(up, "_KEY") {
		return true
	}
	for _, n := range needles {
		if strings.Contains(up, n) {
			return true
		}
	}
	return false
}
