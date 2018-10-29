package webhook

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"time"

	r "gopkg.in/redis.v3"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"github.com/matryer/vice/queues/redis"
	log "github.com/sirupsen/logrus"
)

func Serve() {
	//flog := log.WithFields(log.Fields{"prefix": "webhook"})
	e := echo.New()
	e.Use(middleware.Logger())
	e.POST("/reload", reloadConfig)
	e.Logger.Fatal(e.Start(":1323"))
}

func Reloader(ctx context.Context, reloads chan<- []byte, errs <-chan error) {
	for {
		select {
		case <-ctx.Done():
			log.Println("finished")
			return
		case err := <-errs:
			log.Println("an arror ocurred:", err)
		default:
			reloads <- []byte("reload")
			return
		}
	}
}

func reloadConfig(c echo.Context) error {
	flog := log.WithFields(log.Fields{"prefix": "webhook"})
	providedToken := c.QueryParam("token")
	secretToken := viper.GetString("ConfigWebhookToken")
	if providedToken != secretToken {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	db_url := os.Getenv("REDIS_URL")
	if db_url == "" {
		flog.Fatalln("REDIS_URL not configured.")
	}
	u, _ := url.ParseRequestURI(os.Getenv("REDIS_URL"))
	password, _ := u.User.Password()
	client := r.NewClient(&r.Options{
		Addr:     u.Host,
		Password: password, // no password set
		DB:       0,  // use default DB
	})
	transport := redis.New(redis.WithClient(client))
	defer func() {
		transport.Stop()
		<-transport.Done()
	}()

	reloads := transport.Send("reload")
	Reloader(ctx, reloads, transport.ErrChan())


	return c.String(http.StatusAccepted, "Accepted")

}
