package duo

import (
	"crypto/hmac"
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	version           = "0.1.1"
	defaultUserAgent  = "duo-svc/" + version
	initialBackoffMS  = 1000
	maxBackoffMS      = 32000
	backoffFactor     = 2
	rateLimitHttpCode = 429
)

var spaceReplacer *strings.Replacer = strings.NewReplacer("+", "%20")

func canonParams(params url.Values) string {
	// Values must be in sorted order
	for key, val := range params {
		sort.Strings(val)
		params[key] = val
	}
	// Encode will place Keys in sorted order
	ordered_params := params.Encode()
	// Encoder turns spaces into +, but we need %XX escaping
	return spaceReplacer.Replace(ordered_params)
}

func canonicalize(method string,
	host string,
	uri string,
	params url.Values,
	date string,
) string {
	var canon [5]string
	canon[0] = date
	canon[1] = strings.ToUpper(method)
	canon[2] = strings.ToLower(host)
	canon[3] = uri
	canon[4] = canonParams(params)
	return strings.Join(canon[:], "\n")
}

func canonicalizeV5(method string,
	host string,
	uri string,
	params url.Values,
	body string,
	date string,
) string {
	var canon [7]string
	canon[0] = date
	canon[1] = strings.ToUpper(method)
	canon[2] = strings.ToLower(host)
	canon[3] = uri
	canon[4] = canonParams(params)
	canon[5] = hashString(body)
	canon[6] = hashString("") // additional headers not needed at this time
	return strings.Join(canon[:], "\n")
}

func hashString(to_hash string) string {
	hash := sha512.New()
	hash.Write([]byte(to_hash))
	return hex.EncodeToString(hash.Sum(nil))
}

func jsonToValues(json JSONParams) (url.Values, error) {
	params := url.Values{}
	for key, val := range json {
		s, ok := val.(string)
		if ok {
			params[key] = []string{s}
		} else {
			return nil, errors.New("JSON value not a string")
		}
	}
	return params, nil
}

