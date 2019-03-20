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

func listCaches(region string, results chan<- []*elasticache.CacheCluster) {
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
				results <- result.CacheClusters
			case elasticache.ErrCodeInvalidParameterValueException:
				fmt.Println(elasticache.ErrCodeInvalidParameterValueException, aerr.Error())
				results <- result.CacheClusters
			case elasticache.ErrCodeInvalidParameterCombinationException:
				fmt.Println(elasticache.ErrCodeInvalidParameterCombinationException, aerr.Error())
				results <- result.CacheClusters
			default:
				fmt.Println(aerr.Error())
				results <- result.CacheClusters
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())

		}
		return
	}

	fmt.Println(result.CacheClusters)
	results <- result.CacheClusters

}

func main() {
	regions := listRegions()
	results := make(chan []*elasticache.CacheCluster, len(regions))

	//start up a worker for each region
	for w := 0; w < len(regions); w++ {
		go listCaches(regions[w], results)
	}
	for a := 1; a <= len(regions); a++ {
		<-results
	}
	close(results)

}
