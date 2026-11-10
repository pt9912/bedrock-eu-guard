package fakeaws

import "github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"

// Bedrock implementiert driven.BedrockReader mit Daten je Region.
type Bedrock struct {
	Profiles    map[string][]model.InferenceProfile
	ProfilesErr map[string]error
	Models      map[string][]model.FoundationModel
	ModelsErr   map[string]error
}

// ListInferenceProfiles implementiert driven.BedrockReader.
func (b Bedrock) ListInferenceProfiles(region string) ([]model.InferenceProfile, error) {
	if err := b.ProfilesErr[region]; err != nil {
		return nil, err
	}
	return b.Profiles[region], nil
}

// ListFoundationModels implementiert driven.BedrockReader.
func (b Bedrock) ListFoundationModels(region string) ([]model.FoundationModel, error) {
	if err := b.ModelsErr[region]; err != nil {
		return nil, err
	}
	return b.Models[region], nil
}
