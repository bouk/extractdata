package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/bouk/extractdata/template"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	var hostedZoneId, bind string
	flag.StringVar(&hostedZoneId, "hosted-zone-id", "", "Hosted zone ID")
	flag.StringVar(&bind, "bind", ":8080", "Interface to bind on")
	flag.Parse()

	if hostedZoneId == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	svc := route53.New(session.New(), &aws.Config{
		Region: aws.String("us-east-1"),
	})
	router := httprouter.New()
	router.HandleMethodNotAllowed = false

	router.GET("/", func(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		port := ""
		host := req.Host
		var err error
		if strings.Contains(host, ":") {
			host, port, err = net.SplitHostPort(req.Host)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
		}

		if port == "" {
			bytes := make([]byte, 16)
			rand.Read(bytes)
			urls := make([]string, 0)

			for _, port := range []int64{6379, 11211, 9200} {
				u := &url.URL{
					Scheme:   "http",
					Host:     fmt.Sprintf("%x.%s:%d", bytes, host, port),
					Path:     "/_extract",
					RawQuery: req.URL.RawQuery,
				}
				urls = append(urls, u.String())
			}
			template.Home(rw, urls)
			return
		} else {
			http.NotFound(rw, req)
		}
	})

	router.GET("/_extract", func(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		host, port, err := net.SplitHostPort(req.Host)
		if err != nil || strings.Count(host, ".") <= 1 {
			http.NotFound(rw, req)
			return
		}

		ip := req.FormValue("ip")
		if ip == "" {
			ip = "127.0.0.1"
		}

		go svc.ChangeResourceRecordSets(
			&route53.ChangeResourceRecordSetsInput{
				HostedZoneId: aws.String(hostedZoneId),
				ChangeBatch: &route53.ChangeBatch{
					Changes: []*route53.Change{
						{
							Action: aws.String("CREATE"),
							ResourceRecordSet: &route53.ResourceRecordSet{
								Name: aws.String(host),
								Type: aws.String("A"),
								ResourceRecords: []*route53.ResourceRecord{
									{
										Value: aws.String(ip),
									},
								},
								TTL: aws.Int64(60),
							},
						},
					},
				},
			},
		)

		switch port {
		case "6379", "16379":
			template.RedisExtract(rw)
		case "11211":
			template.MemcachedExtract(rw)
		case "9200":
			template.ElasticsearchExtract(rw)
		default:
			http.Error(rw, "unknown service port", http.StatusNotFound)
		}
	})

	http.ListenAndServe(bind, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		header := rw.Header()
		header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		header.Set("Pragma", "no-cache")
		header.Set("Expires", "0")
		header.Set("Connection", "close")
		router.ServeHTTP(rw, req)
	}))
}
