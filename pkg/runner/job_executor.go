package runner

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nektos/act/pkg/common"
	"github.com/nektos/act/pkg/model"
)

type jobInfo interface {
	matrix() map[string]interface{}
	steps() []*model.Step
	startContainer() common.Executor
	stopContainer() common.Executor
	closeContainer() common.Executor
	interpolateOutputs() common.Executor
	result(result string)
}

//nolint:contextcheck
func newJobExecutor(info jobInfo, sf stepFactory, rc *RunContext) common.Executor {
	steps := make([]common.Executor, 0)
	preSteps := make([]common.Executor, 0)
	var postExecutor common.Executor

	steps = append(steps, func(ctx context.Context) error {
		logger := common.Logger(ctx)
		if len(info.matrix()) > 0 {
			logger.Infof("\U0001F9EA  Matrix: %v", info.matrix())
		}
		return nil
	})

	infoSteps := info.steps()

	if len(infoSteps) == 0 {
		return common.NewDebugExecutor("No steps found")
	}

	preSteps = append(preSteps, func(ctx context.Context) error {
		// Have to be skipped for some Tests
		if rc.Run == nil {
			return nil
		}
		rc.ExprEval = rc.NewExpressionEvaluator(ctx)
		// evaluate environment variables since they can contain
		// GitHub's special environment variables.
		for k, v := range rc.GetEnv() {
			rc.Env[k] = rc.ExprEval.Interpolate(ctx, v)
		}
		return nil
	})

	var setJobError = func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}
		logger := common.Logger(ctx)
		logger.Errorf("%v", err)
		common.SetJobError(ctx, err)
		return err
	}

	for i, stepModel := range infoSteps {
		if stepModel == nil {
			return func(_ context.Context) error {
				return fmt.Errorf("invalid Step %v: missing run or uses key", i)
			}
		}
		if stepModel.ID == "" {
			stepModel.ID = fmt.Sprintf("%d", i)
		}

		step, err := sf.newStep(stepModel, rc)

		if err != nil {
			return common.NewErrorExecutor(err)
		}

		preSteps = append(preSteps, useStepLogger(rc, stepModel, stepStagePre, step.pre().ThenError(setJobError)))

		stepExec := step.main()
		steps = append(steps, useStepLogger(rc, stepModel, stepStageMain, func(ctx context.Context) error {
			err := stepExec(ctx)
			if err != nil {
				_ = setJobError(ctx, err)
			} else if ctx.Err() != nil {
				_ = setJobError(ctx, ctx.Err())
			}
			return nil
		}))

		postExec := useStepLogger(rc, stepModel, stepStagePost, step.post().ThenError(setJobError))
		if postExecutor != nil {
			// run the post executor in reverse order
			postExecutor = postExec.Finally(postExecutor)
		} else {
			postExecutor = postExec
		}
	}

	var stopContainerExecutor common.Executor = func(ctx context.Context) error {
		jobError := common.JobError(ctx)
		var err error
		if rc.Config.AutoRemove || jobError == nil {
			// always allow 1 min for stopping and removing the runner, even if we were cancelled
			ctx, cancel := context.WithTimeout(common.WithLogger(context.Background(), common.Logger(ctx)), time.Minute)
			defer cancel()

			logger := common.Logger(ctx)
			logger.Infof("Cleaning up container for job %s", rc.JobName)
			if err = info.stopContainer()(ctx); err != nil {
				logger.Errorf("Error while stop job container: %v", err)
			}
		}
		return err
	}

	var setJobResultExecutor common.Executor = func(ctx context.Context) error {
		jobError := common.JobError(ctx)
		setJobResult(ctx, info, rc, jobError == nil)
		setJobOutputs(ctx, rc)
		return nil
	}

	pipeline := make([]common.Executor, 0)
	pipeline = append(pipeline, preSteps...)
	pipeline = append(pipeline, steps...)

	jobPipeline := common.NewPipelineExecutor(
		common.NewFieldExecutor("step", "Set up job", common.NewFieldExecutor("stepid", []string{"--setup-job"},
			common.NewPipelineExecutor(common.NewInfoExecutor("\u2B50 Run Set up job"), info.startContainer(), rc.InitializeNodeTool()).
				Then(common.NewFieldExecutor("stepResult", model.StepStatusSuccess, common.NewInfoExecutor("  \u2705  Success - Set up job"))).
				ThenError(setJobError).OnError(common.NewFieldExecutor("stepResult", model.StepStatusFailure, common.NewInfoExecutor("  \u274C  Failure - Set up job"))))),
		common.NewPipelineExecutor(pipeline...).
			Finally(func(ctx context.Context) error { //nolint:contextcheck
				var cancel context.CancelFunc
				if ctx.Err() == context.Canceled {
					// in case of an aborted run, we still should execute the
					// post steps to allow cleanup.
					ctx, cancel = context.WithTimeout(common.WithLogger(context.Background(), common.Logger(ctx)), 5*time.Minute)
					defer cancel()
				}
				return postExecutor(ctx)
			}).
			Finally(common.NewFieldExecutor("step", "Complete job", common.NewFieldExecutor("stepid", []string{"--complete-job"},
				common.NewInfoExecutor("\u2B50 Run Complete job").
					Finally(stopContainerExecutor).
					Finally(
						info.interpolateOutputs().Finally(info.closeContainer()).Then(common.NewFieldExecutor("stepResult", model.StepStatusSuccess, common.NewInfoExecutor("  \u2705  Success - Complete job"))).
							OnError(common.NewFieldExecutor("stepResult", model.StepStatusFailure, common.NewInfoExecutor("  \u274C  Failure - Complete job"))),
					))))).Finally(setJobResultExecutor)

	return func(ctx context.Context) error {
		ctx, cancel := evaluateJobTimeout(ctx, rc)
		defer cancel()

		if rc.Run != nil && rc.Run.Job() != nil {
			job := rc.Run.Job()
			conc := job.Concurrency
			if conc == nil && rc.Run.Workflow != nil {
				conc = rc.Run.Workflow.Concurrency
			}
			if conc != nil && conc.Group != "" {
				if cc := common.GetConcurrencyController(ctx); cc != nil {
					eval := rc.NewExpressionEvaluator(ctx)
					var release context.CancelFunc
					ctx, release = cc.Acquire(ctx, eval.Interpolate(ctx, conc.Group), conc.CancelInProgress)
					defer release()
				}
			}
		}

		err := jobPipeline(ctx)
		printStepSummaries(ctx, rc)
		return err
	}
}

