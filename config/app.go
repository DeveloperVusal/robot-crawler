package config

type AppConfig struct{}

func (ap *AppConfig) Get(key string) string {
	appConfig := map[string]string{
		"projectPath": "",
	}

	return appConfig[key]
}
