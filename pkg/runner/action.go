package runner

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nektos/act/pkg/common"
	"github.com/nektos/act/pkg/model"
)

type actionStep interface {
	step

	getActionModel() *model.Action
	getCompositeRunContext(context.Context) *RunContext
	getCompositeSteps() *compositeSteps
}

type readAction func(ctx context.Context, step *model.Step, actionDir string, actionPath string, readFile actionYamlReader, writeFile fileWriter) (*model.Action, error)

type actionYamlReader func(filename string) (io.Reader, io.Closer, error)

type fileWriter func(filename string, data []byte, perm fs.FileMode) error

type runAction func(step actionStep, actionDir string, remoteAction *remoteAction) common.Executor

//go:embed res/trampoline.js
var trampoline embed.FS

func readActionImpl(ctx context.Context, step *model.Step, actionDir string, actionPath string, readFile actionYamlReader, writeFile fileWriter) (*model.Action, error) {
	logger := common.Logger(ctx)
	allErrors := []error{}
	addError := func(fileName string, err error) {
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to read '%s' from action '%s' with path '%s' of step: %w", fileName, step.String(), actionPath, err))
		} else {
			// One successful read, clear error state
			allErrors = nil
		}
	}
	reader, closer, err := readFile("action.yml")
	addError("action.yml", err)
	if os.IsNotExist(err) {
		reader, closer, err = readFile("action.yaml")
		addError("action.yaml", err)
		if os.IsNotExist(err) {
			_, closer, err := readFile("Dockerfile")
			addError("Dockerfile", err)
			if err == nil {
				closer.Close()
				action := &model.Action{
					Name: "(Synthetic)",
					Runs: model.ActionRuns{
						Using: "docker",
						Image: "Dockerfile",
					},
				}
				logger.Debugf("Using synthetic action %v for Dockerfile", action)
				return action, nil
			}
			if step.With != nil {
				if val, ok := step.With["args"]; ok {
					var b []byte
					if b, err = trampoline.ReadFile("res/trampoline.js"); err != nil {
						return nil, err
					}
					err2 := writeFile(filepath.Join(actionDir, actionPath, "trampoline.js"), b, 0o400)
					if err2 != nil {
						return nil, err2
					}
					action := &model.Action{
						Name: "(Synthetic)",
						Inputs: map[string]model.Input{
							"cwd": {
								Description: "(Actual working directory)",
								Required:    false,
								Default:     path.Join(actionDir, actionPath),
							},
							"command": {
								Description: "(Actual program)",
								Required:    false,
								Default:     val,
							},
						},
						Runs: model.ActionRuns{
							Using: "node12",
							Main:  "trampoline.js",
						},
					}
					logger.Debugf("Using synthetic action %v", action)
					return action, nil
				}
			}
		}
	}
	if allErrors != nil {
		return nil, errors.Join(allErrors...)
	}
	defer closer.Close()

	action, err := model.ReadAction(reader)
	logger.Debugf("Read action %v from '%s'", action, "Unknown")
	return action, err
}

func maybeCopyToActionDir(ctx context.Context, step actionStep, actionDir string, actionPath string, containerActionDir string) error {
	logger := common.Logger(ctx)
	rc := step.getRunContext()
	stepModel := step.getStepModel()

	if stepModel.Type() != model.StepTypeUsesActionRemote {
		return nil
	}

	var containerActionDirCopy string
	containerActionDirCopy = strings.TrimSuffix(containerActionDir, actionPath)
	logger.Debug(containerActionDirCopy)

	if !strings.HasSuffix(containerActionDirCopy, `/`) {
		containerActionDirCopy += `/`
	}

	if rc.Config != nil && rc.Config.ActionCache != nil {
		raction := step.(*stepActionRemote)
		ta, err := rc.Config.ActionCache.GetTarArchive(ctx, raction.cacheDir, raction.resolvedSha, "")
		if err != nil {
			return err
		}
		defer ta.Close()
		return rc.JobContainer.CopyTarStream(ctx, containerActionDirCopy, ta)
	}

	if err := removeGitIgnore(ctx, actionDir); err != nil {
		return err
	}

	return rc.JobContainer.CopyDir(containerActionDirCopy, actionDir+"/", rc.Config.UseGitIgnore)(ctx)
}

