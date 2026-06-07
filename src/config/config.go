package config

type Config struct {
	Logging        LoggingConfig  `yaml:"logging"`
	TicketStatuses []StatusConfig `yaml:"ticket_statuses"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	Dir   string `yaml:"dir"`
}

// StatusConfig represents a single ticket status entry from the config file.
// Position in the slice is semantically significant: first=open, last=solved.
type StatusConfig struct {
	Title string `yaml:"title"`
	Color string `yaml:"color"`
}

// Defaults returns a Config populated with sane defaults and three Czech statuses.
func Defaults() *Config {
	return &Config{
		Logging: LoggingConfig{
			Level: "info",
			Dir:   "/var/log/ticketa",
		},
		TicketStatuses: []StatusConfig{
			{Title: "Otevřeno", Color: "#3498db"},
			{Title: "Probíhá", Color: "#f39c12"},
			{Title: "Vyřešeno", Color: "#2ecc71"},
		},
	}
}
