package config

const (

	// Sources
	SourceNhentai   = "nhentai"
	SourceComicInfo = "comicinfo"
	SourceForm      = "form"
	SourceHentag    = "hentag"

	// Metadata Files
	ComicInfoFile = "ComicInfo.xml"
	HentagFile    = "info.json"
	NHFile        = "info.txt"

	// Source Specific
	// nhentai
	NHDomain  = "nhentai.net"
	NHApi     = "https://nhentai.net/api"
	NHImages  = "https://i.nhentai.net"
	NHGallery = "https://nhentai.net/g"
)

var (
	// Methods
	MethodID    = "id"
	MethodTitle = "title"

	// Metadata
	MetadataFormats = map[string]string{
		SourceComicInfo: ComicInfoFile,
		SourceHentag:    HentagFile,
		SourceNhentai:   NHFile,
	}

	NHFileTypes = map[string]string{
		"w": ".webp",
		"j": ".jpg",
		"p": ".png",
	}
)
