package http

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"

	"github.com/micromdm/nanoaxm/cryptoutil"
	"github.com/micromdm/nanoaxm/storage"
	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
)

//go:embed authcreds.html
var form []byte

// NewAuthCredsSaveFormHandler creates a handler for configuring authentication credentials in store.
// GET requests serve out an HTML form.
// POST handles the submission of said form.
func NewAuthCredsSaveFormHandler(store storage.AuthCredentialsStorer, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write(form)
		case http.MethodPost:
			logger := ctxlog.Logger(r.Context(), logger)

			err := r.ParseMultipartForm(1 << 16) // 65KB
			if err != nil {
				logger.Info("msg", "parsing form", "err", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			ac := storage.AuthCredentials{
				ClientID: r.FormValue("client_id"),
				KeyID:    r.FormValue("key_id"),
			}

			file, _, err := r.FormFile("private_key")
			if err != nil {
				logger.Info("msg", "parsing form file", "err", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			defer file.Close()

			ac.PrivateKeyPEM, err = io.ReadAll(file)
			if err != nil {
				logger.Info("msg", "reading form file", "err", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			_, err = cryptoutil.ECPrivateKeyFromPEM(ac.PrivateKeyPEM)
			if err != nil {
				logger.Info("msg", "parsing private key", "err", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			axmName := r.FormValue("axm_name")

			err = store.StoreAuthCredentials(r.Context(), axmName, ac)
			if err != nil {
				logger.Info("msg", "reading form file", "err", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			logger.Debug("msg", "stored auth credentials", "client_id", ac.ClientID)

			fmt.Fprintf(w, "Saved authentication credentials for AXM name: %s (Client ID %s)\n", axmName, ac.ClientID)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	}
}
