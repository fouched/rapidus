package rapidus

import (
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/fouched/rapidus/cache"
	"github.com/fouched/rapidus/mailer"
	"github.com/fouched/rapidus/render"

	//"github.com/fouched/rapidus/render"
	"github.com/fouched/rapidus/session"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const version = "1.0.0"

var badgerCache *cache.BadgerCache
var badgerConn *badger.DB

type Rapidus struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        render.Render
	Session       *scs.SessionManager
	DB            Database
	config        config // no reason to export this
	EncryptionKey string
	RedisClient   *redis.Client
	Cache         cache.Cache
	Mail          mailer.Mail
	Server        Server
}

type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConfig
}

func (r *Rapidus) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "mail", "data", "public", "tmp", "logs", "middleware"},
	}

	err := r.Init(pathConfig)
	if err != nil {
		return err
	}

	err = r.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// create loggers
	infoLog, errorLog := r.startLoggers()

	// create Rapidus configuration
	r.InfoLog = infoLog
	r.ErrorLog = errorLog
	r.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	r.Version = version
	r.RootPath = rootPath
	r.Mail = r.createMailer()
	r.Routes = r.routes().(*chi.Mux)

	// connect to database if specified
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := r.OpenDB(os.Getenv("DATABASE_TYPE"), r.BuildDSN())
		//TODO: build in retry
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		r.DB = Database{
			Type: os.Getenv("DATABASE_TYPE"),
			Pool: db,
		}
	}

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		r.RedisClient = r.createRedisClient()
	}

	if os.Getenv("CACHE") == "badger" {
		badgerCache = r.createBadgerCache()
		r.Cache = badgerCache
		badgerConn = badgerCache.Conn

		go func() {
			for range time.Tick(12 * time.Hour) {
				_ = badgerCache.Conn.RunValueLogGC(0.7)
			}
		}()
	}

	// setup config
	r.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"), // we probably don't need this with templ
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSIST"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			dsn:      r.BuildDSN(),
			database: os.Getenv("DATABASE_TYPE"),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	// create session
	s := session.Session{
		CookieLifetime: r.config.cookie.lifetime,
		CookiePersist:  r.config.cookie.persist,
		CookieSecure:   r.config.cookie.secure,
		CookieName:     r.config.cookie.name,
		CookieDomain:   r.config.cookie.domain,
		SessionType:    r.config.sessionType,
		DBPool:         r.DB.Pool,
	}

	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}
	r.Server = Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}

	switch r.config.sessionType {
	case "redis":
		s.RedisPool = r.RedisClient
	case "mysql", "mariadb", "postgres", "postgresql":
		s.DBPool = r.DB.Pool
	}

	r.Session = s.InitSession()

	// encryption key
	r.EncryptionKey = os.Getenv("KEY")

	// create renderer
	r.Render = render.Render{Session: r.Session}

	// listen for mail requests
	go r.Mail.ListenForMail()

	return nil
}

func (r *Rapidus) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if it does not exist
		err := r.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Rapidus) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     r.ErrorLog,
		Handler:      r.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	if r.DB.Pool != nil {
		defer r.DB.Pool.Close()
	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	r.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	r.ErrorLog.Fatal(err)
}

func (r *Rapidus) checkDotEnv(path string) error {
	err := r.CreateFileIfNotExist(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

func (r *Rapidus) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

//func (r *Rapidus) createRenderer() {
//	myRenderer := render.Render{Session: r.Session}
//
//	r.Render = &myRenderer
//}

func (r *Rapidus) createMailer() mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   r.RootPath + "/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromAddress: os.Getenv("MAIL_FROM_NAME"),
		FromName:    os.Getenv("MAIL_FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}

	return m
}

func (r *Rapidus) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"),
		)
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}
	}

	return dsn
}

func (r *Rapidus) createRedisClient() *redis.Client {

	// PoolSize int
	// Base number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	// If there is not enough connections in the pool, new connections will be allocated in excess of PoolSize,
	// you can limit it through MaxActiveConns

	return redis.NewClient(&redis.Options{
		Addr:     r.config.redis.host,
		Password: r.config.redis.password,
		DB:       0, // use default DB
	})
}

func (r *Rapidus) createBadgerCache() *cache.BadgerCache {
	cacheClient := cache.BadgerCache{
		Conn: r.createBadgerConn(),
	}

	return &cacheClient
}

func (r *Rapidus) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(r.RootPath + "/tmp/badger"))

	if err != nil {
		return nil
	}
	return db
}
