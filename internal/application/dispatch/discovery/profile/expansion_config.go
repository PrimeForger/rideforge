package profile

// ExpansionConfig defines the configured ExpansionProfile presets for each driver density class.
type ExpansionConfig struct {
	Sparse  ExpansionProfile
	Normal  ExpansionProfile
	Dense   ExpansionProfile
	Default ExpansionProfile
}
