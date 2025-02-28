package pki

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type ServicesService struct {
	client *Client

	validate *validator.Validate
}

type CertificateRequest struct {
	CommonName        string `json:"common_name" validate:"required,max=253"`
	AltNames          string `json:"alt_names,omitempty"`
	IPSANS            string `json:"ip_sans,omitempty"`
	URISANS           string `json:"uri_sans,omitempty"`
	OtherSANS         string `json:"other_sans,omitempty"`
	TTL               string `json:"ttl,omitempty"`
	Format            string `json:"format,omitempty"`
	PrivateKeyFormat  string `json:"private_key_format,omitempty"`
	ExcludeCNFromSANS *bool  `json:"exclude_cn_from_sans,omitempty"`
}

type IssueData struct {
	CaChain        []string `json:"ca_chain,omitempty"`
	Certificate    string   `json:"certificate,omitempty"`
	Expiration     int      `json:"expiration,omitempty"`
	IssuingCa      string   `json:"issuing_ca,omitempty"`
	PrivateKey     string   `json:"private_key,omitempty"`
	PrivateKeyType string   `json:"private_key_type,omitempty"`
	SerialNumber   string   `json:"serial_number,omitempty"`
}

// IssueResponse
type IssueResponse struct {
	RequestID     string    `json:"request_id"`
	LeaseID       string    `json:"lease_id"`
	Renewable     bool      `json:"renewable"`
	LeaseDuration int       `json:"lease_duration"`
	Data          IssueData `json:"data"`
	WrapInfo      *string   `json:"wrap_info,omitempty"`
	Warnings      *string   `json:"warnings,omitempty"`
	Auth          *string   `json:"auth,omitempty"`
}

// RevokeResponse
type RevokeResponse struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Data          struct {
		RevocationTime        int       `json:"revocation_time"`
		RevocationTimeRfc3339 time.Time `json:"revocation_time_rfc3339"`
	} `json:"data"`
	WrapInfo *string `json:"wrap_info,omitempty"`
	Warnings *string `json:"warnings,omitempty"`
	Auth     *string `json:"auth,omitempty"`
}

// SignRequest
type SignRequest struct {
	CSR               string `json:"csr" validation:"required"`
	CommonName        string `json:"common_name" validation:"required"`
	AltNames          string `json:"alt_names"`
	OtherSans         string `json:"other_sans"`
	IPSans            string `json:"ip_sans"`
	URISans           string `json:"uri_sans"`
	TTL               string `json:"ttl,omitempty"`
	Format            string `json:"format" validation:"required"  enum:"pem|der|pem_bundle"`
	ExcludeCNFromSans bool   `json:"exclude_cn_from_sans"`
}

// ServiceOptions
type ServiceOptions struct {
}

// GetRootCA
func (c *ServicesService) GetRootCA(options ...OptionFunc) (*x509.Certificate, *pem.Block, *Response, error) {
	options = append(options, func(req *http.Request) error {
		req.Header.Del("Authorization") // Remove authorization header
		return nil
	})
	return c.getCA("core/pki/api/root/ca/pem", options...)
}

// GetPolicyCA
func (c *ServicesService) GetPolicyCA(options ...OptionFunc) (*x509.Certificate, *pem.Block, *Response, error) {
	options = append(options, func(req *http.Request) error {
		req.Header.Del("Authorization") // Remove authorization header
		return nil
	})
	return c.getCA("core/pki/api/policy/ca/pem", options...)
}

func (c *ServicesService) getCA(path string, options ...OptionFunc) (*x509.Certificate, *pem.Block, *Response, error) {
	req, err := c.client.newServiceRequest(http.MethodGet, path, nil, options)
	if err != nil {
		return nil, nil, nil, err
	}
	resp, err := c.client.do(req, nil)
	if err != nil {
		return nil, nil, resp, err
	}
	if resp == nil {
		return nil, nil, nil, fmt.Errorf("getCA: %w", ErrEmptyResult)
	}
	defer resp.Body.Close()
	pemData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, resp, err
	}
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, resp, ErrCertificateExpected
	}
	pub, err := x509.ParseCertificate(block.Bytes)
	return pub, block, resp, err
}

