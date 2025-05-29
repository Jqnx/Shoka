package metadata

type Director struct {
	source IMetadata
}

func NewDirector(s IMetadata) *Director {
	return &Director{
		source: s,
	}
}

func (d *Director) setSource(s IMetadata) {
	d.source = s
}

func (d *Director) FetchMetadata(data string) Metadata {
	d.source.Unmarshal(data)
	return d.source.getMetadata()
}
