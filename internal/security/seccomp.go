package security

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// SeccompProfile represents a seccomp profile for container security
type SeccompProfile struct {
	DefaultAction string        `json:"defaultAction"`
	Architectures []string      `json:"architectures"`
	Syscalls      []SyscallRule `json:"syscalls"`
}

// SyscallRule defines a syscall rule
type SyscallRule struct {
	Names  []string `json:"names"`
	Action string   `json:"action"`
	Args   []Arg    `json:"args,omitempty"`
}

// Arg defines a syscall argument rule
type Arg struct {
	Index    int    `json:"index"`
	Value    uint64 `json:"value,omitempty"`
	ValueTwo uint64 `json:"valueTwo,omitempty"`
	Op       string `json:"op"`
}

// dangerousSyscalls lists syscalls that should never be allowed in
// sandboxed containers. These are explicitly denied with SCMP_ACT_KILL
// to ensure they cannot be whitelisted by accident.
var dangerousSyscalls = []string{
	"acct",             // process accounting
	"add_key",          // kernel keyring manipulation
	"bpf",              // eBPF programs (kernel attack surface)
	"clock_adjtime",    // adjust system clock
	"clock_settime",    // set system clock
	"create_module",    // load kernel modules
	"delete_module",    // unload kernel modules
	"finit_module",     // load kernel module from fd
	"get_kernel_syms",  // retrieve kernel symbols
	"init_module",      // load kernel module
	"ioperm",           // set port I/O permissions
	"iopl",             // change I/O privilege level
	"kcmp",             // compare two processes
	"kexec_file_load",  // load new kernel for later execution
	"kexec_load",       // load new kernel for later execution
	"keyctl",           // kernel keyring manipulation
	"lookup_dcookie",   // return directory entry cookie
	"mount",            // mount filesystems
	"move_pages",       // move pages across NUMA nodes
	"nfsservctl",       // NFS server control
	"open_by_handle_at", // open file by handle (bypass DAC)
	"perf_event_open",  // performance monitoring (attack surface)
	"pivot_root",       // change root filesystem
	"process_vm_readv", // read another process's memory
	"process_vm_writev", // write another process's memory
	"ptrace",           // process tracing (debugger, injection)
	"query_module",     // query kernel module info
	"quotactl",         // disk quota manipulation
	"reboot",           // reboot the system
	"request_key",      // request key from kernel keyring
	"setdomainname",    // set NIS domain name
	"sethostname",      // set hostname
	"settimeofday",     // set system time
	"swapon",           // enable swap area
	"swapoff",          // disable swap area
	"syslog",           // read/clear kernel message buffer
	"umount2",          // unmount filesystems
	"unshare",          // create new namespaces (escape risk)
	"userfaultfd",      // userfaultfd (attack surface)
	"vhangup",          // simulate hangup on terminal
	"setns",            // enter another namespace (escape risk)
	"chroot",           // change root directory (escape risk)
	"adjtimex",         // adjust kernel clock
	"modify_ldt",       // modify local descriptor table (kernel attack surface)
	"clone3",           // new clone with extensible args (escape risk)
	"move_mount",       // move mount point (escape risk)
	"open_tree",        // open mount tree (escape risk)
	"fsopen",           // open filesystem context (escape risk)
	"fspick",           // pick filesystem (escape risk)
	"fsconfig",         // configure filesystem (escape risk)
	"fsmount",          // mount filesystem (escape risk)
}

