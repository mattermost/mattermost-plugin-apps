package apps

type Wish struct {
	AppID string `json:"app_id"`
	URL   string `json:"url"`
}

func NewWish(appID, url string) *Wish {
	return &Wish{
		AppID: appID,
		URL:   url,
	}
}
