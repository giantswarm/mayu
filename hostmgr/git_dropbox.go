package hostmgr

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/golang/glog"
)

func parentDir(p string) string {
	return path.Join(path.Dir(p), "..")
}

var gitMutex *sync.Mutex

func init() {
	// this sucks, but right now is the easiest way without
	// defining an explicit host -> cluster relationship
	gitMutex = new(sync.Mutex)
}

func isGitRepo(p string) bool {
	if fi, err := os.Stat(path.Join(p, ".git")); err == nil {
		return fi.IsDir()
	}
	return false
}

func gitAddCommit(baseDir string, path string, commitMsg string) error {
	err := gitAdd(baseDir, path)
	if err != nil {
		glog.V(5).Infoln("error git-adding:", err)
		return microerror.Mask(err)
	}
	return gitCommit(baseDir, commitMsg)
}

var DisableGit bool

func gitExec(baseDir string, args ...string) error {
	gitMutex.Lock()
	defer gitMutex.Unlock()
	cmdline := []string{"git"}
	if DisableGit {
		glog.V(4).Infoln("GIT", strings.Join(args, " "))
		return nil
	}
	err := cmdExec(baseDir, append(cmdline, args...)...)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func cmdExec(cwd string, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	glog.V(6).Infoln("running", args)
	cmd.Stdin = os.Stdin
	if glog.V(8) {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return microerror.Mask(err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return microerror.Mask(err)
		}
		multiReader := io.MultiReader(stdout, stderr)
		pipeLogger := func(rdr io.Reader) {
			scanner := bufio.NewScanner(rdr)
			for scanner.Scan() {
				glog.V(8).Infoln(scanner.Text())
			}
		}
		go pipeLogger(multiReader)
	}
	err := cmd.Run()
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func gitAdd(baseDir string, path string) error {
	absPath, err := filepath.Abs(path)
	glog.V(3).Infof("adding file '%s' to '%s'\n", absPath, baseDir)
	if err != nil {
		return microerror.Mask(err)
	}
	err = gitExec(baseDir, "add", absPath)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func gitCommit(baseDir string, commitMsg string) error {
	err := gitExec(baseDir, "commit", "-m", commitMsg)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func gitPush(baseDir string) error {
	err := gitExec(baseDir, "push")
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func gitInit(baseDir string) error {
	err := gitExec(baseDir, "init")
	if err != nil {
		return microerror.Mask(err)
	}
	err = gitExec(baseDir, "config", "--local", "user.name", "mayu commiter")
	if err != nil {
		return microerror.Mask(err)
	}
	err = gitExec(baseDir, "config", "--local", "push.default", "matching")
	if err != nil {
		return microerror.Mask(err)
	}
	err = gitExec(baseDir, "config", "--local", "user.email", "support+noise@giantswarm.io")
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
