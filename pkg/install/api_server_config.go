package install

import (
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/util"
)

type APIServerConfig struct {
	RawData map[string]string `yaml:"api_server"`
}

type Transformer struct {
	transform func(string) string
}

var protectedValues = [14]string{
	"advertise-address",
	"apiserver-count",
	"authorization-policy-file",
	"basic-auth-file",
	"client-ca-file",
	"etcd-cafile",
	"etcd-certfile",
	"etcd-keyfile",
	"etcd-servers",
	"service-account-key-file",
	"service-cluster-ip-range",
	"tls-cert-file",
	"tls-private-key-file",
	"v",
}

var transformValues = map[string]Transformer{
	"admission-control": defaultValue("NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,ResourceQuota"),
	"allow-privileged": defaultValue("true"),
	"anonymous-auth": defaultValue("false"),
	"authorization-mode": defaultValue("RBAC,ABAC"),
	"bind-address": defaultValue("0.0.0.0"),
	"enable-swagger-ui": defaultValue("true"),
	"insecure-bind-address": defaultValue("127.0.0.1"),
	"insecure-port": defaultValue("8080"),
	"runtime-config" : mapWithDefaults("extensions/v1beta1=true,extensions/v1beta1/networkpolicies=true"),
	"secure-port": defaultValue("6443"),
	"v": defaultValue("2"),
}

var ansibleOverridenValues = map[string]string{
	"advertise-address": "{{ internal_ipv4 }}",
	"apiserver-count": "{{ kubernetes_master_apiserver_count }}",
	"authorization-policy-file": "{{ kubernetes_authorization_policy_path }}",
	"basic-auth-file": "{{ kubernetes_basic_auth_path }}",
	"client-ca-file": "{{ kubernetes_certificates_ca_path }}",
	"etcd-cafile": "{{ kubernetes_certificates_ca_path }}",
	"etcd-certfile": "{{ kubernetes_certificates_cert_path }}",
	"etcd-keyfile": "{{ kubernetes_certificates_key_path}}",
	"etcd-servers": "{{ etcd_k8s_cluster_ip_list }}",
	"service-account-key-file": "{{ kubernetes_certificates_service_account_key_path }}",
	"service-cluster-ip-range": "{{ kubernetes_services_cidr }}",
	"tls-cert-file": "{{ kubernetes_certificates_cert_path }}",
	"tls-private-key-file": "{{ kubernetes_certificates_key_path }}",
}

func (config *APIServerConfig) validate() (bool, []error) {
	v := newValidator()
	for _, protectedItem := range protectedValues {
		_, found := config.RawData[protectedItem]
		if found {
			v.addError(fmt.Errorf("Api config value [%s] should not be overriden", protectedItem))
		}
	}
	return v.valid();
}

func defaultValue(defaultValue string) Transformer {
	return Transformer{
		transform: func(inputValue string) string {
			if len(inputValue) > 0 {
				return defaultValue;
			}
			return inputValue;
		},
	}
}

func (config *APIServerConfig) ConfigValues() map[string]string {
	output := make(map[string]string)
	keys := util.MapKeys(config.RawData, transformValues)
	for _, key := range keys {
		trans, ok := transformValues[key]
		if ok {
			output[key] = trans.transform(config.RawData[key])
		} else {
			output[key] = config.RawData[key]
		}
	}
	return output
}

func mapWithDefaults(defaultValue string) Transformer {
	return Transformer{
		transform: func(inputValue string) string {
			newValues := util.StringToMap(inputValue)
			defaults := util.StringToMap(defaultValue)
			finalConfig := util.MergeMaps(newValues, defaults)

			return strings.Join(util.MapToSortedList(finalConfig), ",")
		},
	}
}
