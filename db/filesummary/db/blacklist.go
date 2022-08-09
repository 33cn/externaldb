package db

import "github.com/33cn/externaldb/db"

type blackFile FileSummary

type BlackFile interface {
	Black(op int, hash, note string)
	Recover(op int, hash, note string)
}

func NewBlackFile(f *FileSummary) BlackFile {
	return (*blackFile)(f)
}

func (f *blackFile) Black(op int, hash, note string) {
	if op == db.SeqTypeAdd {
		f.FileBlacklist = hash
		f.FileBlacklistNote = note
		f.FileBlacklistFlag = true
	} else {
		f.FileBlacklist = ""
		f.FileBlacklistNote = ""
		f.FileBlacklistFlag = false
	}
}

func (f *blackFile) Recover(op int, hash, note string) {
	if op == db.SeqTypeAdd {
		f.FileBlacklist = ""
		f.FileBlacklistNote = ""
		f.FileBlacklistFlag = false
	} else {
		f.FileBlacklist = note
		f.FileBlacklistNote = hash
		f.FileBlacklistFlag = true
	}
}
