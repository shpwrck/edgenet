/*
Copyright 2020 Sorbonne Université

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

package team

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	apps_v1alpha "github.com/EdgeNet-project/edgenet/pkg/apis/apps/v1alpha"
	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	appsinformer_v1 "github.com/EdgeNet-project/edgenet/pkg/generated/informers/externalversions/apps/v1alpha"
	"github.com/EdgeNet-project/edgenet/pkg/permission"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// The main structure of controller
type controller struct {
	logger   *log.Entry
	queue    workqueue.RateLimitingInterface
	informer cache.SharedIndexInformer
	handler  HandlerInterface
}

// The main structure of informerEvent
type informerevent struct {
	key      string
	function string
	change   fields
}

// This contains the fields to check whether they are updated
type fields struct {
	enabled bool
	users   userData
	object  objectData
}

type userData struct {
	status  bool
	deleted string
	added   string
}

type objectData struct {
	name           string
	ownerNamespace string
	childNamespace string
}

// Constant variables for events
const create = "create"
const update = "update"
const delete = "delete"
const success = "Successful"

// Start function is entry point of the controller
func Start(kubernetes kubernetes.Interface, edgenet versioned.Interface) {
	var err error
	clientset := kubernetes
	edgenetClientset := edgenet
	teamHandler := &Handler{}
	// Create the team informer which was generated by the code generator to list and watch team resources
	informer := appsinformer_v1.NewTeamInformer(
		edgenetClientset,
		metav1.NamespaceAll,
		0,
		cache.Indexers{},
	)
	// Create a work queue which contains a key of the resource to be handled by the handler
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	var event informerevent
	// Event handlers deal with events of resources. In here, we take into consideration of adding and updating nodes
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// Put the resource object into a key
			event.key, err = cache.MetaNamespaceKeyFunc(obj)
			event.function = create
			log.Infof("Add team: %s", event.key)
			if err == nil {
				// Add the key to the queue
				queue.Add(event)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			event.key, err = cache.MetaNamespaceKeyFunc(newObj)
			event.function = update
			// Find out whether the fields updated
			event.change.users.status = false
			event.change.users.deleted = ""
			event.change.users.added = ""
			if !reflect.DeepEqual(oldObj.(*apps_v1alpha.Team).Spec.Users, newObj.(*apps_v1alpha.Team).Spec.Users) {
				event.change.users.status = true
				sliceDeleted, sliceAdded := dry(oldObj.(*apps_v1alpha.Team).Spec.Users, newObj.(*apps_v1alpha.Team).Spec.Users)
				sliceDeletedJSON, err := json.Marshal(sliceDeleted)
				if err == nil {
					event.change.users.deleted = string(sliceDeletedJSON)
				}
				sliceAddedJSON, err := json.Marshal(sliceAdded)
				if err == nil {
					event.change.users.added = string(sliceAddedJSON)
				}
			}
			log.Infof("Update team: %s", event.key)
			if err == nil {
				queue.Add(event)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc helps to check the existence of the object while it is still contained in the index.
			// Put the resource object into a key
			event.key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			event.function = delete
			event.change.users.status = true
			event.change.users.deleted = ""
			sliceDeletedJSON, err := json.Marshal(obj.(*apps_v1alpha.Team).Spec.Users)
			if err == nil {
				event.change.users.deleted = string(sliceDeletedJSON)
			}
			event.change.object.name = obj.(*apps_v1alpha.Team).GetName()
			event.change.object.ownerNamespace = obj.(*apps_v1alpha.Team).GetNamespace()
			event.change.object.childNamespace = fmt.Sprintf("%s-team-%s", obj.(*apps_v1alpha.Team).GetNamespace(), obj.(*apps_v1alpha.Team).GetName())
			event.change.enabled = obj.(*apps_v1alpha.Team).Spec.Enabled
			log.Infof("Delete team: %s", event.key)
			if err == nil {
				queue.Add(event)
			}
		},
	})
	controller := controller{
		logger:   log.NewEntry(log.New()),
		informer: informer,
		queue:    queue,
		handler:  teamHandler,
	}

	// Create the roles of EdgeNet users
	permission.Clientset = clientset
	permission.CreateTeamRoles()

	// A channel to terminate elegantly
	stopCh := make(chan struct{})
	defer close(stopCh)
	// Run the controller loop as a background task to start processing resources
	go controller.run(stopCh, clientset, edgenetClientset)
	// A channel to observe OS signals for smooth shut down
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}

// Run starts the controller loop
func (c *controller) run(stopCh <-chan struct{}, clientset kubernetes.Interface, edgenetClientset versioned.Interface) {
	// A Go panic which includes logging and terminating
	defer utilruntime.HandleCrash()
	// Shutdown after all goroutines have done
	defer c.queue.ShutDown()
	c.logger.Info("run: initiating")
	c.handler.Init(clientset, edgenetClientset)
	// Run the informer to list and watch resources
	go c.informer.Run(stopCh)

	// Synchronization to settle resources one
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}
	c.logger.Info("run: cache sync complete")
	// Operate the runWorker
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

// To process new objects added to the queue
func (c *controller) runWorker() {
	log.Info("runWorker: starting")
	// Run processNextItem for all the changes
	for c.processNextItem() {
		log.Info("runWorker: processing next item")
	}

	log.Info("runWorker: completed")
}

// This function deals with the queue and sends each item in it to the specified handler to be processed.
func (c *controller) processNextItem() bool {
	log.Info("processNextItem: start")
	// Fetch the next item of the queue
	event, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(event)
	// Get the key string
	keyRaw := event.(informerevent).key
	// Use the string key to get the object from the indexer
	item, exists, err := c.informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		if c.queue.NumRequeues(event.(informerevent).key) < 5 {
			c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, retrying", event.(informerevent).key, err)
			c.queue.AddRateLimited(event.(informerevent).key)
		} else {
			c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", event.(informerevent).key, err)
			c.queue.Forget(event.(informerevent).key)
			utilruntime.HandleError(err)
		}
	}

	if !exists {
		if event.(informerevent).function == delete {
			c.logger.Infof("Controller.processNextItem: object deleted detected: %s", keyRaw)
			c.handler.ObjectDeleted(item, event.(informerevent).change)
		}
	} else {
		if event.(informerevent).function == create {
			c.logger.Infof("Controller.processNextItem: object created detected: %s", keyRaw)
			c.handler.ObjectCreated(item)
		} else if event.(informerevent).function == update {
			c.logger.Infof("Controller.processNextItem: object updated detected: %s", keyRaw)
			c.handler.ObjectUpdated(item, event.(informerevent).change)
		}
	}
	c.queue.Forget(event.(informerevent).key)

	return true
}
