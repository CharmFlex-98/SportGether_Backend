package main // "main" indicating this is main app, not library. Other files cannot import this.
import (
	"flag"
	"log/slog"
	"os"
)

type config struct {
	port int
	env  string
}
type Application struct {
	config config
	logger *slog.Logger
}

func main() {
	config := config{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	flag.IntVar(&config.port, "port", 3000, "Port for listening")
	flag.StringVar(&config.env, "environment", "dev", "environment dev|prd")

	flag.Parse()

	app := Application{
		config: config,
		logger: logger,
	}

	app.serve()
}
