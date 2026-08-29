package config

type Twitch struct {
	Username string
	Oauth    string
	Channel  string
}

type Config struct {
	Twitch
}

func NewConfig() (Config, error) {
	return Config{}, nil
}
