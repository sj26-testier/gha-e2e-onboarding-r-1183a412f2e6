// Diagnostic only: inspect the Buildkite Agent executable path without reading credentials.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

func emit(kind string, values map[string]any) {
	values["kind"] = kind
	data, _ := json.Marshal(values)
	fmt.Println(string(data))
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func main() {
	emit("runtime", map[string]any{"go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH})
	candidate := os.Getenv("BUILDKITE_GHA_AGENT")
	if candidate == "" {
		candidate = "buildkite-agent"
	}
	path, err := exec.LookPath(candidate)
	emit("LookPath", map[string]any{"candidate": candidate, "path": path, "error": errText(err)})
	if err != nil {
		return
	}
	abs, err := filepath.Abs(path)
	emit("Abs", map[string]any{"path": abs, "error": errText(err)})
	if err != nil {
		return
	}
	resolved, err := filepath.EvalSymlinks(abs)
	emit("EvalSymlinks", map[string]any{"input": abs, "path": resolved, "error": errText(err)})
	file, openErr := os.Open(abs)
	emit("open", map[string]any{"path": abs, "error": errText(openErr)})
	if openErr == nil {
		defer file.Close()
		proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetFinalPathNameByHandleW")
		for _, flags := range []uintptr{0, 8} { // DOS normalized, then FILE_NAME_OPENED.
			buf := make([]uint16, 32768)
			n, _, callErr := proc.Call(file.Fd(), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), flags)
			path, message := "", ""
			if n == 0 {
				message = callErr.Error()
			} else if n >= uintptr(len(buf)) {
				message = "result exceeds diagnostic buffer"
			} else {
				path = syscall.UTF16ToString(buf[:n])
			}
			emit("GetFinalPathNameByHandleW", map[string]any{"flags": flags, "path": path, "error": message})
		}
	}
	for component := abs; ; component = filepath.Dir(component) {
		info, statErr := os.Lstat(component)
		mode := ""
		if info != nil {
			mode = info.Mode().String()
		}
		target, linkErr := os.Readlink(component)
		canonical, evalErr := filepath.EvalSymlinks(component)
		_, followErr := os.Stat(component)
		emit("component", map[string]any{
			"path": component, "mode": mode, "lstat_error": errText(statErr),
			"readlink_target": target, "readlink_error": errText(linkErr),
			"canonical": canonical, "eval_error": errText(evalErr), "stat_error": errText(followErr),
		})
		if filepath.Dir(component) == component {
			break
		}
	}
}
