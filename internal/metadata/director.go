package metadata

type Director struct {
	source IMetadata
}

func newDirector(s IMetadata) *Director {
	return &Director{
		source: s,
	}
}

func (d *Director) setSource(s IMetadata) {
	d.source = s
}

func (d *Director) fetchMetadata(data string) Metadata {
	d.source.Unmarshal(data)
	return d.source.getMetadata()
}
