package main


//import "com.abneptis.oss/aws"
//import "com.abneptis.oss/aws/sqs"
import "com.abneptis.oss/aws/s3"
//import "http"
import "flag"
//import "fmt"
import "log"

func main(){
  flag.Parse()
  id, err := GetAWSIdentity()
  if err != nil {
    log.Exitf("Unable to get AWS identity: %v", err)
  }
  ep, err := GetS3Endpoint()
  if err != nil {
    log.Exitf("Unable to construct endpoint: %v", err)
  }
  bucket := s3.NewBucket(ep, flag.Arg(0))
  o, err := bucket.GetKey(id, flag.Arg(1))
  if err != nil {
    log.Exitf("Couldn't get key: %v", err)
  }
  log.Printf("Object: %v", o)
}