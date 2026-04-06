package main

type App struct {
	conf *Config
}

func NewApp(conf *Config) *App {
	return &App{
		conf: conf,
	}
}

func (a *App) Run() error {

	return nil
}
