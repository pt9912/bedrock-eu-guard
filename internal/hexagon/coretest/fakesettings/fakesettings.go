// Package fakesettings ist ein Test-Double für driven.ClaudeSettingsReader.
package fakesettings

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// Reader liefert vorbereitete Settings oder Fehler.
type Reader struct {
	User       *model.ClaudeSettings
	UserErr    error
	Project    map[string]*model.ClaudeSettings
	ProjectErr map[string]error
}

// ReadUser implementiert driven.ClaudeSettingsReader.
func (r Reader) ReadUser() (*model.ClaudeSettings, error) { return r.User, r.UserErr }

// ReadProject implementiert driven.ClaudeSettingsReader.
func (r Reader) ReadProject(rel string) (*model.ClaudeSettings, error) {
	if err := r.ProjectErr[rel]; err != nil {
		return nil, err
	}
	return r.Project[rel], nil
}
