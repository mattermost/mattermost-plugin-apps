package apps

type Wish struct {
	URL string
}

func NewWish(url string) *Wish {
	return &Wish{
		URL: url,
	}
}