func evaluateJobTimeout(ctx context.Context, rc *RunContext) (context.Context, context.CancelFunc) {
	noop := func() {}
	if rc == nil || rc.Run == nil {
		return ctx, noop
	}
	job := rc.Run.Job()
	if job == nil || job.TimeoutMinutes == "" {
		return ctx, noop
	}
	eval := rc.NewExpressionEvaluator(ctx)
	timeout := eval.Interpolate(ctx, job.TimeoutMinutes)
	if timeout == "" {
		return ctx, noop
	}
	minutes, err := strconv.ParseInt(timeout, 10, 64)
	if err != nil || minutes <= 0 {
		return ctx, noop
	}
	return context.WithTimeout(ctx, time.Duration(minutes)*time.Minute)
}

func printStepSummaries(ctx context.Context, rc *RunContext) {
	if rc == nil || len(rc.stepSummaries) == 0 {
		return
	}
	logger := common.Logger(ctx)
	logger.Infof("---- Step summaries ----")
	for stepID, body := range rc.stepSummaries {
		if strings.TrimSpace(body) == "" {
			continue
		}
		logger.Infof("[%s]\n%s", stepID, body)
	}
}

func setJobResult(ctx context.Context, info jobInfo, rc *RunContext, success bool) {
	logger := common.Logger(ctx)
	job := rc.Run.Job()
	mu := jobResultMutex(job)
	mu.Lock()

	jobResult := "success"
	// we have only one result for a whole matrix build, so we need
	// to keep an existing result state if we run a matrix
	if len(info.matrix()) > 0 && job != nil && job.Result != "" {
		jobResult = job.Result
	}

	if !success {
		jobResult = "failure"
		if job != nil {
			eval := rc.NewExpressionEvaluator(ctx)
			coe := eval.Interpolate(ctx, job.ContinueOnError)
			if strings.EqualFold(coe, "true") {
				jobResult = "success"
				logger.Infof("Job '%s' failed but continue-on-error is set; treating as success", rc.JobName)
			}
		}
	}

	// Failure wins across matrix cells.
	if job != nil && job.Result == "failure" && jobResult == "success" && len(info.matrix()) > 0 {
		jobResult = "failure"
	}
	mu.Unlock()

	info.result(jobResult)
	if rc.caller != nil {
		rc.caller.runContext.result(jobResult)
	}

	jobResultMessage := "succeeded"
	if jobResult != "success" {
		jobResultMessage = "failed"
	}
	logger.WithField("jobResult", jobResult).Infof("\U0001F3C1  Job %s", jobResultMessage)
}

func setJobOutputs(ctx context.Context, rc *RunContext) {
	if rc.caller != nil {
		// map outputs for reusable workflows
		callerOutputs := make(map[string]string)

		ee := rc.NewExpressionEvaluator(ctx)

		for k, v := range rc.Run.Workflow.WorkflowCallConfig().Outputs {
			callerOutputs[k] = ee.Interpolate(ctx, ee.Interpolate(ctx, v.Value))
		}

		rc.caller.runContext.Run.Job().Outputs = callerOutputs
	}
}

func useStepLogger(rc *RunContext, stepModel *model.Step, stage stepStage, executor common.Executor) common.Executor {
	return func(ctx context.Context) error {
		ctx = withStepLogger(ctx, stepModel.ID, rc.ExprEval.Interpolate(ctx, stepModel.String()), stage.String())

		rawLogger := common.Logger(ctx).WithField("raw_output", true)
		logWriter := common.NewLineWriter(rc.commandHandler(ctx), func(s string) bool {
			if rc.Config.LogOutput {
				rawLogger.Infof("%s", s)
			} else {
				rawLogger.Debugf("%s", s)
			}
			return true
		})

		oldout, olderr := rc.JobContainer.ReplaceLogWriter(logWriter, logWriter)
		defer rc.JobContainer.ReplaceLogWriter(oldout, olderr)

		return executor(ctx)
	}
}
