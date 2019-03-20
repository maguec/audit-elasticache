package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
)

func listRegions() []string {
	var regions []string
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()
	for _, p := range partitions {
		for id := range p.Regions() {
			regions = append(regions, id)
		}

	}
	return regions
}

func listCaches(region string) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	svc := elasticache.New(sess)
	input := &elasticache.DescribeCacheClustersInput{}
	result, err := svc.DescribeCacheClusters(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elasticache.ErrCodeCacheClusterNotFoundFault:
				fmt.Println(elasticache.ErrCodeCacheClusterNotFoundFault, aerr.Error())
			case elasticache.ErrCodeInvalidParameterValueException:
				fmt.Println(elasticache.ErrCodeInvalidParameterValueException, aerr.Error())
			case elasticache.ErrCodeInvalidParameterCombinationException:
				fmt.Println(elasticache.ErrCodeInvalidParameterCombinationException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)

}

func main() {
	regions := listRegions()
	fmt.Println(regions)
	listCaches(regions[0])
}
