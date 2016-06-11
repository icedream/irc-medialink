package wikipedia

type v1Summary struct {
	Dir       string      `json:"dir"`
	Extract   string      `json:"extract"`
	Lang      string      `json:"lang"`
	Thumbnail v1Thumbnail `json:"thumbnail"`
	Title     string      `json:"title"`
}

type v1Thumbnail struct {
	Height uint64 `json:"height"`
	Source string `json:"source"`
	Width  uint64 `json:"width"`
}
