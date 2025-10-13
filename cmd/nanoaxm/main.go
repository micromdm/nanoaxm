package main

import (
	"flag"
	"fmt"
	stdlog "log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/micromdm/nanoaxm/client"
	axmhttp "github.com/micromdm/nanoaxm/http"
	"github.com/micromdm/nanoaxm/http/proxy"

	"github.com/google/uuid"
	"github.com/micromdm/nanolib/envflag"
	libhttp "github.com/micromdm/nanolib/http"
	"github.com/micromdm/nanolib/http/trace"
	"github.com/micromdm/nanolib/log/stdlogfmt"
)

// overridden by -ldflags -X
var version = "unknown"

const (
	apiUsername = "nanoaxm"
)

func main() {
	var (
		flDebug   = flag.Bool("debug", false, "log debug messages")
		flListen  = flag.String("listen", ":9005", "HTTP listen address")
		flAPIKey  = flag.String("api", "", "API key for API endpoints")
		flVersion = flag.Bool("version", false, "print version and exit")
		flStorage = flag.String("storage", "file", "storage backend")
		flDSN     = flag.String("storage-dsn", "", "storage backend data source name")
		flOptions = flag.String("storage-options", "", "storage backend options")
	)
	envflag.Parse("NANOAXM_", []string{"version"})

	if *flVersion {
		fmt.Println(version)
		return
	}

	if *flAPIKey == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "empty API key\n")
		flag.Usage()
		os.Exit(1)
	}

	logger := stdlogfmt.New(
		stdlogfmt.WithLogger(stdlog.Default()),
		stdlogfmt.WithDebugFlag(*flDebug),
	)

	store, err := newStore(*flStorage, *flDSN, *flOptions, logger)
	if err != nil {
		logger.Info("msg", "creating storage backend", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/version", libhttp.NewJSONVersionHandler(version))

	mwmux := libhttp.NewMWMux(mux)
	mwmux.Use(func(h http.Handler) http.Handler {
		return libhttp.NewSimpleBasicAuthHandler(h, apiUsername, *flAPIKey, "NanoAXM")
	})

	mwmux.Handle("/authcreds", axmhttp.NewAuthCredsSaveFormHandler(store, logger.With("handler", "auth-creds-save-form")))

	proxyLogger := logger.With("handler", "proxy")

	mwmux.Handle("/proxy/business/",
		http.StripPrefix("/proxy/business/",
			axmhttp.DelHeaderMiddleware(
				proxy.NewNameMiddleware(
					proxy.New(
						client.NewTransport(
							http.DefaultTransport,
							http.DefaultClient,
							store,
							uuid.NewString,
						),
						"https://api-business.apple.com",
						proxyLogger,
					),
					proxyLogger,
				),
				"Authorization",
			),
		),
	)

	mwmux.Handle("/proxy/school/",
		http.StripPrefix("/proxy/school/",
			axmhttp.DelHeaderMiddleware(
				proxy.NewNameMiddleware(
					proxy.New(
						client.NewTransport(
							http.DefaultTransport,
							http.DefaultClient,
							store,
							uuid.NewString,
						),
						"https://api-school.apple.com",
						proxyLogger,
					),
					proxyLogger,
				),
				"Authorization",
			),
		),
	)

	// init for newTraceID()
	rand.Seed(time.Now().UnixNano())

	logger.Info("msg", "starting server", "listen", *flListen)
	err = http.ListenAndServe(*flListen, trace.NewTraceLoggingHandler(mux, logger.With("handler", "log"), newTraceID))
	logs := []interface{}{"msg", "server shutdown"}
	if err != nil {
		logs = append(logs, "err", err)
	}
	logger.Info(logs...)
}

// newTraceID generates a new HTTP trace ID for context logging.
// Currently this just makes a random string. This would be better
// served by e.g. https://github.com/oklog/ulid or something like
// https://opentelemetry.io/ someday.
func newTraceID(*http.Request) string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
