package config

type AppConfig struct{}

func (ap *AppConfig) Get(key string) string {
	appConfig := map[string]string{
		"projectPath": "D:\\Development\\golang\\robot-butago",
	}

	return appConfig[key]
}
