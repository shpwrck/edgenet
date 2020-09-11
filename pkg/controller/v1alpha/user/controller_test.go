package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EdgeNet-project/edgenet/pkg/controller/v1alpha/authority"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStartController(t *testing.T) {
	g := UserTestGroup{}
	g.Init()
	authorityHandler := authority.Handler{}
	authorityHandler.Init(g.client, g.edgenetclient)
	g.authorityObj.Spec.Enabled = true
	g.edgenetclient.AppsV1alpha().Authorities().Create(context.TODO(), g.authorityObj.DeepCopy(), metav1.CreateOptions{})
	authorityHandler.ObjectCreated(g.authorityObj.DeepCopy())

	// Run the controller in a goroutine
	go Start(g.client, g.edgenetclient)
	// Create a user
	g.edgenetclient.AppsV1alpha().Users(g.userObj.GetNamespace()).Create(context.TODO(), g.userObj.DeepCopy(), metav1.CreateOptions{})
	// Wait for the status update of created object
	time.Sleep(time.Millisecond * 500)
	// Get the object and check the status
	user, _ := g.edgenetclient.AppsV1alpha().Users(fmt.Sprintf("authority-%s", g.authorityObj.GetName())).Get(context.TODO(), g.userObj.GetName(), metav1.GetOptions{})
	if !user.Spec.Active {
		t.Errorf(errorDict["add-func"])
	}
	// Update a user
	g.userObj.Spec.FirstName = "newName"
	g.edgenetclient.AppsV1alpha().Users(fmt.Sprintf("authority-%s", g.authorityObj.GetName())).Update(context.TODO(), g.userObj.DeepCopy(), metav1.UpdateOptions{})
	user, _ = g.edgenetclient.AppsV1alpha().Users(fmt.Sprintf("authority-%s", g.authorityObj.GetName())).Get(context.TODO(), g.userObj.GetName(), metav1.GetOptions{})
	if user.Spec.FirstName != "newName" {
		t.Errorf(errorDict["upd-func"])
	}
	// Delete a user
	g.edgenetclient.AppsV1alpha().Users(fmt.Sprintf("authority-%s", g.authorityObj.GetName())).Delete(context.TODO(), g.userObj.GetName(), metav1.DeleteOptions{})
	user, _ = g.edgenetclient.AppsV1alpha().Users(fmt.Sprintf("authority-%s", g.authorityObj.GetName())).Get(context.TODO(), g.userObj.GetName(), metav1.GetOptions{})
	if user != nil {
		t.Errorf(errorDict["del-func"])
	}

}