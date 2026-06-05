package health

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Handler(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "unhealthy",
			})
		}
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	}
}
