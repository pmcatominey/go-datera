package dsdk_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"testing"
	"time"

	dsdk "github.com/Datera/go-sdk/pkg/v2/dsdk"
)

func createAi(ctxt context.Context, sdk *dsdk.SDK) (*dsdk.AppInstance, func(), error) {
	vol := &dsdk.Volume{
		Name:          "volume-1",
		Size:          1,
		PlacementMode: "hybrid",
		ReplicaCount:  1,
	}
	si := &dsdk.StorageInstance{
		Name:    "storage-1",
		Volumes: []*dsdk.Volume{vol},
	}
	aiReq := dsdk.AppInstancesCreateRequest{
		Ctxt:             ctxt,
		Name:             fmt.Sprintf("test-%s", dsdk.RandString(10)),
		StorageInstances: []*dsdk.StorageInstance{si},
	}
	resp, apierr, err := sdk.AppInstances.Create(&aiReq)
	if err != nil {
		return nil, func() {}, err
	}
	if apierr != nil {
		return nil, func() {}, fmt.Errorf("%#v", apierr)
	}
	ai := dsdk.AppInstance(*resp)
	return &ai, func() {
		_, _, err = ai.Set(&dsdk.AppInstanceSetRequest{
			Ctxt:       ctxt,
			AdminState: "offline",
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		_, _, err := ai.Delete(&dsdk.AppInstanceDeleteRequest{Ctxt: ctxt})
		if err != nil {
			fmt.Println(err)
			return
		}
	}, nil
}

func createInitiator(ctxt context.Context, sdk *dsdk.SDK) (*dsdk.Initiator, func(), error) {
	init, apierr, err := sdk.Initiators.Create(&dsdk.InitiatorsCreateRequest{
		Ctxt: ctxt,
		Name: "my-test-init",
		Id:   "iqn.1993-08.org.debian:01:58cc6c30e338",
	})
	if err != nil {
		return nil, func() {}, err
	}
	if apierr != nil {
		if apierr.Name == "ConflictError" {
			init, apierr, err = sdk.Initiators.Get(&dsdk.InitiatorsGetRequest{
				Ctxt: ctxt,
				Id:   "iqn.1993-08.org.debian:01:58cc6c30e338",
			})
		} else {
			return nil, func() {}, fmt.Errorf("%#v", apierr)
		}
	}
	return init, func() {
		if init == nil {
			return
		}
		_, _, err = init.Delete(&dsdk.InitiatorDeleteRequest{Ctxt: ctxt})
		if err != nil {
			fmt.Println(err)
			return
		}
	}, nil
}

func TestStorageNodes(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestStorageNodes")
	sns, _, err := sdk.StorageNodes.List(&dsdk.StorageNodesListRequest{Ctxt: sdk.NewContext()})
	if err != nil {
		t.Fatal(err)
	}
	for _, sn := range sns {
		fmt.Printf("StorageNode: %s\n", sn.Uuid)
	}
}

func TestIpPools(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestIpPools")
	anips, _, err := sdk.AccessNetworkIpPools.List(&dsdk.AccessNetworkIpPoolsListRequest{Ctxt: sdk.NewContext()})
	if err != nil {
		t.Fatal(err)
	}
	for _, anip := range anips {
		fmt.Printf("AccessNetworkIpPool: %s\n", anip.Name)
	}
}

func TestStoragePools(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestStoragePools")
	sps, _, err := sdk.StoragePools.List(&dsdk.StoragePoolsListRequest{Ctxt: sdk.NewContext()})
	if err != nil {
		// Can only be accessed in v3.2+
		t.Skip(err)
	}
	for _, sp := range sps {
		fmt.Printf("StoragePool: %s\n", sp.Name)
	}
}

func TestInitiators(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestInitiators")
	inits, _, err := sdk.Initiators.List(&dsdk.InitiatorsListRequest{Ctxt: sdk.NewContext()})
	if err != nil {
		t.Fatal(err)
	}
	for _, init := range inits {
		fmt.Printf("Initiator: %s\n", init.Name)
	}
}

func TestInitiatorGroups(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestInitiatorGroups")
	igs, _, err := sdk.InitiatorGroups.List(&dsdk.InitiatorGroupsListRequest{Ctxt: sdk.NewContext()})
	if err != nil {
		t.Fatal(err)
	}
	for _, ig := range igs {
		fmt.Printf("InitiatorGroup: %s\n", ig.Name)
	}
}

func TestAclPolicy(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestACLPolicy")
	ctxt := sdk.NewContext()
	ai, cleanAi, err := createAi(ctxt, sdk)
	if err != nil {
		t.Fatal(err)
	}
	init, cleanInit, err := createInitiator(ctxt, sdk)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanInit()
	defer cleanAi()
	time.Sleep(time.Second / 2)

	si := ai.StorageInstances[0]
	fmt.Printf("\nACL Policy: %#v\n", si.AclPolicy)
	resp, _, err := si.AclPolicy.Get(&dsdk.AclPolicyGetRequest{Ctxt: ctxt})
	if err != nil {
		t.Fatal(err)
	}
	acl := dsdk.AclPolicy(*resp)
	init.Name = ""
	init.Id = ""
	_, _, err = acl.Set(&dsdk.AclPolicySetRequest{
		Ctxt:       ctxt,
		Initiators: []*dsdk.Initiator{init},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTenants(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestTenants")
	tnts, _, err := sdk.Tenants.List(&dsdk.TenantsListRequest{Ctxt: sdk.NewContext()})
	if err != nil {
		t.Fatal(err)
	}
	for _, tnt := range tnts {
		fmt.Printf("Tenant: %s\n", tnt.Name)
	}
}

func TestSystem(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestSystem")
	sys, _, err := sdk.System.Get(&dsdk.SystemGetRequest{Ctxt: sdk.NewContext()})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("System: %s\n", dsdk.Pretty(sys))
}

// func TestPaging(t *testing.T) {
// 	sdk, err := dsdk.NewSDK(nil, true)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Running: TestPaging")
// 	cleanups := []func(){}
// 	ctxt := sdk.NewContext()
// 	w := 10
// 	workers := make(chan int, w)
// 	startAis := dsdk.NewStringSet(200)
// 	for i := 0; i < w; i++ {
// 		workers <- i
// 	}
// 	for i := 0; i < 200; i++ {
// 		w := <-workers
// 		go func() {
// 			ai, clean, err := createAi(ctxt, sdk)
// 			if err != nil {
// 				fmt.Println(err)
// 				workers <- w
// 				return
// 			}
// 			cleanups = append(cleanups, clean)
// 			startAis.Add(ai.Name)
// 			workers <- w
// 		}()
// 	}
// 	defer func() {
// 		for _, clean := range cleanups {
// 			w := <-workers
// 			go func(cleanup func()) {
// 				cleanup()
// 				workers <- w
// 			}(clean)
// 		}
// 		// Waiting for all workers to complete
// 		for i := 0; i < w; i++ {
// 			<-workers
// 		}
// 	}()
// 	ais, apierr, err := sdk.AppInstances.List(&dsdk.AppInstancesListRequest{
// 		Ctxt:   ctxt,
// 		Params: dsdk.ListParams{Limit: 0},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if apierr != nil {
// 		t.Fatal(fmt.Sprintf("%#v", apierr))
// 	}
// 	fmt.Printf("APPINSTANCES RESP LEN: %d\n", len(ais))
// 	endAis := dsdk.NewStringSet(200)
// 	for _, ai := range ais {
// 		endAis.Add(ai.Name)
// 	}
// 	for _, ai := range startAis.List() {
// 		if !endAis.Contains(ai) {
// 			t.Fatalf("Missing AppInstance %s from List results", ai)
// 		}
// 	}
// }

func TestLogUpload(t *testing.T) {
	sdk, err := dsdk.NewSDK(nil, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running: TestLogUpload")
	ctxt := sdk.NewContext()
	tmp, err := ioutil.TempFile("", "log-upload*.txt")
	fmt.Printf("MIME TYPE: (%s, %s)\n", tmp.Name(), mime.TypeByExtension(tmp.Name()))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	content := []byte(`This was a triumph
I'm making a note here:
Huge success
It's hard to overstate
My satisfaction
Aperture Science
We do what we must
Because we can
For the good of all of us
Except the ones who are dead
But there's no sense crying
Over every mistake
You just keep on trying
Till you run out of cake
And the Science gets done
And you make a neat gun
For the people who are
Still alive

I'm not even angry
I'm being so sincere right now
Even though you broke my heart
And killed me
And tore me to pieces
And threw every piece into a fire
As they burned it hurt because
I was so happy for you!
Now these points of data
Make a beautiful line
And we're out of beta
We're releasing on time
So I'm glad. I got burned
Think of all the things we learned
For the people who are
Still alive

Go ahead and leave me
I think I prefer to stay inside
Maybe you'll find someone else
To help you
Maybe Black Mesa...
That was A joke, ha ha, fat chance
Anyway this cake is great
It's so delicious and moist
Look at me still talking when there's science to do
When I look out there
It makes me glad I'm not you
I've experiments to be run
There is research to be done
On the people who are
Still alive

And believe me I am still alive
I'm doing science and I'm still alive
I feel fantastic and I'm still alive
And while you're dying I'll be still alive
And when you're dead I will be still alive
Still alive
Still alive
`)
	if _, err = tmp.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}
	_, apierr, err := sdk.LogsUpload.Upload(&dsdk.LogsUploadRequest{
		Ctxt:  ctxt,
		Files: []string{tmp.Name()},
	})
	if apierr != nil {
		t.Fatal(apierr)
	}
	if err != nil {
		t.Fatal(err)
	}
}
