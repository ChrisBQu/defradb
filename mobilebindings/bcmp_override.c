// bcmp_wrapper.c
#include <string.h>
#include <errno.h>
#include <spawn.h>
#include <unistd.h>
#include <string.h>
#include <sys/uio.h>
#include <resolv.h>

int bcmp(const void *s1, const void *s2, size_t n) {
    return memcmp(s1, s2, n);
}

// Provides a pointer to thread-local errno
int *__errno_location(void) {
    return &errno;
}

// Stub implementations for posix_spawn functions
int posix_spawnattr_destroy(posix_spawnattr_t *attr) {
    return 0;
}

int posix_spawn_file_actions_destroy(posix_spawn_file_actions_t *actions) {
    return 0;
}

int posix_spawnattr_init(posix_spawnattr_t *attr) {
    return 0;
}

int posix_spawn_file_actions_init(posix_spawn_file_actions_t *actions) {
    return 0;
}

int posix_spawn_file_actions_adddup2(posix_spawn_file_actions_t *actions, int fd, int newfd) {
    return 0;
}

int posix_spawnattr_setpgroup(posix_spawnattr_t *attr, pid_t pgroup) {
    return 0;
}

int posix_spawnattr_setflags(posix_spawnattr_t *attr, short flags) {
    return 0;
}

int posix_spawnattr_setsigdefault(posix_spawnattr_t *attr, const sigset_t *sigdefault) {
    return 0;
}

// Simple stub for gnu_get_libc_version
const char *gnu_get_libc_version(void) {
    return "glibc 2.31";
}

// Simple wrappers for preadv and pwritev using pread and pwrite
ssize_t preadv(int fd, const struct iovec *iov, int iovcnt, off_t offset) {
    return pread(fd, iov->iov_base, iov->iov_len, offset);
}

ssize_t pwritev(int fd, const struct iovec *iov, int iovcnt, off_t offset) {
    return pwrite(fd, iov->iov_base, iov->iov_len, offset);
}

// Basic stub for __res_init
int __res_init(void) {
    return 0;
}

// Minimal implementation of posix_spawnp using fork + exec
int posix_spawnp(pid_t *pid, const char *file, const posix_spawn_file_actions_t *actions,
                 const posix_spawnattr_t *attr, char *const argv[], char *const envp[]) {
    pid_t child_pid = fork();
    if (child_pid == 0) {
        execvp(file, argv);
        _exit(127); // exec failed
    } else if (child_pid < 0) {
        return errno;
    }
    *pid = child_pid;
    return 0;
}

/* Provide a simple __xpg_strerror_r implementation */
char *__xpg_strerror_r(int errnum, char *buf, size_t buflen) {
    /* POSIX strerror_r returns an int, but GNU version returns a char* */
    if (strerror_r(errnum, buf, buflen) == 0) {
        return buf;
    } else {
        return "Unknown error";
    }
}