package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

var globalMediaCertificate *tls.Certificate = nil
var globalMediaPem string

func makeRandomSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}

func GenerateX509Cert(sn *big.Int, from, to time.Time, IPAddresses []net.IP, DNSNames []string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber:          sn,
		Subject:               pkix.Name{Organization: []string{"Self-signed cert"}},
		NotBefore:             from,
		NotAfter:              to,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           IPAddresses,
		DNSNames:              DNSNames,
	}
}

func GenerateX509PEMs(cert *x509.Certificate, key *ecdsa.PrivateKey) (certPem, keyPem []byte, err error) {
	derBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return
	}
	certPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return
	}
	keyPem = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	return
}

func GenerateTLSCert(notBefore, notAfter time.Time, IPAddresses []net.IP, DNSNames []string) (*tls.Certificate, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	sn, err := makeRandomSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	cert := GenerateX509Cert(sn, notBefore, notAfter, IPAddresses, DNSNames)
	certPem, keyPem, err := GenerateX509PEMs(cert, priv)
	if err != nil {
		return nil, nil, err
	}

	finalCert, err := tls.X509KeyPair(certPem, keyPem)
	return &finalCert, certPem, err
}

func generateMediaTLSCert() error {
	if globalMediaCertificate != nil {
		return nil
	}

	now := time.Now()
	notBefore := now.Add(-365 * 24 * time.Hour * 100)
	notAfter := now.Add(365 * 24 * time.Hour * 100)
	log.Debug("generate media cert", "system time", time.Now().String(), "cert notBefore", notBefore.String(), "cert notAfter", notAfter.String())
	finalCert, certPem, err := GenerateTLSCert(notBefore, notAfter, []net.IP{}, []string{Localhost})
	if err != nil {
		return err
	}

	globalMediaCertificate = finalCert
	globalMediaPem = string(certPem)
	return nil
}

func PublicMediaTLSCert() (string, error) {
	err := generateMediaTLSCert()
	if err != nil {
		return "", err
	}

	return globalMediaPem, nil
}

// ToECDSA takes a []byte of D and uses it to create an ecdsa.PublicKey on the elliptic.P256 curve
// this function is basically a P256 curve version of eth-node/crypto.ToECDSA without all the nice validation
func ToECDSA(d []byte) *ecdsa.PrivateKey {
	k := new(ecdsa.PrivateKey)
	k.D = new(big.Int).SetBytes(d)
	k.PublicKey.Curve = elliptic.P256()

	k.PublicKey.X, k.PublicKey.Y = k.PublicKey.Curve.ScalarBaseMult(d)
	return k
}
