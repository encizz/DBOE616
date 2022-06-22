package contacts

import (
	"net/http"

	"github.com/kataras/iris"
)

func DeleteContact(c iris.Context) (code int, data interface{}) {

	action := Route.First
	contact := Route.Second

	switch action {
	case "detail":
		if contact == "" {
			code, data = exceptions.StatusCode(http.StatusForbidden)
		} else {
			code, data = http.StatusOK, iris.Map{"message": "Le contact a été supprimé avec succès avec succès !"}
		}
	default:
		code, data = exceptions.StatusCode(http.StatusNotFound)
	}

	return
}
