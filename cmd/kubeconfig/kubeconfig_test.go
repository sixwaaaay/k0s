package kubeconfig

// Basic imports
import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/k0sproject/k0s/pkg/apis/k0s.k0sproject.io/v1beta1"
	"github.com/k0sproject/k0s/pkg/certificate"
	"github.com/k0sproject/k0s/pkg/config"
	"github.com/k0sproject/k0s/pkg/constant"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CLITestSuite struct {
	suite.Suite
}

func (s *CLITestSuite) TestKubeConfigExternalAddress() {
	yamlData := `
apiVersion: k0s.k0sproject.io/v1beta1
kind: ClusterConfig
spec:
  api:
    externalAddress: 1.2.3.4
`
	c := CmdOpts(config.GetCmdOpts())
	cfgFilePath, err := writeTmpFile(yamlData, "k0s-config")
	s.NoError(err)
	c.CfgFile = cfgFilePath

	cfg, err := v1beta1.ConfigFromString(yamlData, "")
	s.NoError(err)
	a := cfg.Spec.API

	s.Equal("1.2.3.4", a.ExternalAddress)
	clusterAPIURL, err := c.getAPIURL()
	s.NoError(err)
	s.Equal("https://1.2.3.4:6443", clusterAPIURL)
}

func (s *CLITestSuite) TestKubeConfigCreate() {
	c := CmdOpts(config.GetCmdOpts())

	caCert := `
-----BEGIN CERTIFICATE-----
MIIDADCCAeigAwIBAgIUW+2hawM8HgHrfxmDRV51wOq95icwDQYJKoZIhvcNAQEL
BQAwGDEWMBQGA1UEAxMNa3ViZXJuZXRlcy1jYTAeFw0yMTEwMjExMjAxMDBaFw0z
MTEwMTkxMjAxMDBaMBgxFjAUBgNVBAMTDWt1YmVybmV0ZXMtY2EwggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDQqkq9cu49/The1CUQSqFNeGaNNblnHYZo
CFcrJYtuimTPc7Abs9vIp6Ax5wqtqGTYzdg0hZc4dKXFDpvVzn8yU17IUpfDY7Ix
j2q8wBDI7bJCJw5Mw8/lcAqI1ub+DEYrdg6sRvCcByCK9qPlvuabc6YAbB0mmES6
rqCXE/Xr8byW9QYPwD+p6wKZoRXm9WlSwCvFT9OCk2OT8G5o+9RagHhmYgsg9vHf
LrDoPUtu/W5zE+fmIaAHGoWoo9yaBGsavPRkFzjbPI7mK1Zci6phnP/YWtLpyIx1
n+evNhdZj7UE9CMrzyhUU45vriK/Arc7co/WAk6pqO82tWsj38zJAgMBAAGjQjBA
MA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQ6hhbj
kcCmAK0BSTI5bU+hkwO00zANBgkqhkiG9w0BAQsFAAOCAQEAITD7bXxrpbS2kHl4
4Z3MKo59+KXHo9ut4a7L+oGwKX694g0/BrEAGHXRZrF5hEY0q8R0g3TdlHax0A6t
jpKePa+9ifNE+34gCz07xvAclcljk87zUM1mYYu1kgSc0XWeHnzMVXalo+gWzTBL
q8mVPQ4v+nk+MVP06r7GA42GsqTZGhH1xDQF0GLa25UHw4pzEX1olwaBWybl7Wql
K3icRdyke+TCLl+YqsCKG2n95cK4CMMEm8a1KVWRZKwDqLD7rFdemNdmzCNlpFW/
1uC/IeGA0XwM6CLsS7VAe0wbgxbgw0vLvkAnAEl6+VAOqr2ux3js1BbQ5g0d/x1L
nzXu8A==
-----END CERTIFICATE-----
`
	caCertPath, err := writeTmpFile(caCert, "ca-cert")
	s.NoError(err)

	caCertKey := `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0KpKvXLuPf04XtQlEEqhTXhmjTW5Zx2GaAhXKyWLbopkz3Ow
G7PbyKegMecKrahk2M3YNIWXOHSlxQ6b1c5/MlNeyFKXw2OyMY9qvMAQyO2yQicO
TMPP5XAKiNbm/gxGK3YOrEbwnAcgivaj5b7mm3OmAGwdJphEuq6glxP16/G8lvUG
D8A/qesCmaEV5vVpUsArxU/TgpNjk/BuaPvUWoB4ZmILIPbx3y6w6D1Lbv1ucxPn
5iGgBxqFqKPcmgRrGrz0ZBc42zyO5itWXIuqYZz/2FrS6ciMdZ/nrzYXWY+1BPQj
K88oVFOOb64ivwK3O3KP1gJOqajvNrVrI9/MyQIDAQABAoIBAQCMFieDNIuZdkzH
7SjM3S2ZcwF2P+EuxvWbFi5fOx92oNa5J3PNxVwCQ/caSYAzwd+iZd+Gs0Eol7dK
qloYmj9uq+XwGvLkLCRPfXctLMyX+Gw6WToSc0s5P5Ty9UOyvs7FEscbBa03Mtm4
MYkrDpSHPIbvtaWEaamKov4RL0dklHh/HQPXOujCPGiwbqHMIIQ0+sQwEDCVcC+Z
eHw4GCiC4GPgABRyhMO1CHdLxvU7xWUGTsPM4jzVmfWT5ZUx6mGi/v+weE600GU6
qSKt6fP4oygGISRA8ya9rtGQe2qQpeMlw0QhJWdRb2KAwbyQKCr1/f9uWTrhhwJ1
T1kezEUBAoGBAPxNdLjjh+UAQ9vF0UWDiTdC5wLi6YCbrcXvIcWfh7AXDinPOpKN
VxSx6cj9KQS/tWWJS8awDNZ+R/vItHJJsmyBBtY2yTQSa+LM44gRpuhdFdKhPvqN
thZVnRWbrY3xgR26Qb8F9iZy3qjkGYH8KCuBwd0a969HwVv8dqJjDhf5AoGBANO5
IAYdzy4TOkh870diAEjigVpBP/QHTRu4SXQgafqn0jPG2qRCEOpCBEjLQzCvzSf/
QMfIpcvT7xaxkSxil7mzt1X0qAjrrzpCsa6fWbiQqiFHrES10b0j4IT5bRRH4Nvz
lKXe4IKsRZ3Hmi7HRX51V1eBy7fQxRhXDy/R/69RAoGAdYebZPlQ96NM+RbIaqpg
hCadOGH9xhQ/OeIwiD/NVIEY7u8C6PwAYbqTHjaYIgcv+BGiA/dEs7J109tl+4tL
G3JrfeRdi+085pTtNRiL+NhL7yeAD/Vtqi/NkiBIE8Q5kmCOee7MAJMoF+LR4xRU
nhe++EG0uakicLhFh1W/XfkCgYBMEuyKxhM3PvlmKl3fjDsF9Tz9LQzJpgXyu9jI
vQzXX42LxRuygXqKcYYQkdhmmgRhJrokDthj0JbL1KmRBSv3MbfiTrJB4k1n5abq
U59tTa2Tn6kqVxoxl76IiQbEjr8gyPjUUKzixvuMobeorzktIwRrENweBAmNoVp3
mEECwQKBgQDCYi2EubaseSNu25UQY7ij1TsxPZpBvPlQFUtwmpz9MmBvqZcJYsco
z+5UodDFCnUsfprMjfTdY2Vk99PT4++SrJ5iTOn7xgKRrd1MPkBv7SXwnPtxCBAK
yJm2KSue0toWmkBFK8WMTjAvmAw3Z/qUhJRKoqCu3k6Mf8DNl6t+Uw==
-----END RSA PRIVATE KEY-----
`
	caKeyPath, err := writeTmpFile(caCertKey, "ca-key")
	s.NoError(err)

	userReq := certificate.Request{
		Name:   "test-user",
		CN:     "test-user",
		O:      groups,
		CACert: caCertPath,
		CAKey:  caKeyPath,
	}

	k0sVars := constant.GetConfig(os.TempDir())
	certManager := certificate.Manager{
		K0sVars: k0sVars,
	}
	err = os.Mkdir(path.Join(k0sVars.CertRootDir), 0755)

	userCert, err := certManager.EnsureCertificate(userReq, "root")
	s.NoError(err)
	clusterAPIURL, err := c.getAPIURL()
	s.NoError(err)

	data := struct {
		CACert     string
		ClientCert string
		ClientKey  string
		User       string
		JoinURL    string
	}{
		CACert:     base64.StdEncoding.EncodeToString([]byte(caCert)),
		ClientCert: base64.StdEncoding.EncodeToString([]byte(userCert.Cert)),
		ClientKey:  base64.StdEncoding.EncodeToString([]byte(userCert.Key)),
		User:       "test-user",
		JoinURL:    clusterAPIURL,
	}

	var buf bytes.Buffer

	err = userKubeconfigTemplate.Execute(&buf, &data)
	s.NoError(err)
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func writeTmpFile(data string, prefix string) (path string, err error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%v-", prefix))
	if err != nil {
		return "", fmt.Errorf("cannot create temporary file: %v", err)
	}

	text := []byte(data)
	if _, err = tmpFile.Write(text); err != nil {
		return "", fmt.Errorf("failed to write to temporary file: %v", err)
	}

	return tmpFile.Name(), nil
}
