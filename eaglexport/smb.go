package eaglexport

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hirochachacha/go-smb2"
	"github.com/spf13/afero"
)

type SmbFs struct {
	Conn    net.Conn
	Session *smb2.Session
	Share   *smb2.Share
}

type SmbFsOption struct {
	User     string
	Password string
}

func NewSmbFs(address string, shareName string, option SmbFsOption) (*SmbFs, error) {
	if !strings.Contains(address, ":") {
		address += ":445"
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, errors.Wrap(err, "net dial failed")
	}

	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     option.User,
			Password: option.Password,
		},
	}
	session, err := dialer.Dial(conn)
	if err != nil {
		return nil, errors.Wrap(err, "dial failed")
	}

	share, err := session.Mount(shareName)
	if err != nil {
		return nil, errors.Wrap(err, "mount failed")
	}

	return &SmbFs{
		Conn:    conn,
		Session: session,
		Share:   share,
	}, nil
}

func (s *SmbFs) Close() error {
	err := s.Share.Umount()
	if err != nil {
		return errors.Wrap(err, "umount failed")
	}

	err = s.Session.Logoff()
	if err != nil {
		return errors.Wrap(err, "log off failed")
	}

	return nil
}

func (s *SmbFs) Create(name string) (afero.File, error) {
	return s.Share.Create(name)
}

func (s *SmbFs) Mkdir(name string, perm os.FileMode) error {
	return s.Share.Mkdir(name, perm)
}

func (s *SmbFs) MkdirAll(path string, perm os.FileMode) error {
	return s.Share.MkdirAll(path, perm)
}

func (s *SmbFs) Open(name string) (afero.File, error) {
	return s.Share.Open(name)
}

func (s *SmbFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return s.Share.OpenFile(name, flag, perm)
}

func (s *SmbFs) Remove(name string) error {
	return s.Share.Remove(name)
}

func (s *SmbFs) RemoveAll(path string) error {
	return s.Share.RemoveAll(path)
}

func (s *SmbFs) Rename(oldname, newname string) error {
	return s.Share.Rename(oldname, newname)
}

func (s *SmbFs) Stat(name string) (os.FileInfo, error) {
	return s.Share.Stat(name)
}

func (s *SmbFs) Name() string {
	return "smbfs"
}

func (s *SmbFs) Chmod(name string, mode os.FileMode) error {
	return s.Share.Chmod(name, mode)
}

func (s *SmbFs) Chown(name string, uid, gid int) error {
	_ = name
	_ = uid
	_ = gid
	return errors.New("not supported")
}

func (s *SmbFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.Share.Chtimes(name, atime, mtime)
}
