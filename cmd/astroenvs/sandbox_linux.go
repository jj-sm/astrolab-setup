//go:build linux

package main

// The sandbox runs a command in its own user, mount and PID namespaces (no root or setuid
// helper needed, only unprivileged user namespaces), in a root that holds nothing but:
//
//	system folders (/usr, /etc, /lib, ...)   read-only
//	the --ro paths (the venv, its Python)    read-only
//	the --rw folder                          read-write
//	the home folder, /tmp, /dev/shm          empty, in memory
//
// It works like bubblewrap: astroenvs starts itself as `astroenvs __sandbox ...` in the new
// namespaces, builds the root, pivots into it, drops every capability and runs the command
// as a child, staying behind as PID 1 to forward signals and reap orphans.

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	capSysAdmin = 21

	prCapAmbient         = 47
	prCapAmbientClearAll = 4
	prSetNoNewPrivs      = 38

	linuxCapabilityVersion3 = 0x20080522

	// statfs flags that a read-only remount must keep: the kernel refuses to clear them
	// on mounts inherited from the host (EPERM)
	stNosuid     = 0x2
	stNodev      = 0x4
	stNoexec     = 0x8
	stNoatime    = 0x400
	stNodiratime = 0x800
	stRelatime   = 0x1000
)

// sandboxSupported reports whether this kernel lets users create user namespaces
func sandboxSupported() bool {
	b, err := os.ReadFile("/proc/sys/user/max_user_namespaces")
	if err != nil {
		return false
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	return n > 0
}

// sandboxCommand returns a command that runs prog inside a sandbox where only rw (read-write),
// the ro paths (read-only) and the system folders exist
func sandboxCommand(rw string, ro []string, prog string, args ...string) *exec.Cmd {
	a := []string{sandboxArg, "--rw", rw}
	for _, p := range ro {
		a = append(a, "--ro", p)
	}
	a = append(append(a, "--", prog), args...)

	cmd := exec.Command("/proc/self/exe", a...)
	uid, gid := os.Getuid(), os.Getgid()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:     true,
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
		// Same IDs inside, so files keep their owner and nothing runs as root
		UidMappings: []syscall.SysProcIDMap{{ContainerID: uid, HostID: uid, Size: 1}},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: gid, HostID: gid, Size: 1}},
		// Kept across exec for the mounts only; runSandbox drops it before starting prog
		AmbientCaps: []uintptr{capSysAdmin},
	}
	return cmd
}

// runSandbox is the `astroenvs __sandbox` entry point, already inside the new namespaces
func runSandbox(argv []string) {
	var rw string
	var ro, cmdArgs []string
	for i := 0; i < len(argv); i++ {
		switch {
		case argv[i] == "--rw" && i+1 < len(argv):
			i++
			rw = argv[i]
		case argv[i] == "--ro" && i+1 < len(argv):
			i++
			ro = append(ro, argv[i])
		case argv[i] == "--":
			cmdArgs = argv[i+1:]
			i = len(argv)
		}
	}
	die := func(err error) {
		fmt.Fprintln(os.Stderr, "sandbox:", err)
		os.Exit(1)
	}
	if rw == "" || len(cmdArgs) == 0 {
		die(errors.New("usage: __sandbox --rw DIR [--ro PATH]... -- COMMAND [ARGS]..."))
	}
	if err := buildRoot(rw, ro); err != nil {
		die(err)
	}
	if err := dropPrivileges(); err != nil {
		die(err)
	}
	os.Exit(runInit(cmdArgs, rw))
}

// buildRoot assembles the sandbox's root and pivots into it
func buildRoot(rw string, ro []string) error {
	// Nothing mounted from here on may propagate back to the host
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("making mounts private: %w", err)
	}
	// Scratch root on /tmp: /newroot becomes the sandbox, /oldroot keeps the host's tree. Pivoting
	// moves the tmpfs off /tmp, so the host's own /tmp stays reachable under /oldroot
	if err := syscall.Mount("tmpfs", "/tmp", "tmpfs", syscall.MS_NOSUID|syscall.MS_NODEV, "size=1m,mode=0700"); err != nil {
		return fmt.Errorf("mounting the scratch root: %w", err)
	}
	for _, d := range []string{"/tmp/newroot", "/tmp/oldroot"} {
		if err := os.Mkdir(d, 0o700); err != nil {
			return err
		}
	}
	// The sandbox's root, which also holds the (empty, throwaway) home folder
	if err := syscall.Mount("tmpfs", "/tmp/newroot", "tmpfs", syscall.MS_NOSUID|syscall.MS_NODEV, "size=256m,mode=0755"); err != nil {
		return fmt.Errorf("mounting the new root: %w", err)
	}
	if err := syscall.PivotRoot("/tmp", "/tmp/oldroot"); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return err
	}

	for _, p := range []string{"/usr", "/bin", "/sbin", "/lib", "/lib32", "/lib64", "/libx32", "/etc", "/opt", "/sys"} {
		if err := bindSystem(p); err != nil {
			return err
		}
	}
	// DNS: /etc/resolv.conf often points into /run (systemd-resolved, NetworkManager)
	if target, err := evalOnHost("/etc/resolv.conf"); err == nil && strings.HasPrefix(target, "/run/") {
		if err := bind(filepath.Dir(target), true); err != nil {
			return err
		}
	}
	if err := mountSpecial(); err != nil {
		return err
	}
	if err := bind(rw, false); err != nil {
		return err
	}
	// After rw, so a venv inside the shared folder is still read-only
	for _, p := range ro {
		if err := bind(p, true); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join("/newroot", home()), 0o700); err != nil {
		return err
	}

	// Drop the host's tree, then make /newroot the root and discard the scratch tmpfs under it
	if err := syscall.Unmount("/oldroot", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("detaching the host's root: %w", err)
	}
	if err := os.Chdir("/newroot"); err != nil {
		return err
	}
	if err := syscall.PivotRoot(".", "."); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}
	if err := syscall.Unmount(".", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("detaching the scratch root: %w", err)
	}
	return os.Chdir("/")
}

