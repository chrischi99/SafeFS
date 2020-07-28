package proj2

// CS 161 Project 2 Spring 2020
// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder. We will be very upset.

import (
	// You neet to add with
	// go get github.com/cs161-staff/userlib
	"strconv"

	"github.com/cs161-staff/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...
	"encoding/json"

	// Likewise useful for debugging, etc...
	"encoding/hex"

	// UUIDs are generated right based on the cryptographic PRNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"

	// Useful for debug messages, or string manipulation for datastore keys.
	"strings"

	// Want to import errors.
	"errors"

	// Optional. You can remove the "_" there, but please do not touch
	// anything else within the import bracket.
	_ "strconv"
	// if you are looking for fmt, we don't give you fmt, but you can use userlib.DebugMsg.
	// see someUsefulThings() below:
)

// This serves two purposes:
// a) It shows you some useful primitives, and
// b) it suppresses warnings for items not being imported.
// Of course, this function can be deleted.
func someUsefulThings() {
	// Creates a random UUID
	f := uuid.New()
	userlib.DebugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	userlib.DebugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	userlib.DebugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	userlib.DebugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	userlib.DebugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	userlib.DebugMsg("Creation of error %v", errors.New(strings.ToTitle("This is an error")))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var pk userlib.PKEEncKey
	var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("Key is %v, %v", pk, sk)
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

// The structure definition for a user record
type User struct {
	Username string
	Password string
	Sk       []byte

	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

// metadata

type DatastoreUser struct {
	OwnedFilePurpose map[string][]byte
}

type DatastoreFile struct {
	RecordUUID uuid.UUID
	SharedUser []string
}

type UserVerification struct {
	HashPassword     []byte
	HashPasswordSalt []byte
	SkSalt           []byte
}

type RecordBook struct {
	Records []uuid.UUID
}

// data should be wrapped with DataDS before sent to Datastore

type DataDS struct {
	Ciphertext []byte
	Verifytext []byte
}

const (
	// uuid
	UUID_PASSWD       = "password uuid"
	UUID_USERMETADATA = "user metadata uuid"
	UUID_RECORDBOOK   = "record book uuid"
	// key store key
	PASSWD_K = "password verification"
	UM_k     = "user metadata"
	// sk genereating purpose
	USERMETADATA_ENC_PURPOSE = "enc user metadata"
	USERMETADATA_MAC_PURPOSE = "mac user metadata"
	FILEMETADATA_ENC_PURPOSE = "enc file metadata"
	FILEMETADATA_MAC_PURPOSE = "mac file metadata"
	PASSWD_PURPOSE           = "password verification purpose"
	RECORD_BOOK_ENC_PURPOSE  = "record book enc purpose"
	RECORD_BOOK_MAC_PURPOSE  = "record book mac purpose"
	FILE_PART_ENC_PURPOSE    = "file part enc"
	FILE_PART_MAC_PURPOSE    = "file part mac"
	// const value
	HASH_PW_LEN = 256
	SALT_LEN    = 256
	SK_LEN      = 256
)

// helper function

func getPurpose(s string) (p []byte) {
	hashval := userlib.Hash([]byte(s))
	return hashval[:16]
}

func getSymKeys(sk []byte, encpurpose string, macpurpose string) (enckey []byte, mackey []byte) {
	// get metadata key pair
	enckey, _ = userlib.HashKDF(sk[:16], getPurpose(encpurpose))
	mackey, _ = userlib.HashKDF(sk[:16], getPurpose(macpurpose))
	enckey = enckey[:16]
	mackey = mackey[:16]
	return
}

// get/put method for each struct stored in datastore

func putuserverification(username string, signk userlib.DSSignKey, uv *UserVerification) {
	plaintext, _ := json.Marshal(&uv)
	verifytext, _ := userlib.DSSign(signk, plaintext)
	_obj := DataDS{plaintext, verifytext}
	obj, _ := json.Marshal(&_obj)
	UUID, _ := uuid.FromBytes(getPurpose(username + UUID_PASSWD))

	userlib.DatastoreSet(UUID, obj)
	_, _ = getUserVerification(username)
}

func getUserVerification(username string) (uv *UserVerification, err error) {
	var userVerification UserVerification
	uv = &userVerification
	var _obj DataDS
	UUID, _ := uuid.FromBytes(getPurpose(username + UUID_PASSWD))
	pwVerifyKey, verifyKeyExist := userlib.KeystoreGet(username + PASSWD_K)
	obj, ok := userlib.DatastoreGet(UUID)
	if !ok || !verifyKeyExist {
		return nil, errors.New("> user verification not found")
	}
	json.Unmarshal(obj, &_obj)
	// interity verify
	if userlib.DSVerify(pwVerifyKey, _obj.Ciphertext, _obj.Verifytext) != nil {
		return nil, errors.New("> user verification integrity verification failed")
	}
	json.Unmarshal(_obj.Ciphertext, uv)
	return
}

func verifyMac(mack []byte, ciphertext []byte, mactext []byte) (ok bool) {
	curMactext, _ := userlib.HMACEval(mack, ciphertext)
	return userlib.HMACEqual(curMactext, mactext)
}

func getSymData(UUID uuid.UUID, sk []byte, encPurpose string, macPurpose string) (plaintext []byte, err error) {
	var _obj DataDS
	obj, ok := userlib.DatastoreGet(UUID)
	if !ok {
		return nil, errors.New("Metadata not found")
	}
	json.Unmarshal(obj, &_obj)
	encKey, mackey := getSymKeys(sk, encPurpose, macPurpose)
	// verify
	if verifyMac(mackey, _obj.Ciphertext, _obj.Verifytext) == false {
		return nil, errors.New("Verify integrity failed")
	}
	// decrypt
	plaintext = userlib.SymDec(encKey, _obj.Ciphertext)
	return
}

func putSymData(UUID uuid.UUID, sk []byte, encPurpose string, macPurpose string, plaintext []byte) {
	var _obj DataDS
	encKey, macKey := getSymKeys(sk, encPurpose, macPurpose)
	ciphertext := userlib.SymEnc(encKey, userlib.RandomBytes(userlib.AESBlockSize), plaintext)
	verifytext, _ := userlib.HMACEval(macKey, ciphertext)
	_obj.Ciphertext = ciphertext
	_obj.Verifytext = verifytext
	obj, _ := json.Marshal(&_obj)
	userlib.DatastoreSet(UUID, obj)
}

func (userdata *User) getUserMetadata() (um *DatastoreUser, err error) {
	var userMetadata DatastoreUser
	um = &userMetadata
	UUID, _ := uuid.FromBytes(getPurpose(userdata.Username + UUID_USERMETADATA))
	plaintext, e := getSymData(UUID, userdata.Sk, USERMETADATA_ENC_PURPOSE, USERMETADATA_MAC_PURPOSE)
	if e != nil {
		return nil, e
	}
	json.Unmarshal(plaintext, um)
	return
}

func (userdata *User) putUserMetadata(um *DatastoreUser) {
	UUID, _ := uuid.FromBytes(getPurpose(userdata.Username + UUID_USERMETADATA))
	plaintext, _ := json.Marshal(um)
	putSymData(UUID, userdata.Sk, USERMETADATA_ENC_PURPOSE, USERMETADATA_MAC_PURPOSE, plaintext)
}

func (userdata *User) getFileMetadata(filename string) (fm *DatastoreFile, err error) {
	var datastoreFile DatastoreFile
	fm = &datastoreFile
	UUID, _ := uuid.FromBytes(getPurpose(userdata.Username + filename))
	plaintext, e := getSymData(UUID, userdata.Sk, FILEMETADATA_ENC_PURPOSE, FILEMETADATA_MAC_PURPOSE)
	if e != nil {
		return nil, errors.New("getfilemetadata " + e.Error())
	}
	json.Unmarshal(plaintext, fm)
	return
}

func (userdata *User) putFileMetadata(filename string, fm *DatastoreFile) {
	UUID, _ := uuid.FromBytes(getPurpose(userdata.Username + filename))
	plaintext, _ := json.Marshal(fm)
	putSymData(UUID, userdata.Sk, FILEMETADATA_ENC_PURPOSE, FILEMETADATA_MAC_PURPOSE, plaintext)
}

func (userdata *User) getRecordBook(UUID uuid.UUID, purpose []byte) (rb *RecordBook, err error) {
	var recordBook RecordBook
	rb = &recordBook
	sk, _ := userlib.HashKDF(userdata.Sk[:16], []byte(purpose))
	plaintext, e := getSymData(UUID, sk, RECORD_BOOK_ENC_PURPOSE, RECORD_BOOK_MAC_PURPOSE)
	if e != nil {
		return nil, e
	}
	json.Unmarshal(plaintext, rb)
	return
}

func (userdata *User) putRecordBook(filename string, UUID uuid.UUID, purpose []byte, rb *RecordBook) {
	plaintext, _ := json.Marshal(rb)
	sk, _ := userlib.HashKDF(userdata.Sk[:16], []byte(purpose))
	putSymData(UUID, sk, RECORD_BOOK_ENC_PURPOSE, RECORD_BOOK_MAC_PURPOSE, plaintext)
}

func (userdata *User) getFilePart(UUID uuid.UUID, filename string, purpose []byte, i int) (data []byte, err error) {
	sk, _ := userlib.HashKDF(userdata.Sk[:16], []byte(purpose))
	data, e := getSymData(UUID, sk, filename+FILE_PART_ENC_PURPOSE+string(i), filename+FILE_PART_MAC_PURPOSE+string(i))
	if e != nil {
		return nil, e
	}
	return
}

func (userdata *User) putFilePart(filename string, UUID uuid.UUID, purpose []byte, i int, plaintext []byte) {
	sk, _ := userlib.HashKDF(userdata.Sk[:16], []byte(purpose))
	putSymData(UUID, sk, filename+FILE_PART_ENC_PURPOSE+string(i), filename+FILE_PART_MAC_PURPOSE+string(i), plaintext)
}

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the password has strong entropy, EXCEPT
// the attackers may possess a precomputed tables containing
// hashes of common passwords downloaded from the internet.
func InitUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata
	// Check uniqueness
	UUID, _ := uuid.FromBytes(getPurpose(username + UUID_PASSWD))
	if _, ok := userlib.DatastoreGet(UUID); ok != false {
		userlib.DebugMsg("InitUser: User already exists")
		return nil, errors.New("User exists")
	}
	// TODO: 2 pairs of File sharing keys
	// TODO: 2 pairs of Access token keys
	// 1 pair of password verification keys
	pwSignKey, pwVerifyKey, _ := userlib.DSKeyGen()
	// Initialize and store User verification
	hashPw := userlib.Hash([]byte(password))
	hashPasswordSalt := userlib.RandomBytes(SALT_LEN)
	hashPassword := userlib.Argon2Key(hashPw[:], hashPasswordSalt, HASH_PW_LEN)
	skSalt := userlib.RandomBytes(SALT_LEN)
	uv := UserVerification{hashPassword, hashPasswordSalt, skSalt}
	userlib.KeystoreSet(username+PASSWD_K, pwVerifyKey)
	putuserverification(username, pwSignKey, &uv)
	// Initialize and store User metadata
	sk := userlib.Argon2Key(hashPw[:], skSalt, SK_LEN)
	userdataptr.Username = username
	userdataptr.Password = password
	userdataptr.Sk = sk
	userMetaData := DatastoreUser{make(map[string][]byte, 0)}
	userdataptr.putUserMetadata(&userMetaData)
	// TODO: store public keys in keystore
	userlib.KeystoreSet(username+PASSWD_K, pwVerifyKey)
	return userdataptr, nil
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata
	// get user verification
	uv, e := getUserVerification(username)
	if e != nil {
		return nil, e
	}
	// verify password
	_hashPw := userlib.Hash([]byte(password))
	hashPassword := userlib.Argon2Key(_hashPw[:], uv.HashPasswordSalt, HASH_PW_LEN)
	if string(hashPassword) != string(uv.HashPassword) {
		return nil, errors.New("Invalid password")
	}
	userdataptr.Username = username
	userdataptr.Password = password
	userdataptr.Sk = userlib.Argon2Key(_hashPw[:], uv.SkSalt, SK_LEN)
	return userdataptr, nil
}

// This stores a file in the datastore.
//
// The plaintext of the filename + the plaintext and length of the filename
// should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {
	// get datastore user to check whether file exists and get purpose
	userlib.SetDebugStatus(true)
	um, e := userdata.getUserMetadata()
	if e != nil {
		return
	}
	// update
	filenameHash := userlib.Hash([]byte(filename))
	purpose, ok := um.OwnedFilePurpose[hex.EncodeToString(filenameHash[:])]
	var datastoreFile *DatastoreFile
	var recordBook *RecordBook
	if ok {
		// file exists, we only need to modify record book, file_part len - 1
		if datastoreFile, e = userdata.getFileMetadata(filename); e != nil {
			return
		}
		if recordBook, e = userdata.getRecordBook(datastoreFile.RecordUUID, purpose); e != nil {
			return
		}
	} else {
		// file doesn't exist, we need to modify user_metadata(purpose), file_metadata(), record_book, file_part0
		// initialize purpose
		purpose = userlib.RandomBytes(16)
		um.OwnedFilePurpose[hex.EncodeToString(filenameHash[:])] = purpose
		userdata.putUserMetadata(um)

		// initialize and modify file metadata
		_datastoreFile := DatastoreFile{uuid.New(), make([]string, 0)}
		datastoreFile = &_datastoreFile
		userdata.putFileMetadata(filename, datastoreFile)
		// initialize record book
		_recordBook := RecordBook{make([]uuid.UUID, 0)}
		recordBook = &_recordBook
	}
	// both need to modify record_book and file_part
	recordBook.Records = append(recordBook.Records, uuid.New())
	userdata.putRecordBook(filename, datastoreFile.RecordUUID, purpose, recordBook)
	userdata.putFilePart(filename, recordBook.Records[len(recordBook.Records)-1], purpose, len(recordBook.Records)-1, data)
	return
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.
func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	// get datastore user to check whether file exists and get purpose
	um, e := userdata.getUserMetadata()
	if e != nil {
		return e
	}
	// update
	filenameHash := userlib.Hash([]byte(filename))
	purpose, ok := um.OwnedFilePurpose[hex.EncodeToString(filenameHash[:])]
	if !ok {
		// TODO: sharing file support
		return errors.New("File does not exist")
	}
	fm, e := userdata.getFileMetadata(filename)
	if e != nil {
		return e
	}
	rb, e := userdata.getRecordBook(fm.RecordUUID, purpose)
	if e != nil {
		return e
	}
	UUID := uuid.New()
	rb.Records = append(rb.Records, UUID)
	userdata.putFilePart(filename, UUID, purpose, len(rb.Records)-1, data)
	userdata.putRecordBook(filename, fm.RecordUUID, purpose, rb)
	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	um, e := userdata.getUserMetadata()
	if e != nil {
		return nil, e
	}
	filenameHash := userlib.Hash([]byte(filename))
	purpose, ok := um.OwnedFilePurpose[hex.EncodeToString(filenameHash[:])]
	if !ok {
		// TODO: file sharing support
		userlib.DebugMsg("ok is false, %v", purpose)
		return nil, errors.New("File does not exist")
	}
	fm, e := userdata.getFileMetadata(filename)
	if e != nil {
		return nil, errors.New("> getFileMetadata " + e.Error())
	}
	rb, e := userdata.getRecordBook(fm.RecordUUID, purpose)
	if e != nil {
		return nil, errors.New("> getRecordBook " + e.Error())
	}
	for i, UUID := range rb.Records {
		d, e := userdata.getFilePart(UUID, filename, purpose, i)
		if e != nil {
			return nil, errors.New("> getFilePart " + strconv.Itoa(i))
		}
		data = append(data, d...)
	}
	return
}

// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.
func (userdata *User) ShareFile(filename string, recipient string) (
	magic_string string, err error) {

	return
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {
	return nil
}

// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	return
}
