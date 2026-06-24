package models

type FeatureFlag struct {
	ID      int64  `json:"id"`
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
}
