package article

type ArticleVO struct {
	Id         int64  `json:"id"`
	Title      string `json:"title"`
	Abstract   string `json:"abstract"`
	Content    string `json:"content"`
	AuthorId   int64  `json:"authorId"`
	AuthorName string `json:"authorName"`
	Status     uint8  `json:"status"`

	ReadCnt    int64 `json:"readCnt"`
	LikeCnt    int64 `json:"likeCnt"`
	CollectCnt int64 `json:"collectCnt"`
	Liked      bool  `json:"liked"`
	Collected  bool  `json:"collected"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}
