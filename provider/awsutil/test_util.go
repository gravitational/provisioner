package awsutil

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var testZones = []string{"us-west-1", "us-west-2", "us-west-3"}

type MockEC2Service struct{}

// mockRequest mocks all EC2 HTTP requests with a no-op
func mockRequest() *request.Request {
	r := request.Request{}
	hl := request.HandlerList{
		AfterEachFn: func(item request.HandlerListRunItem) bool { return true },
	}

	r.Handlers = request.Handlers{
		Validate:         hl,
		Build:            hl,
		Send:             hl,
		Sign:             hl,
		ValidateResponse: hl,
		Unmarshal:        hl,
		UnmarshalMeta:    hl,
		UnmarshalError:   hl,
		Retry:            hl,
		AfterRetry:       hl,
		Complete:         hl,
	}

	return &r
}

func (svc *MockEC2Service) DescribeAvailabilityZones(*ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error) {
	azs := make([]*ec2.AvailabilityZone, 3)
	for i := 0; i < 3; i++ {
		azs[i] = &ec2.AvailabilityZone{
			ZoneName: &testZones[i],
		}
	}

	return &ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: azs,
	}, nil
}

func (svc *MockEC2Service) DescribeAvailabilityZonesRequest(*ec2.DescribeAvailabilityZonesInput) (*request.Request, *ec2.DescribeAvailabilityZonesOutput) {
	azs := make([]*ec2.AvailabilityZone, 3)
	for i := 0; i < 3; i++ {
		azs[i] = &ec2.AvailabilityZone{
			ZoneName:   &testZones[i],
			RegionName: aws.String("us-west"),
		}
	}

	return mockRequest(), &ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: azs,
	}
}
func (svc *MockEC2Service) DescribeSubnetsRequest(*ec2.DescribeSubnetsInput) (*request.Request, *ec2.DescribeSubnetsOutput) {
	return mockRequest(), &ec2.DescribeSubnetsOutput{
		Subnets: []*ec2.Subnet{
			&ec2.Subnet{AvailabilityZone: aws.String("us-west-1"), CidrBlock: aws.String("10.0.0.0/24")},
			&ec2.Subnet{AvailabilityZone: aws.String("us-west-2"), CidrBlock: aws.String("10.0.1.0/24")},
			&ec2.Subnet{AvailabilityZone: aws.String("us-west-3"), CidrBlock: aws.String("10.0.2.0/24")},
		},
	}
}
func (svc *MockEC2Service) DescribeNatGatewaysRequest(*ec2.DescribeNatGatewaysInput) (*request.Request, *ec2.DescribeNatGatewaysOutput) {
	gwo := ec2.DescribeNatGatewaysOutput{
		NatGateways: []*ec2.NatGateway{
			&ec2.NatGateway{SubnetId: aws.String("s1"), NatGatewayId: aws.String("ngw1")},
			&ec2.NatGateway{SubnetId: aws.String("s2"), NatGatewayId: aws.String("ngw2")},
			&ec2.NatGateway{SubnetId: aws.String("s3"), NatGatewayId: aws.String("ngw3")},
		},
	}

	return mockRequest(), &gwo
}
func (svc *MockEC2Service) DescribeInternetGatewaysRequest(*ec2.DescribeInternetGatewaysInput) (*request.Request, *ec2.DescribeInternetGatewaysOutput) {
	igwID := "igw-1"
	igw := ec2.InternetGateway{
		InternetGatewayId: &igwID,
	}

	gwo := ec2.DescribeInternetGatewaysOutput{
		InternetGateways: []*ec2.InternetGateway{&igw},
	}

	return mockRequest(), &gwo
}
func (svc *MockEC2Service) DescribeVpcsRequest(*ec2.DescribeVpcsInput) (*request.Request, *ec2.DescribeVpcsOutput) {
	return mockRequest(), &ec2.DescribeVpcsOutput{
		Vpcs: []*ec2.Vpc{
			&ec2.Vpc{CidrBlock: aws.String("10.0.0.0/16"), VpcId: aws.String("vpc-1")},
		},
	}
}
