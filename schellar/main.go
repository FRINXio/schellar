package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/frinx/schellar/graph"
	"github.com/frinx/schellar/ifc"
	"github.com/frinx/schellar/scheduler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const defaultPort = "3000"

var config = scheduler.Configuration

func main() {

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

	var err error
	err = scheduler.StartScheduler()
	if err != nil {
		logrus.Fatalf("Error during scheduler startup: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	startApi()

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func startApi() {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
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
