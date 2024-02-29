package main // "main" indicating this is main app, not library. Other files cannot import this.
import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"os"
	"sportgether/internal/models"
	"time"

	"sportgether/internal/mailer"

	firebase "firebase.google.com/go/v4"
	"github.com/cloudinary/cloudinary-go/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/api/option"
)

type sslCertConfig struct {
	certPath string
	certKey  string
}

type config struct {
	port     int
	env      string
	dbConfig struct {
		dsn               string
		maxOpenConnection int
		maxIdleConnection int
		maxIdleTime       time.Duration
	}
	smtp struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Sender   string `json:"sender"`
	}
}

func (c config) getCertConfig() sslCertConfig {
	if c.env == "PRD" {
		return sslCertConfig{
			certPath: "./cert/prd_ca_cert.pem",
			certKey:  "./cert/prd_ca_key.pem",
		}
	} else {
		return sslCertConfig{
			certPath: "cert.crt",
			certKey:  "cert.key",
		}
	}
}

type Application struct {
	config        config
	logger        *slog.Logger
	daos          models.Daos
	firebaseApp   *firebase.App
	cloudinaryApp *cloudinary.Cloudinary
	mailer        mailer.Mailer
}

func main() {
	config := config{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cld := credentials()

	flag.IntVar(&config.port, "port", 3000, "Port for listening")
	config.env = os.Getenv("ENV")
	if config.env == "" {
		config.env = "DEV"
	}

	opt := option.WithCredentialsFile("./data/service-account-file.json")
	firebaseApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	flag.Parse()

	db, err := config.openDatabase()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	initSmtpConfig(&config)

	app := Application{
		config:        config,
		logger:        logger,
		daos:          models.NewDaoHandler(db),
		firebaseApp:   firebaseApp,
		cloudinaryApp: cld,
		mailer:        mailer.New(config.smtp.Host, config.smtp.Port, config.smtp.Username, config.smtp.Password, config.smtp.Sender),
	}

	err = app.serve()
	if err != nil {
		app.logInfo("error: %s, stopping server...", err)
		os.Exit(1)
	}
}

func initSmtpConfig(c *config) {
	var path string
	if c.isProd() {
		path = "./data/smtp_config_prod"
	} else {
		path = "./data/smtp_config_dev"
	}
	readJsonFromFile(path, &c.smtp)
}

func credentials() *cloudinary.Cloudinary {
	cld, _ := cloudinary.New()
	cld.Config.URL.Secure = true
	return cld
}

func (cfg config) openDatabase() (*sql.DB, error) {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.dbConfig.maxOpenConnection)
	db.SetMaxIdleConns(cfg.dbConfig.maxIdleConnection)
	db.SetConnMaxIdleTime(cfg.dbConfig.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (c config) isProd() bool {
	return c.env == "PRD"
}
