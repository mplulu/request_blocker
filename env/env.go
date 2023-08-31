package env

var E *ENV

type ENV struct {
	Host             string     `yaml:"host"`
	TargetURl        string     `yaml:"target_url"`
	LimitRate        *LimitRate `yaml:"limit_rate"`
	TelegramBotToken string     `yaml:"telegram_bot_token"`
	TelegramChatId   string     `yaml:"telegram_chat_id"`
}

type LimitRate struct {
	EnableLog     bool `yaml:"enable_log"`
	EnableBlock   bool `yaml:"enable_block"`
	MaxTotalCount int  `yaml:"max_total_count"`
	MaxCount      int  `yaml:"max_count"`
}
