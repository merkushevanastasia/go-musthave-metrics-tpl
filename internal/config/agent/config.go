package agent

type Config struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	ServerURL      string `env:"ADDRESS"`
	LogLevel       string
}
