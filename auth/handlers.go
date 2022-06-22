package handlers

package auth

import (
	"fmt"
	"main/email"
	"main/models/users"
	"net/http"

	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
)

type LoginForm struct {
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type PasswordForm struct {
	ResetToken string `form:"reset_token" json:"reset_token" binding:"required"`
	Password   string `form:"password" json:"password" binding:"required"`
}

type ResetForm struct {
	Email string `form:"email" json:"email" binding:"required"`
}

// Login:
// Handler which respond on '/auth/login'
func Login(c iris.Context) {
	// Verification des information reçus si le fichier n'est pas vide et si le password contient plus de 6 caractères
	var loginForm LoginForm
	if err := c.ReadBody(&loginForm); err != nil || len(loginForm.Password) == 0 || loginForm.Email == "" {
		c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": "Email ou Mot de passe vide"})
		return
	}

	// Check email validFormat
	if !validEmail(loginForm.Email) {
		c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": "Format d'email incorrect"})
		return
	}

	// Check password
	if err := validPassword(loginForm.Password); err != nil {
		c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": err.Error()})
		return
	}

	// Construction de l'objet CurentUserToken à enregistrer en bdd rédis
	CurrentUserToken, err := GetSQLUserToken(loginForm.Email, loginForm.Password)
	if err != nil {
		c.StopWithJSON(http.StatusForbidden, iris.Map{"message": "Email ou Mot de passe invalide"})
		return
	}

	// Appel de la fonction de TokensModel pour l'enregistrement du UserToken en bdd rédis
	TokenID := CurrentUserToken.Store()
	if TokenID == "" {
		golog.Error("during store access token memory")
		c.StopWithJSON(http.StatusInternalServerError, iris.Map{"message": "Erreur de serveur interne"})
		return
	}

	// Envoie de la réponse en JSON avec status de la réponse et le TokenId
	c.StopWithJSON(http.StatusOK, iris.Map{"token": TokenID})
}

// Reset:
// This function handle the request of a reset password on auth/reset.
func Reset(emailConfig email.Config, frontURL string) func(c iris.Context) {
	return func(c iris.Context) {

		var resetForm ResetForm
		// Check email vide
		if err := c.ReadBody(&resetForm); err != nil || resetForm.Email == "" {
			c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": "Format d'email incorrect"})
			return
		}

		// Check email validate
		if !validEmail(resetForm.Email) {
			c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": "Format d'email incorrect"})
			return
		}

		CurrentUser := users.User{
			Email: resetForm.Email,
		}

		message := "Vous allez recevoir un mail avec les étapes de réinitialisation du mot de passe"

		ResetToken := CurrentUser.GenResetToken()
		if ResetToken == "" {
			c.StopWithJSON(http.StatusCreated, iris.Map{"message": message})
			return
		}

		subject := "Réinitialisation de votre mot de passe"
		text := "Une demande de réinitialisation de votre mot de passe pour notre portail client a été fait avec votre adresse mail. " +
			"Pour compléter ce processus, cliquez sur le bouton ci-dessous :"
		btnText := "REDEFINIR MOT DE PASSE"
		btnURL := frontURL + "/reset-password" + "?reset_token=" + ResetToken

		sucessfullySent := email.New(CurrentUser.Email, subject, subject, text, btnText, btnURL).Send(emailConfig)
		if !sucessfullySent {
			golog.Error("failed send mail to;", CurrentUser.Email)
			c.StopWithJSON(http.StatusInternalServerError, iris.Map{"message": "Le mail de réinitialisation n'a pu être envoyé <a href=''>-RENDEZ-VOUS ICI-</a>"})
			return
		} else {
			fmt.Println("[INFO]", "reset email sucessfully sent to:", CurrentUser.Email)
		}

		c.StopWithJSON(http.StatusCreated, iris.Map{"message": message})
	}
}

// Password:
// is the handler which define the password sent over PasswordForm
func Password(c iris.Context) {

	// Check mot de passe et jeton de réinitialisation
	var passwordForm PasswordForm
	if err := c.ReadBody(&passwordForm); err != nil || len(passwordForm.Password) == 0 {
		c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": "Jeton de réinitialisation ou mot de passe vide"})
		return
	}

	// Check mot de passe format
	if err := validPassword(passwordForm.Password); err != nil {
		c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": err.Error()})
		return
	}

	success := users.DefinePasswordWithResetToken(passwordForm.ResetToken, passwordForm.Password)
	if !success {
		c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": "Jeton de réinitialisation ou mot de passe invalide"})
		return
	}

	c.StopWithJSON(http.StatusCreated, iris.Map{"message": "Votre mot de passe à été définie avec succès !"})
}

func Logout(c iris.Context) {
	var h Header
	if err := c.ReadHeaders(&h); err != nil || h.TokenID == "" || len(h.TokenID) < 20 || len(h.TokenID) > 25 {
		c.StopWithJSON(http.StatusUnauthorized, iris.Map{"message": "Session invalide"})
		return
	}

	if ok := RevokeUserToken(h.TokenID); ok {
		c.StopWithJSON(http.StatusCreated, iris.Map{"message": "Vous avez été déconnecté.e"})
	} else {
		c.StopWithJSON(http.StatusInternalServerError, iris.Map{"message": "Erreur interne au serveur"})
	}
}

// Me:
// return the full userToken values in JSON
// for test purpose only.
// Afterward, this handler must be modify for return only usefull data.
func Me(c iris.Context) {
	var h Header
	if err := c.ReadHeaders(&h); err != nil || h.TokenID == "" || len(h.TokenID) < 20 || len(h.TokenID) > 25 {
		c.StopWithJSON(http.StatusBadRequest, iris.Map{"message": "Session invalide"})
		return
	}

	User, err := GetUserToken(h.TokenID)
	if err != nil {
		c.StopWithJSON(http.StatusUnauthorized, iris.Map{"message": "Votre session a expirée"})
		return
	}

	c.StopWithJSON(http.StatusOK, User)
}
