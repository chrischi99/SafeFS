package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	_ "encoding/hex"
	_ "encoding/json"
	_ "errors"
	"reflect"
	"strconv"
	_ "strings"
	"testing"

	"github.com/cs161-staff/userlib"
	_ "github.com/google/uuid"
)

func TestSome(t *testing.T) {
	someUsefulThings()
}

func clear() {
	// Wipes the storage so one test does not affect another
	userlib.DatastoreClear()
	userlib.KeystoreClear()
}

func TestInit(t *testing.T) {
	clear()
	t.Log("Initialization test")

	someUsefulThings()

	// You can set this to false!
	userlib.SetDebugStatus(true)

	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	// t.Log() only produces output if you run with "go test -v"
	t.Log("Got user", u)
	// If you want to comment the line above,
	// write _ = u here to make the compiler happy
	// You probably want many more tests here.

	// Test Init existent user
	u, err = InitUser("alice", "fubar")
	if err == nil {
		// t.Error says the test fails
		t.Error("Able to initialize existent user", err)
		return
	}

}

func TestGet(t *testing.T) {
	clear()
	t.Log("Initialization test")

	someUsefulThings()

	// You can set this to false!
	userlib.SetDebugStatus(true)

	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	// t.Log() only produces output if you run with "go test -v"
	t.Log("Got user", u)

	// Get non-exist user
	_, err = GetUser("ABC", "BCA")
	if err == nil {
		t.Error("Failed to verify whether user exists")
		return
	}

	// Get from multiple instances
	u2, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to get user", err)
		return
	}
	if u2 == nil {
		t.Error("Failed to reget user", err)
		return
	}

	u3, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to get user", err)
		return
	}
	if u3 == nil {
		t.Error("Failed to reget user", err)
		return
	}

	// Test invalid passwd
	_, err = GetUser("alice", "aabb")
	if err == nil {
		t.Error("Failed to validate user", err)
		return
	}
}

func TestStorage(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestAppend(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	for i := 0; i < 100; i++ {
		u.AppendFile("file1", []byte(strconv.Itoa(i)))
		v = append(v, []byte(strconv.Itoa(i))...)
	}

	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v2, v) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}

}

func TestMultipleInstancesStorage(t *testing.T) {
	clear()
	u1, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	u2, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test")
	u1.StoreFile("file1", v)

	v2, err := u2.LoadFile("file1")
	if err != nil {
		t.Error("Failed to upload and download", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}

	err = u1.AppendFile("file1", []byte("hello world"))
	if err != nil {
		t.Error("Append failed", err)
		return
	}
	v3, err := u1.LoadFile("file1")
	if !reflect.DeepEqual(v3, []byte("This is a testhello world")) {
		t.Error("Downloaded file is not the same, append failed", string(v), string(v2))
		return
	}
	v2, err = u2.LoadFile("file1")
	if err != nil {
		t.Error("Failed to upload and download", err)
		return
	}
	if !reflect.DeepEqual(v2, []byte("This is a testhello world")) {
		t.Error("Downloaded file is not the same", string(v), string(v2))
		return
	}
}

func TestInvalidFile(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	_, err2 := u.LoadFile("this file does not exist")
	if err2 == nil {
		t.Error("Downloaded a ninexistent file", err2)
		return
	}
}

func TestShare(t *testing.T) {
	clear()
	userlib.SetDebugStatus(true)
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	u2.StoreFile("file2", []byte("ABC"))
	v, err = u.LoadFile("file1")
	if !reflect.DeepEqual(v, []byte("ABC")) {
		t.Error("Shared file is not the same", "ABC", string(v))
		return
	}
	v2, err = u2.LoadFile("file2")
	if !reflect.DeepEqual(v2, []byte("ABC")) {
		t.Error("Shared file is not the same", []byte("ABC"), v2)
		return
	}

}

func TestRevoke(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	err = u2.AppendFile("file2", []byte("helloworld"))
	if err != nil {
		t.Error("Shared user failed to append data")
		return
	}

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to load the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, []byte("This is a testhelloworld")) {
		t.Error("Sharing user cannot observe the shard user's append data")
		return
	}

	userlib.SetDebugStatus(true)
	u.RevokeFile("file1", "bob")
	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to load the file after revoking", err)
		return
	}
	userlib.SetDebugStatus(false)

	err = u2.AppendFile("file2", []byte("helloworld"))
	if err != nil {
		t.Error("Shared user failed to append data")
		return
	}
	if reflect.DeepEqual(v, []byte("This is a testhelloworldhelloworld")) {
		t.Error("Revoked user can still append data")
		return
	}

}
