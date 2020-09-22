package client

type InstallCompleteBody struct {
	BotID          string
	BotAccessToken string
	OAuthAppID     string
	OAuthAppSecret string
}

func (c *client) InstallComplete() {
	c.DoPost(c.app.Manifest.InstallCompleteURL, &InstallCompleteBody{
		BotID:          c.app.BotID,
		BotAccessToken: c.app.BotToken,
		OAuthAppID:     c.app.OAuthAppID,
		OAuthAppSecret: c.app.OAuthSecret,
	})
}
