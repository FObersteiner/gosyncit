package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FObersteiner/gosyncit/cmd"
	"github.com/FObersteiner/gosyncit/lib/fileset"
)

func TestSync(t *testing.T) {
	dry := false
	ignorehidden := false

	err := cmd.Sync("A", "B", dry, ignorehidden)
	if err == nil {
		t.Fail()
		t.Log("sync must fail with invalid src/dst input")
	}

	// sync functionality...

	src, err := os.MkdirTemp("", "src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(src)

	dst, err := os.MkdirTemp("", "dst")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dst)

	// --- SRC ---
	// file0 is younger in source
	file0 := filepath.Join(src, "file0")
	if err := os.WriteFile(file0, []byte("content_src"), 0655); err != nil {
		t.Fatal(err)
	}
	mtime := time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file0, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	// file1 is older in source
	file1 := filepath.Join(src, "file1")
	if err := os.WriteFile(file1, []byte("content_src"), 0655); err != nil {
		t.Fatal(err)
	}
	mtime = time.Date(2006, time.March, 1, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file1, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	// file2 is equally old in both src and dst. content of src should take over.
	file2 := filepath.Join(src, "file2")
	// file2 --> write one more byte so simple comparison (n bytes) says 'different'
	if err := os.WriteFile(file2, []byte("content_src_"), 0655); err != nil {
		t.Fatal(err)
	}
	mtime = time.Date(2006, time.March, 1, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file2, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	// file_src_only does not exist in dst
	file_src_only := filepath.Join(src, "file_src_only")
	if err := os.WriteFile(file_src_only, []byte("content_src"), 0655); err != nil {
		t.Fatal(err)
	}
	mtime = time.Date(2006, time.March, 1, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file_src_only, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	// --- DST ---
	// file_dst_only does not exist in src
	file_dst_only := filepath.Join(dst, "file_dst_only")
	if err := os.WriteFile(file_dst_only, []byte("content_dst"), 0655); err != nil {
		t.Fatal(err)
	}
	mtime = time.Date(2006, time.March, 1, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file_dst_only, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	// file0 is older in dst
	file0 = filepath.Join(dst, "file0")
	if err := os.WriteFile(file0, []byte("content_dst"), 0655); err != nil {
		t.Fatal(err)
	}
	mtime = time.Date(2006, time.January, 1, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file0, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	// file1 is younger in dst
	file1 = filepath.Join(dst, "file1")
	if err := os.WriteFile(file1, []byte("content_dst"), 0655); err != nil {
		t.Fatal(err)
	}
	mtime = time.Date(2006, time.April, 1, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file1, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	// file2 is equally old in both src and dst. content of src should take over.
	file2 = filepath.Join(dst, "file2")
	if err := os.WriteFile(file2, []byte("content_dst"), 0655); err != nil {
		t.Fatal(err)
	}
	mtime = time.Date(2006, time.March, 1, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file2, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	// --- sync call ---
	// log.Println(src, dst)
	if err := cmd.Sync(src, dst, dry, ignorehidden); err != nil {
		t.Fatal(err)
	}

	// --- want after sync ---
	fs_src, _ := fileset.New(src)
	_ = fs_src.Populate()
	fs_dst, _ := fileset.New(dst)
	_ = fs_dst.Populate()

	// file_dst_only must be in src:
	if ok := fs_src.Contains("file_dst_only"); !ok {
		t.Log("'file_dst_only' must be in src")
		t.Fail()
	}
	if ok := fs_dst.Contains("file_src_only"); !ok {
		t.Log("'file_src_only' must be in dst")
		t.Fail()
	}

	file0_dst_content, _ := os.ReadFile(filepath.Join(fs_dst.Basepath, "file0"))
	if !bytes.Equal(file0_dst_content, []byte("content_src")) {
		t.Log("file0 is younger in src, so dst must contain content of src")
		t.Fail()
	}
	file1_src_content, _ := os.ReadFile(filepath.Join(fs_src.Basepath, "file1"))
	if !bytes.Equal(file1_src_content, []byte("content_dst")) {
		t.Log("file1 is younger in dst, so src must contain content of dst")
		t.Fail()
	}
	// file2_src_content, _ := os.ReadFile(filepath.Join(fs_src.Basepath, "file2"))
	file2_content_dst, _ := os.ReadFile(filepath.Join(fs_dst.Basepath, "file2"))
	if !bytes.Equal(file2_content_dst, []byte("content_src_")) {
		t.Log("content of src must be used if file names and mtime are equal")
		t.Fail()
	}
}
