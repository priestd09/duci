package runner_test

import (
	"github.com/duck8823/duci/application"
	"github.com/duck8823/duci/application/context"
	"github.com/duck8823/duci/application/service/git/mock_git"
	"github.com/duck8823/duci/application/service/github/mock_github"
	"github.com/duck8823/duci/application/service/logstore/mock_logstore"
	"github.com/duck8823/duci/application/service/runner"
	"github.com/duck8823/duci/infrastructure/clock"
	"github.com/duck8823/duci/infrastructure/docker"
	"github.com/duck8823/duci/infrastructure/docker/mock_docker"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io"
	"net/url"
	"os"
	"path"
	"testing"
	"time"
)

func TestRunnerImpl_Run(t *testing.T) {
	// setup
	ctrl := gomock.NewController(t)

	t.Run("with correct return values", func(t *testing.T) {
		t.Run("when Dockerfile in proj root", func(t *testing.T) {
			// given
			mockGitHub := mock_github.NewMockService(ctrl)
			mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(2).
				Return(nil)

			// and
			mockGit := mock_git.NewMockService(ctrl)
			mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(func(_ interface{}, dir string, _, _, _ interface{}) error {
					if err := os.MkdirAll(dir, 0700); err != nil {
						return err
					}

					dockerfile, err := os.OpenFile(path.Join(dir, "Dockerfile"), os.O_RDWR|os.O_CREATE, 0600)
					if err != nil {
						return err
					}
					defer dockerfile.Close()

					dockerfile.WriteString("FROM alpine\nENTRYPOINT [\"echo\"]")

					return nil
				})

			// and
			mockDocker := mock_docker.NewMockClient(ctrl)
			mockDocker.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq("./Dockerfile")).
				Times(1).
				Return(&MockBuildLog{}, nil)
			mockDocker.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Not("./Dockerfile")).
				Return(nil, errors.New("must not call this"))
			mockDocker.EXPECT().
				Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1).
				Return("", &MockJobLog{}, nil)
			mockDocker.EXPECT().
				ExitCode(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(int64(0), nil)
			mockDocker.EXPECT().
				Rm(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(nil)

			// and
			mockLogStore := mock_logstore.NewMockService(ctrl)
			mockLogStore.EXPECT().
				Append(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(nil)
			mockLogStore.EXPECT().
				Start(gomock.Any()).
				AnyTimes().
				Return(nil)
			mockLogStore.EXPECT().
				Finish(gomock.Any()).
				AnyTimes().
				Return(nil)

			r := &runner.DockerRunner{
				Name:        "test-runner",
				BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
				Git:         mockGit,
				GitHub:      mockGitHub,
				Docker:      mockDocker,
				LogStore:    mockLogStore,
			}

			// and
			repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

			// when
			err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

			// then
			if err != nil {
				t.Errorf("must not error. but: %+v", err)
			}
		})

		t.Run("when Dockerfile in sub directory", func(t *testing.T) {
			// given
			mockGitHub := mock_github.NewMockService(ctrl)
			mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(2).
				Return(nil)

			// and
			mockGit := mock_git.NewMockService(ctrl)
			mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(func(_ interface{}, dir string, _, _, _ interface{}) error {
					if err := os.MkdirAll(path.Join(dir, ".duci"), 0700); err != nil {
						return err
					}

					dockerfile, err := os.OpenFile(path.Join(dir, ".duci/Dockerfile"), os.O_RDWR|os.O_CREATE, 0600)
					if err != nil {
						return err
					}
					defer dockerfile.Close()

					dockerfile.WriteString("FROM alpine\nENTRYPOINT [\"echo\"]")

					return nil
				})

			// and
			mockDocker := mock_docker.NewMockClient(ctrl)
			mockDocker.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(".duci/Dockerfile")).
				Return(&MockBuildLog{}, nil)
			mockDocker.EXPECT().
				Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Not(".duci/Dockerfile")).
				Return(nil, errors.New("must not call this"))
			mockDocker.EXPECT().
				Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1).
				Return("", &MockJobLog{}, nil)
			mockDocker.EXPECT().
				ExitCode(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(int64(0), nil)
			mockDocker.EXPECT().
				Rm(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(nil)

			// and
			mockLogStore := mock_logstore.NewMockService(ctrl)
			mockLogStore.EXPECT().
				Append(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(nil)
			mockLogStore.EXPECT().
				Start(gomock.Any()).
				AnyTimes().
				Return(nil)
			mockLogStore.EXPECT().
				Finish(gomock.Any()).
				AnyTimes().
				Return(nil)

			r := &runner.DockerRunner{
				Name:        "test-runner",
				BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
				Git:         mockGit,
				GitHub:      mockGitHub,
				Docker:      mockDocker,
				LogStore:    mockLogStore,
			}

			// and
			repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

			// when
			err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

			// then
			if err != nil {
				t.Errorf("must not error. but: %+v", err)
			}
		})
	})

	t.Run("with config file", func(t *testing.T) {
		// given
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockGit := mock_git.NewMockService(ctrl)
		mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			DoAndReturn(func(_ interface{}, dir string, _, _, _ interface{}) error {
				if err := os.MkdirAll(path.Join(dir, ".duci"), 0700); err != nil {
					return err
				}

				dockerfile, err := os.OpenFile(path.Join(dir, ".duci/config.yml"), os.O_RDWR|os.O_CREATE, 0600)
				if err != nil {
					return err
				}
				defer dockerfile.Close()

				dockerfile.WriteString("---\nvolumes:\n  - /hello:/hello")

				return nil
			})

		// and
		mockDocker := mock_docker.NewMockClient(ctrl)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&MockBuildLog{}, nil)
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Eq(docker.RuntimeOptions{Volumes: []string{"/hello:/hello"}}), gomock.Any(), gomock.Any()).
			Times(1).
			Return("", &MockJobLog{}, nil)
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Not(docker.RuntimeOptions{Volumes: []string{"/hello:/hello"}}), gomock.Any(), gomock.Any()).
			Return("", nil, errors.New("must not call this"))
		mockDocker.EXPECT().
			ExitCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(int64(0), nil)
		mockDocker.EXPECT().
			Rm(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Append(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Finish(gomock.Any()).
			AnyTimes().
			Return(nil)

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
			Git:         mockGit,
			GitHub:      mockGitHub,
			Docker:      mockDocker,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err != nil {
			t.Errorf("must not error. but: %+v", err)
		}
	})

	t.Run("when failed to git clone", func(t *testing.T) {
		// given
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockGit := mock_git.NewMockService(ctrl)
		mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(errors.New("error"))

		// and
		mockDocker := mock_docker.NewMockClient(ctrl)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
		mockDocker.EXPECT().
			ExitCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(int64(0), nil)
		mockDocker.EXPECT().
			Rm(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Append(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Finish(gomock.Any()).
			AnyTimes().
			Return(nil)

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
			Git:         mockGit,
			GitHub:      mockGitHub,
			Docker:      mockDocker,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err == nil {
			t.Error("must occur error")
		}
	})

	t.Run("when failed store#$tart", func(t *testing.T) {
		// given
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(errors.New("test error"))

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
			GitHub:      mockGitHub,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err == nil {
			t.Error("must occur error")
		}
	})

	t.Run("when workdir not exists", func(t *testing.T) {
		// given
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockGit := mock_git.NewMockService(ctrl)
		mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(nil)

		// and
		mockDocker := mock_docker.NewMockClient(ctrl)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
		mockDocker.EXPECT().
			ExitCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(int64(0), nil)
		mockDocker.EXPECT().
			Rm(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Append(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Finish(gomock.Any()).
			AnyTimes().
			Return(nil)

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: "/path/to/not/exists/dir",
			Git:         mockGit,
			GitHub:      mockGitHub,
			Docker:      mockDocker,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err == nil {
			t.Error("must occur error")
		}
	})

	t.Run("when docker build failure", func(t *testing.T) {
		// given
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockGit := mock_git.NewMockService(ctrl)
		mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(nil)

		// and
		mockDocker := mock_docker.NewMockClient(ctrl)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(nil, errors.New("test"))
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
		mockDocker.EXPECT().
			ExitCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(int64(0), nil)
		mockDocker.EXPECT().
			Rm(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Append(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Finish(gomock.Any()).
			AnyTimes().
			Return(nil)

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
			Git:         mockGit,
			GitHub:      mockGitHub,
			Docker:      mockDocker,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err == nil {
			t.Error("must occur error")
		}
	})

	t.Run("when docker run error", func(t *testing.T) {
		// given
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockGit := mock_git.NewMockService(ctrl)
		mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(nil)

		// and
		mockDocker := mock_docker.NewMockClient(ctrl)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(&MockBuildLog{}, nil)
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return("", nil, errors.New("test"))
		mockDocker.EXPECT().
			ExitCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(int64(0), nil)
		mockDocker.EXPECT().
			Rm(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Append(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Finish(gomock.Any()).
			AnyTimes().
			Return(nil)

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
			Git:         mockGit,
			GitHub:      mockGitHub,
			Docker:      mockDocker,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err == nil {
			t.Error("must occur error")
		}
	})

	t.Run("when fail to remove container", func(t *testing.T) {
		// given
		expected := errors.New("test")

		// and
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockGit := mock_git.NewMockService(ctrl)
		mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(nil)

		// and
		mockDocker := mock_docker.NewMockClient(ctrl)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(".duci/Dockerfile")).
			Return(&MockBuildLog{}, nil)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Not(".duci/Dockerfile")).
			Return(nil, errors.New("must not call this"))
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return("", &MockJobLog{}, nil)
		mockDocker.EXPECT().
			ExitCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(int64(0), nil)
		mockDocker.EXPECT().
			Rm(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(expected)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Append(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Finish(gomock.Any()).
			AnyTimes().
			Return(nil)

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
			Git:         mockGit,
			GitHub:      mockGitHub,
			Docker:      mockDocker,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err.Error() != expected.Error() {
			t.Errorf("err must be %+v, but got %+v", expected, err)
		}
	})

	t.Run("when docker run failure ( with exit code 1 )", func(t *testing.T) {
		// given
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockGit := mock_git.NewMockService(ctrl)
		mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(nil)

		// and
		mockDocker := mock_docker.NewMockClient(ctrl)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(&MockBuildLog{}, nil)
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return("", &MockJobLog{}, nil)
		mockDocker.EXPECT().
			ExitCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(int64(1), nil)
		mockDocker.EXPECT().
			Rm(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Append(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Finish(gomock.Any()).
			AnyTimes().
			Return(nil)

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
			Git:         mockGit,
			GitHub:      mockGitHub,
			Docker:      mockDocker,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err != runner.Failure {
			t.Errorf("error must be %s, but got %s", runner.Failure, err)
		}
	})

	t.Run("when runner timeout", func(t *testing.T) {
		// given
		mockGitHub := mock_github.NewMockService(ctrl)
		mockGitHub.EXPECT().CreateCommitStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(2).
			Return(nil)

		// and
		mockGit := mock_git.NewMockService(ctrl)
		mockGit.EXPECT().Clone(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(nil)

		// and
		application.Config.Job.Timeout = 1

		mockDocker := mock_docker.NewMockClient(ctrl)
		mockDocker.EXPECT().
			Build(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(&MockBuildLog{}, nil)
		mockDocker.EXPECT().
			Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			DoAndReturn(func(ctx context.Context, opts docker.RuntimeOptions, tag string, cmd ...string) (string, docker.Log, error) {
				time.Sleep(3 * time.Second)
				return "container_id", &MockJobLog{}, nil
			})
		mockDocker.EXPECT().
			ExitCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(int64(0), nil)
		mockDocker.EXPECT().
			Rm(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		// and
		mockLogStore := mock_logstore.NewMockService(ctrl)
		mockLogStore.EXPECT().
			Append(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Start(gomock.Any()).
			AnyTimes().
			Return(nil)
		mockLogStore.EXPECT().
			Finish(gomock.Any()).
			AnyTimes().
			Return(nil)

		r := &runner.DockerRunner{
			Name:        "test-runner",
			BaseWorkDir: path.Join(os.TempDir(), "test-runner"),
			Git:         mockGit,
			GitHub:      mockGitHub,
			Docker:      mockDocker,
			LogStore:    mockLogStore,
		}

		// and
		repo := &MockRepo{"duck8823/duci", "git@github.com:duck8823/duci.git"}

		// when
		err := r.Run(context.New("test/task", uuid.New(), &url.URL{}), repo, "master", plumbing.ZeroHash, "Hello World.")

		// then
		if err.Error() != "context deadline exceeded" {
			t.Errorf("error must be runner.Failure, but got %+v", err)
		}
	})
}

type MockRepo struct {
	FullName string
	SSHURL   string
}

func (r *MockRepo) GetFullName() string {
	return r.FullName
}

func (r *MockRepo) GetSSHURL() string {
	return r.SSHURL
}

type MockBuildLog struct {
}

func (l *MockBuildLog) ReadLine() (*docker.LogLine, error) {
	return &docker.LogLine{Timestamp: clock.Now(), Message: []byte("{\"stream\":\"Hello World,\"}")}, io.EOF
}

type MockJobLog struct {
}

func (l *MockJobLog) ReadLine() (*docker.LogLine, error) {
	return &docker.LogLine{Timestamp: clock.Now(), Message: []byte("Hello World,")}, io.EOF
}
