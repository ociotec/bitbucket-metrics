package metrics

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	ProjectsGauge         prometheus.Gauge
	RepositoriesGauge     prometheus.Gauge
	PRsByAuthorGauge      *prometheus.GaugeVec
	PRsByReviewerGauge    *prometheus.GaugeVec
	BranchesByAuthorGauge *prometheus.GaugeVec
	TagsByAuthorGauge     *prometheus.GaugeVec
	CollectTimeGauge      prometheus.Gauge
}

func ListenAndServe(hostname string, port uint16, path string, takeMetrics func(metrics *Metrics)) {
	metrics := Metrics{
		ProjectsGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "bitbucket_projects",
				Help: "Number of Bitbucket projects being monitored",
			},
		),
		RepositoriesGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "bitbucket_repositories",
				Help: "Number of Bitbucket repositories being monitored",
			},
		),
		PRsByAuthorGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "bitbucket_prs_by_author",
				Help: "Number of Bitbucket PRs by author being monitored",
			},
			[]string{"project", "repo", "author"},
		),
		PRsByReviewerGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "bitbucket_prs_by_reviewer",
				Help: "Number of Bitbucket PRs by reviewer being monitored",
			},
			[]string{"project", "repo", "reviewer"},
		),
		BranchesByAuthorGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "bitbucket_branches_by_author",
				Help: "Number of Bitbucket branches by author being monitored",
			},
			[]string{"project", "repo", "author"},
		),
		TagsByAuthorGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "bitbucket_tags_by_author",
				Help: "Number of Bitbucket tags by author being monitored",
			},
			[]string{"project", "repo", "author"},
		),
		CollectTimeGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "bitbucket_collect_time",
				Help: "Bitbucket metrics collect time in milliseconds",
			},
		),
	}
	takeMetrics(&metrics)

	prometheus.MustRegister(
		metrics.ProjectsGauge,
		metrics.RepositoriesGauge,
		metrics.PRsByAuthorGauge,
		metrics.PRsByReviewerGauge,
		metrics.BranchesByAuthorGauge,
		metrics.TagsByAuthorGauge,
		metrics.CollectTimeGauge,
	)

	http.Handle(path, promhttp.Handler())

	log.WithFields(log.Fields{
		"hostname": hostname,
		"port":     port,
		"path":     path,
		"url":      fmt.Sprintf("http://%v:%v%v", hostname, port, path),
	}).Info("Serving metrics via HTTP")
	http.ListenAndServe(":"+fmt.Sprint(port), nil)
}
