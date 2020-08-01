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
// func bytesToUUID(data []byte) (ret uuid.UUID) {
// 	for x := range ret {
// 		ret[x] = data[x]
// 	}
// 	return
// }

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
	OwnedFilePurpose          map[string][]byte
	SharedFileSecrecyUUIDs    map[string]uuid.UUID
	SharedFileSecrecyKeypairs map[string][2][]byte
	ATokenSignSK              userlib.DSSignKey
	ATokenDecSk               userlib.PKEDecKey
}

type DatastoreFile struct {
	RecordUUID uuid.UUID
	SharedUser map[string]uuid.UUID
}

type UserVerification struct {
	HashPassword     []byte
	HashPasswordSalt []byte
	SkSalt           []byte
}

type RecordBook struct {
	Records []uuid.UUID
}

// file sharing
type AccessToken struct {
	UUID          uuid.UUID // uuid to KeySecrecy
	SecrecyDecKey []byte
	SecrecyMacKey []byte
}

type KeySecrecy struct {
	SK   []byte    // sk able to generate keys related to data
	UUID uuid.UUID // uuid to record book
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
	PASSWD_K  = "password verification"
	UM_k      = "user metadata"
	AT_SIGN_K = "access token sign pk"
	AT_ENC_K  = "access token enc pk"
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
	KEY_SECRECY_ENC_PURPOSE  = "secrecy enc"
	KEY_SECRECY_MAC_PURPSOE  = "secrecy mac"
	// const value
	HASH_PW_LEN = 256
	SALT_LEN    = 256
	SK_LEN      = 256
	PURPOSE_LEN = 16
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

func mapKey(name string) (s string) {
	namehash := userlib.Hash([]byte(name))
	s = hex.EncodeToString(namehash[:])
	return
}

// get/put method for each struct stored in datastore

// public encryption/verification
func putUserVerification(username string, signk userlib.DSSignKey, uv *UserVerification) {
	UUID, _ := uuid.FromBytes(getPurpose(username + UUID_PASSWD))
	plaintext, e := json.Marshal(uv)
	if e != nil {
		return
	}
	verifytext, _ := userlib.DSSign(signk, plaintext)
	_obj := DataDS{plaintext, verifytext}
	obj, e := json.Marshal(&_obj)
	if e != nil {
		return
	}
	userlib.DatastoreSet(UUID, obj)
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
	if e := json.Unmarshal(obj, &_obj); e != nil {
		return nil, e
	}
	// interity verify
	if userlib.DSVerify(pwVerifyKey, _obj.Ciphertext, _obj.Verifytext) != nil {
		return nil, errors.New("> user verification integrity verification failed")
	}
	if e := json.Unmarshal(_obj.Ciphertext, uv); e != nil {
		return nil, e
	}
	return
}

// symmetric encryption/verification
func verifyMac(mack []byte, ciphertext []byte, mactext []byte) (ok bool) {
	curMactext, _ := userlib.HMACEval(mack[:16], ciphertext)
	return userlib.HMACEqual(curMactext, mactext)
}

func getSymData(UUID uuid.UUID, sk []byte, encPurpose string, macPurpose string, encKey []byte, macKey []byte, shared bool) (plaintext []byte, err error) {
	var _obj DataDS
	obj, ok := userlib.DatastoreGet(UUID)
	if !ok {
		return nil, errors.New("object not found")
	}
	if e := json.Unmarshal(obj, &_obj); e != nil {
		return nil, e
	}
	if !shared {
		encKey, macKey = getSymKeys(sk, encPurpose, macPurpose)
	}
	// verify
	userlib.DebugMsg("verify inte, mackey: %v", macKey)
	if verifyMac(macKey[:16], _obj.Ciphertext, _obj.Verifytext) == false {
		return nil, errors.New("Verify integrity failed, use sk: ")
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
	obj, e := json.Marshal(&_obj)
	if e != nil {
		return
	}
	userlib.DatastoreSet(UUID, obj)
}

func (userdata *User) getUserMetadata() (um *DatastoreUser, err error) {
	var userMetadata DatastoreUser
	um = &userMetadata
	UUID, _ := uuid.FromBytes(getPurpose(userdata.Username + UUID_USERMETADATA))
	plaintext, e := getSymData(UUID, userdata.Sk, USERMETADATA_ENC_PURPOSE, USERMETADATA_MAC_PURPOSE, nil, nil, false)
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal(plaintext, um)
	if e != nil {
		return nil, e
	}
	return
}

func (userdata *User) putUserMetadata(um *DatastoreUser) {
	UUID, _ := uuid.FromBytes(getPurpose(userdata.Username + UUID_USERMETADATA))
	plaintext, e := json.Marshal(um)
	if e != nil {
		return
	}
	putSymData(UUID, userdata.Sk, USERMETADATA_ENC_PURPOSE, USERMETADATA_MAC_PURPOSE, plaintext)
}

func (userdata *User) getFileMetadata(filename string) (fm *DatastoreFile, err error) {
	var datastoreFile DatastoreFile
	fm = &datastoreFile
	UUID, _ := uuid.FromBytes(getPurpose(userdata.Username + filename))
	plaintext, e := getSymData(UUID, userdata.Sk, FILEMETADATA_ENC_PURPOSE, FILEMETADATA_MAC_PURPOSE, nil, nil, false)
	if e != nil {
		return nil, errors.New("getfilemetadata " + e.Error())
	}
	if e := json.Unmarshal(plaintext, fm); e != nil {
		return nil, e
	}
	return
}

func (userdata *User) putFileMetadata(filename string, fm *DatastoreFile) {
	UUID, _ := uuid.FromBytes(getPurpose(userdata.Username + filename))
	plaintext, e := json.Marshal(fm)
	if e != nil {
		return
	}
	putSymData(UUID, userdata.Sk, FILEMETADATA_ENC_PURPOSE, FILEMETADATA_MAC_PURPOSE, plaintext)
}

func (userdata *User) putFileSecrecy(UUID uuid.UUID, filename string, recipient string, keySecrecy *KeySecrecy) {
	plaintext, e := json.Marshal(keySecrecy)
	if e != nil {
		return
	}
	putSymData(UUID, userdata.Sk, filename+recipient+KEY_SECRECY_ENC_PURPOSE, filename+recipient+KEY_SECRECY_MAC_PURPSOE, plaintext)
}

func (userdata *User) getFileSecrecy(UUID uuid.UUID, filename string, recipient string, keyPair [2][]byte, shared bool) (ks *KeySecrecy, err error) {
	var keySecrecy KeySecrecy
	ks = &keySecrecy
	plaintext, e := getSymData(UUID, userdata.Sk, filename+recipient+KEY_SECRECY_ENC_PURPOSE, filename+recipient+KEY_SECRECY_MAC_PURPSOE, keyPair[0], keyPair[1], shared)
	if e != nil {
		return nil, e
	}
	if e := json.Unmarshal(plaintext, ks); e != nil {
		return nil, e
	}
	return
}

func (userdata *User) getRecordBook(UUID uuid.UUID, purpose []byte, sk []byte, shared bool) (rb *RecordBook, err error) {
	var recordBook RecordBook
	rb = &recordBook
	if !shared {
		sk, _ = userlib.HashKDF(userdata.Sk[:16], purpose)
	}
	plaintext, e := getSymData(UUID, sk[:16], RECORD_BOOK_ENC_PURPOSE, RECORD_BOOK_MAC_PURPOSE, nil, nil, false)
	if e != nil {
		return nil, e
	}
	if e := json.Unmarshal(plaintext, rb); e != nil {
		return nil, e
	}
	return
}

func (userdata *User) putRecordBook(UUID uuid.UUID, purpose []byte, rb *RecordBook, sk []byte, shared bool) {
	plaintext, e := json.Marshal(rb)
	if e != nil {
		return
	}
	if !shared {
		sk, _ = userlib.HashKDF(userdata.Sk[:16], purpose)
	}
	putSymData(UUID, sk, RECORD_BOOK_ENC_PURPOSE, RECORD_BOOK_MAC_PURPOSE, plaintext)
}

func (userdata *User) getFilePart(UUID uuid.UUID, purpose []byte, sk []byte, i int, shared bool) (data []byte, err error) {
	if !shared {
		sk, _ = userlib.HashKDF(userdata.Sk[:16], []byte(purpose))
	}
	data, e := getSymData(UUID, sk, FILE_PART_ENC_PURPOSE+string(i), FILE_PART_MAC_PURPOSE+string(i), nil, nil, false)
	if e != nil {
		return nil, e
	}
	return
}

func (userdata *User) putFilePart(UUID uuid.UUID, purpose []byte, i int, plaintext []byte, sk []byte, shared bool) {
	if !shared {
		sk, _ = userlib.HashKDF(userdata.Sk[:16], []byte(purpose))
	}
	putSymData(UUID, sk, FILE_PART_ENC_PURPOSE+string(i), FILE_PART_MAC_PURPOSE+string(i), plaintext)
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
	if _, ok := userlib.KeystoreGet(username + PASSWD_K); ok {
		return nil, errors.New("User exists")
	}
	// get UUID for userverification
	// UUID, _ := uuid.FromBytes(getPurpose(username + UUID_PASSWD))
	// 2 pairs of Access token keys
	atSignSk, atSignPk, _ := userlib.DSKeyGen()
	atEncPk, atEncSk, _ := userlib.PKEKeyGen()
	if userlib.KeystoreSet(username+AT_SIGN_K, atSignPk) != nil {
		return nil, errors.New("Failed to store atSignPk in keystore")
	}
	if userlib.KeystoreSet(username+AT_ENC_K, atEncPk) != nil {
		return nil, errors.New("Failed to store atEncPk in keystore")
	}
	// 1 pair of password verification keys
	pwSignKey, pwVerifyKey, _ := userlib.DSKeyGen()
	// Initialize and store User verification
	hashPw := userlib.Hash([]byte(password))
	hashPasswordSalt := userlib.RandomBytes(SALT_LEN)
	hashPassword := userlib.Argon2Key(hashPw[:], hashPasswordSalt, HASH_PW_LEN)
	skSalt := userlib.RandomBytes(SALT_LEN)
	uv := UserVerification{hashPassword, hashPasswordSalt, skSalt}
	userlib.KeystoreSet(username+PASSWD_K, pwVerifyKey)
	putUserVerification(username, pwSignKey, &uv)
	// Initialize and store User metadata
	sk := userlib.Argon2Key(hashPw[:], skSalt, SK_LEN)
	userdataptr.Username = username
	userdataptr.Password = password
	userdataptr.Sk = sk
	var userMetaData DatastoreUser
	userMetaData.OwnedFilePurpose = make(map[string][]byte, 0)
	userMetaData.SharedFileSecrecyKeypairs = make(map[string][2][]byte, 0)
	userMetaData.SharedFileSecrecyUUIDs = make(map[string]uuid.UUID, 0)
	userMetaData.ATokenDecSk = atEncSk
	userMetaData.ATokenSignSK = atSignSk
	userdataptr.putUserMetadata(&userMetaData)
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
	if userdata == nil {
		return
	}
	// get datastore user to check whether file exists and get purpose
	um, e := userdata.getUserMetadata()
	if e != nil {
		return
	}
	// update
	purpose, ok := um.OwnedFilePurpose[mapKey(filename)]
	var datastoreFile *DatastoreFile
	var rbUUID uuid.UUID
	var rb *RecordBook
	var shared bool
	sk := make([]byte, 0)
	filek := mapKey(filename)
	if ok {
		// file exists, we only need to modify record book, file_part len - 1
		if datastoreFile, e = userdata.getFileMetadata(filename); e != nil {
			return
		}
		if rb, e = userdata.getRecordBook(datastoreFile.RecordUUID, purpose, nil, false); e != nil {
			return
		}
		rbUUID = datastoreFile.RecordUUID
	}
	if secrecyUUID, ok := um.SharedFileSecrecyUUIDs[mapKey(filename)]; ok {
		keyPair := um.SharedFileSecrecyKeypairs[filek]
		ks, e := userdata.getFileSecrecy(secrecyUUID, "", "", keyPair, true)
		if e != nil {
			return
		}
		if rb, e = userdata.getRecordBook(ks.UUID, nil, ks.SK, true); e != nil {
			return
		}
		sk = ks.SK
		rbUUID = ks.UUID
		shared = true
	} else {
		// file doesn't exist, we need to modify user_metadata(purpose), file_metadata(), record_book, file_part0
		// initialize purpose
		purpose = userlib.RandomBytes(16)
		um.OwnedFilePurpose[mapKey(filename)] = purpose
		userdata.putUserMetadata(um)
		// initialize and modify file metadata
		_datastoreFile := DatastoreFile{uuid.New(), make(map[string]uuid.UUID, 0)}
		datastoreFile = &_datastoreFile
		userdata.putFileMetadata(filename, datastoreFile)
		// initialize record book
		_recordBook := RecordBook{make([]uuid.UUID, 0)}
		rb = &_recordBook
		rbUUID = datastoreFile.RecordUUID
	}
	// both need to modify record_book and file_part
	rb.Records = make([]uuid.UUID, 0)
	rb.Records = append(rb.Records, uuid.New())
	userdata.putRecordBook(rbUUID, purpose, rb, sk, shared)
	userdata.putFilePart(rb.Records[len(rb.Records)-1], purpose, len(rb.Records)-1, data, sk, shared)
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
	var rb *RecordBook
	var rbUUID uuid.UUID
	var shared bool
	filek := mapKey(filename)
	purpose, ok := um.OwnedFilePurpose[filek]
	sk := make([]byte, 0)
	if !ok {
		var secrecyUUID uuid.UUID
		if secrecyUUID, ok = um.SharedFileSecrecyUUIDs[filek]; !ok {
			return errors.New("File does not exist")
		}
		keyPair := um.SharedFileSecrecyKeypairs[filek]
		ks, e := userdata.getFileSecrecy(secrecyUUID, "", "", keyPair, true)
		if e != nil {
			return e
		}
		rbUUID = ks.UUID
		if e != nil {
			return e
		}
		if rb, e = userdata.getRecordBook(rbUUID, nil, ks.SK, true); e != nil {
			return e
		}
		sk = ks.SK
		shared = true
	} else {
		fm, e := userdata.getFileMetadata(filename)
		rbUUID = fm.RecordUUID
		if e != nil {
			return e
		}
		rb, e = userdata.getRecordBook(rbUUID, purpose, nil, false)
		if e != nil {
			return e
		}
	}
	UUID := uuid.New()
	rb.Records = append(rb.Records, UUID)
	userdata.putFilePart(UUID, purpose, len(rb.Records)-1, data, sk, shared)
	userdata.putRecordBook(rbUUID, purpose, rb, sk, shared)
	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	if userdata == nil {
		return nil, errors.New("Bad userdata ptr")
	}
	um, e := userdata.getUserMetadata()
	if e != nil {
		return nil, e
	}

	filek := mapKey(filename)
	purpose, ok := um.OwnedFilePurpose[mapKey(filename)]
	var rb *RecordBook
	var shared bool
	sk := make([]byte, 0)
	if !ok {
		// file sharing
		var secrecyUUID uuid.UUID
		if secrecyUUID, ok = um.SharedFileSecrecyUUIDs[filek]; !ok {
			return nil, errors.New("File does not exist")
		}
		keyPair, ok := um.SharedFileSecrecyKeypairs[filek]
		if !ok {
			return nil, errors.New("Load after share: Fail to get from map")
		}
		ks, e := userdata.getFileSecrecy(secrecyUUID, "", "", keyPair, true)
		if e != nil {
			return nil, errors.New("Load File: get File Secrecy " + e.Error())
		}
		userlib.DebugMsg("> Loadfile call getrecordbook, uuid: %v, sk: %v", ks.UUID, ks.SK)
		if rb, e = userdata.getRecordBook(ks.UUID, nil, ks.SK, true); e != nil {
			return nil, e
		}
		sk = ks.SK
		shared = true
	} else {
		fm, e := userdata.getFileMetadata(filename)
		if e != nil {
			return nil, errors.New("> getFileMetadata " + e.Error())
		}
		rb, e = userdata.getRecordBook(fm.RecordUUID, purpose, nil, false)
		if e != nil {
			return nil, errors.New("> getRecordBook " + e.Error())
		}
	}
	for i, UUID := range rb.Records {
		d, e := userdata.getFilePart(UUID, purpose, sk, i, shared)
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
	if userdata == nil {
		return
	}
	// check file existence
	var obj DataDS
	var at AccessToken
	var encPk userlib.PKEEncKey
	um, e := userdata.getUserMetadata()
	if e != nil {
		return "", e
	}
	purpose, ok := um.OwnedFilePurpose[mapKey(filename)]
	if ok {
		// file owned by user, create secrecy and construct access token
		var keySecrecy KeySecrecy
		fm, e := userdata.getFileMetadata(filename)
		if e != nil {
			return "", errors.New(e.Error() + "Failed to get " + filename + " metadata")
		}
		if _, ok = fm.SharedUser[recipient]; ok {
			return "", errors.New("File already shared with " + recipient)
		}
		fm.SharedUser[recipient] = uuid.New()
		userdata.putFileMetadata(filename, fm)
		// initialize and store secrecy key
		keySecrecy.SK, _ = userlib.HashKDF(userdata.Sk[:16], purpose)
		keySecrecy.UUID = fm.RecordUUID
		// initialize access token
		at.UUID = fm.SharedUser[recipient]
		encPurpose := filename + recipient + KEY_SECRECY_ENC_PURPOSE
		macPurpose := filename + recipient + KEY_SECRECY_MAC_PURPSOE
		at.SecrecyDecKey, at.SecrecyMacKey = getSymKeys(userdata.Sk, encPurpose, macPurpose)
		// userlib.DebugMsg("share file!!: p1: %v\np2: %v\n", encPurpose, macPurpose)
		// put file secrecy
		userdata.putFileSecrecy(fm.SharedUser[recipient], filename, recipient, &keySecrecy)
	} else {
		if at.UUID, ok = um.SharedFileSecrecyUUIDs[mapKey(filename)]; !ok {
			return "", errors.New("Failed to find file " + filename)
		}
		if _, e = userdata.getFileSecrecy(at.UUID, "", "", um.SharedFileSecrecyKeypairs[mapKey(filename)], true); e != nil {
			return "", errors.New("File has been revoked")
		}
		at.SecrecyDecKey = um.SharedFileSecrecyKeypairs[mapKey(filename)][0]
		at.SecrecyMacKey = um.SharedFileSecrecyKeypairs[mapKey(filename)][1]
		at.SecrecyDecKey = at.SecrecyDecKey[:16]
		at.SecrecyMacKey = at.SecrecyMacKey[:16]
	}
	_at := at.UUID.String() + string(at.SecrecyDecKey[:]) + string(at.SecrecyMacKey[:])
	if encPk, ok = userlib.KeystoreGet(recipient + AT_ENC_K); !ok {
		return "", errors.New("Failed to get recipient's enc Pk")
	}
	obj.Ciphertext, e = userlib.PKEEnc(encPk, []byte(_at))
	if e != nil {
		return "", errors.New("Failed to encrypt " + e.Error() + " ")
	}
	obj.Verifytext, _ = userlib.DSSign(um.ATokenSignSK, obj.Ciphertext)
	msg, e := json.Marshal(&obj)
	if e != nil {
		return "", errors.New("Failed to marshal data")
	}
	return hex.EncodeToString(msg), nil
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string, magic_string string) error {
	if userdata == nil {
		return errors.New("Bad userdata ptr")
	}
	var msg DataDS
	var at AccessToken
	data, e := hex.DecodeString(magic_string)
	if e != nil {
		return e
	}
	e = json.Unmarshal(data, &msg)
	if e != nil {
		return e
	}
	// use public key to verify
	verifyKey, ok := userlib.KeystoreGet(sender + AT_SIGN_K)
	if !ok {
		return errors.New("Failed to get verify ks from keystore")
	}
	if e := userlib.DSVerify(verifyKey, msg.Ciphertext, msg.Verifytext); e != nil {
		return e
	}
	// use own private key to decrypt
	um, e := userdata.getUserMetadata()
	if e != nil {
		return e
	}
	_at, _ := userlib.PKEDec(um.ATokenDecSk, msg.Ciphertext)
	s := string(_at[:])
	at.UUID, _ = uuid.Parse(s[:36])
	at.SecrecyDecKey = []byte(s[36:52])
	at.SecrecyMacKey = []byte(s[52:68])
	// update user metadata
	filek := mapKey(filename)
	if _, ok := um.OwnedFilePurpose[filek]; ok {
		return errors.New("File already exists, owned by user")
	}
	if _, ok := um.SharedFileSecrecyUUIDs[filek]; ok {
		return errors.New("File already exists, shared by ohters")
	}
	um.SharedFileSecrecyUUIDs[filek] = at.UUID
	um.SharedFileSecrecyKeypairs[filek] = [2][]byte{at.SecrecyDecKey[:], at.SecrecyMacKey[:]}
	userdata.putUserMetadata(um)
	return nil
}

// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	if userdata == nil {
		return errors.New("Bad userdata ptr")
	}
	var newPurpose = userlib.RandomBytes(PURPOSE_LEN)
	newSk, _ := userlib.HashKDF(userdata.Sk[:16], newPurpose)
	var newRBUUID = uuid.New()
	// check ownership
	um, e := userdata.getUserMetadata()
	if e != nil {
		return e
	}
	filek := mapKey(filename)
	if _, ok := um.OwnedFilePurpose[filek]; !ok {
		return errors.New("Doesn't have this file's ownership")
	}
	// get metadata and file
	fm, e := userdata.getFileMetadata(filename)
	if e != nil {
		return e
	}
	if _, ok := fm.SharedUser[target_username]; !ok {
		return errors.New("Not direct shared user")
	}
	f, e := userdata.LoadFile(filename)
	if e != nil {
		return e
	}
	// update user metadata, file metadata, record book and file
	// user metadata
	um.OwnedFilePurpose[mapKey(filename)] = newPurpose
	// file metadata
	fm.RecordUUID = newRBUUID
	userlib.DatastoreDelete(fm.SharedUser[target_username])
	delete(fm.SharedUser, target_username)
	userdata.putFileMetadata(filename, fm)
	// record book
	rb := RecordBook{make([]uuid.UUID, 0)}
	rb.Records = append(rb.Records, uuid.New())
	userdata.putUserMetadata(um)
	userdata.putFileMetadata(filename, fm)
	userdata.putRecordBook(fm.RecordUUID, newPurpose, &rb, nil, false)
	// update file
	userdata.putFilePart(rb.Records[0], newPurpose, 0, f, nil, false)
	// update other shared user's secrecy
	for username, UUID := range fm.SharedUser {
		var keySecrecy KeySecrecy
		keySecrecy.SK = newSk
		keySecrecy.UUID = newRBUUID
		userdata.putFileSecrecy(UUID, filename, username, &keySecrecy)
	}
	return
}
