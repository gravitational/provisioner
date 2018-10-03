package provisioner

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gravitational/provisioner/provider/awsutil"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

const (
	// AWSOperationTimeout is the amount of time to wait for calls to AWS to complete
	AWSOperationTimeout = 30 * time.Second
)

// Loader governs process of inspecting VPC and generating TerraForm template
// We support 2 modes of loading
//   * load with an existing VPC: in this case, we will re-use nat gateway,
//     internet gateway, only subnet and security group will be created
//   * load without a VPC: in this case, a whole new VPC will be created.

type Loader struct {
	LoaderConfig
	EC2 awsutil.EC2Service
	*s3.S3
}

// LoaderConfig holds configuration for Loader to work
type LoaderConfig struct {
	VPCID         string
	Region        string
	TemplatePath  string
	ClusterBucket string
}

// NewLoader initializes Loader from a LoadConfig and related AWS Service
func NewLoader(config LoaderConfig) (*Loader, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	svc := ec2.New(sess)
	s3 := s3.New(sess)
	return &Loader{
		S3:           s3,
		EC2:          svc,
		LoaderConfig: config,
	}, nil
}

// loadTemplate loads a Terraform template from a file into an instance of
// template.Template to generate the final script depending on creating a new
// VPC or using an existing VPC
func (l *Loader) loadTemplate() (*template.Template, error) {
	out, err := ioutil.ReadFile(l.TemplatePath)
	if err != nil {
		return nil, trace.ConvertSystemError(err)
	}
	t, err := template.New("terraform").Funcs(template.FuncMap{"counter": counter}).Parse(string(out))
	if err != nil {
		return nil, trace.ConvertSystemError(err)
	}

	return t, nil
}

// templateForNewVPC generates final TerraForm script when creating a VPC
func (l *Loader) templateForNewVPC() ([]byte, error) {
	tpl, err := l.loadTemplate()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	var vpc *awsutil.VPC
	vpc, err = awsutil.NewVPC(l.EC2, l.Region)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, vpc.GenVars())
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return buf.Bytes(), nil
}

// load generates final TerraForm script when using an existing VPC
func (l *Loader) load() ([]byte, error) {
	if l.VPCID == "" {
		return l.templateForNewVPC()
	}

	natGateways, err := l.loadNatGateways()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	log.Debugf("loaded nat gateways: %v", natGateways)

	// collect public subnets information associated
	// with nat gateways, so we can set routing properly
	var subnetIDs []string
	for _, gateway := range natGateways {
		subnetIDs = append(subnetIDs, *gateway.SubnetId)
	}

	subnets, err := l.loadSubnets(subnetIDs)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	log.Debugf("loaded subnets: %v", subnets)

	// detect regions
	var regionName string
	for _, subnet := range subnets {
		regionName, err = l.loadRegion(*subnet.AvailabilityZone)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		break
	}

	// compute subnet ranges
	privateSubnets, publicSubnets, err := l.computeSubnetRanges(len(subnets))
	if err != nil {
		return nil, trace.Wrap(err)
	}
	log.Debugf("computed subnet ranges: %v", privateSubnets, publicSubnets)

	tpl, err := l.loadTemplate()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	awsVars := map[string]interface{}{}
	vars := map[string]interface{}{
		"variables": map[string]interface{}{
			"aws": awsVars,
		},
	}

	awsVars["subnets"] = privateSubnets
	awsVars["public_subnets"] = publicSubnets
	awsVars["region"] = regionName
	awsVars["vpc_id"] = l.VPCID

	// build nat gateway per AZ map
	natGatewayIDs := []string{}
	for _, gateway := range natGateways {
		natGatewayIDs = append(natGatewayIDs, *gateway.NatGatewayId)
	}

	awsVars["nat_gateways"] = natGatewayIDs

	// build availability zones deterministic array
	var azNames []string
	for _, subnet := range subnets {
		azNames = append(azNames, *subnet.AvailabilityZone)
	}
	awsVars["azs"] = azNames
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, vars)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return buf.Bytes(), nil
}

// counter is a template helper function to generate an increasing counter starting from 0
func counter() func() int {
	i := -1
	return func() int {
		i++
		return i
	}
}

