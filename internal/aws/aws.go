package aws

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func UpdateKubeconfig(region, clusterName string) error {
	fmt.Printf("[AWS] Updating local kubeconfig for EKS cluster '%s'...\n", clusterName)
	cmd := exec.Command("aws", "eks", "update-kubeconfig", "--region", region, "--name", clusterName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update kubeconfig: %w", err)
	}

	return nil
}

// CheckCredentials verifies that AWS credentials can be resolved from the environment.
func CheckCredentials() error {
	fmt.Println("Checking for AWS credentials...")
	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	if _, err := sess.Config.Credentials.Get(); err != nil {
		return fmt.Errorf("could not resolve AWS credentials. Please run 'aws configure' or set up credentials via environment variables or an IAM role")
	}

	fmt.Println("✓ AWS credentials found.")
	return nil
}

// CheckClusterExists checks if an EKS cluster with the given name already exists in the specified region.
func CheckClusterExists(region, clusterName string) (bool, error) {
	fmt.Printf("--> Checking if EKS cluster '%s' already exists in region %s...\n", clusterName, region)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return false, fmt.Errorf("failed to create AWS session for cluster check: %w", err)
	}

	eksSvc := eks.New(sess)

	_, err = eksSvc.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == eks.ErrCodeResourceNotFoundException {
			return false, nil // Cluster does not exist, which is the desired state for provisioning.
		}
		return false, fmt.Errorf("error describing EKS cluster: %w", err) // Another error occurred.
	}

	return true, nil // No error means the cluster was found.
}

// CheckEC2InstanceExists checks if an EC2 instance with the given Name tag exists in the specified region.
func CheckEC2InstanceExists(region, instanceName string) (bool, string, string, error) {
	fmt.Printf("--> Checking if EC2 instance '%s' exists in region %s...\n", instanceName, region)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return false, "", "", fmt.Errorf("failed to create AWS session for ec2 check: %w", err)
	}

	ec2Svc := ec2.New(sess)

	result, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String(instanceName)},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running")},
			},
		},
	})

	if err != nil {
		return false, "", "", fmt.Errorf("error describing EC2 instances: %w", err)
	}

	hasInstance := false
	var instanceID string
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			hasInstance = true
			instanceID = *instance.InstanceId
			if instance.PublicIpAddress != nil {
				return true, *instance.PublicIpAddress, instanceID, nil
			}
			if instance.PrivateIpAddress != nil {
				return true, *instance.PrivateIpAddress, instanceID, nil
			}
		}
	}

	if hasInstance {
		return true, "", instanceID, nil // Found but no IP assigned?
	}

	return false, "", "", nil
}

// NewSSMClient creates and returns a new AWS Systems Manager (SSM) client.
func NewSSMClient(region string) (*ssm.SSM, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session for SSM: %w", err)
	}
	return ssm.New(sess), nil
}

// EnsureSSMRoleAttached checks if the SSM role exists, creates it if not, and attaches it to the instance.
func EnsureSSMRoleAttached(region, instanceID string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session for IAM: %w", err)
	}

	iamSvc := iam.New(sess)
	ec2Svc := ec2.New(sess)

	roleName := "StackPulseSSMRole"
	profileName := "StackPulseSSMInstanceProfile"

	// 1. Create or get Role
	_, err = iamSvc.GetRole(&iam.GetRoleInput{RoleName: aws.String(roleName)})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == iam.ErrCodeNoSuchEntityException {
			trustPolicy := `{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {"Service": "ec2.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}
				]
			}`
			_, err = iamSvc.CreateRole(&iam.CreateRoleInput{
				RoleName:                 aws.String(roleName),
				AssumeRolePolicyDocument: aws.String(trustPolicy),
			})
			if err != nil {
				return fmt.Errorf("failed to create IAM role: %w", err)
			}

			_, err = iamSvc.AttachRolePolicy(&iam.AttachRolePolicyInput{
				RoleName:  aws.String(roleName),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
			})
			if err != nil {
				return fmt.Errorf("failed to attach SSM policy to role: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check IAM role: %w", err)
		}
	}

	// 2. Create or get Instance Profile
	_, err = iamSvc.GetInstanceProfile(&iam.GetInstanceProfileInput{InstanceProfileName: aws.String(profileName)})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == iam.ErrCodeNoSuchEntityException {
			_, err = iamSvc.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
				InstanceProfileName: aws.String(profileName),
			})
			if err != nil {
				return fmt.Errorf("failed to create instance profile: %w", err)
			}

			_, err = iamSvc.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
				InstanceProfileName: aws.String(profileName),
				RoleName:            aws.String(roleName),
			})
			if err != nil {
				return fmt.Errorf("failed to add role to instance profile: %w", err)
			}

			fmt.Println("--> ⏳ Waiting 10 seconds for IAM instance profile propagation...")
			time.Sleep(10 * time.Second)
		} else {
			return fmt.Errorf("failed to check instance profile: %w", err)
		}
	}

	// 3. Attach to EC2 Instance
	_, err = ec2Svc.AssociateIamInstanceProfile(&ec2.AssociateIamInstanceProfileInput{
		InstanceId: aws.String(instanceID),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(profileName),
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already has an IAM instance profile") {
			fmt.Println("--> ℹ️  Instance already has an IAM profile attached. Ensure it contains the SSM policy.")
			return nil
		}
		return fmt.Errorf("failed to associate instance profile to EC2: %w", err)
	}

	fmt.Println("--> ✅ IAM Role attached successfully! Waiting for SSM Agent to register (approx. 45s)...")
	time.Sleep(45 * time.Second)

	return nil
}
