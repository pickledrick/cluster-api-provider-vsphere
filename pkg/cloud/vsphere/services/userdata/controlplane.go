/*
Copyright 2019 The Kubernetes Authors.

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

package userdata

import (
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
)

const (
	cloudConfig = `[Global]
secret-name = "{{ .SecretName }}"
secret-namespace = "{{ .SecretNamespace }}"
insecure-flag = "1" # set to 1 if the vCenter uses a self-signed cert
datacenters = "{{ .Datacenter }}"

[VirtualCenter "{{ .Server }}"]

[Workspace]
server = "{{ .Server }}"
datacenter = "{{ .Datacenter }}"
folder = "{{ .Folder }}"
default-datastore = "{{ .Datastore }}"
resourcepool-path = "{{ .ResourcePool }}"

[Disk]
scsicontrollertype = pvscsi

[Network]
public-network = "{{ .Network }}"
`

	controlPlaneCloudInit = `{{.Header}}
{{if .SSHAuthorizedKeys}}ssh_authorized_keys:{{range .SSHAuthorizedKeys}}
- "{{.}}"{{end}}{{end}}

write_files:
-   path: /etc/test/pki/ca.crt
    encoding: "base64"
    owner: root:root
    permissions: '0640'
    content: |
      {{.CACert | Base64Encode}}

-   path: /etc/test/pki/ca.key
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.CAKey | Base64Encode}}

-   path: /etc/test/pki/etcd/ca.crt
    encoding: "base64"
    owner: root:root
    permissions: '0640'
    content: |
      {{.EtcdCACert | Base64Encode}}

-   path: /etc/test/pki/etcd/ca.key
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.EtcdCAKey | Base64Encode}}

-   path: /etc/test/pki/front-proxy-ca.crt
    encoding: "base64"
    owner: root:root
    permissions: '0640'
    content: |
      {{.FrontProxyCACert | Base64Encode}}

-   path: /etc/test/pki/front-proxy-ca.key
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.FrontProxyCAKey | Base64Encode}}

-   path: /etc/test/pki/sa.pub
    encoding: "base64"
    owner: root:root
    permissions: '0640'
    content: |
      {{.SaCert | Base64Encode}}

-   path: /etc/test/pki/sa.key
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.SaKey | Base64Encode}}

-   path: /etc/kubernetes/vsphere.conf
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.CloudConfig | Base64Encode}}

-   path: /tmp/kubeadm.yaml
    owner: root:root
	permissions: '0640'
	content: |
      ---
{{.ClusterConfiguration | Indent 6}}
      ---
{{.InitConfiguration | Indent 6}}
kubeadm:
  operation: init
  config: /tmp/kubeadm.yaml

`

	controlPlaneJoinCloudInit = `{{.Header}}
{{if .SSHAuthorizedKeys}}ssh_authorized_keys:{{range .SSHAuthorizedKeys}}
- "{{.}}"{{end}}{{end}}

write_files:
-   path: /etc/test/pki/ca.crt
    encoding: "base64"
    owner: root:root
    permissions: '0640'
    content: |
      {{.CACert | Base64Encode}}

-   path: /etc/test/pki/ca.key
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.CAKey | Base64Encode}}

-   path: /etc/test/pki/etcd/ca.crt
    encoding: "base64"
    owner: root:root
    permissions: '0640'
    content: |
      {{.EtcdCACert | Base64Encode}}

-   path: /etc/test/pki/etcd/ca.key
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.EtcdCAKey | Base64Encode}}

-   path: /etc/test/pki/front-proxy-ca.crt
    encoding: "base64"
    owner: root:root
    permissions: '0640'
    content: |
      {{.FrontProxyCACert | Base64Encode}}

-   path: /etc/test/pki/front-proxy-ca.key
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.FrontProxyCAKey | Base64Encode}}

-   path: /etc/test/pki/sa.pub
    encoding: "base64"
    owner: root:root
    permissions: '0640'
    content: |
      {{.SaCert | Base64Encode}}

-   path: /etc/test/pki/sa.key
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.SaKey | Base64Encode}}

-   path: /etc/kubernetes/vsphere.conf
    encoding: "base64"
    owner: root:root
    permissions: '0600'
    content: |
      {{.CloudConfig | Base64Encode}}

-   path: /tmp/kubeadm-controlplane-join-config.yaml
    owner: root:root
    permissions: '0640'
    content: |
{{.JoinConfiguration | Indent 6}}
kubeadm:
  operation: join
  config: /tmp/kubeadm-controlplane-join-config.yaml
`
)

func isKeyPairValid(cert, key string) bool {
	return cert != "" && key != ""
}

// ControlPlaneInput defines the context to generate a controlplane instance user data.
type ControlPlaneInput struct {
	baseUserData

	SSHAuthorizedKeys []string

	CACert               string
	CAKey                string
	EtcdCACert           string
	EtcdCAKey            string
	FrontProxyCACert     string
	FrontProxyCAKey      string
	SaCert               string
	SaKey                string
	CloudConfig          string
	ClusterConfiguration string
	InitConfiguration    string
	CertDir              string
}

// ContolPlaneJoinInput defines context to generate controlplane instance user data for controlplane node join.
type ContolPlaneJoinInput struct {
	baseUserData

	SSHAuthorizedKeys []string

	CACert            string
	CAKey             string
	EtcdCACert        string
	EtcdCAKey         string
	FrontProxyCACert  string
	FrontProxyCAKey   string
	SaCert            string
	SaKey             string
	BootstrapToken    string
	ELBAddress        string
	CloudConfig       string
	JoinConfiguration string
	CertDir           string
}

// CloudConfigInput defines parameters required to generate the
// vSphere Cloud Provider cloud config file
type CloudConfigInput struct {
	SecretName      string
	SecretNamespace string
	Server          string
	Datacenter      string
	ResourcePool    string
	Folder          string
	Datastore       string
	Network         string
}

func (cpi *ControlPlaneInput) validateCertificates() error {
	if !isKeyPairValid(cpi.CACert, cpi.CAKey) {
		return errors.New("CA cert material in the ControlPlaneInput is missing cert/key")
	}

	if !isKeyPairValid(cpi.EtcdCACert, cpi.EtcdCAKey) {
		return errors.New("ETCD CA cert material in the ControlPlaneInput is  missing cert/key")
	}

	if !isKeyPairValid(cpi.FrontProxyCACert, cpi.FrontProxyCAKey) {
		return errors.New("FrontProxy CA cert material in ControlPlaneInput is  missing cert/key")
	}

	if !isKeyPairValid(cpi.SaCert, cpi.SaKey) {
		return errors.New("ServiceAccount cert material in ControlPlaneInput is  missing cert/key")
	}

	return nil
}

func (cpi *ContolPlaneJoinInput) validateCertificates() error {
	if !isKeyPairValid(cpi.CACert, cpi.CAKey) {
		return errors.New("CA cert material in the ContolPlaneJoinInput is  missing cert/key")
	}

	if !isKeyPairValid(cpi.EtcdCACert, cpi.EtcdCAKey) {
		return errors.New("ETCD cert material in the ContolPlaneJoinInput is  missing cert/key")
	}

	if !isKeyPairValid(cpi.FrontProxyCACert, cpi.FrontProxyCAKey) {
		return errors.New("FrontProxy cert material in ContolPlaneJoinInput is  missing cert/key")
	}

	if !isKeyPairValid(cpi.SaCert, cpi.SaKey) {
		return errors.New("ServiceAccount cert material in ContolPlaneJoinInput is  missing cert/key")
	}

	return nil
}

// NewControlPlane returns the user data string to be used on a controlplane instance.
func NewControlPlane(input *ControlPlaneInput) (string, error) {
	input.Header = cloudConfigHeader
	if err := input.validateCertificates(); err != nil {
		return "", errors.Wrapf(err, "ControlPlaneInput is invalid")
	}

	fMap := map[string]interface{}{
		"Base64Encode": templateBase64Encode,
		"Indent":       templateYAMLIndent,
	}

	userData, err := generateWithFuncs("controlplane", controlPlaneCloudInit, funcMap(fMap), input)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate user data for new control plane machine")
	}

	return userData, err
}

// JoinControlPlane returns the user data string to be used on a new contrplplane instance.
func JoinControlPlane(input *ContolPlaneJoinInput) (string, error) {
	input.Header = cloudConfigHeader

	if err := input.validateCertificates(); err != nil {
		return "", errors.Wrapf(err, "ControlPlaneInput is invalid")
	}

	fMap := map[string]interface{}{
		"Base64Encode": templateBase64Encode,
		"Indent":       templateYAMLIndent,
	}

	userData, err := generateWithFuncs("controlplane", controlPlaneJoinCloudInit, funcMap(fMap), input)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate user data for machine joining control plane")
	}
	return userData, err
}

// NewCloudConfig returns the string content for the vSphere Cloud Provider cloud config file
func NewCloudConfig(input *CloudConfigInput) (string, error) {
	fMap := map[string]interface{}{}

	userData, err := generateWithFuncs("cloudprovider", cloudConfig, funcMap(fMap), input)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate user data for new control plane machine")
	}

	return userData, nil
}

func templateBase64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func templateYAMLIndent(i int, input string) string {
	split := strings.Split(input, "\n")
	ident := "\n" + strings.Repeat(" ", i)
	return strings.Repeat(" ", i) + strings.Join(split, ident)
}