// loadRegion fetchs AWS region from an availability zone
func (l *Loader) loadRegion(az string) (string, error) {
	params := &ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{
			{ // Required
				Name: aws.String("zone-name"),
				Values: []*string{
					aws.String(az),
				},
			},
		},
	}
	req, out := l.EC2.DescribeAvailabilityZonesRequest(params)
	if err := req.Send(); err != nil {
		return "", trace.Wrap(err)
	}
	if len(out.AvailabilityZones) == 0 {
		return "", trace.NotFound("no AZ with name %v found", az)
	}
	return aws.StringValue(out.AvailabilityZones[0].RegionName), nil
}

// loadSubnets fetchs ec2.Subnet struct with a given array of subnet id
func (l *Loader) loadSubnets(subnetIDs []string) ([]*ec2.Subnet, error) {
	filter := &ec2.Filter{
		Name:   aws.String("subnet-id"),
		Values: []*string{},
	}
	for _, subnetID := range subnetIDs {
		filter.Values = append(filter.Values, aws.String(subnetID))
	}
	params := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{filter},
	}
	req, out := l.EC2.DescribeSubnetsRequest(params)
	if err := req.Send(); err != nil {
		return nil, trace.Wrap(err)
	}
	if len(out.Subnets) == 0 {
		return nil, trace.NotFound("no subnets with ids %v found", subnetIDs)
	}
	return out.Subnets, nil
}

func (l *Loader) loadAllSubnets() ([]*ec2.Subnet, error) {
	params := &ec2.DescribeSubnetsInput{}
	req, out := l.EC2.DescribeSubnetsRequest(params)
	if err := req.Send(); err != nil {
		return nil, trace.Wrap(err)
	}
	return out.Subnets, nil
}

func (l *Loader) computeSubnetRanges(count int) ([]string, []string, error) {
	vpc, err := l.loadVPC()
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}
	log.Debugf("vpc %v subnet: %v", vpc, count)
	subnets, err := l.loadAllSubnets()
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}
	log.Debugf("all subnets: %v", subnets)
	var cidrs []string
	for _, subnet := range subnets {
		cidrs = append(cidrs, *subnet.CidrBlock)
	}

	var out []string
	for i := 0; i < count*2; i++ {
		next, err := awsutil.SelectVPCSubnet(*vpc.CidrBlock, cidrs)
		if err != nil {
			return nil, nil, trace.Wrap(err)
		}
		out = append(out, next)
		cidrs = append(cidrs, next)
	}

	return out[:len(out)/2], out[len(out)/2:], nil
}

// loadNatGateways finds all NAT gateways in current vpc
func (l *Loader) loadNatGateways() ([]*ec2.NatGateway, error) {
	params := &ec2.DescribeNatGatewaysInput{
		Filter: []*ec2.Filter{
			{ // Required
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(l.VPCID),
				},
			},
		},
	}
	req, out := l.EC2.DescribeNatGatewaysRequest(params)
	if err := req.Send(); err != nil {
		return nil, trace.Wrap(err)
	}
	if len(out.NatGateways) == 0 {
		return nil, trace.NotFound("no nat gateways found")
	}
	return out.NatGateways, nil
}

// loadVPC fetchs ec2.Vpc structs from our vpc id
func (l *Loader) loadVPC() (*ec2.Vpc, error) {
	params := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{ // Required
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(l.VPCID),
				},
			},
		},
	}
	req, out := l.EC2.DescribeVpcsRequest(params)
	if err := req.Send(); err != nil {
		return nil, trace.Wrap(err)
	}
	if len(out.Vpcs) == 0 {
		return nil, trace.NotFound("no VPC with id %v found", l.VPCID)
	}

	return out.Vpcs[0], nil
}

// UpsertBucket upserts bucket if it does not exist
func (l *Loader) UpsertBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), AWSOperationTimeout)
	defer cancel()

	input := &s3.CreateBucketInput{
		Bucket: aws.String(l.ClusterBucket),
		ACL:    aws.String("private"),
	}
	_, err := l.CreateBucketWithContext(ctx, input)
	err = awsutil.ConvertS3Error(err, "bucket %s already exists", aws.String(l.ClusterBucket))
	if err != nil {
		if !trace.IsAlreadyExists(err) {
			return err
		}
	}
	ver := &s3.PutBucketVersioningInput{
		Bucket: aws.String(l.ClusterBucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String("Enabled"),
		},
	}
	_, err = l.PutBucketVersioningWithContext(ctx, ver)
	err = awsutil.ConvertS3Error(err, "failed to set versioning state for bucket %s", aws.String(l.ClusterBucket))
	if err != nil {
		return err
	}
	return nil
}

