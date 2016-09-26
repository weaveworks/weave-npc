package controller

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/iptables"

	"github.com/weaveworks/weave-npc/pkg/util/ipset"
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

	ipt iptables.Interface
	ips ipset.Interface

	nss         map[string]*ns // ns name -> ns struct
	nsSelectors *selectorSet   // selector string -> nsSelector
}

func New(ipt iptables.Interface, ips ipset.Interface) NetworkPolicyController {
	c := &controller{
		ipt: ipt,
		ips: ips,
		nss: make(map[string]*ns)}

	c.nsSelectors = newSelectorSet(ips, c.onNewNsSelector)

	return c
}

func (npc *controller) onNewNsSelector(selector *selector) error {
	for _, ns := range npc.nss {
		if ns.namespace != nil {
			if selector.matches(ns.namespace.ObjectMeta.Labels) {
				if err := ns.ips.AddEntry(selector.spec.ipsetName, string(ns.allPods.ipsetName)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (npc *controller) withNS(name string, f func(ns *ns) error) error {
	ns, found := npc.nss[name]
	if !found {
		newNs, err := newNS(name, npc.ipt, npc.ips, npc.nsSelectors)
		if err != nil {
			return err
		}
		npc.nss[name] = newNs
		ns = newNs
	}
	if err := f(ns); err != nil {
		return err
	}
	if ns.empty() {
		if err := ns.destroy(); err != nil {
			return err
		}
		delete(npc.nss, name)
	}
	return nil
}

func (npc *controller) AddPod(obj *api.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT AddPod %#v", obj)
	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.addPod(obj), "add pod")
	})
}

func (npc *controller) UpdatePod(oldObj, newObj *api.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT UpdatePod %#v %#v", oldObj, newObj)
	return npc.withNS(oldObj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.updatePod(oldObj, newObj), "update pod")
	})
}

func (npc *controller) DeletePod(obj *api.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT DeletePod %#v", obj)
	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.deletePod(obj), "delete pod")
	})
}

func (npc *controller) AddNetworkPolicy(obj *extensions.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT AddNetworkPolicy %#v", obj)
	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.addNetworkPolicy(obj), "add network policy")
	})
}

func (npc *controller) UpdateNetworkPolicy(oldObj, newObj *extensions.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT UpdateNetworkPolicy %#v %#v", oldObj, newObj)
	return npc.withNS(oldObj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.updateNetworkPolicy(oldObj, newObj), "update network policy")
	})
}

func (npc *controller) DeleteNetworkPolicy(obj *extensions.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT DeleteNetworkPolicy %#v", obj)
	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.deleteNetworkPolicy(obj), "delete network policy")
	})
}

func (npc *controller) AddNamespace(obj *api.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT AddNamespace %#v", obj)
	return npc.withNS(obj.ObjectMeta.Name, func(ns *ns) error {
		return errors.Wrap(ns.addNamespace(obj), "add namespace")
	})
}

func (npc *controller) UpdateNamespace(oldObj, newObj *api.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT UpdateNamespace%#v %#v", oldObj, newObj)
	return npc.withNS(oldObj.ObjectMeta.Name, func(ns *ns) error {
		return errors.Wrap(ns.updateNamespace(oldObj, newObj), "update namespace")
	})
}

func (npc *controller) DeleteNamespace(obj *api.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	log.Infof("EVENT DeleteNamespace %#v", obj)
	return npc.withNS(obj.ObjectMeta.Name, func(ns *ns) error {
		return errors.Wrap(ns.deleteNamespace(obj), "delete namespace")
	})
}
