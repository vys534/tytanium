package encryption

import (
	"crypto/sha512"
	"golang.org/x/crypto/hkdf"
	"io"
)

func DeriveKey(k []byte, nonce []byte) ([32]byte, error) {
	var key [32]byte

	// TODO: remove experiment
	//d := sha3.NewShake256()
	//_, err := d.Write(k)
	//if err != nil {
	//	return [64]byte{}, err
	//}
	//_, err = d.Write(nonce)
	//if err != nil {
	//	return [64]byte{}, err
	//}
	//
	//_, err = d.Read(key[:])
	//if err != nil {
	//	return [64]byte{}, err
	//}

	// SHA2-512 related code
	kdf := hkdf.New(sha512.New, k, nonce, nil)
	_, err := io.ReadFull(kdf, key[:])
	if err != nil {
		return [32]byte{}, err
	}

	return key, nil
}
