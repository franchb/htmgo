package process

import (
	"fmt"
	"github.com/franchb/htmgo/cli/htmgo/v2/internal"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type CmdWithFlags struct {
	Flags []RunFlag
	Name  string
	Cmd   *exec.Cmd
}

type RawCommand struct {
	Name  string
	Args  string
	Flags []RunFlag
}

func NewRawCommand(name string, args string, flags ...RunFlag) RawCommand {
	if name == "" {
		name = args
	}
	c := RawCommand{Name: name, Args: args, Flags: flags}
	if c.Flags == nil {
		c.Flags = make([]RunFlag, 0)
	}
	return c
}

var workingDir string

// commands holds started processes by name; guard every access with commandsMu.
// Only commands whose Start has returned are stored, so Cmd.Process is never
// written while another goroutine may read it.
var commands = make(map[string]CmdWithFlags)
var commandsMu sync.Mutex

func AppendRunning(cmd *exec.Cmd, raw RawCommand) {
	slog.Debug("running", slog.String("command", strings.Join(cmd.Args, " ")),
		slog.String("dir", cmd.Dir),
		slog.String("cwd", GetWorkingDir()))

	commandsMu.Lock()
	defer commandsMu.Unlock()
	commands[raw.Name] = CmdWithFlags{Flags: raw.Flags, Name: raw.Name, Cmd: cmd}
}

// snapshotCommands returns a copy of the running commands that is safe to
// iterate without holding commandsMu.
func snapshotCommands() []CmdWithFlags {
	commandsMu.Lock()
	defer commandsMu.Unlock()
	snapshot := make([]CmdWithFlags, 0, len(commands))
	for _, cmd := range commands {
		snapshot = append(snapshot, cmd)
	}
	return snapshot
}

func resetCommands(keep []CmdWithFlags) {
	commandsMu.Lock()
	defer commandsMu.Unlock()
	commands = make(map[string]CmdWithFlags, len(keep))
	for _, cmd := range keep {
		commands[cmd.Name] = cmd
	}
}

// removeCommand forgets name, unless it has since been replaced by another run.
func removeCommand(name string, cmd *exec.Cmd) {
	commandsMu.Lock()
	defer commandsMu.Unlock()
	if existing, ok := commands[name]; ok && existing.Cmd == cmd {
		delete(commands, name)
	}
}

func GetWorkingDir() string {
	if workingDir == "" {
		wd, _ := os.Getwd()
		return wd
	}
	return workingDir
}

func SetWorkingDir(dir string) {
	workingDir = dir
}

func GetPathRelativeToCwd(path string) string {
	return filepath.Join(GetWorkingDir(), path)
}

func shouldSkipKilling(flags []RunFlag, skipFlag []RunFlag) bool {
	for _, flag := range flags {
		if slices.Contains(skipFlag, flag) {
			return true
		}
	}
	return false
}

func StartLogger() {
	if internal.GetLogLevel() != slog.LevelDebug {
		return
	}
	go func() {
		for {
			time.Sleep(time.Second * 5)
			items := make([]map[string]string, 0)
			for _, cmd := range snapshotCommands() {
				data := make(map[string]string)
				data["command"] = fmt.Sprintf("%s %s", cmd.Cmd.Path, strings.Join(cmd.Cmd.Args, " "))
				if cmd.Cmd.Process != nil {
					data["pid"] = fmt.Sprintf("%d", cmd.Cmd.Process.Pid)
				}
				items = append(items, data)
			}

			fmt.Printf("Running processes:\n")
			for i, item := range items {
				fmt.Printf("%d: %+v\n", i, item)
			}
			fmt.Printf("\n")
		}
	}()
}

// stopped records commands the CLI killed on purpose, so Run can tell a
// deliberate restart from a crash. On Unix a killed process reports
// "signal: killed", but on Windows taskkill makes it exit with status 1.
var stopped sync.Map // *exec.Cmd -> struct{}

func kill(command CmdWithFlags) error {
	if command.Cmd != nil {
		stopped.Store(command.Cmd, struct{}{})
	}
	return KillProcess(command)
}

// Stop kills the named process if it is running and waits up to timeout for
// it to exit, so files it holds open (on Windows, its own executable) are
// released before the caller continues.
func Stop(name string, timeout time.Duration) {
	existing := GetProcessByName(name)
	if existing == nil || existing.Cmd == nil || existing.Cmd.Process == nil {
		return
	}
	slog.Debug("stopping process", slog.String("command", name))
	_ = kill(*existing)
	pid := int32(existing.Cmd.Process.Pid)
	deadline := time.Now().Add(timeout)
	for PidExists(pid) && time.Now().Before(deadline) {
		time.Sleep(25 * time.Millisecond)
	}
}

func GetProcessByName(name string) *CmdWithFlags {
	commandsMu.Lock()
	defer commandsMu.Unlock()
	if cmd, ok := commands[name]; ok {
		return &cmd
	}
	return nil
}

func OnShutdown() {
	// request for shutdown
	for _, cmd := range snapshotCommands() {
		if cmd.Cmd != nil && cmd.Cmd.Process != nil {
			cmd.Cmd.Process.Signal(os.Interrupt)
		}
	}
	// give it a second
	time.Sleep(time.Second * 1)
	// force kill
	KillAll()
}

func KillAll(skipFlag ...RunFlag) {

	tries := 0
	updatedCommands := make(map[string]CmdWithFlags)
	for {
		tries++
		allFinished := true
		for _, cmd := range snapshotCommands() {
			if cmd.Cmd.Process == nil {
				allFinished = false
				if tries > 50 {
					args := strings.Join(cmd.Cmd.Args, " ")
					slog.Debug("process is not running after 50 tries, breaking.", slog.String("command", args))
					allFinished = true
					break
				} else {
					time.Sleep(time.Millisecond * 50)
					continue
				}
			} else {
				updatedCommands[cmd.Name] = cmd
			}
		}
		if allFinished {
			break
		}
	}

	running := make([]CmdWithFlags, 0, len(updatedCommands))
	for _, command := range updatedCommands {
		if command.Cmd != nil && command.Cmd.Process != nil {
			running = append(running, command)
		}
	}
	resetCommands(running)

	for _, command := range running {
		if shouldSkipKilling(command.Flags, skipFlag) {
			continue
		}
		err := kill(command)
		if err != nil {
			continue
		}
	}

	for {
		finished := true
		for _, c := range running {
			if c.Cmd.Process == nil {
				continue
			}
			if shouldSkipKilling(c.Flags, skipFlag) {
				continue
			}
			exists := PidExists(int32(c.Cmd.Process.Pid))
			if exists {
				_ = kill(c)
				finished = false
			}
		}

		if finished {
			break
		} else {
			slog.Debug("waiting for all processes to exit\n")
			time.Sleep(time.Millisecond * 5)
		}
	}

	resetCommands(nil)
	slog.Debug("all processes killed\n")
}

func RunOrExit(command RawCommand) {
	command.Flags = append(command.Flags, ExitOnError)
	_ = Run(command)
}

type RunFlag int

const (
	ExitOnError RunFlag = iota
	Silent
	KillOnlyOnExit
)

func RunMany(commands ...RawCommand) error {
	for _, command := range commands {
		err := Run(command)
		if err != nil {
			if slices.Contains(command.Flags, ExitOnError) {
				os.Exit(1)
			}
			return err
		}
	}
	return nil
}

var mutex = &sync.Mutex{}

func Run(command RawCommand) error {
	mutex.Lock()

	parts := strings.Fields(command.Args)

	args := make([]string, 0)
	if len(parts) > 1 {
		args = parts[1:]
	}

	path := parts[0]

	existing := GetProcessByName(command.Name)

	if existing != nil {
		slog.Debug("process already running, killing it", slog.String("command", command.Name))
		_ = kill(*existing)
		time.Sleep(time.Millisecond * 50)
	} else {
		slog.Debug("no existing process found for %s, safe to run...", slog.String("command", command.Name))
	}

	cmd := exec.Command(path, args...)

	PrepareCommand(cmd)

	if slices.Contains(command.Flags, Silent) {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Start before publishing cmd so that cmd.Process is fully written before
	// KillAll/KillProcess can read it from another goroutine.
	err := cmd.Start()
	if err == nil {
		AppendRunning(cmd, command)
	}

	mutex.Unlock()

	if err == nil {
		err = cmd.Wait()
	}

	slog.Debug("command finished",
		slog.String("command", command.Name),
		slog.String("args", command.Args),
		slog.String("dir", cmd.Dir),
		slog.String("cwd", GetWorkingDir()),
		slog.String("error", fmt.Sprintf("%v", err)))

	removeCommand(command.Name, cmd)
	_, wasStopped := stopped.LoadAndDelete(cmd)

	if err == nil || wasStopped {
		return nil
	}

	if strings.Contains(err.Error(), "signal: killed") {
		return nil
	}

	if slices.Contains(command.Flags, ExitOnError) {
		slog.Error("Error running command: ",
			slog.String("error", err.Error()),
			slog.String("command", command.Name))
		os.Exit(1)
	}

	return err
}