func sign(ikey string,
	skey string,
	method string,
	host string,
	uri string,
	date string,
	params url.Values,
) string {
	canon := canonicalize(method, host, uri, params, date)
	mac := hmac.New(sha512.New, []byte(skey))
	mac.Write([]byte(canon))
	sig := hex.EncodeToString(mac.Sum(nil))
	auth := ikey + ":" + sig
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func signV5(ikey string,
	skey string,
	method string,
	host string,
	uri string,
	date string,
	params url.Values,
	body string,
) string {
	canon := canonicalizeV5(method, host, uri, params, body, date)
	mac := hmac.New(sha512.New, []byte(skey))
	mac.Write([]byte(canon))
	sig := hex.EncodeToString(mac.Sum(nil))
	auth := ikey + ":" + sig
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

type DuoApi struct {
	ikey       string
	skey       string
	host       string
	userAgent  string
	apiClient  httpClient
	authClient httpClient
	sleepSvc   sleepService
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
type sleepService interface {
	Sleep(duration time.Duration)
}
type timeSleepService struct{}

func (svc timeSleepService) Sleep(duration time.Duration) {
	time.Sleep(duration + (time.Duration(rand.Intn(1000)) * time.Millisecond))
}

type apiOptions struct {
	timeout   time.Duration
	insecure  bool
	proxy     func(*http.Request) (*url.URL, error)
	transport func(*http.Transport)
}

// Optional parameter for NewDuoApi, used to configure timeouts on API calls.
func SetTimeout(timeout time.Duration) func(*apiOptions) {
	return func(opts *apiOptions) {
		opts.timeout = timeout
		return
	}
}

// Optional parameter for testing only.  Bypasses all TLS certificate validation.
func SetInsecure() func(*apiOptions) {
	return func(opts *apiOptions) {
		opts.insecure = true
	}
}

// Optional parameter for NewDuoApi, used to configure an HTTP Connect proxy
// server for all outbound communications.
func SetProxy(proxy func(*http.Request) (*url.URL, error)) func(*apiOptions) {
	return func(opts *apiOptions) {
		opts.proxy = proxy
	}
}

// SetTransport enables additional control over the HTTP transport used to connect to the Duo API.
func SetTransport(transport func(*http.Transport)) func(*apiOptions) {
	return func(opts *apiOptions) {
		opts.transport = transport
	}
}

// Build and return a DuoApi struct.
// ikey is your Duo integration key
// skey is your Duo integration secret key
// host is your Duo host
// userAgent allows you to specify the user agent string used when making
// the web request to Duo.  Information about the client will be
// appended to the userAgent.
// options are optional parameters.  Use SetTimeout() to specify a timeout value
// for Rest API calls.  Use SetProxy() to specify proxy settings for Duo API calls.
//
// Example: duoapi.NewDuoApi(ikey,skey,host,userAgent,duoapi.SetTimeout(10*time.Second))
func NewDuoApi(ikey string,
	skey string,
	host string,
	userAgent string,
	options ...func(*apiOptions),
) *DuoApi {
	opts := apiOptions{proxy: http.ProxyFromEnvironment}
	for _, o := range options {
		o(&opts)
	}

	// Certificate pinning
	// certPool := x509.NewCertPool()

	transport := &http.Transport{
		Proxy: opts.proxy,
		TLSClientConfig: &tls.Config{
			// RootCAs:            certPool,
			InsecureSkipVerify: opts.insecure,
		},
	}
	if opts.transport != nil {
		opts.transport(transport)
	}

	if userAgent != "" {
		userAgent += " "
	}
	userAgent += defaultUserAgent

	return &DuoApi{
		ikey:      ikey,
		skey:      skey,
		host:      host,
		userAgent: userAgent,
		apiClient: &http.Client{
			Timeout:   opts.timeout,
			Transport: transport,
		},
		authClient: &http.Client{
			Transport: transport,
		},
		sleepSvc: timeSleepService{},
	}
}

type requestOptions struct {
	timeout bool
}

type DuoApiOption func(*requestOptions)

// Pass to Request or SignedRequest to configure a timeout on the request
func UseTimeout(opts *requestOptions) {
	opts.timeout = true
}

func (duoapi *DuoApi) buildOptions(options ...DuoApiOption) *requestOptions {
	opts := &requestOptions{}
	for _, o := range options {
		o(opts)
	}
	return opts
}

// API calls will return a StatResult object.  On success, Stat is 'OK'.
// On error, Stat is 'FAIL', and Code, Message, and Message_Detail
// contain error information.
type StatResult struct {
	Stat           string
	Code           *int32
	Message        *string
	Message_Detail *string
}

// SetCustomHTTPClient allows one to set a completely custom http client that
// will be used to make network calls to the duo api
func (duoapi *DuoApi) SetCustomHTTPClient(c *http.Client) {
	duoapi.apiClient = c
	duoapi.authClient = c
}

// Make an unsigned Duo Rest API call.  See Duo's online documentation
// for the available REST API's.
// method is POST or GET
// uri is the URI of the Duo Rest call
// params HTTP query parameters to include in the call.
// options Optional parameters.  Use UseTimeout to toggle whether the
// Duo Rest API call should timeout or not.
//
// Example: duo.Call("GET", "/auth/v2/ping", nil, duoapi.UseTimeout)
func (duoapi *DuoApi) Call(method string,
	uri string,
	params url.Values,
	options ...DuoApiOption,
) (*http.Response, []byte, error) {
	url := url.URL{
		Scheme:   "https",
		Host:     duoapi.host,
		Path:     uri,
		RawQuery: params.Encode(),
	}
	headers := make(map[string]string)
	headers["User-Agent"] = duoapi.userAgent

	return duoapi.makeRetryableHttpCall(method, url, headers, nil, options...)
}

// Make a signed Duo Rest API call.  See Duo's online documentation
// for the available REST API's.
// method is POST or GET
// uri is the URI of the Duo Rest call
// params HTTP query parameters to include in the call.
// options Optional parameters.  Use UseTimeout to toggle whether the
// Duo Rest API call should timeout or not.
//
// Example: duo.SignedCall("GET", "/auth/v2/check", nil, duoapi.UseTimeout)
func (duoapi *DuoApi) SignedCall(method string,
	uri string,
	params url.Values,
	options ...DuoApiOption,
) (*http.Response, []byte, error) {
	now := time.Now().UTC().Format(time.RFC1123Z)
	auth_sig := sign(duoapi.ikey, duoapi.skey, method, duoapi.host, uri, now, params)

	url := url.URL{
		Scheme: "https",
		Host:   duoapi.host,
		Path:   uri,
	}
	method = strings.ToUpper(method)

	if method == "GET" {
		url.RawQuery = params.Encode()
	}

	headers := make(map[string]string)
	headers["User-Agent"] = duoapi.userAgent
	headers["Authorization"] = auth_sig
	headers["Date"] = now
	var requestBody io.ReadCloser = nil
	if method == "POST" || method == "PUT" {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
		requestBody = io.NopCloser(strings.NewReader(params.Encode()))
	}

	return duoapi.makeRetryableHttpCall(method, url, headers, requestBody, options...)
}

type JSONParams map[string]interface{}

// Make a signed Duo Rest API call that takes JSON as an argument.
// See Duo's online documentation for the available REST API's.
// method is one of GET, POST, PATCH, PUT, DELETE
// uri is the URI of the Duo Rest call
// json is the JSON parameters to include in the call.
// options Optional parameters.  Use UseTimeout to toggle whether the
// Duo Rest API call should timeout or not.
//
//	Example:
//	params := duoapi.JSONParams{
//		"user_id":         userid,
//		"activation_code": activationCode,
//	}
//	JSONSignedCall("POST", "/auth/v2/enroll_status", params, duoapi.UseTimeout)
func (duoapi *DuoApi) JSONSignedCall(method string,
	uri string,
	params JSONParams,
	options ...DuoApiOption,
) (*http.Response, []byte, error) {
	body_methods := make(map[string]struct{})
	body_methods["POST"] = struct{}{}
	body_methods["PUT"] = struct{}{}
	body_methods["PATCH"] = struct{}{}
	_, params_go_in_body := body_methods[method]

	now := time.Now().UTC().Format(time.RFC1123Z)
	var body string
	api_url := url.URL{
		Scheme: "https",
		Host:   duoapi.host,
		Path:   uri,
	}

	url_values := url.Values{}
	if params_go_in_body {
		body_bytes, err := json.Marshal(params)
		if err != nil {
			return nil, nil, err
		}
		body = string(body_bytes[:])
	} else {
		body = ""
		var err error
		url_values, err = jsonToValues(params)
		if err != nil {
			return nil, nil, err
		}
		api_url.RawQuery = url_values.Encode()
	}

	auth_sig := signV5(duoapi.ikey, duoapi.skey, method, duoapi.host, uri, now, url_values, body)

	method = strings.ToUpper(method)
	headers := make(map[string]string)
	headers["User-Agent"] = duoapi.userAgent
	headers["Authorization"] = auth_sig
	headers["Date"] = now
	var requestBody io.ReadCloser = nil
	if params_go_in_body {
		headers["Content-Type"] = "application/json"
		requestBody = io.NopCloser(strings.NewReader(body))
	}

	return duoapi.makeRetryableHttpCall(method, api_url, headers, requestBody, options...)
}

func (duoapi *DuoApi) makeRetryableHttpCall(
	method string,
	url url.URL,
	headers map[string]string,
	body io.ReadCloser,
	options ...DuoApiOption,
) (*http.Response, []byte, error) {
	opts := duoapi.buildOptions(options...)

	client := duoapi.authClient
	if opts.timeout {
		client = duoapi.apiClient
	}

	backoffMs := initialBackoffMS
	for {
		request, err := http.NewRequest(method, url.String(), nil)
		if err != nil {
			return nil, nil, err
		}

		if headers != nil {
			for k, v := range headers {
				request.Header.Set(k, v)
			}
		}
		if body != nil {
			request.Body = body
		}

		resp, err := client.Do(request)
		var body []byte
		if err != nil {
			return resp, body, err
		}

		if backoffMs > maxBackoffMS || resp.StatusCode != rateLimitHttpCode {
			body, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			return resp, body, err
		}

		resp.Body.Close()

		duoapi.sleepSvc.Sleep(time.Millisecond * time.Duration(backoffMs))
		backoffMs *= backoffFactor
	}
}

// Client provides access to Duo's Admin or Accounts API.

type Client struct {
	DuoApi
}

type ListResultMetadata struct {
	NextOffset   json.Number `json:"next_offset"`
	PrevOffset   json.Number `json:"prev_offset"`
	TotalObjects json.Number `json:"total_objects"`
}

type ListResult struct {
	Metadata ListResultMetadata `json:"metadata"`
}

func (l *ListResult) metadata() ListResultMetadata {
	return l.Metadata
}

// New initializes an admin API Client struct.
func New(base DuoApi) *Client {
	return &Client{base}
}

// Account type represents a Duo subaccount, which maps to VCD/Zerto organizations.
type Account struct {
	Name      string `json:"name"`
	AccountId string `json:"account_id"`
}

// GetAccountResult models responses containing a single account.
type GetAccountResult struct {
	StatResult
	Response Account
}

// Common URL options

// Limit sets the optional limit parameter for an API request.
func Limit(limit uint64) func(*url.Values) {
	return func(opts *url.Values) {
		opts.Set("limit", strconv.FormatUint(limit, 10))
	}
}

// Offset sets the optional offset parameter for an API request.
func Offset(offset uint64) func(*url.Values) {
	return func(opts *url.Values) {
		opts.Set("offset", strconv.FormatUint(offset, 10))
	}
}

type responsePage interface {
	metadata() ListResultMetadata
	getResponse() interface{}
	appendResponse(interface{})
}

type pageFetcher func(params url.Values) (responsePage, error)

func (c *Client) retrieveItems(
	params url.Values,
	fetcher pageFetcher,
) (responsePage, error) {
	if params.Get("offset") == "" {
		params.Set("offset", "0")
	}

	if params.Get("limit") == "" {
		params.Set("limit", "100")
		accumulator, firstErr := fetcher(params)

		if firstErr != nil {
			return nil, firstErr
		}

		params.Set("offset", accumulator.metadata().NextOffset.String())
		for params.Get("offset") != "" {
			nextResult, err := fetcher(params)
			if err != nil {
				return nil, err
			}
			nextResult.appendResponse(accumulator.getResponse())
			accumulator = nextResult
			params.Set("offset", accumulator.metadata().NextOffset.String())
		}
		return accumulator, nil
	}

	return fetcher(params)
}

// See Duo's online documentation for the available REST API's.
// Create an new Duo account.
//
//	Example:
//	params := duoapi.JSONParams{
//		"name":         name,
//	}
//	CreateAccount(testName string)
func (c *Client) createAccount(a Account) (*GetAccountResult, error) {
	path := "/accounts/v1/account/create"
	params := JSONParams{
		"name": a.Name,
	}

	_, body, err := c.JSONSignedCall(http.MethodPost, path, params, UseTimeout)
	if err != nil {
		return nil, err
	}

	res := &GetAccountResult{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
