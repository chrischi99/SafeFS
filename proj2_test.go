package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	_ "encoding/hex"
	_ "encoding/json"
	"errors"
	_ "errors"
	"reflect"
	"strconv"
	_ "strings"
	"testing"

	"github.com/cs161-staff/userlib"
	"github.com/google/uuid"
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

func TestBadInit(t *testing.T) {
	clear()
	t.Log("Bad Init")
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	ds := userlib.DatastoreGetMap()
	for k, v := range ds {
		v = append(v, v...)
		userlib.DatastoreSet(k, v)
	}
	_, e := GetUser("alice", "fubar")
	if e == nil {
		t.Error("Failed to detec verification violation")
	}
}

func TestBadInit2(t *testing.T) {
	clear()
	t.Log("Bad Init")
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	ds := userlib.DatastoreGetMap()
	for k := range ds {
		userlib.DatastoreDelete(k)
	}
	_, e := GetUser("alice", "fubar")
	if e == nil {
		t.Error("Failed to detec verification violation")
	}
}

func TestDuplicateInit(t *testing.T) {
	clear()
	t.Log("DuplicationInit test")
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	_, err = InitUser("alice", "fubar2")
	if err == nil {
		t.Error("Failed to prevent initialize dupilcate user", err)
		return
	}
}

func TestAttackerInitUser(t *testing.T) {
	clear()
	t.Log("DuplicationInit test")
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	userlib.DatastoreClear()
	_, err = InitUser("alice", "fubar")
	if err == nil {
		// t.Error says the test fails
		t.Error("Attack able to delete then initialize user", err)
		return
	}

	_, err = InitUser("Bob", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	dsmap := userlib.DatastoreGetMap()
	for k := range dsmap {
		userlib.DatastoreSet(k, []byte("AABBCC"))
	}

	_, err = InitUser("alice", "fubar")
	if err == nil {
		// t.Error says the test fails
		t.Error("Attack able to delete then initialize user", err)
		return
	}
}

func TestGet(t *testing.T) {
	clear()
	t.Log("Get test")

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

	f := []byte("This is a test")
	u.StoreFile("file1", f)

	v, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, f) {
		t.Error("Downloaded file is not the same", v, f)
		return
	}

	f = []byte("This is another test")
	u.StoreFile("file1", f)
	v, err2 = u.LoadFile("file1")

	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, f) {
		t.Error("Downloaded file is not the same", v, f)
		return
	}

}

