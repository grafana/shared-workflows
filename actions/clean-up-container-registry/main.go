package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type refDate struct {
	ref  name.Reference
	date time.Time
}

type keepReason int

const (
	keepReasonExcluded keepReason = iota
	keepReasonLatest
)

func (kr keepReason) String() string {
	switch kr {
	case keepReasonExcluded:
		return "excluded by filter"
	case keepReasonLatest:
		return "new enough"
	default:
		return "Unknown"
	}
}

type keptTag struct {
	refDate
	keepReason
}

type repoSearchResult struct {
	tagsToKeep   []keptTag
	tagsToRemove []refDate
}

type Config struct {
	ImageRef    string
	ImageRepo   name.Repository
	ExcludeTags []glob.Glob
	TagFilter   []glob.Glob
	KeepLatest  int
	DryRun      bool
}

func buildGlobs(filters []string) ([]glob.Glob, error) {
	globs := make([]glob.Glob, 0, len(filters))
	for _, filter := range filters {
		g, err := glob.Compile(filter)
		if err != nil {
			return nil, err
		}
		globs = append(globs, g)
	}
	return globs, nil
}

func main() {
	config := Config{}

	app := &cli.App{
		Name:  "docker-registry-cleanup",
		Usage: "Clean up Docker images in a registry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "image-name",
				Usage:       "The name of the Docker image to clean up",
				Required:    true,
				Destination: &config.ImageRef,
				Action: func(c *cli.Context, imageName string) error {
					repo, err := name.NewRepository(imageName)
					if err != nil {
						return err
					}

					config.ImageRepo = repo

					return nil
				},
			},
			&cli.StringSliceFlag{
				Name:  "exclude-tags",
				Usage: "Tags to exclude from deletion",
				Action: func(c *cli.Context, excludeTags []string) error {
					globs, err := buildGlobs(excludeTags)
					if err != nil {
						return err
					}

					config.ExcludeTags = globs
					return nil
				},
			},
			&cli.StringSliceFlag{
				Name:  "tag-filter",
				Usage: "Glob pattern to filter tags",
				Action: func(c *cli.Context, tagFilter []string) error {
					globs, err := buildGlobs(tagFilter)
					if err != nil {
						return err
					}

					config.TagFilter = globs
					return nil
				},
			},
			&cli.IntFlag{
				Name:        "keep-latest",
				Usage:       "Number of latest images to keep",
				Value:       6,
				Destination: &config.KeepLatest,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Usage:       "Run the action in dry-run mode",
				Value:       true,
				Destination: &config.DryRun,
			},
		},
		Action: func(_ *cli.Context) error {
			return run(config)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(config Config) error {
	imageName := config.ImageRef
	imageRepo := config.ImageRepo

	excludeTags := config.ExcludeTags
	tagFilters := config.TagFilter
	keepLatest := config.KeepLatest
	dryRun := config.DryRun

	remoteTags, err := listRemoteTags(imageRepo)
	if err != nil {
		return fmt.Errorf("failed to list tags for %s: %v", imageName, err)
	}
	fmt.Printf("Found %d tags for %s\n", len(remoteTags), imageName)

	tags, err := getTagsToRemove(remoteTags, tagFilters, excludeTags, keepLatest, imageRepo)
	if err != nil {
		return fmt.Errorf("failed to determine tags to remove: %v", err)
	}

	if dryRun {
		fmt.Printf("Dry run mode enabled. Would have removed the following tags:\n%s", tags)
		return nil
	}

	err = removeTags(tags)
	if err != nil {
		return err
	}

	return nil
}

func listRemoteTags(imageRepo name.Repository) ([]string, error) {
	return remote.List(imageRepo, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}

func (rs repoSearchResult) String() string {
	var sb strings.Builder
	sb.WriteString("Tags to remove:\n")

	for _, tag := range rs.tagsToRemove {
		sb.WriteString(fmt.Sprintf("  - %s\n", tag.ref.Name()))
	}

	sb.WriteString("\nTags to keep:\n")
	for _, tag := range rs.tagsToKeep {
		sb.WriteString(fmt.Sprintf("  - %s (%s)\n", tag.ref.Name(), tag.keepReason))
	}

	return sb.String()
}

func removeTags(tags repoSearchResult) error {
	for _, tag := range tags.tagsToRemove {
		if err := remote.Delete(tag.ref, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
			return fmt.Errorf("failed to delete tag %s: %v", tag.ref.Name(), err)
		}

		fmt.Printf("Deleted tag %s...\n", tag.ref.Name())
	}
	return nil
}

func getTagsToRemove(allTags []string, tagFilters []glob.Glob, excludes []glob.Glob, keepLatest int, imageRepo name.Repository) (repoSearchResult, error) {
	var (
		tagsToConsiderRemoving []refDate
		tagsKept               []keptTag
		mu                     sync.Mutex
	)

	eg := errgroup.Group{}
	eg.SetLimit(100)

	for _, tag := range allTags {
		eg.Go(func() error {
			fmt.Printf("Processing tag %s...\n", tag)
			ref, err := name.ParseReference(fmt.Sprintf("%s:%s", imageRepo, tag))
			if err != nil {
				return err
			}

			if !matchesFilters(tag, tagFilters, true) {
				return nil
			}

			creationDate, err := getCreationDate(ref)
			if err != nil {
				return fmt.Errorf("failed to get creation date for %s: %v", ref.Name(), err)
			}

			mu.Lock()
			defer mu.Unlock()

			if matchesFilters(tag, excludes, false) {
				tagsKept = append(tagsKept, keptTag{
					refDate:    refDate{ref, creationDate},
					keepReason: keepReasonExcluded,
				})
				return nil
			}

			tagsToConsiderRemoving = append(tagsToConsiderRemoving, refDate{ref, creationDate})
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return repoSearchResult{}, err
	}

	if len(tagsToConsiderRemoving) <= keepLatest {
		keepLatest = len(tagsToConsiderRemoving)
	}

	slices.SortFunc(tagsToConsiderRemoving, func(l, r refDate) int {
		return l.date.Compare(r.date) * -1
	})

	slices.SortFunc(tagsKept, func(l, r keptTag) int {
		return l.refDate.date.Compare(r.refDate.date) * -1
	})

	tagsToRemove := tagsToConsiderRemoving[keepLatest:]
	for _, tag := range tagsToConsiderRemoving[:keepLatest] {
		tagsKept = append(tagsKept, keptTag{refDate: tag, keepReason: keepReasonLatest})
	}

	return repoSearchResult{
		tagsToKeep:   tagsKept,
		tagsToRemove: tagsToRemove,
	}, nil
}

func matchesFilters(input string, filters []glob.Glob, matchesIfNoFilters bool) bool {
	if len(filters) == 0 {
		return matchesIfNoFilters
	}

	for _, filter := range filters {
		if filter.Match(input) {
			return true
		}
	}

	return false
}

func getCreationDate(ref name.Reference) (time.Time, error) {
	img, err := getImage(ref)
	if err != nil {
		return time.Time{}, err
	}

	configFile, err := img.ConfigFile()
	if err != nil {
		return time.Time{}, err
	}

	return configFile.Created.Time, nil
}

func getImage(ref name.Reference) (v1.Image, error) {
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, err
	}

	return img, nil
}