// GetRootCRL
func (c *ServicesService) GetRootCRL(options ...OptionFunc) (*pkix.CertificateList, *pem.Block, *Response, error) {
	options = append(options, func(req *http.Request) error {
		req.Header.Del("Authorization") // Remove authorization header
		return nil
	})
	return c.getCRL("core/pki/api/root/crl/pem", options...)
}

// GetPolicyCRL
func (c *ServicesService) GetPolicyCRL(options ...OptionFunc) (*pkix.CertificateList, *pem.Block, *Response, error) {
	options = append(options, func(req *http.Request) error {
		req.Header.Del("Authorization") // Remove authorization header
		return nil
	})
	return c.getCRL("core/pki/api/policy/crl/pem", options...)
}

func (c *ServicesService) getCRL(path string, options ...OptionFunc) (*pkix.CertificateList, *pem.Block, *Response, error) {
	req, err := c.client.newServiceRequest(http.MethodGet, path, nil, options)
	if err != nil {
		return nil, nil, nil, err
	}
	resp, err := c.client.do(req, nil)
	if err != nil {
		return nil, nil, resp, err
	}
	if resp == nil {
		return nil, nil, resp, fmt.Errorf("getCRL: %w", ErrEmptyResult)
	}
	defer resp.Body.Close()
	pemData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, resp, err
	}
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "X509 CRL" {
		return nil, nil, resp, ErrCRLExpected
	}
	pub, err := x509.ParseCRL(block.Bytes)
	return pub, block, resp, err
}

func (d *IssueData) GetCertificate() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(d.Certificate))
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, ErrCertificateExpected
	}
	return x509.ParseCertificate(block.Bytes)
}

func (d *IssueData) GetPrivateKey() (interface{}, error) {
	block, _ := pem.Decode([]byte(d.PrivateKey))
	if block == nil {
		return nil, ErrInvalidPrivateKey
	}
	switch d.PrivateKeyType {
	case "rsa":
		if block.Type != "RSA PRIVATE KEY" {
			return nil, ErrInvalidPrivateKey
		}
		private, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return private, nil
	case "ec":
		if block.Type != "EC PRIVATE KEY" {
			return nil, ErrInvalidPrivateKey
		}
		private, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return private, nil
	}
	return nil, ErrInvalidPrivateKey
}

// Sign
func (c *ServicesService) Sign(logicalPath, roleName string, signRequest SignRequest, options ...OptionFunc) (*IssueResponse, *Response, error) {
	if err := c.validate.Struct(signRequest); err != nil {
		return nil, nil, err
	}
	req, err := c.client.newServiceRequest(http.MethodPost, "core/pki/api/"+logicalPath+"/sign/"+roleName, &signRequest, options)
	if err != nil {
		return nil, nil, err
	}
	var responseStruct struct {
		IssueResponse
		ErrorResponse
	}
	resp, err := c.client.do(req, &responseStruct)
	if err != nil {
		return nil, resp, err
	}
	if resp == nil {
		return nil, resp, fmt.Errorf("Sign: %w", ErrEmptyResult)
	}
	return &responseStruct.IssueResponse, resp, nil
}

// IssueCertificate
func (c *ServicesService) IssueCertificate(logicalPath, roleName string, request CertificateRequest, options ...OptionFunc) (*IssueResponse, *Response, error) {
	req, err := c.client.newServiceRequest(http.MethodPost, "core/pki/api/"+logicalPath+"/issue/"+roleName, &request, options)
	if err != nil {
		return nil, nil, err
	}
	var responseStruct struct {
		IssueResponse
		ErrorResponse
	}
	resp, err := c.client.do(req, &responseStruct)
	if err != nil {
		return nil, resp, err
	}
	if resp == nil {
		return nil, resp, fmt.Errorf("IssueCertificate: %w", ErrEmptyResult)
	}
	return &responseStruct.IssueResponse, resp, nil
}

