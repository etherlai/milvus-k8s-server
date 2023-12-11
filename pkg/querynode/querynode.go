package querynode

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sclient "k8s.io/client-go/kubernetes"
	"milvus-k8s-server/pkg/common/k8s"
	"milvus-k8s-server/pkg/common/kv"
	"milvus-k8s-server/pkg/configs"
	"milvus-k8s-server/pkg/session"
)

type QueryNodeManager struct {
	K8sClient  *k8sclient.Clientset
	MetaClient kv.MetaKV
}

type QueryNodeK8sInfo struct {
	PodName   string            `json:"podName"`
	Addr      string            `json:"addr"`
	Selectors map[string]string `json:"selectors,omitempty"`
	K8sNode   string            `json:"k8sNode"`
}

func NewQueryNodeManager(c *configs.Config) (*QueryNodeManager, error) {
	K8sClient, err := k8s.CreateClientFromClusterConfig()
	if err != nil {
		return nil, err
	}
	etcdClient, err := kv.ConnectEtcd(c)
	if err != nil {
		return nil, err
	}
	return &QueryNodeManager{
		K8sClient:  K8sClient,
		MetaClient: etcdClient,
	}, nil
}

func (q *QueryNodeManager) GetAllQueryNodes() ([]*QueryNodeK8sInfo, error) {
	labelSelector := labels.Set{"component": "querynode"}.AsSelector().String()
	pods, err := q.K8sClient.CoreV1().Pods("milvus").List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	sessions, err := session.ListSessionsByPrefix(q.MetaClient, "querynode")
	if err != nil {
		return nil, err
	}
	sessionIPs := make(map[string]struct{})
	for _, s := range sessions {
		sessionIPs[s.Address] = struct{}{}
	}
	queryNodes := make([]*QueryNodeK8sInfo, 0)
	for _, pod := range pods.Items {
		if _, ok := sessionIPs[pod.Status.PodIP]; ok {
			qn := &QueryNodeK8sInfo{
				PodName:   pod.Name,
				Addr:      fmt.Sprintf("%s:21123", pod.Status.PodIP),
				Selectors: pod.Labels,
				K8sNode:   pod.Spec.NodeName,
			}
			queryNodes = append(queryNodes, qn)
		}
	}
	return queryNodes, nil
}
