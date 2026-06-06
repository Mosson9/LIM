package appstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// certNode is a generated certificate plus its key.
type certNode struct {
	cert *x509.Certificate
	der  []byte
	key  *ecdsa.PrivateKey
}

func mkCert(t *testing.T, cn string, parent *certNode, isCA bool) *certNode {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}
	if isCA {
		tmpl.KeyUsage |= x509.KeyUsageCertSign
	}
	signer, signerKey := tmpl, key
	if parent != nil {
		signer, signerKey = parent.cert, parent.key
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, signer, &key.PublicKey, signerKey)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return &certNode{cert: cert, der: der, key: key}
}

// signJWS builds an Apple-style JWS for the payload, signed by leaf, with the
// given x5c chain (leaf → … → root).
func signJWS(t *testing.T, payload jwt.MapClaims, leaf *certNode, chain []*certNode) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, payload)
	var x5c []string
	for _, c := range chain {
		x5c = append(x5c, base64.StdEncoding.EncodeToString(c.der))
	}
	tok.Header["x5c"] = x5c
	s, err := tok.SignedString(leaf.key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func rootPEM(t *testing.T, root *certNode) []byte {
	t.Helper()
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: root.der})
}

func validPayload() jwt.MapClaims {
	return jwt.MapClaims{
		"transactionId":         "2000000123456789",
		"originalTransactionId": "2000000123456789",
		"productId":             "app.lim.ios.plus.yearly",
		"bundleId":              "app.lim.ios",
		"environment":           "Sandbox",
		"type":                  "Auto-Renewable Subscription",
		"purchaseDate":          float64(time.Now().UnixMilli()),
		"expiresDate":           float64(time.Now().Add(365 * 24 * time.Hour).UnixMilli()),
	}
}

func TestVerifyValidChain(t *testing.T) {
	root := mkCert(t, "Test Apple Root", nil, true)
	inter := mkCert(t, "Test Apple WWDR", root, true)
	leaf := mkCert(t, "Test Apple Leaf", inter, false)

	jws := signJWS(t, validPayload(), leaf, []*certNode{leaf, inter, root})
	v, err := New([][]byte{rootPEM(t, root)}, "app.lim.ios", "Sandbox")
	if err != nil {
		t.Fatal(err)
	}
	txn, err := v.Verify(jws)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if txn.ProductID != "app.lim.ios.plus.yearly" {
		t.Errorf("productId = %q", txn.ProductID)
	}
	if txn.ExpiresDate.Before(time.Now()) {
		t.Errorf("expiresDate not in the future: %v", txn.ExpiresDate)
	}
}

func TestVerifyUntrustedRootRejected(t *testing.T) {
	root := mkCert(t, "Real Root", nil, true)
	inter := mkCert(t, "Inter", root, true)
	leaf := mkCert(t, "Leaf", inter, false)
	jws := signJWS(t, validPayload(), leaf, []*certNode{leaf, inter, root})

	// Verifier trusts a DIFFERENT root → chain must fail.
	other := mkCert(t, "Other Root", nil, true)
	v, _ := New([][]byte{rootPEM(t, other)}, "app.lim.ios", "")
	if _, err := v.Verify(jws); err == nil {
		t.Fatal("expected verification to fail for untrusted root")
	}
}

func TestVerifyTamperedSignatureRejected(t *testing.T) {
	root := mkCert(t, "Root", nil, true)
	inter := mkCert(t, "Inter", root, true)
	leaf := mkCert(t, "Leaf", inter, false)
	jws := signJWS(t, validPayload(), leaf, []*certNode{leaf, inter, root})

	tampered := jws[:len(jws)-3] + "AAA"
	v, _ := New([][]byte{rootPEM(t, root)}, "", "")
	if _, err := v.Verify(tampered); err == nil {
		t.Fatal("expected verification to fail for tampered signature")
	}
}

func TestVerifyBundleMismatchRejected(t *testing.T) {
	root := mkCert(t, "Root", nil, true)
	inter := mkCert(t, "Inter", root, true)
	leaf := mkCert(t, "Leaf", inter, false)
	jws := signJWS(t, validPayload(), leaf, []*certNode{leaf, inter, root})

	v, _ := New([][]byte{rootPEM(t, root)}, "com.evil.app", "")
	if _, err := v.Verify(jws); err == nil {
		t.Fatal("expected verification to fail for bundle id mismatch")
	}
}

func TestVerifyNoRootConfigured(t *testing.T) {
	v, _ := New(nil, "", "")
	if v.HasRoots() {
		t.Fatal("expected no roots")
	}
	if _, err := v.Verify("x.y.z"); err == nil {
		t.Fatal("expected error when no root configured")
	}
}