// RevokeCertificateBySerial
func (c *ServicesService) RevokeCertificateBySerial(logicalPath, serial string, options ...OptionFunc) (*RevokeResponse, *Response, error) {
	revokeRequest := struct {
		SerialNumber string `json:"serial_number"`
	}{
		SerialNumber: serial,
	}
	req, err := c.client.newServiceRequest(http.MethodPost, "core/pki/api/"+logicalPath+"/revoke", revokeRequest, options)
	if err != nil {
		return nil, nil, err
	}
	var responseStruct struct {
		RevokeResponse
		ErrorResponse
	}
	resp, err := c.client.do(req, &responseStruct)
	if err != nil {
		return nil, resp, err
	}
	if resp == nil {
		return nil, nil, fmt.Errorf("RevokeCertificateBySerial: %w", ErrEmptyResult)
	}
	return &responseStruct.RevokeResponse, resp, nil
}

// GetCertificateBySerial
func (c *ServicesService) GetCertificateBySerial(logicalPath, serial string, options ...OptionFunc) (*IssueResponse, *Response, error) {
	req, err := c.client.newServiceRequest(http.MethodGet, "core/pki/api/"+logicalPath+"/cert/"+serial, nil, options)
	if err != nil {
		return nil, nil, err
	}
	var responseStruct struct {
		IssueResponse
		ErrorResponse
	}
	resp, err := c.client.do(req, &responseStruct)
	if err != nil {
		return nil, resp, err
	}
	if resp == nil {
		return nil, resp, fmt.Errorf("GetCertificateBySerial: %w", ErrEmptyResult)
	}
	return &responseStruct.IssueResponse, resp, nil
}

type QueryOptions struct {
	OrganizationID     *string `url:"organizationId,omitempty"`
	CommonName         *string `url:"commonName,omitempty"`
	CommonNameExact    *string `url:"commonName:exact,omitempty"`
	CommonNameContains *string `url:"commonName:contains,omitempty"`
	CommonNameMissing  *bool   `url:"commonName:missing,omitempty"`
	CommonNameExists   *bool   `url:"commonName:exists,omitempty"`

	AltName         *string `url:"altName,omitempty"`
	AltNameExact    *string `url:"altName:exact,omitempty"`
	AltNameContains *string `url:"altName:contains,omitempty"`
	AltNameMissing  *bool   `url:"altName:missing,omitempty"`
	AltNameExists   *bool   `url:"altName:exists,omitempty"`

	SerialNumber *string `url:"serialNumber,omitempty"`

	IssuedAt       *string `url:"issuedAt,omitempty"`
	ExpiresAt      *string `url:"expiresAt,omitempty"`
	KeyType        *string `url:"keyType,omitempty"`
	KeyLength      *string `url:"keyLength,omitempty"`
	KeyUsage       *string `url:"keyUsage,omitempty"`
	ExtKeyUsage    *string `url:"extKeyUsage,omitempty"`
	SubjectKeyId   *string `url:"subjectKeyId,omitempty"`
	AuthorityKeyId *string `url:"authorityKeyId,omitempty"`

	Status    *string `url:"_status,omitempty"`
	RevokedAt *string `url:"revokedAt,omitempty"`

	Operation *string `url:"_operation,omitempty"`
	Count     *string `url:"_count,omitempty"`
	Page      *string `url:"_page,omitempty"`
	Sort      *string `url:"_sort,omitempty"`
}

// CertificateList list serial numbers of non-revoked certificates including the Issuing CA
type CertificateList struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Data          struct {
		Keys []string `json:"keys"`
	} `json:"data"`
	WrapInfo string `json:"wrap_info,omitempty"`
	Warnings string `json:"warnings,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

// GetCertificates
func (c *ServicesService) GetCertificates(logicalPath string, opt *QueryOptions, options ...OptionFunc) (*CertificateList, *Response, error) {
	req, err := c.client.newServiceRequest(http.MethodGet, "core/pki/api/"+logicalPath+"/certs", opt, options)
	if err != nil {
		return nil, nil, err
	}
	var responseStruct struct {
		CertificateList
		ErrorResponse
	}
	resp, err := c.client.do(req, &responseStruct)
	if err != nil {
		return nil, resp, err
	}
	if resp == nil {
		return nil, nil, fmt.Errorf("GetCertificates: %w", ErrEmptyResult)
	}
	return &responseStruct.CertificateList, resp, nil
}
