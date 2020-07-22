package gitlab

import (
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"strings"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeGitLab, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeGitLab, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeGitLab)
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r), nil
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

type adapter struct {
	*native.Adapter
	registry        *model.Registry
	url             string
	username        string
	token           string
	clientGitlabAPI *Client
}

func newAdapter(registry *model.Registry) *adapter {
	return &adapter{
		registry:        registry,
		url:             registry.URL,
		clientGitlabAPI: NewClient(registry),
		Adapter:         native.NewAdapter(registry),
	}
}

func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeGitLab,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeImage,
		},
		SupportedResourceFilters: []*model.FilterStyle{
			{
				Type:  model.FilterTypeName,
				Style: model.FilterStyleTypeText,
			},
			{
				Type:  model.FilterTypeTag,
				Style: model.FilterStyleTypeText,
			},
		},
		SupportedTriggers: []model.TriggerType{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}, nil
}

// FetchArtifacts fetches images
func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	var resources []*model.Resource
	var projects []*Project
	var err error
	nameFilter := ""
	tagFilter := ""
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			nameFilter = filter.Value.(string)
			break
		} else if filter.Type == model.FilterTypeTag {
			tagFilter = filter.Value.(string)
			break
		}
	}

	projects = a.searchByPattern(nameFilter)
	if len(projects) == 0 {
		projects, err = a.clientGitlabAPI.getProjects()
		if err != nil {
			return nil, err
		}
	}
	var pathPatterns []string

	if paths, ok := util.IsSpecificPath(nameFilter); ok {
		pathPatterns = paths
	} else {
		pathPatterns = append(pathPatterns, nameFilter)
	}

	for _, project := range projects {
		if !existPatterns(project.FullPath, pathPatterns) {
			continue
		}
		repositories, err := a.clientGitlabAPI.getRepositories(project.ID)
		if err != nil {
			return nil, err
		}
		if len(repositories) == 0 {
			continue
		}
		for _, repository := range repositories {
			if !existPatterns(repository.Path, pathPatterns) {
				continue
			}
			vTags, err := a.clientGitlabAPI.getTags(project.ID, repository.ID)
			if err != nil {
				return nil, err
			}
			if len(vTags) == 0 {
				continue
			}
			tags := []string{}
			for _, vTag := range vTags {
				if len(tagFilter) > 0 {
					if ok, _ := util.Match(strings.ToLower(vTag.Name), strings.ToLower(tagFilter)); !ok {
						continue
					}
				}
				tags = append(tags, vTag.Name)
			}
			info := make(map[string]interface{})
			info["location"] = repository.Location
			info["path"] = repository.Path

			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name:     strings.ToLower(repository.Path),
						Metadata: info,
					},
					Vtags: tags,
				},
			})
		}
	}
	return resources, nil
}

func (a *adapter) searchByPattern(pattern string) []*Project {
	var projects []*Project
	var err error
	if len(pattern) > 0 {

		names, ok := util.IsSpecificPath(pattern)
		if ok {
			for _, name := range names {
				substrings := strings.Split(name, "/")
				if len(substrings) != 2 {
					continue
				}
				var projectsByName, err = a.clientGitlabAPI.getProjectsByName(substrings[1])
				if err != nil {
					return nil
				}
				if projectsByName == nil {
					continue
				}
				projects = append(projects, projectsByName...)
			}
		} else {
			substrings := strings.Split(pattern, "/")
			if len(substrings) != 2 {
				return projects
			}
			projectName := substrings[1]
			if projectName == "*" {
				return projects
			}
			projectName = strings.Trim(projectName, "*")

			if strings.Contains(projectName, "*") {
				return projects
			}
			projects, err = a.clientGitlabAPI.getProjectsByName(projectName)
			if err != nil {
				return projects
			}
		}
	}
	return projects
}
func existPatterns(path string, patterns []string) bool {
	correct := false
	if len(patterns) > 0 {
		for _, pathPattern := range patterns {
			if ok, _ := util.Match(strings.ToLower(pathPattern), strings.ToLower(path)); ok {
				correct = true
				break
			}
		}
	} else {
		correct = true
	}
	return correct
}
