package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context/ctxhttp"
)

const URL = "someUrl"
const QPS = 300
const TEST_TIME_SECONDS = 60 * 20
const LOG_SUCCESS = false
const LOG_ERROR = false

func main() {
	// init logrus
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	// exemplary and supplementary requests to select from
	input := []string{
		/*0*/ `{"puts":[{"type":"xml","value":"<VAST version=\\\"3.0\\\"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><5YnPFbYABozbp1eJG4BYABAYoBA1VTRJIBAQbwUpgBAaABAagBAbABALgBA8ABBMgBAtABANgBAOABAPABAIoCO3VmKCdhJywgMjUyOTg4NSwgMTU3ODMxMzc2NSk7dWYoJ3InLCA5NzUxNzc3MSwgLh4A8PWSArUCIXlqd21vZ2kyLUx3S0VNdUJ3QzRZQUNDYzhWc3dBRGdBUUFSSTdVaFE2ZEduQmxnQVlNSUdhQUJ3S0hnR2dBRThpQUVHa0FFQW1BRUFvQUVCcUFFRHNBRUF1UUh6cldxa0FBQVVRTUVCODYxcXBBQUFGRURKQWZmNTBxUXI0ZW9fMlFFQUFBQUFBQUR3UC1BQkFQVUJBQUFBQUpnQ0FLQUNBTFVDQUFBQUFMMENBQUFBQU9BQ0FPZ0NBUGdDQUlBREFaZ0RBYWdEdHZpOENyb0RDVk5KVGpNNk5EY3pPT0FELVJpSUJBQ1FCQUNZQkFIQkIFRQkBCHlRUQkJAQEUTmdFQVBFEY0BkCw0QkFDSUJZSWxxUVUBEQEUPHdQdy4umgKJASFMZzlaRFE2OQEkblBGYklBUW9BRBVIVFVRRG9KVTBsT016bzBOek00UVBrWVMReAxQQV9VEQwMQUFBVx0MAFkdDABhHQwAYx0M8FJlQUEuwgI_aHR0cDovL3ByZWJpZC5vcmcvZGV2LWRvY3Mvc2hvdy12aWRlby13aXRoLWEtZGZwLXZpZGVvLXRhZy5odG1s2AIA4AKtmEjqAjNodAVKSHRlc3QubG9jYWxob3N0Ojk5OTkFFDgvcGFnZXMvaW5zdHJlYW0FPmjyAhMKD0NVU1RPTV9NT0RFTF9JRBIA8gIaChYyFgAgTEVBRl9OQU1FAR0IHgoaNh0ACEFTVAE-4ElGSUVEEgCAAwCIAwGQAwCYAxegAwGqAwDAA-CoAcgDANgDAOADAOgDAPgDAYAEAJIEDS91dC92Mw398F6YBACiBAsxMC43NS43NC42OagEtCyyBBIIARACGIAFIOADKAEoAjAAOAO4BADABADIBADSBA45MzI1I1NJTjM6NDczONoEAggB4AQA8ATLgcAuiAUBmAUAoAX______wEDFAHABQDJBWnbFPA_0gUJCQkMeAAA2AUB4AUB8AXDlQv6BQQIABAAkAYBmAYAuAYAwQYJJSjwP9AG9S_aBhYKEAkRGQFQEAAYAOAGBPIGAggAgAcBiAcAoAdA%26s%3D68b9d39d60a72307a201e479000a8c7be5508188]]></VASTAdTagURI><Impression><![CDATA[http://sin3-ib.adnxs.com/vast_track/v2?info=aAAAAAMArgAFAQklKBNeAAAAABEx74AO4IBoExklKBNeAAAAACDLgcAuKAAw7Ug47UhA0-hISLuv1AFQ6dGnBljDlQtiAi0taAFwAXgAgAEBiAEBkAGABZgB4AOgAQCoAcuBwC6wAQE.&s=07e6e5f2f03cc92e899c3ddbf4e2988e966caaa2&event_type=1]]></Impression><Creatives></Creatives></Wrapper></Ad></VAST>","ttlseconds":30}]}`,
		/*1*/ `{"puts":[{"type":"json","value":true,"ttlseconds":30}]}`,
		/*2*/ `{"puts":[{"type":"xml","value":"plain text","ttlseconds":30}]}`,
		/*3*/ `{"puts":[{"type":"xml","value":"2","ttlseconds":30}]}`,
		/*4*/ `{"puts":[{"type":}]}`,
		/*5*/ `{"puts":[]}`,
		/*6*/ `{}`,
		/*7*/ `{"puts":[{"type":"xml","value":"","ttlseconds":30}]}`,
		/*8*/ `{"puts":[{"type":"xml","value":"<tag>YourXMLcontentgoeshere.</tag>","ttlseconds":3600,"ttlseconds":30}]}`,
		/*9*/ `{"puts":[{"type":"xml","value":"<tag>YourXMLcontentgoeshere.</tag>","ttlseconds":30}]}`,
	}

	// Run sequantially second by second
	for i := 0; i < TEST_TIME_SECONDS; i++ {
		run(input)
	}
}

func run(input []string) {

	var counter int = 0
	var errCounter int = 0
	var waitGroup sync.WaitGroup
	ticker := time.NewTicker(time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				waitGroup.Add(QPS)
				for i := 0; i < QPS; i++ {
					go aParallelCall(input[9], &waitGroup, &counter, &errCounter)
				}
				waitGroup.Wait()
			}
		}
	}()
	time.Sleep(time.Second)
	ticker.Stop()
	done <- true
	logrus.Infof("QPS = %d; Errors = %d", counter, errCounter)
}

func aParallelCall(reqBody string, wg *sync.WaitGroup, counter *int, errCounter *int) {
	wasError := aCall(reqBody, URL)
	if wasError {
		*errCounter = *errCounter + 1
	}
	*counter = *counter + 1
	wg.Done()
}

func aCall(reqBody, url string) bool {
	wasError := false

	httpClientPtr, httpRequest, err := buildClientAndRequest(reqBody, url)
	if nil != err {
		if LOG_ERROR {
			logrus.Errorf("%v", err)
		}
		wasError = true
		return wasError
	}

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	resp, err := ctxhttp.Do(ctx, httpClientPtr, httpRequest)
	if nil != err {
		if LOG_ERROR {
			logrus.Errorf("client.Do() call error: %v", err)
		}
		wasError = true
		return wasError
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		if LOG_ERROR {
			logrus.Errorf("Response status code = %d", resp.StatusCode)
		}
		wasError = true
		return wasError
	}

	// Print response
	if LOG_SUCCESS {
		logSuccess(resp, reqBody)
	}

	return wasError
}

func buildClientAndRequest(reqBody, url string) (*http.Client, *http.Request, error) {

	httpClient := http.Client{}
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpRequest, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	if nil != err {
		return nil, nil, fmt.Errorf("\"%s\" >> creating HTTP request >> %v", reqBody, err)
	}

	//httpRequest.Header.Add("Host", "prebid.adnxs.com")
	httpRequest.Host = "prebid.adnxs.com"

	return &httpClient, httpRequest, nil
}

func logSuccess(resp *http.Response, reqBody string) {
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("\"%s\" >> Could not convert io.ReadCloser response to string >> %v", reqBody, err)
		return
	}
	logrus.Infof("[SUCCESS] \"%s\" >> \"%s\" ", reqBody, buf)
}
