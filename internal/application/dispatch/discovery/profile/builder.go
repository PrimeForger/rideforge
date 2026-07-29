package profile

type Builder struct {
	profile SearchProfile
}

func NewBuilder() *Builder {

	return &Builder{
		profile: SearchProfile{
			name: "default",
		},
	}
}

func (b *Builder) SetName(
	name string,
) {
	b.profile.name = name
}

func (b *Builder) SetExpansion(
	expansion ExpansionPolicy,
) {
	b.profile.expansion = expansion
}

func (b *Builder) Build() SearchProfile {
	return b.profile
}
