package config

const (
	// Sources
	SourceNhentai   = "nhentai"
	SourceComicInfo = "comicinfo"
)

var (

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
