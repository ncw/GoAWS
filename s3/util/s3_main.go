package s3_util

import (
	"errors"
	"github.com/abneptis/GoAWS"
	. "github.com/abneptis/GoAWS/flags"
	"github.com/abneptis/GoAWS/s3"
	. "github.com/abneptis/GoAWS/util/common"
	"net/url"
)

import (
	"flag"
	"fmt"
	"os"
	"path"
)

// Safety warning
// These are globals to allow the code to be more readable,
// since the tool is "single-tasked" it has no threading issues.
//
// You are of course encouraged to take a more thread-safe approach
// if you intend to use multiple threads.

var flag_endpoint_url string = ""
var signer *aws.Signer
var service *s3.Service

// Convenience method to clean up calls.
func DefaultS3Service() (id *aws.Signer, s *s3.Service, err error) {
	id, err = DefaultSigner()
	if err == nil {
		url_, err := url.Parse(flag_endpoint_url)
		if err == nil {
			s = s3.NewService(url_)
		}
	}
	return
}

func init() {
	AddModule("s3", func() {
		flag.StringVar(&flag_endpoint_url, "s3-endpoint", "https://s3.amazonaws.com/", "Endpoint to use for S3 calls")
	})

	Modules["s3"].Setup = func() (err error) {
		signer, service, err = DefaultS3Service()
		return
	}

	// awstool.s3.ls
	Modules["s3"].Calls["ls"] = func(args []string) (err error) {
		if len(args) != 1 {
			return errors.New("USAGE: ls BUCKET")
		}
		keys := make(chan string)
		go func() {
			for i := range keys {
				fmt.Printf("%s\n", i)
			}
		}()
		err = service.Bucket(args[0]).ListKeys(signer, "", "", "", keys)
		if err != nil {
			close(keys)
		}
		return
	}

	// awstool.s3.buckets
	Modules["s3"].Calls["buckets"] = func(args []string) (err error) {
		if len(args) != 0 {
			return errors.New("USAGE: buckets")
		}
		lb, err := service.ListBuckets(signer)
		if err == nil {
			for b := range lb {
				fmt.Println(lb[b])
			}
		}
		return
	}

	Modules["s3"].Calls["cat"] = func(args []string) (err error) {
		if len(args) != 2 {
			return errors.New("Usage: get BUCKET KEY")
		}
		if err == nil {
			_, err = service.Bucket(args[0]).GetKey(signer, args[1], os.Stdout)
		}
		return
	}
	Modules["s3"].Calls["exists"] = func(args []string) (err error) {
		if len(args) == 2 {
			fmt.Printf("Usage: exists BUCKET KEY\n")
			os.Exit(1)
		}
		err = service.Bucket(args[0]).Exists(signer, args[1])
		return
	}
	Modules["s3"].Calls["drop"] = func(args []string) (err error) {
		if len(args) != 1 {
			return errors.New("Usage: drop BUCKET")
		}
		err = service.DeleteBucket(signer, args[0])
		return
	}
	Modules["s3"].Calls["create"] = func(args []string) (err error) {
		if len(args) != 1 {
			return errors.New("Usage: create BUCKET")
		}
		err = service.CreateBucket(signer, args[0])
		return
	}
	Modules["s3"].Calls["rm"] = func(args []string) (err error) {
		if len(args) < 2 {
			return errors.New("Usage: rm BUCKET KEY [KEY2...]")
		}
		bucket := args[0]
		args = args[1:]
		for i := range args {
			err = service.Bucket(bucket).Delete(signer, args[i])
			if err != nil {
				break
			}
		}
		return
	}
	Modules["s3"].Calls["put"] = func(args []string) (err error) {
		if len(args) < 3 {
			return errors.New("Usage: put BUCKET PREFIX FILE [FILE2...]")
		}
		bucket := args[0]
		prefix := args[1]
		keys := args[2:]
		for i := range keys {
			err = service.Bucket(bucket).PutLocalFile(signer, path.Join(prefix, path.Base(keys[i])), keys[i])
			if err != nil {
				break
			}
		}
		return
	}

}

func Nil() {}
