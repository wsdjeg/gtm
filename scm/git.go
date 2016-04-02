package scm

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func GitRootPath(path ...string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}

	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Unable to parse repository path, %s %s", string(b), err)
	}

	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", fmt.Errorf("Unable to parse repository path, %s", err)
	}

	return s, nil
}

func GitBranch(path ...string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return "", fmt.Errorf("Unable to parse branch name, %s %s", string(b), err)
	}
	return strings.TrimSpace(string(b)), nil

}

func GitEmail(path ...string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "user.email")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return "", fmt.Errorf("Unable to get user email, %s %s", string(b), err)
	}
	return strings.TrimSpace(string(b)), nil
}

func GitCommitMsg(path ...string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--oneline", "--raw")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		// if there are no git commits yet it will fail
		// ignoring this error
		return "", nil
	}
	return string(b), err
}

func GitParseMessage(m string) (uuid, msg string, files []string) {
	l := strings.Split(m, "\n")
	files = make([]string, 0)
	for i, v := range l {
		if i == 0 {
			s := strings.SplitN(v, " ", 2)
			uuid = s[0]
			msg = s[1]
		} else {
			if strings.TrimSpace(v) != "" {
				s := strings.Split(v, "\t")
				files = append(files, s[1])
			}
		}
	}
	return
}

func GitAddNote(n string, nameSpace string, path ...string) error {
	cmd := exec.Command("git", "notes", fmt.Sprintf("--ref=%s", nameSpace), "add", "-f", "-m", n)
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	if b, err := cmd.Output(); err != nil {
		return fmt.Errorf("Unable to add git note, %s %s", string(b), err)
	}
	return nil
}

func GitSetRewriteRef(ref string, path ...string) error {
	cmd := exec.Command("git", "config", "-l")
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return fmt.Errorf("Unable to run git config -l notes.rewriteref, %s %s", string(b), err)
	}
	if !strings.Contains(string(b), ref+"\n") {
		cmd := exec.Command("git", "config", "--add", "notes.rewriteref", ref)
		if len(path) > 0 {
			cmd.Dir = path[0]
		}
		if b, err := cmd.Output(); err != nil {
			return fmt.Errorf("Unable to run git config --add notes.rewriteref %s, %s %s", ref, string(b), err)
		}
	}
	return nil
}

func GitTracked(f string, path ...string) (bool, error) {
	cmd := exec.Command("git", "ls-files", f)
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return false, fmt.Errorf("Unable to determine git tracked status for %s, %s %s", f, string(b), err)
	}
	return strings.TrimSpace(string(b)) != "", nil
}

func GitModified(f string, path ...string) (bool, error) {
	cmd := exec.Command("git", "ls-files", "-m", f)
	if len(path) > 0 {
		cmd.Dir = path[0]
	}
	var (
		b   []byte
		err error
	)
	if b, err = cmd.Output(); err != nil {
		return false, fmt.Errorf("Unable to determine git modified status for %s, %s %s", f, string(b), err)
	}
	return strings.TrimSpace(string(b)) != "", nil
}

func GitInitHook(hook, command string, wd ...string) error {
	var (
		p   string
		err error
	)

	if len(wd) > 0 {
		p = wd[0]
	} else {
		p, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	fp := path.Join(p, ".git", "hooks", hook)

	var output string
	if _, err := os.Stat(fp); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(fp)
		if err != nil {
			return err
		}
		output = string(b)

		if strings.Contains(output, command+"\n") {
			// if file already exists this will make sure it's executable
			if err := os.Chmod(fp, 0755); err != nil {
				return err
			}
			return nil
		}
	}

	if err = ioutil.WriteFile(
		fp, []byte(fmt.Sprintf("%s\n%s\n", output, command)), 0755); err != nil {
		return err
	}
	// if file already exists this will make sure it's executable
	if err := os.Chmod(fp, 0755); err != nil {
		return err
	}

	return nil
}

func GitIgnore(ignore string, wd ...string) error {
	var (
		p   string
		err error
	)

	if len(wd) > 0 {
		p = wd[0]
	} else {
		p, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	fp := path.Join(p, ".gitignore")

	var output string
	if _, err := os.Stat(fp); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(fp)
		if err != nil {
			return err
		}
		output = string(b)

		if strings.Contains(output, ignore+"\n") {
			return nil
		}
	}

	if err = ioutil.WriteFile(
		fp, []byte(fmt.Sprintf("%s\n%s\n", output, ignore)), 0644); err != nil {
		return err
	}
	return nil
}
