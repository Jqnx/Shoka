package metadata

type IMetadata interface {
	Unmarshal(data string)
	// setTitle(title string)
	// setSummary()
	// setPageCount()
	// setURLs()
	// setGenre()
	// setSeries()
	// setCharacters()
	// setTags()
	// setWriter()
	// setLanguage()
	getMetadata() Metadata
}

type Metadata struct {
	Title      string
	Summary    string
	PageCount  int
	URL        string
	Genre      string
	Series     string
	Characters string
	Tags       []Tag
	Writer     string
	Language   string
}

type Tag struct {
	Tag string
}

func getBuilder(builderType string) IMetadata {
	if builderType == "comicinfo" {
		return newComicInfo()
	}
	return nil
}