// evalOnHost resolves symlinks in a host path, reading the host's tree under /oldroot
func evalOnHost(p string) (string, error) {
	for i := 0; i < 40; i++ {
		target, err := os.Readlink("/oldroot" + p)
		if err != nil {
			if errors.Is(err, syscall.EINVAL) { // not a symlink: check the parents
				parent := filepath.Dir(p)
				if parent == p {
					return p, nil
				}
				rp, err := evalOnHost(parent)
				if err != nil {
					return "", err
				}
				return filepath.Join(rp, filepath.Base(p)), nil
			}
			return "", err
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(p), target)
		}
		p = filepath.Clean(target)
	}
	return "", fmt.Errorf("%s: too many symlinks", p)
}

// bindSystem mirrors a top-level system folder read-only, or the symlink it is (/bin -> usr/bin)
func bindSystem(p string) error {
	st, err := os.Lstat("/oldroot" + p)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink("/oldroot" + p)
		if err != nil {
			return err
		}
		return os.Symlink(target, "/newroot"+p)
	}
	return bind(p, true)
}

// bind mounts the host path p at the same path in the sandbox, with its submounts
func bind(p string, readonly bool) error {
	src, err := evalOnHost(p)
	if err != nil {
		return fmt.Errorf("%s: %w", p, err)
	}
	st, err := os.Stat("/oldroot" + src)
	if err != nil {
		return fmt.Errorf("%s: %w", p, err)
	}
	dst := "/newroot" + p
	if st.IsDir() {
		err = os.MkdirAll(dst, 0o755)
	} else if err = os.MkdirAll(filepath.Dir(dst), 0o755); err == nil {
		var f *os.File
		if f, err = os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
			f.Close()
		}
	}
	if err != nil {
		return fmt.Errorf("%s: %w", p, err)
	}
	if err := syscall.Mount("/oldroot"+src, dst, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("binding %s: %w", p, err)
	}
	if readonly {
		return remountReadOnly(dst)
	}
	return nil
}

// remountReadOnly makes dst and every mount below it read-only (a bind's remount only
// changes the top one). Failing on dst itself is an error; on a submount it is skipped,
// since the host's permissions still apply there
func remountReadOnly(dst string) error {
	mounts, err := mountPoints()
	if err != nil {
		return err
	}
	for _, mp := range mounts {
		if mp != dst && !strings.HasPrefix(mp, dst+"/") {
			continue
		}
		var st syscall.Statfs_t
		flags := uintptr(syscall.MS_BIND | syscall.MS_REMOUNT | syscall.MS_RDONLY)
		if err := syscall.Statfs(mp, &st); err == nil {
			for _, f := range []struct{ st, ms uintptr }{
				{stNosuid, syscall.MS_NOSUID}, {stNodev, syscall.MS_NODEV}, {stNoexec, syscall.MS_NOEXEC},
				{stNoatime, syscall.MS_NOATIME}, {stNodiratime, syscall.MS_NODIRATIME}, {stRelatime, syscall.MS_RELATIME},
			} {
				if uintptr(st.Flags)&f.st != 0 {
					flags |= f.ms
				}
			}
		}
		if err := syscall.Mount("", mp, "", flags, ""); err != nil && mp == dst {
			return fmt.Errorf("making %s read-only: %w", strings.TrimPrefix(dst, "/newroot"), err)
		}
	}
	return nil
}

// mountPoints lists the current mount points, read from the host's /proc (paths are
// relative to the current root, so they start with /newroot or /oldroot)
func mountPoints() ([]string, error) {
	f, err := os.Open("/oldroot/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var mps []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) > 4 {
			mps = append(mps, unescapeMountinfo(fields[4]))
		}
	}
	return mps, sc.Err()
}

// unescapeMountinfo decodes the octal escapes mountinfo uses for spaces and the like (\040)
func unescapeMountinfo(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) {
			if n, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(n))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// mountSpecial sets up /proc (this PID namespace only), a minimal /dev and an empty /tmp