func TestStorageSameFileName(t *testing.T) {
	clear()
	u1, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	u2, err := InitUser("Bob", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	f1 := []byte("This is a test")
	u1.StoreFile("file1", f1)
	f2 := []byte("This is another test")
	u2.StoreFile("file1", f2)

	v1, e := u1.LoadFile("file1")
	if e != nil || !reflect.DeepEqual(v1, f1) {
		t.Error("Failed to load file or content wrong", e)
		return
	}

	v2, e := u2.LoadFile("file1")
	if e != nil || !reflect.DeepEqual(v2, f2) {
		t.Error("Failed to load file or content wrong", e)
		return
	}
}

func userMetadata() (uuids map[uuid.UUID][]byte) {
	uuids = make(map[uuid.UUID][]byte)
	dmap := userlib.DatastoreGetMap()
	for k, v := range dmap {
		uuids[k] = v
	}
	return
}

func userFileData(userMetaUuids map[uuid.UUID][]byte) (uuids map[uuid.UUID][]byte) {
	uuids = make(map[uuid.UUID][]byte)
	dmap := userlib.DatastoreGetMap()
	for k, v := range dmap {
		if _, ok := userMetaUuids[k]; !ok {
			uuids[k] = v
		}
	}
	return
}

func appendData(u1 map[uuid.UUID][]byte, u2 map[uuid.UUID][]byte) (uuids map[uuid.UUID][]byte) {
	uuids = make(map[uuid.UUID][]byte)
	dmap := userlib.DatastoreGetMap()
	for k, v := range dmap {
		if _, ok := u1[k]; !ok {
			if _, ok := u2[k]; !ok {
				uuids[k] = v
			}
		}
	}
	return
}

func sharingData(u1 map[uuid.UUID][]byte, u2 map[uuid.UUID][]byte) (uuids map[uuid.UUID][]byte) {
	uuids = make(map[uuid.UUID][]byte)
	dmap := userlib.DatastoreGetMap()
	for k, v := range dmap {
		if _, ok := u1[k]; !ok {
			if _, ok := u2[k]; !ok {
				uuids[k] = v
			}
		}
	}
	return
}

func TestAppend(t *testing.T) {
	clear()
	userlib.SetDebugStatus(false)
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	umeta := userMetadata()

	t.Log("umeta key")
	for k := range umeta {
		t.Log(k.String())
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	fmeta := userFileData(umeta)

	t.Log("fmeta key")
	for k := range fmeta {
		t.Log(k.String())
	}

	// modify file data
	for k, v := range fmeta {
		v = append(v, v...)
		userlib.DatastoreSet(k, v)
	}

	// var uv DataDS
	// json.Unmarshal(v, &uv)
	// uv.Ciphertext = append(uv.Ciphertext, uv.Ciphertext...)
	// v, _ = json.Marshal(uv)
	// userlib.DatastoreSet(k, v)
	_, err2 := u.LoadFile("file1")
	if err2 == nil {
		t.Error("failed to upload and download", err2)
		return
	}

	// modify append data
	u.StoreFile("file2", v)
	fmeta = userFileData(umeta)

	t.Log("fmeta 2")
	for k := range fmeta {
		t.Log(k.String())
	}

	for i := 0; i < 100; i++ {
		u.AppendFile("file2", []byte(strconv.Itoa(i)))
		v = append(v, []byte(strconv.Itoa(i))...)
	}
	v2, err2 := u.LoadFile("file2")
	if err2 != nil {
		t.Error("failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v2, v) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}

	adata := appendData(umeta, fmeta)
	t.Log("append dta")
	for k, v := range adata {
		t.Log(k.String())
		v = append(v, v...)
		userlib.DatastoreSet(k, v)
	}

	_, err2 = u.LoadFile("file1")
	if err2 == nil {
		t.Error("failed to upload and download", err2)
		return
	}

}
func TestAppend2(t *testing.T) {
	clear()
	userlib.SetDebugStatus(false)
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	umeta := userMetadata()

	t.Log("umeta key")
	for k := range umeta {
		t.Log(k.String())
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	fmeta := userFileData(umeta)

	t.Log("fmeta key")
	for k := range fmeta {
		t.Log(k.String())
	}

	// modify file data
	for k := range fmeta {
		userlib.DatastoreDelete(k)
	}

	_, err2 := u.LoadFile("file1")
	if err2 == nil {
		t.Error("failed to upload and download", err2)
		return
	}

	// modify append data
	u.StoreFile("file2", v)
	fmeta = userFileData(umeta)

	t.Log("fmeta 2")
	for k := range fmeta {
		t.Log(k.String())
	}

	for i := 0; i < 100; i++ {
		u.AppendFile("file2", []byte(strconv.Itoa(i)))
		v = append(v, []byte(strconv.Itoa(i))...)
	}
	v2, err2 := u.LoadFile("file2")
	if err2 != nil {
		t.Error("failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v2, v) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}

	adata := appendData(umeta, fmeta)
	t.Log("append dta")
	for k := range adata {
		userlib.DatastoreDelete(k)
	}
	u.StoreFile("file2", v)
	_, err2 = u.LoadFile("file1")
	if err2 == nil {
		t.Error("failed to upload and download", err2)
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
	udata := userMetadata()
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	fdata := userFileData(udata)

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
	sharedata := sharingData(udata, fdata)

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

	// modify metadata then try to share

	for k, v := range fdata {
		v = append(v, v...)
		userlib.DatastoreSet(k, v)
	}

	_, e := u.ShareFile("file1", "bob")

	if e == nil {
		t.Error("Able to share after file metada broke")
		return
	}

	// modify sharedata
	for k, v := range sharedata {
		v = append(v, v...)
		userlib.DatastoreSet(k, v)
	}
	u2.StoreFile("file2", v)
	v, err = u2.LoadFile("file2")
	if err == nil {
		t.Error("Fail to detect malicious modification", err)
		return
	}

}

func TestShare2(t *testing.T) {
	clear()
	userlib.SetDebugStatus(false)
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	udata := userMetadata()
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	fdata := userFileData(udata)

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
	sharedata := sharingData(udata, fdata)

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

	// modify metadata then try to share

	for k := range fdata {
		userlib.DatastoreDelete(k)
	}

	_, e := u.ShareFile("file1", "bob")

	if e == nil {
		t.Error("Able to share after file metada broke")
		return
	}

	// modify sharedata
	for k := range sharedata {
		t.Log(k)
		userlib.DatastoreDelete(k)
	}

	u2.StoreFile("file2", v)

	v, err = u2.LoadFile("file2")
	if err == nil {
		t.Error("Fail to detect malicious modification", err)
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

	u.RevokeFile("file1", "bob")
	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to load the file after revoking", err)
		return
	}

	err = u2.AppendFile("file2", []byte("helloworld"))
	if err == nil {
		v, e := u2.LoadFile("file2")
		if e != nil {
			t.Error("error", e)
			return
		}
		if !reflect.DeepEqual(v, []byte("This is a testhelloworldhelloworld")) {
			t.Error("Revoked user can still append data", v)
			return
		}
	}

}
func TestShareHierarchy(t *testing.T) {
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
	u3, err2 := InitUser("charles", "foodbar")
	if err2 != nil {
		t.Error("Failed to initialize charles", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var v3 []byte
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

	magic_string, err = u2.ShareFile("file2", "charles")
	if err != nil {
		t.Error("Failed to share the a file w/ charles", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string)
	if err != nil {
		t.Error("charles failed to receive the share message", err)
		return
	}

	v3, err = u3.LoadFile("file3")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v3) || !reflect.DeepEqual(v2, v3) {
		t.Error("Shared file is not the same", v, v3)
		return
	}
}

func TestShareHierarchyAppend(t *testing.T) {
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
	u3, err2 := InitUser("charles", "foodbar")
	if err2 != nil {
		t.Error("Failed to initialize charles", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	//var v3 []byte
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

	vAp1 := []byte(" olala")
	err = u2.AppendFile("file2", vAp1)
	if err != nil {
		t.Error("Failed to append to file2", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(append(v, vAp1...), v2) {
		t.Error("Shared file is not the same", string(append(v, vAp1...)), string(v2))
		return
	}

	magic_string, err = u2.ShareFile("file2", "charles")
	if err != nil {
		t.Error("Failed to share the a file w/ charles", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string)
	if err != nil {
		t.Error("charles failed to receive the share message", err)
		return
	}
}

func TestShareHierarchyStore(t *testing.T) {
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
	u3, err2 := InitUser("charles", "foodbar")
	if err2 != nil {
		t.Error("Failed to initialize charles", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	//var v3 []byte
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

	vS1 := []byte("bb")
	u2.StoreFile("file2", vS1)

	v, err = u.LoadFile("file1")
	v2, err = u2.LoadFile("file2")
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", string(v), string(v2))
		return
	}

	magic_string, err = u2.ShareFile("file2", "charles")
	if err != nil {
		t.Error("Failed to share the a file w/ charles", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string)
	if err != nil {
		t.Error("charles failed to receive the share message", err)
		return
	}

	v, err = u.LoadFile("file1")
	v2, err = u2.LoadFile("file2")
	v3, err := u3.LoadFile("file3")
	if !reflect.DeepEqual(v, v2) || !reflect.DeepEqual(v3, v2) {
		t.Error("Shared file is not the same", string(v), string(v3))
		return
	}
}

func spawnOneLevel(t *testing.T, name string) (users []*User) {
	for i := 0; i < 3; i++ {
		uname := name + "-" + strconv.Itoa(i)
		u, e := InitUser(uname, uname)
		if e != nil {
			t.Error("Init user " + name + " failed")
		}
		users = append(users, u)
	}
	return
}

func shareWithOneLevel(t *testing.T, parent *User, filename string, l1 []*User) error {
	for _, user := range l1 {
		at, e := parent.ShareFile(filename, user.Username)
		if e != nil {
			t.Error("Failed to share with " + user.Username)
			return e
		}
		e = user.ReceiveFile(filename+parent.Username, parent.Username, at)
		if e != nil {
			t.Error("Failed to receive at " + user.Username)
			return e
		}
	}
	return nil
}

func revokeWithOneLevel(t *testing.T, parent *User, filename string, l1 []*User) error {
	for _, user := range l1 {
		e := parent.RevokeFile(filename, user.Username)
		if e != nil {
			t.Error("Failed to share with " + user.Username)
			return e
		}
		// t.Log(parent.Username + " revoke " + user.Username + " " + filename)
	}
	return nil
}

func verifyFileOneLevel(t *testing.T, parent *User, filename string, equal bool, content []byte, l1 []*User) error {
	for _, user := range l1 {
		fname := filename + parent.Username
		f, e := user.LoadFile(filename + parent.Username)
		if e != nil && !equal {
			continue
		}
		if e != nil {
			t.Error(user.Username, ": Failed to Load file", fname, e)
			return e
		}
		if equal != reflect.DeepEqual(content, f) {
			t.Error(user.Username+": file content wrong:", string(f[:]), " v.s. ", string(content))
			return errors.New("bad content")
		}
	}
	return nil
}

func TestBadAccessToken(t *testing.T) {
	clear()

	t.Log("> Init user a")
	ua, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	t.Log("> Init user b")
	ub, err := InitUser("Bob", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	t.Log("usera store a new file abc")
	ua.StoreFile("abc", []byte("Hello world"))
	at, _ := ua.ShareFile("abc", ub.Username)

	at = at + "a"
	t.Log("userb try to receive bad magic string")
	if ub.ReceiveFile("c", ua.Username, at) == nil {
		t.Error("Failed to check validness of the magic string")
		return
	}
}

func TestReceiveAfterRevoke(t *testing.T) {
	clear()

	t.Log("> Init user a")
	ua, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	t.Log("> Init user b")
	ub, err := InitUser("Bob", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	t.Log("usera store a new file abc")
	ua.StoreFile("abc", []byte("Hello world"))
	at, _ := ua.ShareFile("abc", ub.Username)

	t.Log("usera revoke file abc")
	ua.RevokeFile("abc", ub.Username)

	t.Log("userb try receive shared file using old access token")
	ub.ReceiveFile("cc", ua.Username, at)

	ua.StoreFile("abc", []byte("Goodbye world"))
	v, _ := ub.LoadFile("cc")
	t.Log("userb get content", string(v))

	if reflect.DeepEqual(v, []byte("Goodbye world")) {
		t.Error("Able to get old content after revoke")
		return
	}

}

func TestRevokeSubTree(t *testing.T) {
	clear()
	userlib.SetDebugStatus(false)

	root, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	t.Log("> Init root")

	f := []byte("ABC")
	root.StoreFile("a", f)

	levelOne1 := spawnOneLevel(t, "level-1")
	t.Log("> share with level-1")
	if shareWithOneLevel(t, root, "a", levelOne1) != nil {
		return
	}

	levelOne2 := spawnOneLevel(t, "level-2")
	t.Log("> share with level-2")
	if shareWithOneLevel(t, root, "a", levelOne2) != nil {
		return
	}

	levelTwo1 := spawnOneLevel(t, "level-2.1")
	t.Log("> share with level-2.1")
	if shareWithOneLevel(t, levelOne2[0], "a"+root.Username, levelTwo1) != nil {
		return
	}

	t.Log("> verify levelOne1")
	if verifyFileOneLevel(t, root, "a", true, f, levelOne1) != nil {
		return
	}
	t.Log("> verify levelOne2")
	if verifyFileOneLevel(t, root, "a", true, f, levelOne2) != nil {
		return
	}
	t.Log("> verify levelTwo1")
	if verifyFileOneLevel(t, levelOne2[0], "a"+root.Username, true, f, levelTwo1) != nil {
		return
	}

	t.Log("> revoke levelOne1")
	revokeWithOneLevel(t, root, "a", levelOne1)

	t.Log("> change file content")
	root.AppendFile("a", f)
	f = append(f, f...)
	t.Log("> verify file content levelOne2")
	verifyFileOneLevel(t, root, "a", true, f, levelOne2)
	t.Log("> verify file content levelOne1")
	verifyFileOneLevel(t, root, "a", false, f, levelOne1)
	t.Log("> verify file content levelTwo1")
	verifyFileOneLevel(t, levelOne2[0], "a"+root.Username, true, f, levelTwo1)

	t.Log("> levelTwo1 append content")
	levelTwo1[0].AppendFile("a"+root.Username+levelOne2[0].Username, f)
	f = append(f, f...)
	t.Log("> verify file content levelOne2")
	verifyFileOneLevel(t, root, "a", true, f, levelOne2)
	t.Log("> verify file content levelOne1")
	verifyFileOneLevel(t, root, "a", false, f, levelOne1)

	t.Log("> levelTwo store new content")
	f = []byte("CBA")
	levelTwo1[0].StoreFile("a"+root.Username+levelOne2[0].Username, f)
	t.Log("> verify file content levelOne2")
	verifyFileOneLevel(t, root, "a", true, f, levelOne2)
	t.Log("> verify file content levelOne1")
	verifyFileOneLevel(t, root, "a", false, f, levelOne1)

}

func TestBadRevoke(t *testing.T) {
	clear()
	u1, _ := InitUser("a", "a")
	u2, _ := InitUser("b", "b")
	u1.StoreFile("af", []byte("asdf"))
	at, _ := u1.ShareFile("af", u2.Username)
	u2.ReceiveFile("sf", u1.Username, at)

	e := u1.RevokeFile("bf", u2.Username)
	if e == nil {
		t.Error("Able to revoke nonexistent file")
		return
	}

	u1.RevokeFile("af", u2.Username)
	e = u2.ReceiveFile("sf", u1.Username, at)
	if e == nil {
		t.Error("Able to use revoked access token")
		return
	}
}

func TestShareMITM(t *testing.T) {
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
	magic_string = magic_string + "a"
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err == nil {
		t.Error("accepted a malicious magic string", err)
		return
	}
}

func TestShareMultiUser(t *testing.T) {
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
	u3, err2 := GetUser("alice", "fubar")
	if err2 != nil {
		t.Error("Failed to get alice", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	//var v3 []byte
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

	v = []byte("This is a test")
	u2.StoreFile("file1", v)
	err = u2.ReceiveFile("file1", "alice", magic_string)
	if err == nil {
		t.Error("received file with the name that already exists", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	vAp1 := []byte(" olala")
	err = u2.AppendFile("file2", vAp1)
	if err != nil {
		t.Error("Failed to append to file2", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	v3, err := u3.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) || !reflect.DeepEqual(v3, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
	err = u3.RevokeFile("file1", "bob")
	if err != nil {
		t.Error("Failed to revoke bob's access after sharing", err)
		return
	}
	v2, err = u2.LoadFile("file2")
	f := append(v, vAp1...)
	if reflect.DeepEqual(v2, f) {
		t.Error("bob was able to read file after revoking", err)
		return
	}
}

//func TestIntegrityUser(t *testing.T) {
//	u1, err := InitUser("alice","fubar")
//	u1, err = GetUser("alice", "fubar")
//	if err != nil {
//		t.Error("Unexpected Error:", err)
//		return
//	}
//	t.Log(u1)
//	// Generate user-pass specific key (PBKDF)
//	PBKDFkey := userlib.Argon2Key([]byte("fubar"), []byte("alice"), 16)
//
//	// Location Key
//	// Create a deterministic UUID/location for User using HashKDF on `key`
//	locKey, err := userlib.HashKDF(PBKDFkey, []byte("uuid User"))
//	locKey = locKey[:userlib.AESKeySize]
//	location := bytesToUUID(locKey)
//	u1.Password = "12"
//	d, _ := json.Marshal(&u1)
//	userlib.DatastoreSet(location, d)
//	_, err = GetUser("alice", "fubar")
//	if err == nil {
//		t.Error("Failed to detect Database integrity error", err)
//		return
//	}
//}

// func TestNilUser(t *testing.T) {
// 	clear()
// 	var u *User
// 	u.StoreFile("ABC", []byte("aa"))
// 	if _, e := u.LoadFile("ABC"); e == nil {
// 		t.Error("Nil User able to store")
// 		return
// 	}
// }

// func TestBadUser(t *testing.T) {
// 	clear()
// 	var u User
// 	u.StoreFile("ABC", []byte("aa"))
// 	if _, e := u.LoadFile("ABC"); e == nil {
// 		t.Error("Bad User able to store")
// 		return
// 	}
// }

func TestINvalidFile(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	_, err2 := u.LoadFile("this file does not exist")
	if err2 == nil {
		t.Error("Downloaded a nonexistent file", err2)
		return
	}
	err = u.AppendFile("file1", []byte("v"))
	if err == nil {
		t.Error("Appended to a nonexistent file", err)
		return
	}
}

func TestStorage2(t *testing.T) {
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
	v = []byte("This is a test")
	u.StoreFile("file1", v)

	v2, err2 = u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestInvalidFile2(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	_, err2 := u.LoadFile("this file does not exist")
	if err2 == nil {
		t.Error("Downloaded a nonexistent file", err2)
		return
	}
	err = u.AppendFile("file1", []byte("v"))
	if err == nil {
		t.Error("Appended to a nonexistent file", err)
		return
	}
}

func TestShareHierarchy2(t *testing.T) {
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
	u3, err2 := InitUser("charles", "foodbar")
	if err2 != nil {
		t.Error("Failed to initialize charles", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var v3 []byte
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

	magic_string, err = u2.ShareFile("file2", "charles")
	if err != nil {
		t.Error("Failed to share the a file w/ charles", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string)
	if err != nil {
		t.Error("charles failed to receive the share message", err)
		return
	}

	v3, err = u3.LoadFile("file3")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v3) || !reflect.DeepEqual(v2, v3) {
		t.Error("Shared file is not the same", v, v3)
		return
	}
}

func TestShareHierarchyStore2(t *testing.T) {
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
	u3, err2 := InitUser("charles", "foodbar")
	if err2 != nil {
		t.Error("Failed to initialize charles", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	//var v3 []byte
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

	vS1 := []byte("bb")
	u2.StoreFile("file2", vS1)

	v, err = u.LoadFile("file1")
	v2, err = u2.LoadFile("file2")
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", string(v), string(v2))
		return
	}

	magic_string, err = u2.ShareFile("file2", "charles")
	if err != nil {
		t.Error("Failed to share the a file w/ charles", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string)
	if err != nil {
		t.Error("charles failed to receive the share message", err)
		return
	}

	v, err = u.LoadFile("file1")
	v2, err = u2.LoadFile("file2")
	v3, err := u3.LoadFile("file3")
	if !reflect.DeepEqual(v, v2) || !reflect.DeepEqual(v3, v2) {
		t.Error("Shared file is not the same", string(v), string(v3))
		return
	}
}

func TestShareRevoke(t *testing.T) {
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
	_, err3 := InitUser("charles", "foodbar")
	if err3 != nil {
		t.Error("Failed to initialize charles", err2)
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
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", string(v), string(v2))
		return
	}

	err = u.RevokeFile("file1", "bob")
	if err != nil {
		t.Error("alice failed to revoke bob's access to file1", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", string(v), string(v2))
		return
	}
	// Append file
	err = u.AppendFile("file1", []byte("!!"))
	if err != nil {
		t.Error("alice failed to append to file1", err)
		return
	}

	v = append(v, []byte("!!")...)
	vL, err := u.LoadFile("file1")
	if !reflect.DeepEqual(vL, v) {
		t.Error("Failed to append to the file", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if reflect.DeepEqual(v, v2) {
		t.Error("Revoked file is the same after update", string(v), string(v2))
		return
	}

	// Store file
	u.StoreFile("file1", []byte("bbb"))

	v = []byte("bbb")
	vL, err = u.LoadFile("file1")
	if !reflect.DeepEqual(vL, v) {
		t.Error("Failed to append to the file", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if reflect.DeepEqual(v, v2) {
		t.Error("Revoked file is the same after update", string(v), string(v2))
		return
	}

	// Share
	magic_string, err = u2.ShareFile("file2", "charles")
	if err == nil {
		t.Error("Was able to share a revoked file", err)
		return
	}
}
func TestGet4(t *testing.T) {
	clear()
	t.Log("Initialization test")

	// You can set this to false!
	userlib.SetDebugStatus(true)

	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	// If you want to comment the line above,
	// write _ = u here to make the compiler happy
	// You probably want many more tests here.
	u, err = GetUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to get user", err)
		return
	}
	// t.Log() only produces output if you run with "go test -v"
	t.Log("Got user", u)

	u, err = GetUser("alice", "foobar")
	if err == nil {
		// t.Error says the test fails
		t.Error("Got a user with wrong password", err)
		return
	}

	u, err = GetUser("bob", "foobar")
	if err == nil {
		// t.Error says the test fails
		t.Error("Got a user that's nonexistent", err)
		return
	}
}

func TestStorage4(t *testing.T) {
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
	v = []byte("This is a test")
	u.StoreFile("file1", v)

	v2, err2 = u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestAppend4(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This")
	u.StoreFile("file1", v)

	v = []byte(" is")
	err = u.AppendFile("file1", v)
	if err != nil {
		t.Error("Failed to append to file1", err)
		return
	}

	v = []byte(" a")
	err = u.AppendFile("file1", v)
	if err != nil {
		t.Error("Failed to append to file1", err)
		return
	}

	v = []byte(" test")
	err = u.AppendFile("file1", v)
	if err != nil {
		t.Error("Failed to append to file1", err)
		return
	}

	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload, append, and download", err2)
		return
	}
	if !reflect.DeepEqual([]byte("This is a test"), v2) {
		t.Error("Downloaded file is not the same", []byte("This is a test"), v2)
		return
	}

	v = []byte("This")
	u.StoreFile("file1", v)
	v2, err2 = u.LoadFile("file1")
	if err2 != nil || !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestInvalidFile4(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	_, err2 := u.LoadFile("this file does not exist")
	if err2 == nil {
		t.Error("Downloaded a nonexistent file", err2)
		return
	}
	err = u.AppendFile("file1", []byte("v"))
	if err == nil {
		t.Error("Appended to a nonexistent file", err)
		return
	}
}

func TestShare4(t *testing.T) {
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
}

func TestShareHierarchy3(t *testing.T) {
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
	u3, err2 := InitUser("charles", "foodbar")
	if err2 != nil {
		t.Error("Failed to initialize charles", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var v3 []byte
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

	magic_string, err = u2.ShareFile("file2", "charles")
	if err != nil {
		t.Error("Failed to share the a file w/ charles", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string)
	if err != nil {
		t.Error("charles failed to receive the share message", err)
		return
	}

	v3, err = u3.LoadFile("file3")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v3) || !reflect.DeepEqual(v2, v3) {
		t.Error("Shared file is not the same", v, v3)
		return
	}
}

func TestShareHierarchyAppend3(t *testing.T) {
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
	u3, err2 := InitUser("charles", "foodbar")
	if err2 != nil {
		t.Error("Failed to initialize charles", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	//var v3 []byte
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

	vAp1 := []byte(" olala")
	err = u2.AppendFile("file2", vAp1)
	if err != nil {
		t.Error("Failed to append to file2", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(append(v, vAp1...), v2) {
		t.Error("Shared file is not the same", string(append(v, vAp1...)), string(v2))
		return
	}

	magic_string, err = u2.ShareFile("file2", "charles")
	if err != nil {
		t.Error("Failed to share the a file w/ charles", err)
		return
	}
	err = u3.ReceiveFile("file3", "bob", magic_string)
	if err != nil {
		t.Error("charles failed to receive the share message", err)
		return
	}
}

func TestShareMITM4(t *testing.T) {
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
	magic_string = magic_string + "a"
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err == nil {
		t.Error("accepted a malicious magic string", err)
		return
	}
}

func TestShareMultiUser5(t *testing.T) {
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
	u3, err2 := GetUser("alice", "fubar")
	if err2 != nil {
		t.Error("Failed to get alice", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	//var v3 []byte
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

	v = []byte("This is a test")
	u2.StoreFile("file1", v)
	err = u2.ReceiveFile("file1", "alice", magic_string)
	if err == nil {
		t.Error("received file with the name that already exists", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	vAp1 := []byte(" olala")
	err = u2.AppendFile("file2", vAp1)
	if err != nil {
		t.Error("Failed to append to file2", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	v3, err := u3.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) || !reflect.DeepEqual(v3, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
	err = u3.RevokeFile("file1", "bob")
	if err != nil {
		t.Error("Failed to revoke bob's access after sharing", err)
		return
	}
	v2, err = u2.LoadFile("file2")
	if err == nil {
		t.Error("bob was able to read file after revoking", err)
		return
	}
}

func TestShareMalicious(t *testing.T) {
	clear()
	alice, _ := InitUser("alice", "fubar")
	bob, _ := InitUser("bob", "clifbar")
	eve, _ := InitUser("eve", "pobar")
	alice.StoreFile("Test", []byte("12345"))
	m, err := alice.ShareFile("Test", "bob")
	if err != nil {
		t.Error("Failed to share Test file", err)
	}
	err = bob.ReceiveFile("SharedTest", "alice", m)
	if err != nil {
		t.Error("Failed to receive the SharedTest file", err)
	}
	aliceData, err := alice.LoadFile("Test")
	if !reflect.DeepEqual([]byte("12345"), aliceData) || err != nil {
		t.Error("Failed to load alice's Test file", err)
		return
	}
	bobData, err := bob.LoadFile("SharedTest")
	if !reflect.DeepEqual(aliceData, bobData) || err != nil {
		t.Error("Failed to share alice's Test file with bob", err)
		return
	}
	eveData, err := eve.LoadFile("Test")
	if reflect.DeepEqual(aliceData, eveData) || reflect.DeepEqual(bobData, eveData) || err == nil {
		t.Error("shared the data with Eve", err)
		return
	}
	eveData, err = eve.LoadFile("SharedTest")
	if reflect.DeepEqual(aliceData, eveData) || reflect.DeepEqual(bobData, eveData) || err == nil {
		t.Error("shared the data with Eve", err)
		return
	}
	eveData, err = eve.LoadFile("SharedTester")
	if reflect.DeepEqual(aliceData, eveData) || reflect.DeepEqual(bobData, eveData) || err == nil {
		t.Error("shared the data with Eve", err)
		return
	}
	t.Log("Successfully shared the file with bob:", "Test")
	m, err = bob.ShareFile("SharedTest", "eve")
	err = eve.ReceiveFile("SharedTester", "bob", m)
	if err != nil {
		t.Error("failed to share file with eve", err)
		return
	}
	bobData, err = bob.LoadFile("SharedTest")
	aliceData, err = alice.LoadFile("Test")
	eveData, err = eve.LoadFile("SharedTester")
	if !reflect.DeepEqual(eveData, bobData) || err != nil {
		t.Error("Failed to share bob's Test file with eve", err)
		return
	}
	if !reflect.DeepEqual(aliceData, eveData) {
		t.Error("Failed to share alice's Test file with eve", err)
		return
	}
	t.Log("Successfully shared the file with eve:", "Test")

	aliceData, err = alice.LoadFile("SharedTest")
	if reflect.DeepEqual(aliceData, bobData) || err == nil {
		t.Error("Accidentally updated the file name in Alice", err)
		return
	}
	aliceData, err = alice.LoadFile("Test")
	err = bob.RevokeFile("Sha", "bob")
	if err == nil {
		t.Error("Revoked access for a file not present", "Sha")
		return
	}
	_, err = bob.ShareFile("Sha", "eve")
	if err == nil {
		t.Error("shared access to a file not present", "Sha")
		return
	}
	err = bob.AppendFile("SharedTest", []byte("7"))
	if err != nil {
		t.Error("failed to append to file", "SharedTest")
		return
	}
	bobData, err = bob.LoadFile("SharedTest")
	aliceData, err = alice.LoadFile("Test")
	eveData, err = eve.LoadFile("SharedTester")
	if !reflect.DeepEqual(bobData, aliceData) || !reflect.DeepEqual(aliceData, eveData) || !reflect.DeepEqual(bobData, eveData) {
		t.Error("Did not update all shared files", "SharedTest")
		return
	}

	err = alice.RevokeFile("Test", "bob")
	if err != nil {
		t.Error("Failed to revoke access by the owner", err)
		return
	}
	_, err1 := bob.LoadFile("SharedTest")
	_, err2 := eve.LoadFile("SharedTester")
	if err1 == nil || err2 == nil {
		t.Error("Failed to revoke access by the owner", "Test")
		return
	}
	err = bob.ReceiveFile("SharedTest", "alice", m)
	_, err1 = bob.LoadFile("SharedTest")
	if err1 == nil {
		t.Error("Failed to keep revoked access by the owner", "Test")
		return
	}
	newAliceData, err := alice.LoadFile("Test")
	if !reflect.DeepEqual(newAliceData, aliceData) {
		t.Error("Modified the contents of original file when revoking access by the owner", "Test")
		return
	}
}

func TestMalicious(t *testing.T) {
	clear()
	alice, _ := InitUser("alice", "fubar")
	userlib.DatastoreClear()
	alice.StoreFile("Test", []byte("12345"))
	_, err := alice.LoadFile("Test")
	if err == nil {
		t.Error("Failed to detect Database integrity error", err)
		return
	}
	alice, err = GetUser("alice", "fubar")
	if err == nil {
		t.Error("Failed to detect Database integrity error", err)
		return
	}
	alice, err = InitUser("alice", "fubar")
	if err == nil {
		t.Error("create a user who key has been generated", err)
		return
	}
	userlib.KeystoreClear()
	alice, err = InitUser("alice", "fubar")
	if err != nil {
		t.Error("failed to create a user", err)
		return
	}
}