// essentialSyscalls lists syscalls required for basic container operation:
// file I/O, memory management, process management, networking, and signals.
var essentialSyscalls = []string{
	// File I/O
	"read", "write", "open", "openat", "close", "stat", "fstat", "lstat",
	"newfstatat", "lseek", "access", "faccessat", "creat",
	"rename", "renameat", "renameat2",
	"mkdir", "mkdirat", "rmdir",
	"link", "linkat", "unlink", "unlinkat",
	"symlink", "symlinkat", "readlink", "readlinkat",
	"chmod", "fchmod", "fchmodat",
	"chown", "fchown", "lchown", "fchownat",
	"truncate", "ftruncate",
	"getdents", "getdents64",
	"getcwd", "chdir", "fchdir",
	"fcntl", "flock",
	"fsync", "fdatasync", "sync", "syncfs",
	"sendfile",
	"fallocate",
	"statfs", "fstatfs",
	"statx",
	"readahead",
	"copy_file_range",
	"splice", "tee", "vmsplice",
	"sync_file_range",
	"fadvise64",
	"name_to_handle_at",
	"umask",

	// Memory management
	"mmap", "mprotect", "munmap", "brk", "mremap",
	"msync", "mincore", "madvise",
	"mlock", "munlock", "mlockall", "munlockall", "mlock2",
	"pkey_mprotect", "pkey_alloc", "pkey_free",
	"memfd_create", "membarrier",
	"remap_file_pages",

	// Process management
	"clone", "fork", "vfork", "execve", "execveat",
	"exit", "exit_group",
	"wait4", "waitid",
	"getpid", "getppid", "gettid",
	"getuid", "getgid", "geteuid", "getegid",
	"setuid", "setgid",
	"setreuid", "setregid",
	"setresuid", "getresuid", "setresgid", "getresgid",
	"setfsuid", "setfsgid",
	"getgroups", "setgroups",
	"setpgid", "getpgrp", "getpgid", "setsid", "getsid",
	"capget", "capset",
	"prctl", "arch_prctl",
	"uname",
	"getrlimit", "setrlimit", "prlimit64",
	"getrusage", "sysinfo", "times",
	"getpriority", "setpriority",
	"sched_yield",
	"sched_setparam", "sched_getparam",
	"sched_setscheduler", "sched_getscheduler",
	"sched_get_priority_max", "sched_get_priority_min",
	"sched_rr_get_interval",
	"sched_setaffinity", "sched_getaffinity",
	"sched_setattr", "sched_getattr",
	"getcpu",
	"personality",

	// Signals
	"kill", "tkill", "tgkill",
	"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",
	"rt_sigpending", "rt_sigtimedwait", "rt_sigqueueinfo",
	"rt_sigsuspend", "rt_tgsigqueueinfo",
	"sigaltstack",
	"pause",

	// Networking
	"socket", "connect", "accept", "accept4",
	"sendto", "recvfrom", "sendmsg", "recvmsg",
	"sendmmsg", "recvmmsg",
	"shutdown", "bind", "listen",
	"getsockname", "getpeername",
	"socketpair",
	"setsockopt", "getsockopt",

	// IPC
	"pipe", "pipe2",
	"dup", "dup2", "dup3",
	"shmget", "shmat", "shmdt", "shmctl",
	"semget", "semop", "semctl", "semtimedop",
	"msgget", "msgsnd", "msgrcv", "msgctl",
	"mq_open", "mq_unlink", "mq_timedsend", "mq_timedreceive",
	"mq_notify", "mq_getsetattr",

	// Polling / events
	"select", "pselect6", "poll", "ppoll",
	"epoll_create", "epoll_create1",
	"epoll_ctl", "epoll_wait", "epoll_pwait",
	"eventfd", "eventfd2",
	"signalfd", "signalfd4",
	"timerfd_create", "timerfd_settime", "timerfd_gettime",

	// Time
	"gettimeofday", "time",
	"clock_gettime", "clock_getres", "clock_nanosleep",
	"nanosleep",
	"getitimer", "setitimer", "alarm",
	"timer_create", "timer_settime", "timer_gettime",
	"timer_getoverrun", "timer_delete",
	"utime", "utimes", "utimensat", "futimesat",

	// I/O multiplexing
	"ioctl",
	"io_setup", "io_destroy", "io_getevents", "io_submit", "io_cancel",
	"preadv", "pwritev", "preadv2", "pwritev2",
	"inotify_init", "inotify_init1",
	"inotify_add_watch", "inotify_rm_watch",
	"fanotify_init", "fanotify_mark",

	// Extended attributes
	"setxattr", "lsetxattr", "fsetxattr",
	"getxattr", "lgetxattr", "fgetxattr",
	"listxattr", "llistxattr", "flistxattr",
	"removexattr", "lremovexattr", "fremovexattr",

	// Miscellaneous safe syscalls
	"getrandom",
	"restart_syscall",
	"set_tid_address",
	"set_robust_list", "get_robust_list",
	"futex",
	"set_thread_area", "get_thread_area",
	"seccomp",
	"mbind", "set_mempolicy", "get_mempolicy",
	"ioprio_set", "ioprio_get",
	"rseq",
	"mknod", "mknodat",
}

