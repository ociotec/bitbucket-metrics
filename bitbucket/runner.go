package bitbucket

import (
	"bitbucket-metrics/config"
	"bitbucket-metrics/metrics"
	"time"

	log "github.com/sirupsen/logrus"
)

type Runner struct {
	config  *config.Config
	request *Request
	metrics *metrics.Metrics
}

func NewRunner(config *config.Config, request *Request, metrics *metrics.Metrics) {
	runner := Runner{
		config:  config,
		request: request,
		metrics: metrics,
	}
	go runner.Run()
}

func (runner *Runner) Run() {
	runner.collectMetrics()

	periodInSeconds := runner.config.Bitbucket.Metrics.PeriodInSeconds
	ticker := time.NewTicker(time.Duration(periodInSeconds) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		runner.collectMetrics()
	}
}

type PRKey struct {
	project string
	repo    string
	person  string
}

func (runner *Runner) collectMetrics() {
	start := time.Now()
	log.Info("Collecting metrics...")
	projects, err := Projects(runner.request, runner.config.Bitbucket.Projects.Include)
	if err == nil {
		projecstCount := len(projects)
		reposCount := 0
		prsByAuthor := map[PRKey]int{}
		prsByReviewer := map[PRKey]int{}
		for _, project := range projects {
			log.WithFields(log.Fields{
				"project": project,
			}).Info("Collecting repos...")
			repos, err := Repos(runner.request, project.Key)
			if err == nil {
				reposCount += len(repos)
				for _, repo := range repos {
					log.WithFields(log.Fields{
						"project": project,
						"repo":    repo,
					}).Info("Collecting PRs...")
					prs, err := PRs(runner.request, project.Key, repo.Name)
					if err == nil {
						for _, pr := range prs {
							prKey := PRKey{
								project: project.Key,
								repo:    repo.Name,
								person:  pr.Author,
							}
							if _, ok := prsByAuthor[prKey]; !ok {
								prsByAuthor[prKey] = 0
							}
							prsByAuthor[prKey] += 1

							for _, reviewer := range pr.Reviewers {
								prKey := PRKey{
									project: project.Key,
									repo:    repo.Name,
									person:  reviewer,
								}
								if _, ok := prsByReviewer[prKey]; !ok {
									prsByReviewer[prKey] = 0
								}
								prsByReviewer[prKey] += 1
							}
						}
					}
				}
			}
		}
		runner.metrics.ProjectsGauge.Set(float64(projecstCount))
		runner.metrics.RepositoriesGauge.Set(float64(reposCount))
		for key, value := range prsByAuthor {
			runner.metrics.PRsByAuthorGauge.WithLabelValues(
				key.project,
				key.repo,
				key.person,
			).Set(float64(value))
		}
		for key, value := range prsByReviewer {
			runner.metrics.PRsByReviewerGauge.WithLabelValues(
				key.project,
				key.repo,
				key.person,
			).Set(float64(value))
		}
	}
	elapsed := time.Since(start)
	log.Infof("Metrics collected in %v", elapsed)
}
