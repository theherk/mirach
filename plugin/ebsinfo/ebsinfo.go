package ebsinfo

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.eng.cleardata.com/dash/mirach/plugin"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/envinfo"
)

type EbsInfoGroup struct {
	// Volumes
	Volumes []Volume
}

type Volume struct {
	Encrypted          bool   `json:"encrypted"`
	Size               int    `json:"size"`
	CreateTime         string `json:"created"`
	VolumeID           string `json:"volume_id"`
	AttachedInstanceID string `json:"instance_id"`
}

// use instance id off of env info if in aws describe volumes
func (e *EbsInfoGroup) GetInfo() {
	instanceID := envinfo.Env.CloudProviderInfo["instance-id"]
	region := envinfo.Env.CloudProviderInfo["region"]
	sess, err := session.NewSession()
	if err != nil {
		jww.DEBUG.Printf("ebsinfo plugin encountered %s", err)
	}
	svc := ec2.New(sess, &aws.Config{Region: aws.String(region)})
	instance, err := e.getEc2Instance(svc, instanceID)
	if err != nil {
		jww.ERROR.Println(
			"ebsinfo plugin encountered an error describing instances, ensure that it has permissions to perform ec2.DescribeInstances",
		)
		fmt.Println(err)
	}

	fmt.Println(instance)
	fmt.Println("here")
	volumes, err := e.getInstanceVolumes(svc, instance)
	fmt.Println(volumes)
}

func (e *EbsInfoGroup) String() string {
	s, _ := json.Marshal(e)
	return string(s)
}

// get the volumes on this ec2
func (e *EbsInfoGroup) getInstanceVolumes(svc *ec2.EC2, instance *ec2.Instance) ([]Volume, error) {
	volumeIDs := []*string{}
	for _, bdm := range instance.BlockDeviceMappings {
		volumeID := bdm.Ebs.VolumeId
		volumeIDs = append(volumeIDs, volumeID)
	}
	res, err := svc.DescribeVolumes(
		&ec2.DescribeVolumesInput{
			VolumeIds: volumeIDs,
		},
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(res)
	return []Volume{}, nil
}

// return this ec2
func (e *EbsInfoGroup) getEc2Instance(svc *ec2.EC2, instanceID string) (*ec2.Instance, error) {
	instances := []*string{&instanceID}
	res, err := svc.DescribeInstances(
		&ec2.DescribeInstancesInput{
			InstanceIds: instances,
		},
	)
	if err != nil {
		return nil, err
	}
	return res.Reservations[0].Instances[0], nil
}

//get info
func GetInfo() plugin.InfoGroup {
	info := new(EbsInfoGroup)
	info.GetInfo()
	return info
}

func String() string {
	info := new(EbsInfoGroup)
	info.GetInfo()
	return info.String()
}
