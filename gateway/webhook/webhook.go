package webhook

import (
	"net/http"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
)

func Serve() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.POST("/reload", reloadConfig)
	e.Logger.Fatal(e.Start(":1323"))
}

func reloadConfig(c echo.Context) error {
	providedToken := c.QueryParam("token")
	secretToken := viper.GetString("ConfigWebhookToken")
	if providedToken != secretToken {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	return c.String(http.StatusAccepted, "Accepted")

}
