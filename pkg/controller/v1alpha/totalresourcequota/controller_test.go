package totalresourcequota

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStartController(t *testing.T) {
	g := TRQTestGroup{}
	g.Init()
	// Run the controller in a goroutine
	go Start(g.client, g.edgenetclient)
	// Create a resource request
	g.TRQObj.Spec.Enabled = true
	g.edgenetclient.AppsV1alpha().TotalResourceQuotas().Create(context.TODO(), g.TRQObj.DeepCopy(), metav1.CreateOptions{})
	// Wait for the status update of created object
	time.Sleep(time.Millisecond * 500)
	// Get the object and check the status
	TRQ, _ := g.edgenetclient.AppsV1alpha().TotalResourceQuotas().Get(context.TODO(), g.TRQObj.GetName(), metav1.GetOptions{})
	if TRQ.Status.State != success {
		t.Error(errorDict["add-func"])
	}
	// Update the TRQ
	g.TRQObj.Spec.Enabled = false
	g.edgenetclient.AppsV1alpha().TotalResourceQuotas().Update(context.TODO(), g.TRQObj.DeepCopy(), metav1.UpdateOptions{})
	TRQ, _ = g.edgenetclient.AppsV1alpha().TotalResourceQuotas().Get(context.TODO(), g.TRQObj.GetName(), metav1.GetOptions{})
	if TRQ.Status.State == success {
		t.Error(errorDict["upd-func"])
	}
}