// GetEssentialSyscalls returns a copy of the essential syscalls list.
// Used by the runtime security_profile package for "strict" mode fallback.
func GetEssentialSyscalls() []string {
	result := make([]string, len(essentialSyscalls))
	copy(result, essentialSyscalls)
	return result
}

// GetDangerousSyscalls returns a copy of the dangerous syscalls list.
func GetDangerousSyscalls() []string {
	result := make([]string, len(dangerousSyscalls))
	copy(result, dangerousSyscalls)
	return result
}

// DefaultSeccompProfile returns a default restrictive seccomp profile.
// The profile uses a deny-by-default action (SCMP_ACT_ERRNO) and
// explicitly allows only the syscalls needed for containerized workloads.
// Dangerous syscalls are explicitly killed with SCMP_ACT_KILL to
// prevent any accidental override.
func DefaultSeccompProfile() *SeccompProfile {
	return &SeccompProfile{
		DefaultAction: "SCMP_ACT_ERRNO",
		Architectures: []string{
			"SCMP_ARCH_X86_64",
			"SCMP_ARCH_X86",
			"SCMP_ARCH_X32",
			"SCMP_ARCH_AARCH64",
		},
		Syscalls: []SyscallRule{
			{
				Names:  essentialSyscalls,
				Action: "SCMP_ACT_ALLOW",
			},
			{
				Names:  dangerousSyscalls,
				Action: "SCMP_ACT_KILL",
			},
		},
	}
}

// ValidateSeccompProfile checks that a seccomp profile has a valid
// deny-by-default configuration and does not allow any dangerous syscalls.
func ValidateSeccompProfile(profile *SeccompProfile) error {
	if profile == nil {
		return fmt.Errorf("seccomp profile is nil")
	}

	// Must have a deny-by-default action
	if profile.DefaultAction != "SCMP_ACT_ERRNO" && profile.DefaultAction != "SCMP_ACT_KILL" {
		return fmt.Errorf("seccomp profile must have deny-by-default action (SCMP_ACT_ERRNO or SCMP_ACT_KILL), got %s", profile.DefaultAction)
	}

	if len(profile.Architectures) == 0 {
		return fmt.Errorf("seccomp profile must specify at least one architecture")
	}

	if len(profile.Syscalls) == 0 {
		return fmt.Errorf("seccomp profile must have at least one syscall rule")
	}

	// Build set of dangerous syscalls for fast lookup
	dangerous := make(map[string]bool, len(dangerousSyscalls))
	for _, s := range dangerousSyscalls {
		dangerous[s] = true
	}

	// Check that no allow rule includes a dangerous syscall
	for _, rule := range profile.Syscalls {
		if rule.Action == "SCMP_ACT_ALLOW" {
			for _, name := range rule.Names {
				if dangerous[name] {
					return fmt.Errorf("seccomp profile allows dangerous syscall: %s", name)
				}
			}
		}
	}

	return nil
}

// AllowedSyscalls returns a sorted list of all syscall names that are
// explicitly allowed in the profile.
func (p *SeccompProfile) AllowedSyscalls() []string {
	var allowed []string
	for _, rule := range p.Syscalls {
		if rule.Action == "SCMP_ACT_ALLOW" {
			allowed = append(allowed, rule.Names...)
		}
	}
	sort.Strings(allowed)
	return allowed
}

// DeniedSyscalls returns a sorted list of all syscall names that are
// explicitly denied (KILL) in the profile.
func (p *SeccompProfile) DeniedSyscalls() []string {
	var denied []string
	for _, rule := range p.Syscalls {
		if rule.Action == "SCMP_ACT_KILL" {
			denied = append(denied, rule.Names...)
		}
	}
	sort.Strings(denied)
	return denied
}

// LoadSeccompProfile loads a seccomp profile from a file
func LoadSeccompProfile(path string) (*SeccompProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read seccomp profile: %w", err)
	}

	var profile SeccompProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse seccomp profile: %w", err)
	}

	return &profile, nil
}

// SaveSeccompProfile saves a seccomp profile to a file
func SaveSeccompProfile(profile *SeccompProfile, path string) error {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal seccomp profile: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write seccomp profile: %w", err)
	}

	return nil
}
