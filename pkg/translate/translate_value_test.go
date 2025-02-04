// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translate

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/kr/pretty"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/version"
)

func TestValueToProto(t *testing.T) {

	tests := []struct {
		desc      string
		valueYAML string
		want      string
		wantErr   string
	}{
		{
			desc: "K8s resources translation",
			valueYAML: `
galley:
  enabled: false
pilot:
  enabled: true
  resources:
    requests:
      cpu: 1000m
      memory: 1G
  replicaCount: 1
  nodeSelector:
    beta.kubernetes.io/os: linux
  autoscaleEnabled: true
  autoscaleMax: 3
  autoscaleMin: 1
  traceSampling: 1.0
  image: pilot
  env:
    GODEBUG: gctrace=1
  podAntiAffinityLabelSelector:
    - labelSelector:
        matchLabels:
          testK1: testV1
global:
  hub: docker.io/istio
  istioNamespace: istio-system
  policyNamespace: istio-policy
  tag: 1.2.3
  telemetryNamespace: istio-telemetry
  proxy:
    ReadinessInitialDelaySeconds: 2
mixer:
  policy:
    enabled: true
    image: mixer
    replicaCount: 1
  telemetry:
    enabled: false
`,
			want: `
hub: docker.io/istio
tag: 1.2.3
defaultNamespace: istio-system
telemetry:
 components:
   namespace: istio-telemetry
   telemetry:
     common:
       enabled: false
 enabled: false
policy:
 components:
   namespace: istio-policy
   policy:
     common:
       enabled: true
       k8s:
         replicaCount: 1
       values:
         image: mixer
 enabled: true
configManagement:
 components:
   galley:
     common:
       enabled: false
 enabled: false
security:
 components:
   namespace: istio-system
   certManager:
     common:
       enabled: false
   nodeAgent:
     common:
       enabled: false
   citadel:
     common:
       enabled: false
 enabled: false
gateways:
 components:
   ingressGateway:
     common:
       enabled: false
   egressGateway:
     common:
       enabled: false
 enabled: false
trafficManagement:
 components:
   pilot:
     common:
       enabled: true
       k8s:
         affinity:
           podAntiAffinity:
             requiredDuringSchedulingIgnoredDuringExecution:
             - labelSelector:
                   matchLabels:
                     testK1: testV1
         replicaCount: 1
         env:
         - name: GODEBUG
           value: gctrace=1
         hpaSpec:
            maxReplicas: 3
            minReplicas: 1
            scaleTargetRef: {}
         nodeSelector:
            beta.kubernetes.io/os: linux
         resources:
            requests:
              cpu: 1000m
              memory: 1G
       values:
          image: pilot
          traceSampling: 1
   proxy:
     common:
       values:
         readinessInitialDelaySeconds: 2
 enabled: true
autoInjection:
 components:
   injector:
      common:
       enabled: false
 enabled: false
`,
		},
		{
			desc: "All Enabled",
			valueYAML: `
certManager:
  enabled: true
galley:
  enabled: true
global:
  hub: docker.io/istio
  istioNamespace: istio-system
  policyNamespace: istio-policy
  tag: 1.2.3
  telemetryNamespace: istio-telemetry
mixer:
  policy:
    enabled: true
  telemetry:
    enabled: true
pilot:
  enabled: true
nodeAgent:
  enabled: true
gateways:
  enabled: true
  istio-ingressgateway:
    resources:
      requests:
        cpu: 1000m
        memory: 1G
    enabled: true
sidecarInjectorWebhook:
  enabled: false
`,
			want: `
hub: docker.io/istio
tag: 1.2.3
defaultNamespace: istio-system
telemetry:
  components:
    namespace: istio-telemetry
    telemetry:
      common:
        enabled: true
  enabled: true
policy:
  components:
    namespace: istio-policy
    policy:
      common:
        enabled: true
  enabled: true
configManagement:
  components:
    galley:
      common:
        enabled: true
  enabled: true 
security:
  components:
    namespace: istio-system
    certManager:
      common:
        enabled: true
    nodeAgent:
      common:
        enabled: true
    citadel:
      common:
        enabled: false
  enabled: true
trafficManagement:
   components:
     pilot:
       common:
         enabled: true
   enabled: true
autoInjection:
  components:
    injector:
      common:
        enabled: false
  enabled: false
gateways:
  components:
    ingressGateway:
      common:
        enabled: true
        k8s:
          resources:
            requests:
              cpu: 1000m
              memory: 1G 
    egressGateway:
          common:
            enabled: false
  enabled: true
`,
		},
		{
			desc: "Some components Disabled",
			valueYAML: `
galley:
  enabled: false
pilot:
  enabled: true
global:
  hub: docker.io/istio
  istioNamespace: istio-system
  policyNamespace: istio-policy
  tag: 1.2.3
  telemetryNamespace: istio-telemetry
mixer:
  policy:
    enabled: true
telemetry:
    enabled: false
`,
			want: `
hub: docker.io/istio
tag: 1.2.3
defaultNamespace: istio-system
telemetry:
 components:
   namespace: istio-telemetry
   telemetry:
     common:
       enabled: false
 enabled: false
policy:
 components:
   namespace: istio-policy
   policy:
     common:
       enabled: true
 enabled: true
configManagement:
 components:
   galley:
     common:
       enabled: false
 enabled: false
security:
 components:
   namespace: istio-system
   certManager:
     common:
       enabled: false
   nodeAgent:
     common:
       enabled: false
   citadel:
     common:
       enabled: false
 enabled: false
gateways:
 components:
   ingressGateway:
     common:
       enabled: false
   egressGateway:
     common:
       enabled: false
 enabled: false
trafficManagement:
 components:
   pilot:
     common:
       enabled: true
 enabled: true
autoInjection:
 components:
   injector:
      common:
       enabled: false
 enabled: false
`,
		},
	}
	tr, err := NewReverseTranslator(version.NewMinorVersion(1, 3))
	if err != nil {
		t.Fatal("fail to get helm value.yaml translator")
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			valueStruct := v1alpha2.Values{}
			err := yaml.Unmarshal([]byte(tt.valueYAML), &valueStruct)
			if err != nil {
				t.Fatalf("unmarshal(%s): got error %s", tt.desc, err)
			}
			scope.Debugf("value struct: \n%s\n", pretty.Sprint(valueStruct))
			got, err := tr.TranslateFromValueToSpec(&valueStruct)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("ValuesToProto(%s)(%v): gotErr:%s, wantErr:%s", tt.desc, tt.valueYAML, gotErr, wantErr)
			}
			if tt.wantErr == "" {
				ms := jsonpb.Marshaler{}
				gotString, err := ms.MarshalToString(got)
				if err != nil {
					t.Errorf("error when marshal translated IstioControlPlaneSpec: %s", err)
				}
				cpYaml, _ := yaml.JSONToYAML([]byte(gotString))
				if want := tt.want; !util.IsYAMLEqual(gotString, want) {
					t.Errorf("ValuesToProto(%s): got:\n%s\n\nwant:\n%s\nDiff:\n%s\n", tt.desc, string(cpYaml), want, util.YAMLDiff(gotString, want))
				}
			}
		})
	}
}
