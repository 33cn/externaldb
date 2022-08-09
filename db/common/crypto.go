package common

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"errors"

	"github.com/33cn/chain33/common"
)

const (
	CharSet             = "UTF-8"
	Base64Format        = "UrlSafeNoPadding"
	RsaAlgorithmKeyType = "PKCS8"
	RsaAlgorithmSign    = crypto.SHA256
	// 定义支持的加密算法
	RsaCrypto = "rsa"
	Encrypt   = "encrypt"
	Template  = "template"
)

type XRsa struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

var keyLength = 128

// 生成密钥对
func CreateKeys() (string, string) {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil {
		panic(err)
	}
	derStream := MarshalPKCS8PrivateKey(privateKey)
	privkey := common.ToHex(derStream)

	// 生成公钥文件
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	pubkey := common.ToHex(derPkix)
	return privkey, pubkey
}
func NewXRsa(privateKey string, publicKey string) (*XRsa, error) {
	pub, err := pubStrToPubKey(publicKey)
	if err != nil {
		return nil, err
	}

	pri, err := privStrToPrivKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &XRsa{publicKey: pub, privateKey: pri}, nil
}

// 公钥加密
func (r *XRsa) PublicEncrypt(data string) (string, error) {
	partLen := r.publicKey.N.BitLen()/8 - 11
	chunks := split([]byte(data), partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		bytes, err := rsa.EncryptPKCS1v15(rand.Reader, r.publicKey, chunk)
		if err != nil {
			return "", err
		}
		buffer.Write(bytes)
	}
	return base64.RawURLEncoding.EncodeToString(buffer.Bytes()), nil
}

// 私钥解密
func (r *XRsa) PrivateDecrypt(encrypted string) (string, error) {
	partLen := r.publicKey.N.BitLen() / 8
	raw, err := base64.RawURLEncoding.DecodeString(encrypted)
	chunks := split(raw, partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, chunk)
		if err != nil {
			return "", err
		}
		buffer.Write(decrypted)
	}
	return buffer.String(), err
}

//pubStrToPubKey：将hex格式的公钥转换成rsa.PublicKey类型
func pubStrToPubKey(pubstr string) (*rsa.PublicKey, error) {
	pubkey, err := common.FromHex(pubstr)
	if err != nil {
		return nil, err
	}
	pubInterface, err := x509.ParsePKIXPublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)

	return pub, nil
}

//privStrToPrivKey：将hex格式的私钥转换成rsa.PrivateKey类型
func privStrToPrivKey(privStr string) (*rsa.PrivateKey, error) {
	privkey, err := common.FromHex(privStr)
	if err != nil {
		return nil, err
	}

	priv, err := x509.ParsePKCS8PrivateKey(privkey)
	if err != nil {
		return nil, err
	}
	pri, ok := priv.(*rsa.PrivateKey)
	if ok {
		return pri, nil
	}
	return nil, errors.New("private key not supported")
}

//  通过公钥加密
func PublicEncrypt(pubstr string, data string) (string, error) {

	pubKey, err := pubStrToPubKey(pubstr)
	if err != nil {
		return "", err
	}
	partLen := pubKey.N.BitLen()/8 - 11
	chunks := split([]byte(data), partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		bytes, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, chunk)
		if err != nil {
			return "", err
		}
		buffer.Write(bytes)
	}
	return base64.RawURLEncoding.EncodeToString(buffer.Bytes()), nil
}

// 私钥解密
func PrivateDecrypt(privstr string, encrypted string) (string, error) {
	privKey, err := privStrToPrivKey(privstr)
	if err != nil {
		return "", err
	}
	partLen := keyLength / 8
	raw, err := base64.RawURLEncoding.DecodeString(encrypted)
	chunks := split(raw, partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, chunk)
		if err != nil {
			return "", err
		}
		buffer.Write(decrypted)
	}
	return buffer.String(), err
}

// 数据加签
func (r *XRsa) Sign(data string) (string, error) {
	h := RsaAlgorithmSign.New()
	h.Write([]byte(data))
	hashed := h.Sum(nil)
	sign, err := rsa.SignPKCS1v15(rand.Reader, r.privateKey, RsaAlgorithmSign, hashed)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sign), err
}

// 数据验签
func (r *XRsa) Verify(data string, sign string) error {
	h := RsaAlgorithmSign.New()
	h.Write([]byte(data))
	hashed := h.Sum(nil)
	decodedSign, err := base64.RawURLEncoding.DecodeString(sign)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(r.publicKey, RsaAlgorithmSign, hashed, decodedSign)
}
func MarshalPKCS8PrivateKey(key *rsa.PrivateKey) []byte {
	info := struct {
		Version             int
		PrivateKeyAlgorithm []asn1.ObjectIdentifier
		PrivateKey          []byte
	}{}
	info.Version = 0
	info.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 1)
	info.PrivateKeyAlgorithm[0] = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	info.PrivateKey = x509.MarshalPKCS1PrivateKey(key)
	k, _ := asn1.Marshal(info)
	return k
}
func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}

// DecodeString returns the bytes represented by the base64 string s.
func DecodeString(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)

}

// EncodeToString returns the base64 encoding of src.
func EncodeToString(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)

}
