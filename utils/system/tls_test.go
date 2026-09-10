package system

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTLSCredentials_AllCases(t *testing.T) {
	_, err := LoadTLSCredentials("invalid-cert.pem", "invalid-key.pem", "any.pem")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client cert/key")

	certFile := writeTempFile(t, validCert)
	keyFile := writeTempFile(t, validKey)
	defer func() {
		assert.NoError(t, os.Remove(certFile))
	}()
	defer func() {
		assert.NoError(t, os.Remove(keyFile))
	}()

	_, err = LoadTLSCredentials(certFile, keyFile, "invalid-ca.pem")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read CA cert")

	invalidCAFile := writeTempFile(t, "INVALID PEM")
	defer func() {
		assert.NoError(t, os.Remove(invalidCAFile))
	}()

	_, err = LoadTLSCredentials(certFile, keyFile, invalidCAFile)
	assert.Error(t, err)
	assert.Equal(t, "failed to append CA certs", err.Error())

	caFile := writeTempFile(t, validCert)
	defer func() {
		assert.NoError(t, os.Remove(caFile))
	}()

	cfg, err := LoadTLSCredentials(certFile, keyFile, caFile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Len(t, cfg.Certificates, 1)
	assert.NotNil(t, cfg.RootCAs)
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "tls-test-*.pem")
	assert.NoError(t, err)

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)
	assert.NoError(t, tmpfile.Close())

	return tmpfile.Name()
}

const validCert = `-----BEGIN CERTIFICATE-----
MIIC9TCCAd2gAwIBAgIUDBHhwJIWjEGVebd7wKLbPDpBTsQwDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJbG9jYWxob3N0MCAXDTI1MDUxNDA5NTQ1OFoYDzIxMjUw
NDIwMDk1NDU4WjAUMRIwEAYDVQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQC2CFFvEkrYYF/73XVMM4ZNwJYSY27uoJRD+9mth5Ju
iQaCO01IVNWvdnAK/RCHUfGBIpDvaOYY/HMRT93GdG11Rq8KTTZQsgPCy/rcxrWZ
CEqOg+VZNbPFg3RjX995vYPzJwqsSDS20adw+BXOpg8n0kDSXVYr5jx4nRAT03Qy
g2FaYk5/MIjeUkN8GqcFyZupk2oGbKzMVA9PhrpTrZN89JgNQNrjQ3EXkh0cYySr
clT0oytbXFWDSPwIxdfedJ012Lj0mnbRVDx4gvZrwfdywnIrGzzXqvnRxnkHHgfi
M0czzmrkNSGP6LLTXpTlVBzZ9O8r/83iIKOWLxsHmNYPAgMBAAGjPTA7MBoGA1Ud
EQQTMBGCCWxvY2FsaG9zdIcEfwAAATAdBgNVHQ4EFgQUGpPq16irqmsD2UWM5f1Y
jQkgpu8wDQYJKoZIhvcNAQELBQADggEBAF59wJt4BSDq4RJMsu8GKNcU7lLHkgWr
9UGdvf9rt6I02Vz1e2BE9ceKhCxRCp3UHEb9DDgJHZ0ADwDNQioAUzY1x2Po91Nv
v+rsJ0JPIX0KrShv9HF8VonJvYlz3L9b5c1j5tutCVGl8Vf39EI3iMHpWpH6X3rr
XoqH5zjAolKInnlM2whSfTmJMjVddK2EG5iz4OfxKv6IL0hlnf2+v08FusDamCL9
Eoj5ZgMtqpEHUSDLZzxO1hvHct9KRU5aY+UhSOshu3GR6Fugf+gB7e3wq3ezKxQ0
pY1Yi5MFsEY5UsO4pI7bAj6vgKqrlsoYIiQ6qs264PRfHe3ZOJC/93Y=
-----END CERTIFICATE-----`

const validKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC2CFFvEkrYYF/7
3XVMM4ZNwJYSY27uoJRD+9mth5JuiQaCO01IVNWvdnAK/RCHUfGBIpDvaOYY/HMR
T93GdG11Rq8KTTZQsgPCy/rcxrWZCEqOg+VZNbPFg3RjX995vYPzJwqsSDS20adw
+BXOpg8n0kDSXVYr5jx4nRAT03Qyg2FaYk5/MIjeUkN8GqcFyZupk2oGbKzMVA9P
hrpTrZN89JgNQNrjQ3EXkh0cYySrclT0oytbXFWDSPwIxdfedJ012Lj0mnbRVDx4
gvZrwfdywnIrGzzXqvnRxnkHHgfiM0czzmrkNSGP6LLTXpTlVBzZ9O8r/83iIKOW
LxsHmNYPAgMBAAECggEACPo+YJQjin9aTIF+tvbgO/ZwZaYaGFJlbY2UDemz8EKk
G7PRWxdgAIrT+h7CpsHaVMw/yfw/mOIzye9JQ0UxCXR7zoNsrIHTC1OPDZNj7REK
cvPl/EaDQA2nJagv99Y/wIlfjovzGZnGax0OdPDIqjtFhM/N9P/3ty/GgAvQ3RV7
FaXAMKOkO8Fy6e9V1Xoj1FP/qNojm3wnwmk/zoaNedz98hZUV7nEFLiAPdbzQMRY
SjXQgSk7gzYaFxBeevGVyDyC7Nj7AZk5Rmy5XcqDxUYhnu8yaY2mC0EHURhP+8UJ
R0zrRO8XyX9Z6lLaWwSjzGUVrTHxVGUgmPnS54OtdQKBgQD/SKlz/5LTbLIeDmks
AYhg+KFyJxZshM7JmWGsFeQj9WRcUZJpGPZ1IGAAMEPc3lLmowrKMQUJZSEllcIE
/7uxiVDT+CFbcoKpjfYeBMLB6uyzSP+GrIlDRxzoFJRYQqdG5rSvS98xBTS/3Klb
5yPoQ8GHfjA5cTIZFYFiOVAw6wKBgQC2iwyLlKGxNu8tn0FR3zPissEMMzQNK3zv
wcK4UrHR0KGo/RGc/dcychm1PnPc4qHPJBxbEZIvUYYAGOx+o2t+gW8T4tTqbepI
RJ5Co874Fc1pQjksoznts55CxtH+L9rOIBIY6KjIU1QXyvdUl6wpJvfgriUGj3Ef
UYg6dkuGbQKBgAutkULjMB5H3KYPVrRSpaB5/zivnRD9yk/imls67SLP+PVYLfBs
2ellv76CdrhF21j9oGK7d1WEsM19WlDMOhPXCkGIGk6KoHuNKPMamKYyTv2smzPX
9LeFK0damaan9esCZsWWHPGrIUydlYnEuxnG77V5Ck+2Y+pN14tcv9RdAoGAatiI
x0qAOhJFfRayTRGwdQjcJh/yX6MMxelL6Ee+/Wh4t0kpfhK2WzieA5BCkQ+2VmB0
mHl4b2nwXS45fwZ4bNumAKXMqksbzqEbYTYwdtWMHgg9HvuLdK6l+8AUOgwYrn3n
Gd1UrazYk/ShQEpm4s+EV2aXFXfwZrx6WH3VRyECgYEAhjIzUY38xMAvuc82oJ/F
J1jY7sQWsn5l3n8PDCGNeFra8N+diJijHGABqQPkPegIevRxMCTSb/hY29rvLgNY
nqvW3VWm02eMSViV1cSvk1Qi0lgUs1fftNxk9ZhKpevsPvXT+JUQ0FxhHsKXB0GA
9h4z1uc2WgRXIrflm7USuho=
-----END PRIVATE KEY-----`
