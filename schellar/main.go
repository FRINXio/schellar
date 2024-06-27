package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/frinx/schellar/graph"
	"github.com/frinx/schellar/ifc"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/frinx/schellar/scheduler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const defaultPort = "3000"
const defaultPlaygoundQueryEndpoint = "/query"

var config = scheduler.Configuration

func main() {

	setupLogging()

	if err := scheduler.StartScheduler(); err != nil {
		logrus.Fatalf("Error during scheduler startup: %v", err)
	}

	port := getPort()
	startApi()

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func setupLogging() {
	logLevel := ifc.GetEnvOrDefault("LOG_LEVEL", "DEBUG")
	switch strings.ToLower(logLevel) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return port
}

func startApi() {

	playgroundQeryEndpoint := os.Getenv("PLAYGROUND_QUERY_ENDPOINT")
	if playgroundQeryEndpoint == "" {
		playgroundQeryEndpoint = defaultPlaygoundQueryEndpoint
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		oc := graphql.GetOperationContext(ctx)

		userHeader := oc.Headers.Get("from")
		roleHeader := []string{oc.Headers.Get("x-auth-user-roles")}
		groupHeader := []string{oc.Headers.Get("x-auth-user-groups")}

		logrus.WithFields(logrus.Fields{
			"timestamp": timestamp,
			"operation": oc.OperationName,
			"user":      userHeader,
			"roles":     roleHeader,
			"groups":    groupHeader,
		}).Info("Audit: ")

		if userHeader == "" {
			return func(ctx context.Context) *graphql.Response {
				logrus.Warnf("Missing header From")
				return &graphql.Response{
					Errors: []*gqlerror.Error{
						gqlerror.Errorf("Missing header From"),
					},
				}
			}
		}

		return next(ctx)
	})

	http.Handle("/", playground.ApolloSandboxHandler("GraphQL playground", playgroundQeryEndpoint))
	http.Handle("/query", srv)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/liveness", getLiveness)
}

func getLiveness(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("getLiveness r r=%v", r)

	if r.Method == http.MethodOptions {
		return
	}
	w.Write([]byte("OK"))
}