func runActionImpl(step actionStep, actionDir string, remoteAction *remoteAction) common.Executor {
	rc := step.getRunContext()
	stepModel := step.getStepModel()

	return func(ctx context.Context) error {
		logger := common.Logger(ctx)
		actionPath := ""
		if remoteAction != nil && remoteAction.Path != "" {
			actionPath = remoteAction.Path
		}

		action := step.getActionModel()
		logger.Debugf("About to run action %v", action)

		err := setupActionEnv(ctx, step, remoteAction)
		if err != nil {
			return err
		}

		actionLocation := path.Join(actionDir, actionPath)
		actionName, containerActionDir := getContainerActionPaths(stepModel, actionLocation, rc)

		logger.Debugf("type=%v actionDir=%s actionPath=%s workdir=%s actionCacheDir=%s actionName=%s containerActionDir=%s", stepModel.Type(), actionDir, actionPath, rc.Config.Workdir, rc.ActionCacheDir(), actionName, containerActionDir)

		x := action.Runs.Using
		switch {
		case x.IsNode():
			if err := maybeCopyToActionDir(ctx, step, actionDir, actionPath, containerActionDir); err != nil {
				return err
			}
			containerArgs := []string{rc.GetNodeToolFullPath(ctx), path.Join(containerActionDir, action.Runs.Main)}
			logger.Debugf("executing remote job container: %s", containerArgs)

			rc.ApplyExtraPath(ctx, step.getEnv())

			return rc.execJobContainer(containerArgs, *step.getEnv(), "", "")(ctx)
		case x.IsDocker():
			if remoteAction == nil {
				actionDir = ""
				actionPath = containerActionDir
			}
			return execAsDocker(ctx, step, actionName, actionDir, actionPath, remoteAction == nil, "entrypoint")
		case x.IsComposite():
			if err := maybeCopyToActionDir(ctx, step, actionDir, actionPath, containerActionDir); err != nil {
				return err
			}

			return execAsComposite(step)(ctx)
		default:
			return fmt.Errorf("The runs.using key must be one of: %v, got %s", []string{
				model.ActionRunsUsingDocker,
				model.ActionRunsUsingNode12,
				model.ActionRunsUsingNode16,
				model.ActionRunsUsingNode20,
				model.ActionRunsUsingNode24,
				model.ActionRunsUsingComposite,
			}, action.Runs.Using)
		}
	}
}

func setupActionEnv(ctx context.Context, step actionStep, _ *remoteAction) error {
	rc := step.getRunContext()

	// A few fields in the environment (e.g. GITHUB_ACTION_REPOSITORY)
	// are dependent on the action. That means we can complete the
	// setup only after resolving the whole action model and cloning
	// the action
	rc.withGithubEnv(ctx, step.getGithubContext(ctx), *step.getEnv())
	populateEnvsFromSavedState(step.getEnv(), step, rc)
	populateEnvsFromInput(ctx, step.getEnv(), step.getActionModel(), rc)

	return nil
}

