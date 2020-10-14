package apps

type Wish struct {
	URL string `json:"url"`
}

func NewWish(url string) *Wish {
	return &Wish{
		URL: url,
	}
}
