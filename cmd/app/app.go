package main

type App struct {
	conf *Config
}

func New(conf *Config) *App {
	return &App{
		conf: conf,
	}
}

func (a *App) Run() error {

	return nil
}
