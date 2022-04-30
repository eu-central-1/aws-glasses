package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pborman/getopt"
)

func ToStrings(vs []ec2types.AccountAttributeValue) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = aws.ToString(v.AttributeValue)
	}
	return vsm
}

func main() {
	optProfile := getopt.StringLong("profile", 0, "default", "Profile defined in ~/.aws/config")
	optHelp := getopt.BoolLong("help", '?', "Show Help")
	getopt.Parse()

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile(*optProfile),
	)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	stsClient := sts.NewFromConfig(cfg)

	identity, err := stsClient.GetCallerIdentity(
		context.TODO(),
		&sts.GetCallerIdentityInput{},
	)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	fmt.Printf(
		"Account: %s\nUserID: %s\nARN: %s\n",
		aws.ToString(identity.Account),
		aws.ToString(identity.UserId),
		aws.ToString(identity.Arn),
	)

	// Create new EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Call to get detailed information on AWS account
	attributes, err := ec2Client.DescribeAccountAttributes(
		context.TODO(),
		&ec2.DescribeAccountAttributesInput{},
	)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Attributes: ")
		for _, element := range attributes.AccountAttributes {
			fmt.Println(
				aws.ToString(element.AttributeName),
				" :",
				strings.Join(ToStrings(element.AttributeValues), ", "),
			)
		}
	}

	orgClient := organizations.NewFromConfig(cfg)

	// Call to get accounts inside AWS organisations
	accounts, err := orgClient.ListAccounts(
		context.TODO(),
		&organizations.ListAccountsInput{},
	)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Accounts: ")
		for i, account := range accounts.Accounts {
			fmt.Printf(
				"Nr %d ID: %s\n",
				i,
				aws.ToString(account.Id),
			)
		}
	}

}
