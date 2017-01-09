/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric/common/metadata"
	"golang.org/x/crypto/sha3"
)

type alg struct {
	hashFun func([]byte) string
}

const defaultAlg = "sha256"

var availableIDgenAlgs = map[string]alg{
	defaultAlg: {GenerateIDfromTxSHAHash},
}

// ComputeCryptoHash should be used in openchain code so that we can change the actual algo used for crypto-hash at one place
func ComputeCryptoHash(data []byte) (hash []byte) {
	hash = make([]byte, 64)
	sha3.ShakeSum256(hash, data)
	return
}

// GenerateBytesUUID returns a UUID based on RFC 4122 returning the generated bytes
func GenerateBytesUUID() []byte {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		panic(fmt.Sprintf("Error generating UUID: %s", err))
	}

	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80

	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40

	return uuid
}

// GenerateIntUUID returns a UUID based on RFC 4122 returning a big.Int
func GenerateIntUUID() *big.Int {
	uuid := GenerateBytesUUID()
	z := big.NewInt(0)
	return z.SetBytes(uuid)
}

// GenerateUUID returns a UUID based on RFC 4122
func GenerateUUID() string {
	uuid := GenerateBytesUUID()
	return idBytesToStr(uuid)
}

// CreateUtcTimestamp returns a google/protobuf/Timestamp in UTC
func CreateUtcTimestamp() *timestamp.Timestamp {
	now := time.Now().UTC()
	secs := now.Unix()
	nanos := int32(now.UnixNano() - (secs * 1000000000))
	return &(timestamp.Timestamp{Seconds: secs, Nanos: nanos})
}

//GenerateHashFromSignature returns a hash of the combined parameters
func GenerateHashFromSignature(path string, args []byte) []byte {
	return ComputeCryptoHash(args)
}

// GenerateIDfromTxSHAHash generates SHA256 hash using Tx payload
func GenerateIDfromTxSHAHash(payload []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(payload))
}

// GenerateIDWithAlg generates an ID using a custom algorithm
func GenerateIDWithAlg(customIDgenAlg string, payload []byte) (string, error) {
	if customIDgenAlg == "" {
		customIDgenAlg = defaultAlg
	}
	var alg = availableIDgenAlgs[customIDgenAlg]
	if alg.hashFun != nil {
		return alg.hashFun(payload), nil
	}
	return "", fmt.Errorf("Wrong ID generation algorithm was given: %s", customIDgenAlg)
}

func idBytesToStr(id []byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:])
}

// FindMissingElements identifies the elements of the first slice that are not present in the second
// The second slice is expected to be a subset of the first slice
func FindMissingElements(all []string, some []string) (delta []string) {
all:
	for _, v1 := range all {
		for _, v2 := range some {
			if strings.Compare(v1, v2) == 0 {
				continue all
			}
		}
		delta = append(delta, v1)
	}
	return
}

// ToChaincodeArgs converts string args to []byte args
func ToChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// ArrayToChaincodeArgs converts array of string args to array of []byte args
func ArrayToChaincodeArgs(args []string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

const testchainid = "**TEST_CHAINID**"
const testorgid = "**TEST_ORGID**"

//GetTestChainID returns the CHAINID constant in use by orderer
func GetTestChainID() string {
	return testchainid
}

//GetTestOrgID returns the ORGID constant in use by gossip join message
func GetTestOrgID() string {
	return testorgid
}

//GetSysCCVersion returns the version of all system chaincodes
//This needs to be revisited on policies around system chaincode
//"upgrades" from user and relationship with "fabric" upgrade. For
//now keep it simple and use the fabric's version stamp
func GetSysCCVersion() string {
	return metadata.Version
}
