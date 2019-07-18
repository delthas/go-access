/*
Package access has simple functions for checking whether a *nix user has the
permissions to access a file (or folder).
*/
package access

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Read permission (r)
const Read = os.FileMode(4)

// Write permission (w)
const Write = os.FileMode(2)

// Execute permission (x)
const Execute = os.FileMode(1)

// ErrNoProgress is returned by some clients of an io.Reader when
// many calls to Read have failed to return any data or error,
// usually the sign of a broken io.Reader implementation.

// PermissionError is returned by Uid and Username when a user
// does not have sufficient permissions to access the requested file or folder.
//
// It contains data about the file for which the permission check failed,
// which can be different from the requested file.
type PermissionError struct {
	// path of the file/folder for which the permission check failed
	File string
	// permissions of the file
	FileMode os.FileMode
	// uid of the file
	FileUid int
	// gid of the file
	FileGid int
	// uid of the user whose permission is checked
	Uid int
	// gid of the premiary and all secondary groups of the user whose permission is checked
	Gid []int
	// permissions requested for the file (can be different from the one requested in Uid or Username)
	WantMode os.FileMode
}

func (p *PermissionError) Error() string {
	return fmt.Sprintf("unsufficient permissions of user (uid %d, gid %d) for file [%s] (uid %d, gid %d): want mode %o, file has mode %o", p.Uid, p.Gid, p.File, p.FileUid, p.FileGid, p.WantMode, p.FileMode)
}

// Uid checks whether a user has the permissions to access a file.
//
// - uid is the *nix uid of the user
//
// - mode is the requested permission on the file, for example Read, Write, and/or Execute
//
// - path is the path of the file/folder
//
// - returns a PermissionError if the user does not have access the requested access to the file
//
// - returns a non-nil error if the user does not exist (in which case the returned
//   error is a UnknownUserIdError), or if an underlying error occurs when reading permissions
//
// - if the error is nil, the user has the requested access to the file
func Uid(uid int, mode os.FileMode, path string) error {
	u, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return err
	}
	return access(u, uid, mode, path)
}

// Username checks whether a user has the permissions to access a file.
//
// - username is the *nix username of the user
//
// - mode is the requested permission on the file, for example Read, Write, and/or Execute
//
// - path is the path of the file/folder
//
// - returns a PermissionError if the user does not have access the requested access to the file
//
// - returns a non-nil error if the user does not exist (in which case the returned
//   error is a UnknownUserIdError), or if an underlying error occurs when reading permissions
//
// - if the error is nil, the user has the requested access to the file
func Username(username string, mode os.FileMode, path string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	return access(u, uid, mode, path)
}

func contains(a []int, i int) bool {
	for _, e := range a {
		if e == i {
			return true
		}
	}
	return false
}

// path is absolute, contains no . or ..
func checkPath(uid int, gid []int, mode os.FileMode, path string) error {
	for len(path) > 0 {
		fi, err := os.Lstat(path)
		if err != nil {
			return err
		}
		fm := fi.Mode()
		s := fi.Sys().(*syscall.Stat_t)

		if uid != 0 && fm&mode != mode && (fm&(mode<<6) != mode<<6 || uint32(uid) != s.Uid) && (fm&(mode<<3) != mode<<3 || !contains(gid, int(s.Gid))) {
			return &PermissionError{
				File:     path,
				FileMode: fm,
				FileUid:  int(s.Uid),
				FileGid:  int(s.Gid),
				Uid:      uid,
				Gid:      gid,
				WantMode: mode,
			}
		}
		mode = 1 // x

		i := strings.LastIndexFunc(path, func(r rune) bool {
			return r == os.PathSeparator
		})
		if i < 0 { // should never happen
			return errors.New("absolute path not containing any slash: " + path)
		}
		path = path[:i]
	}
	return nil
}

func access(user *user.User, uid int, mode os.FileMode, path string) error {
	gs, err := user.GroupIds()
	if err != nil {
		return err
	}
	gi := make([]int, len(gs))
	for i, g := range gs {
		gi[i], err = strconv.Atoi(g)
		if err != nil {
			return err
		}
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	// some code adapted from filepath.walkSymlinks

	volLen := 0
	pathSeparator := string(os.PathSeparator)

	if volLen < len(path) && os.IsPathSeparator(path[volLen]) {
		volLen++
	}
	vol := path[:volLen]
	dest := vol
	linksWalked := 0
	for start, end := volLen, volLen; start < len(path); start = end {
		for start < len(path) && os.IsPathSeparator(path[start]) {
			start++
		}
		end = start
		for end < len(path) && !os.IsPathSeparator(path[end]) {
			end++
		}
		// The next path component is in path[start:end].
		if end == start {
			// No more path components.
			break
		} else if path[start:end] == "." {
			// Ignore path component ".".
			continue
		} else if path[start:end] == ".." {
			// Back up to previous component if possible.
			// Note that volLen includes any leading slash.

			// Set r to the index of the last slash in dest,
			// after the volume.
			var r int
			for r = len(dest) - 1; r >= volLen; r-- {
				if os.IsPathSeparator(dest[r]) {
					break
				}
			}
			if r < volLen || dest[r+1:] == ".." {
				// Either path has no slashes
				// (it's empty or just "C:")
				// or it ends in a ".." we had to keep.
				// Either way, keep this "..".
				if len(dest) > volLen {
					dest += pathSeparator
				}
				dest += ".."
			} else {
				// Discard everything since the last slash.
				dest = dest[:r]
			}
			continue
		}

		// Ordinary path component. Add it to result.

		if len(dest) > 0 && !os.IsPathSeparator(dest[len(dest)-1]) {
			dest += pathSeparator
		}

		l := len(dest)

		dest += path[start:end]

		// Check perms on symlink.

		if err := checkPath(uid, gi, 1, dest[:l]); err != nil {
			return err
		}

		// Resolve symlink.

		fi, err := os.Lstat(dest)
		if err != nil {
			return err
		}

		if fi.Mode()&os.ModeSymlink == 0 {
			if !fi.Mode().IsDir() && end < len(path) {
				return syscall.ENOTDIR
			}
			continue
		}

		// Found symlink.

		linksWalked++
		if linksWalked > 255 {
			return errors.New("access: too many links")
		}

		link, err := os.Readlink(dest)
		if err != nil {
			return err
		}

		path = link + path[end:]

		if len(link) > 0 && os.IsPathSeparator(link[0]) {
			// Symlink to absolute path.
			dest = link[:1]
			end = 1
		} else {
			// Symlink to relative path; replace last
			// path component in dest.
			var r int
			for r = len(dest) - 1; r >= volLen; r-- {
				if os.IsPathSeparator(dest[r]) {
					break
				}
			}
			if r < volLen {
				dest = vol
			} else {
				dest = dest[:r]
			}
			end = 0
		}
	}

	// all symlinks resolved, check access on final path
	if err := checkPath(uid, gi, mode, dest); err != nil {
		return err
	}

	return nil
}
