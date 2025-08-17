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

type ProjectRepoPersonKey struct {
	project string
	repo    string
	person  string
}

func (runner *Runner) collectPRs(project Project, repo Repo, prsByAuthor, prsByReviewer map[ProjectRepoPersonKey]int) {
	log.WithFields(log.Fields{
		"project": project.Key,
		"repo":    repo.Name,
	}).Info("Collecting PRs...")
	prs, err := PRs(runner.request, project.Key, repo.Name)
	if err == nil {
		for _, pr := range prs {
			prKey := ProjectRepoPersonKey{
				project: project.Key,
				repo:    repo.Name,
				person:  pr.Author,
			}
			if _, ok := prsByAuthor[prKey]; !ok {
				prsByAuthor[prKey] = 0
			}
			prsByAuthor[prKey] += 1

			for _, reviewer := range pr.Reviewers {
				prKey := ProjectRepoPersonKey{
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

func (runner *Runner) collectReferences(project Project, repo Repo, references []Reference, referencesByAuthor map[ProjectRepoPersonKey]int) {
	for _, reference := range references {
		prKey := ProjectRepoPersonKey{
			project: project.Key,
			repo:    repo.Name,
			person:  reference.Author,
		}
		if _, ok := referencesByAuthor[prKey]; !ok {
			referencesByAuthor[prKey] = 0
		}
		referencesByAuthor[prKey] += 1
	}
}

func (runner *Runner) collectBranchesAndTags(project Project, repo Repo, branchesByAuthor, tagsByAuthor map[ProjectRepoPersonKey]int) {
	log.WithFields(log.Fields{
		"project": project.Key,
		"repo":    repo.Name,
	}).Info("Collecting branches & tags...")
	branches, tags, err := References(runner.request, project.Key, repo.Name)
	if err == nil {
		runner.collectReferences(project, repo, branches, branchesByAuthor)
		runner.collectReferences(project, repo, tags, tagsByAuthor)
	}
}

func (runner *Runner) collectMetrics() {
	start := time.Now()
	log.Info("Collecting metrics...")
	projects, err := Projects(runner.request, runner.config.Bitbucket.Projects.Include)
	if err == nil {
		projecstCount := len(projects)
		reposCount := 0
		prsByAuthor := map[ProjectRepoPersonKey]int{}
		prsByReviewer := map[ProjectRepoPersonKey]int{}
		branchesByAuthor := map[ProjectRepoPersonKey]int{}
		tagsByAuthor := map[ProjectRepoPersonKey]int{}
		for _, project := range projects {
			log.WithFields(log.Fields{
				"project": project.Key,
			}).Info("Collecting repos...")
			repos, err := Repos(runner.request, project.Key)
			if err == nil {
				reposCount += len(repos)
				for _, repo := range repos {
					runner.collectPRs(project, repo, prsByAuthor, prsByReviewer)
					runner.collectBranchesAndTags(project, repo, branchesByAuthor, tagsByAuthor)
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
		for key, value := range branchesByAuthor {
			runner.metrics.BranchesByAuthorGauge.WithLabelValues(
				key.project,
				key.repo,
				key.person,
			).Set(float64(value))
		}
		for key, value := range tagsByAuthor {
			runner.metrics.TagsByAuthorGauge.WithLabelValues(
				key.project,
				key.repo,
				key.person,
			).Set(float64(value))
		}
	}
	elapsed := time.Since(start)
	runner.metrics.CollectTimeGauge.Set(float64(elapsed.Milliseconds()))
	log.Infof("Metrics collected in %v", elapsed)
}
