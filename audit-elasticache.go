package main

import (
	"encoding/csv"
	"fmt"
	//	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"os"
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

func listCaches(region string, results chan<- []*elasticache.CacheCluster, outf *csv.Writer) {
	sess, sesserr := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if sesserr != nil {
		fmt.Println(os.Stderr, sesserr.Error())
	}
	svc := elasticache.New(sess)
	input := &elasticache.DescribeCacheClustersInput{}
	result, err := svc.DescribeCacheClusters(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elasticache.ErrCodeCacheClusterNotFoundFault:
				//fmt.Println(os.Stderr, "DAMMIT1", elasticache.ErrCodeCacheClusterNotFoundFault, aerr.Error())
				results <- result.CacheClusters
			case elasticache.ErrCodeInvalidParameterValueException:
				//fmt.Println(os.Stderr, "DAMMIT2", elasticache.ErrCodeInvalidParameterValueException, aerr.Error())
				results <- result.CacheClusters
			case elasticache.ErrCodeInvalidParameterCombinationException:
				//fmt.Println(os.Stderr, "DAMMIT3", elasticache.ErrCodeInvalidParameterCombinationException, aerr.Error())
				results <- result.CacheClusters
			default:
				//fmt.Println(os.Stderr, aerr.Error())
				results <- result.CacheClusters
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(os.Stderr, "DAMMIT" < err.Error())

		}
		return
	}

	//	cw := cloudwatch.New(sess)
	//	cws, _ := cw.ListMetrics(&cloudwatch.ListMetricsInput{
	//		MetricName: aws.String("CurrItems"),
	//		Namespace:  aws.String("AWS/ElastiCache"),
	//		Dimensions: []*cloudwatch.DimensionFilter{
	//			&cloudwatch.DimensionFilter{
	//				Name: aws.String(dimension),
	//			},
	//		},
	//	})
	//fmt.Println("Metrics", result.Metrics)

	for _, r := range result.CacheClusters {
		//fmt.Println(r)
		if err := outf.Write([]string{
			*r.CacheClusterId,
			*r.CacheClusterStatus,
			region,
			*r.PreferredAvailabilityZone,
			*r.CacheNodeType,
			*r.CacheParameterGroup.CacheParameterGroupName,
			*r.EngineVersion}); err != nil {
			fmt.Println("error writing record to csv:", err)
		}
	}

	results <- result.CacheClusters

}

func main() {
	regions := listRegions()
	results := make(chan []*elasticache.CacheCluster, len(regions))

	outf := csv.NewWriter(os.Stdout)
	if err := outf.Write([]string{
		"name",
		"status",
		"region",
		"availability_zone",
		"node_type",
		"parameter_group",
		"version"}); err != nil {
		fmt.Println("error writing record to csv:", err)
	}

	//start up a worker for each region
	for w := 0; w < len(regions); w++ {
		go listCaches(regions[w], results, outf)
	}
	for a := 1; a <= len(regions); a++ {
		<-results
	}

	close(results)
	outf.Flush()

}
