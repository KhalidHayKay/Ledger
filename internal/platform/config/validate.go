package config

import "errors"

func validate() error {
	if Env.App.Environment == "" {
		return errors.New("APP_ENV not set")
	}

	if Env.App.Url == "" {
		return errors.New("APP_URL not set")
	}

	if Env.App.Port == "" {
		return errors.New("APP_PORT not set")
	}

	missingSplitDBConfig :=
		Env.DB.Connection == "" ||
			Env.DB.Host == "" ||
			Env.DB.Port == "" ||
			Env.DB.Database == "" ||
			Env.DB.Username == "" ||
			Env.DB.Password == ""

	if Env.DB.url == "" && missingSplitDBConfig {
		return errors.New(
			"DB_URL or all DB_CONNECTION, DB_HOST, DB_PORT, DB_DATABASE, DB_USERNAME, DB_PASSWORD must be set",
		)
	}

	if Env.Redis.Host == "" {
		return errors.New("REDIS_HOST not set")
	}

	if Env.Redis.Port == "" {
		return errors.New("REDIS_PORT not set")
	}

	if Env.Redis.Password == "" {
		return errors.New("REDIS_PASSWORD not set")
	}

	return nil
}