// GetKey downloads key if it exists, otherwise returns NotFound
func (l *Loader) GetKey(ctx context.Context, bucketName, bucketKey string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(bucketKey),
	}
	result, err := l.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, awsutil.ConvertS3Error(err)
	}
	defer result.Body.Close()
	return ioutil.ReadAll(result.Body)
}

func (l *Loader) PutBytes(bucketName, bucketKey string, data []byte) error {
	return l.PutKey(bucketName, bucketKey, bytes.NewReader(data), int64(len(data)), "text/plain")
}

// PutKey puts key to the remote storage
func (l *Loader) PutKey(bucketName, bucketKey string, out io.ReadSeeker, contentSize int64, contentType string) error {
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(bucketKey),
		Body:          out,
		ContentLength: aws.Int64(contentSize),
		ContentType:   aws.String(contentType),
	}
	_, err := l.PutObject(params)
	if err != nil {
		return awsutil.ConvertS3Error(err, "failed to write key %s to bucket %s", bucketKey, bucketName)
	}
	return nil
}

func (l *Loader) initVars(bucketKey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), AWSOperationTimeout)
	defer cancel()

	err := l.UpsertBucket()
	if err != nil {
		return trace.Wrap(err)
	}
	_, err = l.GetKey(ctx, l.ClusterBucket, bucketKey)
	if err == nil {
		log.Printf("found vars in s3://%v/%v", l.ClusterBucket, bucketKey)
		return nil
	}
	if !trace.IsNotFound(err) {
		return trace.Wrap(err, "failed to load key: %v", bucketKey)
	}
	log.Printf("key is not found, going to generate data from AWS")
	data, err := l.load()

	if err != nil {
		return trace.Wrap(err, "failed to load data from AWS")
	}
	err = l.PutBytes(l.ClusterBucket, bucketKey, data)
	if err != nil {
		return trace.Wrap(err)
	}
	log.Printf("uploaded vars to s3://%v/%v", l.ClusterBucket, bucketKey)
	return nil
}

func (l *Loader) sync(paths []string, targetDir string) error {
	log.Debug("starting sync operation")
	ctx, cancel := context.WithTimeout(context.Background(), AWSOperationTimeout)
	defer cancel()

	log.WithField("targetDir", targetDir).Debug("creating target directory")
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		return trace.ConvertSystemError(err)
	}

	for _, path := range paths {
		params := &s3.ListObjectsInput{
			Bucket: aws.String(l.ClusterBucket),
			Prefix: aws.String(path),
		}
		log.WithFields(log.Fields{
			"Bucket": params.Bucket,
			"Prefix": params.Prefix,
		}).Debug("listing objects")
		resp, err := l.ListObjectsWithContext(ctx, params)
		if err != nil {
			return trace.Wrap(err)
		}
		for _, key := range resp.Contents {
			targetFile := filepath.Join(targetDir, filepath.Base(*key.Key))
			log.Printf("syncing s3://%v/%v to %v", l.ClusterBucket, *key.Key, targetFile)
			data, err := l.GetKey(ctx, l.ClusterBucket, *key.Key)
			if err != nil {
				return trace.Wrap(err)
			}
			err = ioutil.WriteFile(targetFile, data, 0644)
			if err != nil {
				return trace.Wrap(err)
			}
		}

	}

	log.Debug("sync complete")
	return nil
}

func (l *Loader) rm(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), AWSOperationTimeout)
	defer cancel()

	params := &s3.DeleteObjectInput{
		Bucket: aws.String(l.ClusterBucket),
		Key:    aws.String(key),
	}
	_, err := l.DeleteObjectWithContext(ctx, params)
	if err == nil {
		log.Printf("removed key s3://%v/%v", l.ClusterBucket, key)
	}
	return awsutil.ConvertS3Error(err, "failed to remove key %s", key)
}
