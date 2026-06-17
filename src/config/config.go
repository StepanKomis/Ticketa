package config

type Config struct {
	Logging        LoggingConfig  `yaml:"logging"`
	TicketStatuses []StatusConfig `yaml:"ticket_statuses"`
	Activity       ActivityConfig `yaml:"activity"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	Dir   string `yaml:"dir"`
}

// ActivityConfig určuje, kam se zapisuje souborový activity log a jak dlouho se rotované soubory uchovávají.
type ActivityConfig struct {
	LogFile    string `yaml:"log_file"`
	LogMaxDays int    `yaml:"log_max_days"`
}

// StatusConfig představuje jeden stav tiketu v konfiguračním souboru.
// Pozice v poli je sémanticky významná: první=otevřeno, poslední=vyřešeno.
type StatusConfig struct {
	Title string `yaml:"title"`
	Color string `yaml:"color"`
}

// Defaults vrátí Config s rozumnými výchozími hodnotami a třemi českými stavy.
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
		Activity: ActivityConfig{
			LogFile:    "logs/activity.log",
			LogMaxDays: 30,
		},
	}
}
