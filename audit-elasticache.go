package main

import (
	"encoding/csv"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"os"
	"time"
)

//func grabStats(region string, clusterId string) {
func grabStats(region string, instance string) float64 {
	s := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(region)}}))
	svc := cloudwatch.New(s)
	endTime := time.Now()
	duration, _ := time.ParseDuration("-1h")
	startTime := endTime.Add(duration)
	namespace := "AWS/ElastiCache"
	metricname := "CurrItems"
	metricid := "noeffingidea"
	period := int64(3600)
	stat := "Average"
	query := &cloudwatch.MetricDataQuery{
		Id: &metricid,
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				Namespace:  &namespace,
				MetricName: &metricname,
				Dimensions: []*cloudwatch.Dimension{
					//Dimensions: []*cloudwatch.DimensionFilter{
					&cloudwatch.DimensionFilter{
						Name:  aws.String("CacheClusterId"),
						Value: aws.String(instance),
					},
				},
			},
			Period: &period,
			Stat:   &stat,
		},
	}
	resp, err := svc.GetMetricData(&cloudwatch.GetMetricDataInput{
		EndTime:           &endTime,
		StartTime:         &startTime,
		MetricDataQueries: []*cloudwatch.MetricDataQuery{query},
	})

	if err != nil {
		fmt.Println("Got error getting metric data")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	j := *resp.MetricDataResults[0]
	return (*j.Values[0])

}

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

	for _, r := range result.CacheClusters {
		//fmt.Println(r)
		keyCount := grabStats(region, *r.CacheClusterId)
		if err := outf.Write([]string{
			*r.CacheClusterId,
			*r.CacheClusterStatus,
			region,
			*r.PreferredAvailabilityZone,
			*r.CacheNodeType,
			*r.CacheParameterGroup.CacheParameterGroupName,
			fmt.Sprintf("%f", keyCount),
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
		"key_count",
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