func mountSpecial() error {
	for _, d := range []string{"proc", "dev", "tmp"} {
		if err := os.MkdirAll("/newroot/"+d, 0o755); err != nil {
			return err
		}
	}
	if err := syscall.Mount("proc", "/newroot/proc", "proc", syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, ""); err != nil {
		return fmt.Errorf("mounting /proc: %w", err)
	}
	if err := syscall.Mount("tmpfs", "/newroot/tmp", "tmpfs", syscall.MS_NOSUID|syscall.MS_NODEV, "size=1g,mode=1777"); err != nil {
		return fmt.Errorf("mounting /tmp: %w", err)
	}

	// /dev: only the harmless devices, not the host's terminals, disks or shared memory
	if err := syscall.Mount("tmpfs", "/newroot/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_NOEXEC, "size=1m,mode=0755"); err != nil {
		return fmt.Errorf("mounting /dev: %w", err)
	}
	for _, d := range []string{"null", "zero", "full", "random", "urandom", "tty"} {
		f, err := os.OpenFile("/newroot/dev/"+d, os.O_CREATE|os.O_WRONLY, 0o666)
		if err != nil {
			return err
		}
		f.Close()
		if err := syscall.Mount("/oldroot/dev/"+d, "/newroot/dev/"+d, "", syscall.MS_BIND, ""); err != nil {
			return fmt.Errorf("binding /dev/%s: %w", d, err)
		}
	}
	for _, d := range []string{"pts", "shm"} {
		if err := os.Mkdir("/newroot/dev/"+d, 0o755); err != nil {
			return err
		}
	}
	// A private pseudo-terminal instance, for Jupyter's terminals. Without it they just won't open
	if err := syscall.Mount("devpts", "/newroot/dev/pts", "devpts", syscall.MS_NOSUID|syscall.MS_NOEXEC, "newinstance,ptmxmode=0666,mode=0620"); err != nil {
		fmt.Fprintln(os.Stderr, "sandbox: no terminals (mounting /dev/pts failed):", err)
	} else if err := os.Symlink("pts/ptmx", "/newroot/dev/ptmx"); err != nil {
		return err
	}
	if err := syscall.Mount("tmpfs", "/newroot/dev/shm", "tmpfs", syscall.MS_NOSUID|syscall.MS_NODEV, "size=256m,mode=1777"); err != nil {
		return fmt.Errorf("mounting /dev/shm: %w", err)
	}
	links := map[string]string{"fd": "/proc/self/fd", "stdin": "/proc/self/fd/0", "stdout": "/proc/self/fd/1", "stderr": "/proc/self/fd/2"}
	for name, target := range links {
		if err := os.Symlink(target, "/newroot/dev/"+name); err != nil {
			return err
		}
	}
	return nil
}

type capHeader struct {
	version uint32
	pid     int32
}

type capData struct {
	effective, permitted, inheritable uint32
}

// dropPrivileges removes every capability from every thread, so nothing started from here
// can change mounts (and undo the sandbox) or gain privileges through setuid binaries
func dropPrivileges() error {
	if _, _, e := syscall.AllThreadsSyscall(syscall.SYS_PRCTL, prSetNoNewPrivs, 1, 0); e != 0 {
		return fmt.Errorf("no_new_privs: %w", e)
	}
	if _, _, e := syscall.AllThreadsSyscall(syscall.SYS_PRCTL, prCapAmbient, prCapAmbientClearAll, 0); e != 0 {
		return fmt.Errorf("clearing ambient capabilities: %w", e)
	}
	hdr := capHeader{version: linuxCapabilityVersion3}
	var data [2]capData
	_, _, e := syscall.AllThreadsSyscall(syscall.SYS_CAPSET, uintptr(unsafe.Pointer(&hdr)), uintptr(unsafe.Pointer(&data[0])), 0)
	runtime.KeepAlive(&hdr)
	runtime.KeepAlive(&data)
	if e != 0 {
		return fmt.Errorf("dropping capabilities: %w", e)
	}
	return nil
}

// runInit starts the command and acts as the namespace's PID 1 until it exits: it passes on
// the signals that stop it and reaps the orphans PID 1 inherits
func runInit(args []string, dir string) int {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	for _, kv := range os.Environ() {
		// The host's runtime dir doesn't exist in here, and Jupyter would try to use it
		if !strings.HasPrefix(kv, "XDG_RUNTIME_DIR=") && !strings.HasPrefix(kv, "TMPDIR=") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	cmd.Env = append(cmd.Env, "TMPDIR=/tmp")

	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "sandbox:", err)
		return 1
	}
	go func() {
		for s := range sigs {
			cmd.Process.Signal(s)
		}
	}()
	for {
		var ws syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &ws, 0, nil)
		if err == syscall.EINTR {
			continue
		}
		if err != nil {
			return 1
		}
		if pid == cmd.Process.Pid {
			if ws.Signaled() {
				return 128 + int(ws.Signal())
			}
			return ws.ExitStatus()
		}
	}
}
