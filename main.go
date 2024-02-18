// SPDX-FileCopyrightText: (c) Mauve Mailorder Software GmbH & Co. KG, 2022. Licensed under [MIT](LICENSE) license.
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"net/http"

	"github.com/MauveSoftware/ilo_exporter/pkg/chassis"
	"github.com/MauveSoftware/ilo_exporter/pkg/client"
	"github.com/MauveSoftware/ilo_exporter/pkg/manager"
	"github.com/MauveSoftware/ilo_exporter/pkg/system"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	initConfig()
	startServer()
}

func startServer() {
	logrus.Infof("Starting iLO exporter (Version: %s)", version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>iLO5 Exporter (Version ` + version + `)</title></head>
			<body>
			<h1>iLO Exporter by Mauve Mailorder Software</h1>
			<h2>Example</h2>
			<p>Metrics for host 172.16.0.200</p>
			<p><a href="` + conf.Web.MetricsPath + `?host=172.16.0.200">` + r.Host + conf.Web.MetricsPath + `?host=172.16.0.200</a></p>
			<h2>More information</h2>
			<p><a href="https://github.com/MauveSoftware/ilo_exporter">github.com/MauveSoftware/ilo_exporter</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc(conf.Web.MetricsPath, errorHandler(handleMetricsRequest))

	logrus.Infof("Listening for %s on %s (TLS: %v)", conf.Web.MetricsPath, conf.Web.ListenAddress, conf.Tls.Enabled)
	if conf.Tls.Enabled {
		logrus.Fatal(http.ListenAndServeTLS(conf.Web.ListenAddress, conf.Tls.CertChainPath, conf.Tls.KeyPath, nil))
		return
	}

	logrus.Fatal(http.ListenAndServe(conf.Web.ListenAddress, nil))
}

func errorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)

		if err != nil {
			logrus.Errorln(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	host := q.Get("host")

	// If username and/or password is provided as a request parameter, use
	// them. Otherwise fallback to the ones provided at startup (if any).
	user := q.Get("user")
	if user == "" {
		user = conf.Api.Username
	}
	pass := q.Get("pass")
	if pass == "" {
		pass = conf.Api.Password
	}

	if host == "" {
		return fmt.Errorf("no host defined")
	}

	reg := prometheus.NewRegistry()

	cl := client.NewClient(
		host, user, pass, conf.Api.Debug,
		client.WithMaxConcurrentRequests(conf.Api.MaxConcurrentRequests),
		client.WithInsecure())

	ctx := r.Context()
	reg.MustRegister(system.NewCollector(ctx, cl))
	reg.MustRegister(manager.NewCollector(ctx, cl))
	reg.MustRegister(chassis.NewCollector(ctx, cl))

	l := logrus.New()
	l.Level = logrus.ErrorLevel

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      l,
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
	return nil
}
