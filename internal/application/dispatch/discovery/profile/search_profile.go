package profile

type SearchProfile struct {
	name string

	expansion ExpansionPolicy
}

func (p SearchProfile) Name() string {
	return p.name
}

func (p SearchProfile) Expansion() ExpansionPolicy {
	return p.expansion
}
