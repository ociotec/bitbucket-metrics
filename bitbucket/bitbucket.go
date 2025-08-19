package bitbucket

import (
	"errors"
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
)

const API_PATH = "rest/api/latest"

func Init(bitbucketBaseURL, username, password string, apiPageSize int) *Request {
	request, err := NewRequest(bitbucketBaseURL, username, password, apiPageSize)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("Cannot create the request")
	}

	version, err := version(request)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("Cannot get Bitbucket version")
	}
	log.Infof("Bitbucket v%s", version)
	request.BitbucketVersion = version

	return request
}

func version(request *Request) (string, error) {
	result, err := request.Run("GET", API_PATH, "application-properties")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Cannot run the request")
		return "", err
	}

	version, ok := result["version"].(string)
	if !ok {
		log.WithFields(log.Fields{
			"json": result,
		}).Error("Cannot extract Bitbucket version")
		message := fmt.Sprintf("Cannot extract Bitbucket version from JSON '%v'", result)
		return "", errors.New(message)
	}

	return version, nil
}

func paginatedValues(request *Request, path string, params map[string]string, valueProcessor func(map[string]any)) error {
	lastPage := false
	start := 0
	for !lastPage {
		args := map[string]any{
			"limit": request.PageSize,
			"start": start,
		}
		for name, value := range params {
			args[name] = value
		}
		result, err := request.RunWithArgs("GET", args, API_PATH, path)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Cannot run the request")
			return err
		}
		log.Debugf("Result: %v", result)

		// Parse required fields
		valuesJSON, okValuesJSON := result["values"].([]any)
		if okValuesJSON {
			for _, value := range valuesJSON {
				valueJSON, okValueJSON := value.(map[string]any)
				if !okValueJSON {
					continue
				}
				valueProcessor(valueJSON)
			}
		}

		// Check if there are more pages to get
		var ok bool
		lastPage, ok = result["isLastPage"].(bool)
		if !ok {
			lastPage = true
		}
		if !lastPage {
			// And the next page start offset
			nextPageStart, ok := result["nextPageStart"].(float64)
			if !ok {
				lastPage = true
			} else {
				start = int(nextPageStart)
			}
		}
	}
	return nil
}

type Project struct {
	Key         string
	Name        string
	Description string
}

func Projects(request *Request, includeProjects []string) (map[string]Project, error) {
	projects := map[string]Project{}
	paginatedValues(request, "projects", nil, func(valueJSON map[string]any) {
		key, okKey := valueJSON["key"].(string)
		name, okName := valueJSON["name"].(string)
		description, okDescription := valueJSON["description"].(string)
		if okKey && okName && okDescription {
			if includeProjects == nil || slices.Contains(includeProjects, key) {
				projects[key] = Project{
					Key:         key,
					Name:        name,
					Description: description,
				}
			}
		}
	})
	return projects, nil
}

type Repo struct {
	Name string
}

func Repos(request *Request, project string) (map[string]Repo, error) {
	repos := map[string]Repo{}
	path := fmt.Sprintf("projects/%s/repos", project)
	paginatedValues(request, path, nil, func(valueJSON map[string]any) {
		name, okName := valueJSON["name"].(string)
		if okName {
			repos[name] = Repo{
				Name: name,
			}
		}
	})
	return repos, nil
}

type PR struct {
	Name      string
	State     string
	Author    string
	Reviewers []string
}

func PRs(request *Request, project string, repo string) ([]PR, error) {
	var prs []PR
	path := fmt.Sprintf("projects/%s/repos/%s/pull-requests", project, repo)
	params := map[string]string{
		"state": "ALL",
	}
	err := paginatedValues(request, path, params, func(valueJSON map[string]any) {
		name, okName := valueJSON["title"].(string)
		state, okState := valueJSON["state"].(string)
		author, okAuthor := "", false
		authorStruct, okAuthorStruct := valueJSON["author"].(map[string]any)
		if okAuthorStruct {
			user, okUser := authorStruct["user"].(map[string]any)
			if okUser {
				author, okAuthor = user["slug"].(string)
			}
		}
		reviewers, okReviewers := []string{}, true
		reviewersStruct, okReviewersStruct := valueJSON["reviewers"].([]any)
		if !okReviewersStruct {
			okReviewers = false
		} else {
			for _, reviewerUserStruct := range reviewersStruct {
				reviewerUser, okReviewerUser := reviewerUserStruct.(map[string]any)
				if !okReviewerUser {
					okReviewers = false
				} else {
					reviewerStruct, reviewerStructOk := reviewerUser["user"].(map[string]any)
					if !reviewerStructOk {
						okReviewers = false
					} else {
						reviewer, reviewerOk := reviewerStruct["slug"].(string)
						if !reviewerOk {
							okReviewers = false
						} else {
							reviewers = append(reviewers, reviewer)
						}
					}
				}
			}
		}
		if okName && okState && okAuthor && okReviewers {
			log.WithFields(log.Fields{
				"project":   project,
				"repo":      repo,
				"PR":        name,
				"state":     state,
				"author":    author,
				"reviewers": reviewers,
			}).Debug("PR collected")
			prs = append(prs, PR{
				Name:      name,
				State:     state,
				Author:    author,
				Reviewers: reviewers,
			})
		}
	})
	if err != nil {
		return nil, err
	}
	return prs, nil
}

type Reference struct {
	Name   string
	Author string
}

func References(request *Request, project string, repo string) ([]Reference, []Reference, error) {
	var branches, tags []Reference
	path := fmt.Sprintf("projects/%s/repos/%s/ref-change-activities", project, repo)
	err := paginatedValues(request, path, nil, func(valueJSON map[string]any) {
		author, okAuthor := "", false
		authorStruct, okAuthorStruct := valueJSON["user"].(map[string]any)
		if okAuthorStruct {
			author, okAuthor = authorStruct["name"].(string)
		}
		refName, okRefName := "", false
		refType, okRefType := "", false
		refChangeStruct, okRefChangeStruct := valueJSON["refChange"].(map[string]any)
		if okRefChangeStruct {
			refStruct, okRefStruct := refChangeStruct["ref"].(map[string]any)
			if okRefStruct {
				refName, okRefName = refStruct["displayId"].(string)
				refType, okRefType = refStruct["type"].(string)
			}
		}
		if okAuthor && okRefName && okRefType {
			log.WithFields(log.Fields{
				"project":   project,
				"repo":      repo,
				"reference": refName,
				"type":      refType,
				"author":    author,
			}).Debug("Reference collected")
			reference := Reference{
				Name:   refName,
				Author: author,
			}
			switch refType {
			case "BRANCH":
				branches = append(branches, reference)
			case "TAG":
				tags = append(tags, reference)
			default:
				log.WithFields(log.Fields{
					"project":   project,
					"repo":      repo,
					"reference": refName,
					"type":      refType,
					"author":    author,
				}).Error("Reference of unknown type")
			}
		}
	})
	if err != nil {
		return nil, nil, err
	}
	return branches, tags, nil
}
