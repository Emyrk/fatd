package fat

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"

	"golang.org/x/crypto/ed25519"
)

// Entry has variables and methods common to all fat0 entries.
type Entry struct {
	Metadata json.RawMessage `json:"metadata,omitempty"`

	factom.Entry `json:"-"`
}

// UnmarshalEntry unmarshals the content of the factom.Entry into the provided
// variable v, disallowing all unknown fields.
func (e Entry) UnmarshalEntry(v interface{}) error {
	return json.Unmarshal(e.Content, v)
}

func (e Entry) MetadataJSONLen() int {
	if e.Metadata == nil {
		return 0
	}
	return len(`,"metadata":`) + len(e.Metadata)
}
func (e *Entry) MarshalEntry(v interface{}) error {
	var err error
	e.Content, err = json.Marshal(v)
	return err
}

// ValidExtIDs validates the structure of the ExtIDs of the factom.Entry to
// make sure that it has a valid timestamp salt and a valid set of
// RCD/signature pairs.
func (e Entry) ValidExtIDs(numRCDSigPairs int) error {
	if numRCDSigPairs == 0 || len(e.ExtIDs) != 2*numRCDSigPairs+1 {
		return fmt.Errorf("invalid number of ExtIDs")
	}
	if err := e.validTimestamp(); err != nil {
		return err
	}
	extIDs := e.ExtIDs[1:]
	for i := 0; i < len(extIDs)/2; i++ {
		rcd := extIDs[i*2]
		if len(rcd) != factom.RCDSize {
			return fmt.Errorf("ExtIDs[%v]: invalid RCD size", i+1)
		}
		if rcd[0] != factom.RCDType {
			return fmt.Errorf("ExtIDs[%v]: invalid RCD type", i+1)
		}
		sig := extIDs[i*2+1]
		if len(sig) != factom.SignatureSize {
			return fmt.Errorf("ExtIDs[%v]: invalid signature size", i+1)
		}
	}
	return e.validSignatures()
}
func (e Entry) validTimestamp() error {
	sec, err := strconv.ParseInt(string(e.ExtIDs[0]), 10, 64)
	if err != nil {
		return fmt.Errorf("timestamp salt: %v", err)
	}
	ts := time.Unix(sec, 0)
	diff := e.Timestamp.Time().Sub(ts)
	if -12*time.Hour > diff || diff > 12*time.Hour {
		return fmt.Errorf("timestamp salt expired")
	}
	return nil
}
func (e Entry) validSignatures() error {
	// Compose the signed message data using exactly allocated bytes.
	numRcdSigPairs := len(e.ExtIDs) / 2
	maxRcdSigIDSalt := numRcdSigPairs - 1
	maxRcdSigIDSaltStrLen := jsonlen.Uint64(uint64(maxRcdSigIDSalt))
	timeSalt := e.ExtIDs[0]
	maxMsgLen := maxRcdSigIDSaltStrLen + len(timeSalt) + len(e.ChainID) + len(e.Content)
	msg := make([]byte, maxMsgLen)
	i := maxRcdSigIDSaltStrLen
	i += copy(msg[i:], timeSalt[:])
	i += copy(msg[i:], e.ChainID[:])
	copy(msg[i:], e.Content)

	rcdSigs := e.ExtIDs[1:] // Skip over timestamp salt in ExtID[0]
	for rcdSigID := 0; rcdSigID < numRcdSigPairs; rcdSigID++ {
		// Prepend the RCD Sig ID Salt to the message data
		rcdSigIDSaltStr := strconv.FormatUint(uint64(rcdSigID), 10)
		start := maxRcdSigIDSaltStrLen - len(rcdSigIDSaltStr)
		copy(msg[start:], rcdSigIDSaltStr)

		msgHash := sha512.Sum512(msg[start:])
		pubKey := []byte(rcdSigs[rcdSigID*2][1:]) // Omit RCD Type byte
		sig := rcdSigs[rcdSigID*2+1]
		if !ed25519.Verify(pubKey, msgHash[:], sig) {
			return fmt.Errorf("ExtIDs[%v]: invalid signature", rcdSigID*2+2)
		}
	}
	return nil
}

// Sign the RCD/Sig ID Salt + Timestamp Salt + Chain ID Salt + Content of the
// factom.Entry and add the RCD + signature pairs for the given addresses to
// the ExtIDs. This clears any existing ExtIDs.
func (e *Entry) Sign(signingSet ...factom.RCDPrivateKey) {
	// Set the Entry's timestamp so that the signatures will verify against
	// this time salt.
	timeSalt := newTimestampSalt()
	e.Timestamp = new(factom.Time)
	*e.Timestamp = factom.Time(time.Now())

	// Compose the signed message data using exactly allocated bytes.
	maxRcdSigIDSaltStrLen := jsonlen.Uint64(uint64(len(signingSet)))
	maxMsgLen := maxRcdSigIDSaltStrLen + len(timeSalt) + len(e.ChainID) + len(e.Content)
	msg := make(factom.Bytes, maxMsgLen)
	i := maxRcdSigIDSaltStrLen
	i += copy(msg[i:], timeSalt[:])
	i += copy(msg[i:], e.ChainID[:])
	copy(msg[i:], e.Content)

	// Generate the ExtIDs for each address in the signing set.
	e.ExtIDs = make([]factom.Bytes, 1, len(signingSet)*2+1)
	e.ExtIDs[0] = timeSalt
	for rcdSigID, a := range signingSet {
		// Compose the RcdSigID salt and prepend it to the message.
		rcdSigIDSalt := strconv.FormatUint(uint64(rcdSigID), 10)
		start := maxRcdSigIDSaltStrLen - len(rcdSigIDSalt)
		copy(msg[start:], rcdSigIDSalt)

		msgHash := sha512.Sum512(msg[start:])
		sig := ed25519.Sign(a.PrivateKey(), msgHash[:])
		e.ExtIDs = append(e.ExtIDs, a.RCD(), sig)
	}
}
func newTimestampSalt() []byte {
	timestamp := time.Now().Add(time.Duration(-rand.Int63n(int64(1 * time.Hour))))
	return []byte(strconv.FormatInt(timestamp.Unix(), 10))
}

// FAAddress computes the FAAddress corresponding to the rcdSigID'th RCD/Sig
// pair.
func (e Entry) FAAddress(rcdSigID int) factom.FAAddress {
	id := rcdSigID*2 + 1
	return factom.FAAddress(sha256d(e.ExtIDs[id]))
}

// sha256d computes two rounds of the sha256 hash.
func sha256d(data []byte) [sha256.Size]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}
