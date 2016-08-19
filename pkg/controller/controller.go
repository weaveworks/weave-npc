package controller

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"sync"
)

type NetworkPolicyController interface {
	AddNamespace(ns *api.Namespace) error
	UpdateNamespace(oldObj, newObj *api.Namespace) error
	DeleteNamespace(ns *api.Namespace) error

	AddPod(obj *api.Pod) error
	UpdatePod(oldObj, newObj *api.Pod) error
	DeletePod(obj *api.Pod) error

	AddNetworkPolicy(obj *extensions.NetworkPolicy) error
	UpdateNetworkPolicy(oldObj, newObj *extensions.NetworkPolicy) error
	DeleteNetworkPolicy(obj *extensions.NetworkPolicy) error
}

type controller struct {
	sync.Mutex

	nss         map[string]*ns       // ns name -> ns struct
	nsSelectors map[string]*selector // selector string -> nsSelector
}

func New() NetworkPolicyController {
	return &controller{
		nss:         make(map[string]*ns),
		nsSelectors: make(map[string]*selector)}
}

func (npc *controller) withNS(name string, f func(ns *ns) error) error {
	ns, found := npc.nss[name]
	if !found {
		ns = newNS(name, npc.nsSelectors)
		npc.nss[name] = ns
	}
	if err := f(ns); err != nil {
		return err
	}
	if ns.empty() {
		delete(npc.nss, name)
	}
	return nil
}

func (npc *controller) AddPod(obj *api.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return ns.addPod(obj)
	})
}

func (npc *controller) UpdatePod(oldObj, newObj *api.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(oldObj.ObjectMeta.Namespace, func(ns *ns) error {
		return ns.updatePod(oldObj, newObj)
	})
}

func (npc *controller) DeletePod(obj *api.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return ns.deletePod(obj)
	})
}

func (npc *controller) AddNetworkPolicy(obj *extensions.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return ns.addNetworkPolicy(obj)
	})
}

func (npc *controller) UpdateNetworkPolicy(oldObj, newObj *extensions.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(oldObj.ObjectMeta.Namespace, func(ns *ns) error {
		return ns.updateNetworkPolicy(oldObj, newObj)
	})
}

func (npc *controller) DeleteNetworkPolicy(obj *extensions.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return ns.deleteNetworkPolicy(obj)
	})
}

func (npc *controller) AddNamespace(obj *api.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(obj.ObjectMeta.Name, func(ns *ns) error {
		return ns.addNamespace(obj)
	})
}

func (npc *controller) UpdateNamespace(oldObj, newObj *api.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(oldObj.ObjectMeta.Name, func(ns *ns) error {
		return ns.updateNamespace(oldObj, newObj)
	})
}

func (npc *controller) DeleteNamespace(obj *api.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	return npc.withNS(obj.ObjectMeta.Name, func(ns *ns) error {
		return ns.deleteNamespace(obj)
	})
}
