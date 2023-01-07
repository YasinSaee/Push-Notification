package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http/pprof"

	"strings"

	"github.com/YasinSaee/Push-Notification/connection/mongo"
	"github.com/YasinSaee/Push-Notification/context"
	session "github.com/ipfans/echo-session"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	ConfigPath   = flag.String("c", "config.json", "config path")
	EC           = echo.New()
	sessionStore = session.NewCookieStore([]byte("This should be change %Secure%"))
	Middlewares  = []echo.MiddlewareFunc{
		session.Sessions("SHOP", sessionStore),
		fixURL,
		// NOTE: For development only
		//middleware.GzipWithConfig(middleware.GzipConfig{}),
		middleware.Secure(),
	}
)

type Config struct {
	Notification NotificationConfig `json:"notification"`
	Auth         AuthConfig         `json:"authentication"`
}

type NotificationConfig struct {
	Port          int    `json:"port"`
	MongoDB       string `json:"mongo_db" valid:"required"`
	MongoHost     string `json:"mongo_host" required:"required"`
	MongoPort     int    `json:"mongo_port" required:"required"`
	Domain        string `json:"domain" required:"required"`
	Debug         bool   `json:"debug"`
	DefaultLang   string `json:"default_lang"`
	RefreshConfig string `json:"refresh_config"`
}

// ConfigStruct struct config
type AuthConfig struct {
	AccessToken  TokenConfig `json:"access_token"`
	RefreshToken TokenConfig `json:"refresh_token"`
	OtpToken     TokenConfig `json:"otp_token"`
	UseRedis     bool        `json:"use_redis"`
}

type TokenConfig struct {
	ExpireTime int64  `json:"expire_time"`
	SecretKey  string `json:"secret_key"`
}

var (
	Cgf = Config{
		Auth: AuthConfig{
			UseRedis: false,
		},
	}
	MongoConn *mongo.Conn
)

func loadConfig() {
	file, err := ioutil.ReadFile("config/config.json")

	if err == nil {
		if err := json.Unmarshal(file, &Cgf); err != nil {
			panic(err)
		}
	}
}

func Run() {
	var (
		err error
	)
	loadConfig()

	if Cgf.Notification.Debug {
		Middlewares = append(Middlewares, middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: middleware.DefaultCORSConfig.AllowMethods,
		}))
		loadProfTools()
	}
	EC.Use(Middlewares...)
	if !Cgf.Notification.Debug {
		EC.Use(middleware.Recover())
	}

	EC.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &context.GlobalContext{Context: c}
			return next(cc)
		}
	})

	// Initial mongodb connection
	MongoConn, err = mongo.InitMongoConnection(mongo.Config{DBName: Cgf.Notification.MongoDB, URI: Cgf.Notification.MongoHost})
	if err != nil {
		panic(err)
	}

	EC.Logger.Fatal(EC.Start(fmt.Sprintf(":%d", Cgf.Notification.Port)))
}

func fixURL(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		if len(req.URL.Path) < 2 {
			return next(c)
		}
		if string(req.URL.Path[len(req.URL.Path)-1]) != "/" {
			return next(c)
		}
		path := strings.TrimRight(req.URL.Path, "/")
		return c.Redirect(302, path)
	}
}

func loadProfTools() {
	EC.POST("/debug/prof", ProfileHandler())
}

func ProfileHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		println("proffiling . . .")
		pprof.Profile(ctx.Response().Writer, ctx.Request())
		return nil
	}
}