// https://github.com/nektos/act/issues/228#issuecomment-629709055
// files in .gitignore are not copied in a Docker container
// this causes issues with actions that ignore other important resources
// such as `node_modules` for example
func removeGitIgnore(ctx context.Context, directory string) error {
	gitIgnorePath := path.Join(directory, ".gitignore")
	if _, err := os.Stat(gitIgnorePath); err == nil {
		// .gitignore exists
		common.Logger(ctx).Debugf("Removing %s before docker cp", gitIgnorePath)
		err := os.Remove(gitIgnorePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: break out parts of function to reduce complexicity
//
//nolint:gocyclo

func shouldRunPreStep(step actionStep) common.Conditional {
	return func(ctx context.Context) bool {
		log := common.Logger(ctx)

		if step.getActionModel() == nil {
			log.Debugf("skip pre step for '%s': no action model available", step.getStepModel())
			return false
		}

		return true
	}
}

func hasPreStep(step actionStep) common.Conditional {
	return func(_ context.Context) bool {
		action := step.getActionModel()
		if action == nil {
			return false
		}
		return action.Runs.Using.IsComposite() ||
			(action.Runs.Using.IsNode() &&
				action.Runs.Pre != "") ||
			(action.Runs.Using.IsDocker() &&
				action.Runs.PreEntrypoint != "")
	}
}

func runPreStep(step actionStep) common.Executor {
	return func(ctx context.Context) error {
		logger := common.Logger(ctx)
		logger.Debugf("run pre step for '%s'", step.getStepModel())

		rc := step.getRunContext()
		stepModel := step.getStepModel()
		action := step.getActionModel()

		// defaults in pre steps were missing, however provided inputs are available
		populateEnvsFromInput(ctx, step.getEnv(), action, rc)

		// todo: refactor into step
		var actionDir string
		var actionPath string
		var remoteAction *stepActionRemote
		if remote, ok := step.(*stepActionRemote); ok {
			actionPath = newRemoteAction(stepModel.Uses).Path
			actionDir = fmt.Sprintf("%s/%s", rc.ActionCacheDir(), safeFilename(stepModel.Uses))
			remoteAction = remote
		} else {
			actionDir = filepath.Join(rc.Config.Workdir, stepModel.Uses)
			actionPath = ""
		}

		actionLocation := ""
		if actionPath != "" {
			actionLocation = path.Join(actionDir, actionPath)
		} else {
			actionLocation = actionDir
		}

		actionName, containerActionDir := getContainerActionPaths(stepModel, actionLocation, rc)

		x := action.Runs.Using
		switch {
		case x.IsNode():
			if err := maybeCopyToActionDir(ctx, step, actionDir, actionPath, containerActionDir); err != nil {
				return err
			}

			containerArgs := []string{rc.GetNodeToolFullPath(ctx), path.Join(containerActionDir, action.Runs.Pre)}
			logger.Debugf("executing remote job container: %s", containerArgs)

			rc.ApplyExtraPath(ctx, step.getEnv())

			return rc.execJobContainer(containerArgs, *step.getEnv(), "", "")(ctx)

		case x.IsDocker():
			if remoteAction == nil {
				actionDir = ""
				actionPath = containerActionDir
			}
			return execAsDocker(ctx, step, actionName, actionDir, actionPath, remoteAction == nil, "pre-entrypoint")

		case x.IsComposite():
			if step.getCompositeSteps() == nil {
				step.getCompositeRunContext(ctx)
			}

			if steps := step.getCompositeSteps(); steps != nil && steps.pre != nil {
				return steps.pre(ctx)
			}
			return fmt.Errorf("missing steps in composite action")

		default:
			return nil
		}
	}
}

func shouldRunPostStep(step actionStep) common.Conditional {
	return func(ctx context.Context) bool {
		log := common.Logger(ctx)
		stepResults := step.getRunContext().getStepsContext()
		stepResult := stepResults[step.getStepModel().ID]

		if stepResult == nil {
			log.WithField("stepResult", model.StepStatusSkipped).Debugf("skipping post step for '%s'; step was not executed", step.getStepModel())
			return false
		}

		if stepResult.Conclusion == model.StepStatusSkipped {
			log.WithField("stepResult", model.StepStatusSkipped).Debugf("skipping post step for '%s'; main step was skipped", step.getStepModel())
			return false
		}

		if step.getActionModel() == nil {
			log.WithField("stepResult", model.StepStatusSkipped).Debugf("skipping post step for '%s': no action model available", step.getStepModel())
			return false
		}

		return true
	}
}

func hasPostStep(step actionStep) common.Conditional {
	return func(_ context.Context) bool {
		action := step.getActionModel()
		return action.Runs.Using.IsComposite() ||
			(action.Runs.Using.IsNode() &&
				action.Runs.Post != "") ||
			(action.Runs.Using.IsDocker() &&
				action.Runs.PostEntrypoint != "")
	}
}

func runPostStep(step actionStep) common.Executor {
	return func(ctx context.Context) error {
		logger := common.Logger(ctx)
		logger.Debugf("run post step for '%s'", step.getStepModel())

		rc := step.getRunContext()
		stepModel := step.getStepModel()
		action := step.getActionModel()

		// todo: refactor into step
		var actionDir string
		var actionPath string
		var remoteAction *stepActionRemote
		if remote, ok := step.(*stepActionRemote); ok {
			actionPath = newRemoteAction(stepModel.Uses).Path
			actionDir = fmt.Sprintf("%s/%s", rc.ActionCacheDir(), safeFilename(stepModel.Uses))
			remoteAction = remote
		} else {
			actionDir = filepath.Join(rc.Config.Workdir, stepModel.Uses)
			actionPath = ""
		}

		actionLocation := ""
		if actionPath != "" {
			actionLocation = path.Join(actionDir, actionPath)
		} else {
			actionLocation = actionDir
		}

		actionName, containerActionDir := getContainerActionPaths(stepModel, actionLocation, rc)

		x := action.Runs.Using
		switch {
		case x.IsNode():
			populateEnvsFromSavedState(step.getEnv(), step, rc)
			populateEnvsFromInput(ctx, step.getEnv(), step.getActionModel(), rc)

			containerArgs := []string{rc.GetNodeToolFullPath(ctx), path.Join(containerActionDir, action.Runs.Post)}
			logger.Debugf("executing remote job container: %s", containerArgs)

			rc.ApplyExtraPath(ctx, step.getEnv())

			return rc.execJobContainer(containerArgs, *step.getEnv(), "", "")(ctx)

		case x.IsDocker():
			if remoteAction == nil {
				actionDir = ""
				actionPath = containerActionDir
			}
			return execAsDocker(ctx, step, actionName, actionDir, actionPath, remoteAction == nil, "post-entrypoint")

		case x.IsComposite():
			if err := maybeCopyToActionDir(ctx, step, actionDir, actionPath, containerActionDir); err != nil {
				return err
			}

			if steps := step.getCompositeSteps(); steps != nil && steps.post != nil {
				return steps.post(ctx)
			}
			return fmt.Errorf("missing steps in composite action")

		default:
			return nil
		}
	}
}
