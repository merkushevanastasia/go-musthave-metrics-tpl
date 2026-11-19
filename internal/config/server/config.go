package server

type Config struct {
	ServerURL       string `env:"ADDRESS"`
	LogLevel        string
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}
