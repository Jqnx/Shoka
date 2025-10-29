package nhentai

type NHSearch struct {
	Metadata []Metadata `json:"result"`
	NumPages int        `json:"num_pages"`
	PerPage  int        `json:"per_page"`
}

type Metadata struct {
	ID         int    `json:"id"`
	MediaID    string `json:"media_id"`
	Title      Title  `json:"title"`
	Images     Images `json:"images"`
	Scanlator  string `json:"scanlator"`
	UploadDate int    `json:"upload_date"`
	Tags       []Tag  `json:"tags"`
	PageCount  int    `json:"num_pages"`
	FavCount   int    `json:"num_favorites"`
}

type Title struct {
	English  string `json:"english"`
	Japanese string `json:"japanese"`
	Pretty   string `json:"pretty"`
}

type Page struct {
	Type   string `json:"t"`
	Width  int    `json:"w"`
	Height int    `json:"h"`
}

type Images struct {
	Pages     []Page `json:"pages"`
	Cover     Page   `json:"cover"`
	Thumbnail Page   `json:"thumbnail"`
}

type Tag struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Count int    `json:"count"`
}
