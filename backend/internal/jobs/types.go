package jobs

const (
	JobTypeMetadata = "metadata"
	JobTypeIndex    = "index"
)

type MetadataPayload struct{}

type IndexPayload struct{}
