package server

type Config struct {
	ServerURL string `env:"ADDRESS"`
	LogLevel  string
}
