// Package appstore verifies Apple StoreKit 2 signed transactions (JWS) so the
// backend can grant Plus entitlements only for genuine, Apple-signed purchases.
//
// StoreKit 2 hands the app a `VerificationResult` whose `jwsRepresentation` is a
// JWS (ES256) signed by Apple, with the signing certificate chain in the `x5c`
// header. The client posts that JWS to POST /api/v1/subscription/verify; this
// package validates the certificate chain up to a trusted Apple root, checks the
// signature, and decodes the transaction payload (product, expiry, environment).
//
// Production must supply Apple's root ("Apple Root CA - G3", available from
// https://www.apple.com/certificateauthority/) via LIM_APPLE_ROOT_CERT. Tests
// inject their own generated root to exercise the full verification path.
package appstore

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Transaction is the subset of Apple's JWSTransactionDecodedPayload we use.
type Transaction struct {
	TransactionID         string
	OriginalTransactionID string
	ProductID             string
	BundleID              string
	Environment           string // "Production" | "Sandbox"
	Type                  string // e.g. "Auto-Renewable Subscription"
	PurchaseDate          time.Time
	ExpiresDate           time.Time
}

// Verifier checks signed transactions against a trusted root pool.
type Verifier struct {
	roots    *x509.CertPool
	bundleID string // if non-empty, payload.bundleId must match
	env      string // if non-empty, payload.environment must match
}

// New builds a Verifier. rootPEMs are one or more PEM-encoded root certificates
// (Apple's root in production). bundleID/env, when non-empty, are enforced.
func New(rootPEMs [][]byte, bundleID, env string) (*Verifier, error) {
	pool := x509.NewCertPool()
	for _, pem := range rootPEMs {
		if !pool.AppendCertsFromPEM(pem) {
			return nil, errors.New("appstore: invalid root certificate PEM")
		}
	}
	return &Verifier{roots: pool, bundleID: bundleID, env: env}, nil
}

// HasRoots reports whether any trusted root is configured.
func (v *Verifier) HasRoots() bool {
	return v.roots != nil && len(v.roots.Subjects()) > 0 //nolint:staticcheck // Subjects ok for count
}

// Notification is the subset of App Store Server Notification V2 we act on.
type Notification struct {
	Type        string // e.g. DID_RENEW, EXPIRED, REFUND, SUBSCRIBED
	Subtype     string
	UUID        string
	BundleID    string
	Environment string
	Transaction *Transaction // decoded from data.signedTransactionInfo (nil if absent)
}

// verifyClaims validates a JWS (chain + ES256 signature) and returns its claims.
func (v *Verifier) verifyClaims(jws string) (jwt.MapClaims, error) {
	if !v.HasRoots() {
		return nil, errors.New("appstore: no trusted root configured (set LIM_APPLE_ROOT_CERT)")
	}
	var claims jwt.MapClaims
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"ES256"}))
	_, err := parser.ParseWithClaims(jws, &claims, func(t *jwt.Token) (any, error) {
		leaf, err := v.verifyChain(t)
		if err != nil {
			return nil, err
		}
		pub, ok := leaf.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("appstore: leaf key is not ECDSA")
		}
		return pub, nil
	})
	if err != nil {
		return nil, fmt.Errorf("appstore: verify: %w", err)
	}
	return claims, nil
}

// Verify validates a JWS signed transaction and returns its decoded payload.
func (v *Verifier) Verify(jws string) (*Transaction, error) {
	claims, err := v.verifyClaims(jws)
	if err != nil {
		return nil, err
	}
	txn := toTransaction(claims)
	if v.bundleID != "" && txn.BundleID != "" && txn.BundleID != v.bundleID {
		return nil, fmt.Errorf("appstore: bundle id mismatch: %q", txn.BundleID)
	}
	if v.env != "" && txn.Environment != "" && txn.Environment != v.env {
		return nil, fmt.Errorf("appstore: environment mismatch: %q", txn.Environment)
	}
	return txn, nil
}

// VerifyNotification validates an App Store Server Notification V2 signedPayload
// and decodes it, including the nested signed transaction (also verified).
func (v *Verifier) VerifyNotification(signedPayload string) (*Notification, error) {
	claims, err := v.verifyClaims(signedPayload)
	if err != nil {
		return nil, err
	}
	n := &Notification{
		Type:    str(claims, "notificationType"),
		Subtype: str(claims, "subtype"),
		UUID:    str(claims, "notificationUUID"),
	}
	if data, ok := claims["data"].(map[string]any); ok {
		dc := jwt.MapClaims(data)
		n.BundleID = str(dc, "bundleId")
		n.Environment = str(dc, "environment")
		if sti := str(dc, "signedTransactionInfo"); sti != "" {
			if txn, err := v.Verify(sti); err == nil {
				n.Transaction = txn
			}
		}
	}
	return n, nil
}

func toTransaction(claims jwt.MapClaims) *Transaction {
	return &Transaction{
		TransactionID:         str(claims, "transactionId"),
		OriginalTransactionID: str(claims, "originalTransactionId"),
		ProductID:             str(claims, "productId"),
		BundleID:              str(claims, "bundleId"),
		Environment:           str(claims, "environment"),
		Type:                  str(claims, "type"),
		PurchaseDate:          ms(claims, "purchaseDate"),
		ExpiresDate:           ms(claims, "expiresDate"),
	}
}

// verifyChain validates the x5c certificate chain in the JWS header up to a
// trusted root and returns the leaf certificate.
func (v *Verifier) verifyChain(t *jwt.Token) (*x509.Certificate, error) {
	raw, ok := t.Header["x5c"].([]any)
	if !ok || len(raw) == 0 {
		return nil, errors.New("appstore: missing x5c header")
	}
	var certs []*x509.Certificate
	for _, e := range raw {
		b64, ok := e.(string)
		if !ok {
			return nil, errors.New("appstore: malformed x5c entry")
		}
		der, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("appstore: x5c base64: %w", err)
		}
		c, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, fmt.Errorf("appstore: x5c parse: %w", err)
		}
		certs = append(certs, c)
	}

	leaf := certs[0]
	intermediates := x509.NewCertPool()
	for _, c := range certs[1:] {
		intermediates.AddCert(c)
	}
	if _, err := leaf.Verify(x509.VerifyOptions{
		Roots:         v.roots,
		Intermediates: intermediates,
		// Apple's chain isn't a TLS server chain; accept any EKU.
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		return nil, fmt.Errorf("appstore: chain verify: %w", err)
	}
	return leaf, nil
}

func str(m jwt.MapClaims, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

// ms reads an Apple millisecond-epoch field into a time.Time.
func ms(m jwt.MapClaims, k string) time.Time {
	switch v := m[k].(type) {
	case float64:
		return time.UnixMilli(int64(v))
	case int64:
		return time.UnixMilli(v)
	default:
		return time.Time{}
	}
}
