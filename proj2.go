package proj2

// CS 161 Project 2 Spring 2020
// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder. We will be very upset.

import (
	// You neet to add with
	// go get github.com/cs161-staff/userlib
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
	SkSalt   []byte

	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

type DatastoreUser struct {
	OwnedFilePurpose map[string][]byte
}

type DataDS struct {
	ciphertext []byte
	mactext    []byte
}

type UserVerification struct {
	HashPassword     []byte
	HashPasswordSalt []byte
	SkSalt           []byte
	HashPasswordSig  []byte
}

const (
	PASSWD_K                 = "password verification"
	UM_k                     = "user metadata"
	USERMETADATA_PURPOSE     = "user metadata    "
	MAC_USERMETADATA_PURPOSE = "mac user metadata"
	HASH_PW_LEN              = 512
	SALT_LEN                 = 256
	SK_LEN                   = 256
)

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

	// Generate UUID
	hashUserName := userlib.Hash([]byte(username + PASSWD_K))
	uvid, e := uuid.FromBytes(hashUserName[:16])

	if e != nil {
		return nil, errors.New("InitUser: Failed to create uuid1")
	}

	hashUserName = userlib.Hash([]byte(username + UM_k))
	umid, e := uuid.FromBytes(hashUserName[:16])

	if e != nil {
		return nil, errors.New("InitUser: Failed to uuid2")
	}

	// Check uniqueness
	if _, ok := userlib.DatastoreGet(uvid); ok != false {
		userlib.DebugMsg("InitUser: User already exists")
		return nil, errors.New("User exists")
	}

	// TODO: 2 pairs of File sharing keys

	// TODO: 2 pairs of Access token keys

	// 1 pair of password verification keys
	pwSignKey, pwVerifyKey, e := userlib.DSKeyGen()

	if e != nil {
		userlib.DebugMsg("InitUser: Failed to create pw keys")
		return nil, e
	}

	// User verification initialization

	hashPasswordSalt := userlib.RandomBytes(256)
	hashPw := userlib.Hash([]byte(password))
	hashPassword := userlib.Argon2Key(hashPw[:], hashPasswordSalt, 512)
	skSalt := userlib.RandomBytes(256)

	var text []byte
	text = append(text, hashPassword...)
	text = append(text, hashPasswordSalt...)
	text = append(text, skSalt...)

	hashPasswordSig, e := userlib.DSSign(pwSignKey, text)

	if e != nil {
		return nil, errors.New("Failed to create hashpasswd signature")
	}

	userlib.DebugMsg("%v, %v\n", text, hashPasswordSig)

	uv := UserVerification{hashPassword, hashPasswordSalt, skSalt, hashPasswordSig}

	// User metadata initialization

	sk := userlib.Argon2Key(hashPw[:], skSalt, 16)
	userMetaData := DatastoreUser{make(map[string][]byte)}
	plaintext, e := json.Marshal(userMetaData)
	if e != nil {
		return nil, errors.New("Failed to marshall userMetadata")
	}
	userMetaDataSk, e := userlib.HashKDF(sk, []byte(USERMETADATA_PURPOSE))
	if e != nil {
		return nil, errors.New("Failed to generate usermetadata sk")
	}
	userMetaDataMac, e := userlib.HashKDF(sk, []byte(MAC_USERMETADATA_PURPOSE))
	if e != nil {
		return nil, errors.New("Failed to generate usermetadata mac key")
	}
	ciphertext := userlib.SymEnc(userMetaDataSk[:16], userlib.RandomBytes(userlib.AESBlockSize), plaintext)
	if e != nil {
		return nil, errors.New("Failed to encrypt usermetadata")
	}
	mactext, e := userlib.HMACEval(userMetaDataMac[:16], ciphertext)
	if e != nil {
		return nil, errors.New("Failed to hmac usermetadata")
	}
	um := DataDS{ciphertext, mactext}

	// TODO: store public keys in keystore

	userlib.KeystoreSet(username+PASSWD_K, pwVerifyKey)

	// TODO: store verification in datastore

	if c, e := json.Marshal(uv); e != nil {
		return nil, e
	} else {
		userlib.DatastoreSet(uvid, c)
	}

	if c, e := json.Marshal(um); e != nil {
		return nil, e
	} else {
		userlib.DatastoreSet(umid, c)
	}

	userdata.Username = username
	userdata.Password = password
	userdata.SkSalt = skSalt

	return &userdata, nil
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata
	pwVerifyKey, verifyKeyExist := userlib.KeystoreGet(username + PASSWD_K)

	if verifyKeyExist == false {
		return nil, errors.New("GetUser: User does not exist")
	}

	// Generate UUID
	hashUserName := userlib.Hash([]byte(username + PASSWD_K))
	uid, e := uuid.FromBytes(hashUserName[:16])

	if e != nil {
		userlib.DebugMsg("GetUser: Failed to create uuid")
		return nil, e
	}

	// Check existence

	data, ok := userlib.DatastoreGet(uid)

	if ok == false {
		return nil, errors.New("GetUser: Failed to get user from datastore")
	}

	// verify pw

	userVerification := UserVerification{}
	if json.Unmarshal(data, &userVerification) != nil {
		return nil, errors.New("GetUser: Failed to unmarshal pw verification")
	}

	// verify integrity

	var text []byte
	text = append(text, userVerification.HashPassword...)
	text = append(text, userVerification.HashPasswordSalt...)
	text = append(text, userVerification.SkSalt...)

	if userlib.DSVerify(pwVerifyKey, text, userVerification.HashPasswordSig) != nil {
		userlib.DebugMsg("%v, %v", text, userVerification.HashPasswordSig)
		return nil, errors.New("GetUser: Verification failed")
	}

	hashPw := userlib.Hash([]byte(password))
	hashPassword := userlib.Argon2Key(hashPw[:], userVerification.HashPasswordSalt, HASH_PW_LEN)

	if string(hashPassword) != string(userVerification.HashPassword) {
		return nil, errors.New("GetUser: Invalid passwd")
	}

	// construct User struct
	userdataptr.Username = username
	userdataptr.Password = password
	userdataptr.SkSalt = userVerification.SkSalt

	return userdataptr, nil
}

// This stores a file in the datastore.
//
// The plaintext of the filename + the plaintext and length of the filename
// should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {

	//TODO: This is a toy implementation.
	UUID, _ := uuid.FromBytes([]byte(filename + userdata.Username)[:16])
	packaged_data, _ := json.Marshal(data)
	userlib.DatastoreSet(UUID, packaged_data)
	//End of toy implementation

	return
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.
func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {

	//TODO: This is a toy implementation.
	UUID, _ := uuid.FromBytes([]byte(filename + userdata.Username)[:16])
	packaged_data, ok := userlib.DatastoreGet(UUID)
	if !ok {
		return nil, errors.New(strings.ToTitle("File not found!"))
	}
	json.Unmarshal(packaged_data, &data)
	return data, nil
	//End of toy implementation

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
