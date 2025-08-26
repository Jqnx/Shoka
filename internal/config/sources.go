package config

var (
	// Sources
	SourceNhentai       = "nhentai"
	SourceNhentaiSearch = "nhsearch"
	SourceComicInfo     = "comicinfo"

	// Methods
	MethodID    = "id"
	MethodTitle = "title"

	// Source Specific
	// nhentai
	NHDomain    = "nhentai.net"
	NHApi       = "https://nhentai.net/api"
	NHImages    = "https://i.nhentai.net"
	NHGallery   = "https://nhentai.net/g"
	NHFileTypes = map[string]string{
		"w": ".webp",
		"j": ".jpg",
		"p": ".png",
	}

	// ComicInfo
	ComicInfoFile = "ComicInfo.xml"
)
