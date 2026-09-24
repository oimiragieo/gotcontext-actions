package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/oimiragieo/gotcontext-actions/internal/common"
	"github.com/oimiragieo/gotcontext-actions/internal/exprparser"
	"github.com/oimiragieo/gotcontext-actions/internal/model"
	"gopkg.in/yaml.v3"
)

func (rc *RunContext) containerImage(ctx context.Context) string {
	job := rc.Run.Job()

	c := job.Container()
	if c != nil {
		return rc.ExprEval.Interpolate(ctx, c.Image)
	}

	return ""
}

func (rc *RunContext) runsOnImage(ctx context.Context) string {
	if rc.Run.Job().RunsOn() == nil {
		common.Logger(ctx).Errorf("'runs-on' key not defined in %s", rc.String())
	}

	for _, platformName := range rc.runsOnPlatformNames(ctx) {
		image := rc.Config.Platforms[strings.ToLower(platformName)]
		if image != "" {
			return image
		}
	}

	return ""
}

func (rc *RunContext) runsOnPlatformNames(ctx context.Context) []string {
	job := rc.Run.Job()

	if job.RunsOn() == nil {
		return []string{}
	}

	// Evaluate a deep copy so concurrent matrix cells do not race on job.RawRunsOn.
	rawCopy := cloneYAMLNode(&job.RawRunsOn)
	if err := rc.ExprEval.EvaluateYamlNode(ctx, rawCopy); err != nil {
		common.Logger(ctx).Errorf("Error while evaluating runs-on: %v", err)
		return []string{}
	}

	tmp := *job
	tmp.RawRunsOn = *rawCopy
	names := tmp.RunsOn()
	if names == nil {
		return []string{}
	}
	return names
}

func cloneYAMLNode(n *yaml.Node) *yaml.Node {
	if n == nil {
		return &yaml.Node{}
	}
	cp := *n
	if len(n.Content) == 0 {
		return &cp
	}
	cp.Content = make([]*yaml.Node, len(n.Content))
	for i, c := range n.Content {
		cp.Content[i] = cloneYAMLNode(c)
	}
	return &cp
}

func (rc *RunContext) platformImage(ctx context.Context) string {
	if containerImage := rc.containerImage(ctx); containerImage != "" {
		return containerImage
	}

	return rc.runsOnImage(ctx)
}

func (rc *RunContext) options(ctx context.Context) string {
	job := rc.Run.Job()
	c := job.Container()
	if c != nil {
		return rc.ExprEval.Interpolate(ctx, c.Options)
	}

	return rc.Config.ContainerOptions
}

func (rc *RunContext) isEnabled(ctx context.Context) (bool, error) {
	job := rc.Run.Job()
	l := common.Logger(ctx)
	runJob, runJobErr := EvalBool(ctx, rc.ExprEval, job.If.Value, exprparser.DefaultStatusCheckSuccess)
	jobType, jobTypeErr := job.Type()

	if runJobErr != nil {
		return false, fmt.Errorf("  \u274C  Error in if-expression: \"if: %s\" (%s)", job.If.Value, runJobErr)
	}

	if jobType == model.JobTypeInvalid {
		return false, jobTypeErr
	}

	if !runJob {
		rc.result("skipped")
		l.WithField("jobResult", "skipped").Debugf("Skipping job '%s' due to '%s'", job.Name, job.If.Value)
		return false, nil
	}

	if jobType != model.JobTypeDefault {
		return true, nil
	}

	img := rc.platformImage(ctx)
	if img == "" {
		names := rc.runsOnPlatformNames(ctx)
		for _, platformName := range names {
			l.Infof("\U0001F6A7  Skipping unsupported platform -- Try running with `-P %+v=...`", platformName)
		}
		if rc.Config.StrictPlatforms {
			if len(names) == 0 {
				return false, fmt.Errorf("job '%s' has no mapped runs-on platform image (strict-platforms)", job.Name)
			}
			return false, fmt.Errorf("job '%s' runs-on unsupported platform(s) %v (strict-platforms)", job.Name, names)
		}
		return false, nil
	}
	return true, nil
}
