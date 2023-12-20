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
	"milvus-k8s-server/pkg/log"
	"milvus-k8s-server/pkg/session"
	"strings"
)

const (
	colocationKey = "dce.netease.com/workload-type"
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
	nodes, err := q.K8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodesMap := make(map[string]map[string]string)
	for _, node := range nodes.Items {
		nodesMap[node.Name] = node.Labels
	}
	sessions, err := session.ListSessionsByPrefix(q.MetaClient, "querynode")
	if err != nil {
		return nil, err
	}
	sessionIPs := make(map[string]struct{})
	for _, s := range sessions {
		addr := strings.Split(s.Address, ":")[0]
		sessionIPs[addr] = struct{}{}
	}
	queryNodes := make([]*QueryNodeK8sInfo, 0)
	for _, pod := range pods.Items {
		if _, ok := sessionIPs[pod.Status.PodIP]; ok {
			selectors := make(map[string]string)
			if nodeLabels, ok := nodesMap[pod.Spec.NodeName]; ok {
				selectors = nodeLabels
			} else {
				log.Logger.Warnf("match node failed, pod name: %s, node name: %s", pod.Name, pod.Spec.NodeName)
				continue
			}
			if val, ok := pod.Labels[colocationKey]; ok {
				if _, exist := selectors[colocationKey]; !exist {
					selectors[colocationKey] = val
				}
			}
			qn := &QueryNodeK8sInfo{
				PodName:   pod.Name,
				Addr:      fmt.Sprintf("%s:21123", pod.Status.PodIP),
				Selectors: selectors,
				K8sNode:   pod.Spec.NodeName,
			}
			log.Logger.Infof("get querynode, name: %s, addr: %s, node name: %s", qn.PodName, qn.Addr, qn.K8sNode)
			queryNodes = append(queryNodes, qn)
		} else {
			log.Logger.Warnf("match pod failed, pod name: %s, pod ip: %s", pod.Name, pod.Status.PodIP)
		}
	}
	return queryNodes, nil
}
