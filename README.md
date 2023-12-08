# Pricefetcher microservice

This is a simple microservice to fetch prices of crypto tokens and return them as a JSON or GRPC (for scalability) depending on preference, this microservice was built using

- golang
- gRPC
- protcol buffers
- make to bootstrap the project

To run this repo a docker file is avaliable to build and run locally then run the command `make run ` to run the app

---

a MockPriceFetcher function was written in the service.go file with a hardcoded map of tokens and their price it returns the price of the token which is a float value as shown below

```
var priceMocks = map[string]float64{
	"BTC": 20_000.0,
	"ETH": 2_000.0,
	"SAI": 10_000.0,
}

func MockPriceFetcher(ctx context.Context, ticker string) (float64, error) {
	//mimic the HTTP process
	time.Sleep(100 * time.Millisecond)
	price, ok := priceMocks[ticker]
	if !ok {
		return price, fmt.Errorf("the given ticker (%s) is not supported", ticker)
	}

	return price, nil
}
```

which is then returned in the parent function FetchPrice

```
func (s *priceService) FetchPrice(ctx context.Context, ticker string) (float64, error) {
	return MockPriceFetcher(ctx, ticker)
}
```

---

## Logger

The logging function makes use of the logrus package and returns a log when the price is fetched like so

```
func (s loggingService) FetchPrice(ctx context.Context, ticker string) (price float64, err error) {
	defer func(begin time.Time) {
		logrus.WithFields(logrus.Fields{
			"requestID": ctx.Value("requestID"),
			"took":      time.Since(begin),
			"err":       err,
			"price":     price,
		}).Info("fetchPrice")
	}(time.Now())

	//return the price fetched after logging
	return s.priceService.FetchPrice(ctx, ticker)
}
```

---

# JSON TRANSPORT

this app returns both JSON and GRPC depnding on preference, for JSON a simple api route is set up as shown below

```
type JSONAPIServer struct {
	listenAddr string
	svc        PriceService
}

func NewJSONAPIServer(listenAddr string, svc PriceService) *JSONAPIServer {
	return &JSONAPIServer{
		listenAddr: listenAddr,
		svc:        svc,
	}
}

func (s *JSONAPIServer) Run() {
	http.HandleFunc("/", makeHTTPHandlerFunc(s.handleFetchPrice))
	http.ListenAndServe(s.listenAddr, nil)
}
```

because this app makes use of no frameworks we create a handler func to handle the route controller

```
type APIFunc func(context.Context, http.ResponseWriter, *http.Request) error

func makeHTTPHandlerFunc(apiFn APIFunc) http.HandlerFunc {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "requestID", rand.Intn(100000000))

	return func(w http.ResponseWriter, r *http.Request) {
		if err := apiFn(ctx, w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
	}
}
```

and a function for writing JSON

```
func writeJSON(w http.ResponseWriter, s int, v any) error {
	w.WriteHeader(s)

	return json.NewEncoder(w).Encode(v)
}
```

the controller function for fetching the price of a token

```
func (s *JSONAPIServer) handleFetchPrice(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	//request the ticker value from the query url
	ticker := r.URL.Query().Get("ticker")

	//fetch the price from FetchPrice()
	price, err := s.svc.FetchPrice(ctx, ticker)
	if err != nil {
		return err
	}
	priceResp := types.PriceResponse{
		Price:  price,
		Ticker: ticker,
	}
	return writeJSON(w, http.StatusOK, &priceResp)

```

first it requests the "ticker" query from the url endpoint and passes it as the parameter for the FetchPrice function along with the context and returns a JSON of type PriceResponse

```
type PriceResponse struct {
	Ticker string  `json:"ticker"`
	Price  float64 `json:"price"`
}
```

---

# gRPC TRANSPORT

gRPC (google Remote Protocol Call) is a framework developed by Google. it uses a binary serialization format known as protocol buffers (protobuf) to efficiently serialize strutured data between client and server and is a better alternative to JSON as it is compact and faster to serialize/deserialize providing a strongly typed schema and ensuring data consistency between client and server. It isn't really necessary for a small microservice like this but for a large microservice it is a better alternative to JSON, to use this implementation we define our message type in a `.proto` file as shown

```
syntax = "proto3";

option go_package = "github.com/RaymondAmenah/pricefetcher-microservice/proto";


service PriceFetcher{
    rpc FetchPrice(PriceRequest) returns (PriceResponse);
}

message PriceRequest {
string ticker = 1;
}

message PriceResponse {
    string ticker = 1;
    float price = 2;
}

```

then we generate a `service_grpc.pb.go` and `service.pb.go` files containng the necessary client and server code using `protoc` like below or in this case running `make proto`

```
protoc --go_out=. --go_opt=paths=source_relative \
   --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  	 proto/service.proto
```

we spin up a grpc server over a tcp connection like this

```
func makeGRPCServerAndRun(listenAddr string, svc PriceService) error {
	grpcPriceFetcher := NewGRPCPriceFetcherServer(svc)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil
	}
	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	proto.RegisterPriceFetcherServer(server, grpcPriceFetcher)

	return server.Serve(ln)
}

type GRPCPriceFetcherServer struct {
	svc PriceService
	proto.UnimplementedPriceFetcherServer
}

func NewGRPCPriceFetcherServer(svc PriceService) *GRPCPriceFetcherServer {
	return &GRPCPriceFetcherServer{
		svc: svc,
	}
}
```

using the provided interface given in the generated file the priceFetcher service becomes

```
type GRPCPriceFetcherServer struct {
	svc PriceService
	proto.UnimplementedPriceFetcherServer
}

func NewGRPCPriceFetcherServer(svc PriceService) *GRPCPriceFetcherServer {
	return &GRPCPriceFetcherServer{
		svc: svc,
	}
}

func (s *GRPCPriceFetcherServer) FetchPrice(ctx context.Context, req *proto.PriceRequest) (*proto.PriceResponse, error) {
	reqId := rand.Intn(1000000)
	ctx = context.WithValue(ctx, "requestID", reqId)
	price, err := s.svc.FetchPrice(ctx, req.Ticker)
	if err != nil {
		return nil, err
	}

	resp := &proto.PriceResponse{
		Ticker: req.Ticker,
		Price:  float32(price),
	}

	return resp, err
}

```

all thats left is to initialize a client with an active connection for the grpc server

```
func NewGRPCClient(remoteAddr string) (proto.PriceFetcherClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := proto.NewPriceFetcherClient(conn)

	return c, nil
}
```

the client is run as a goroutine in the main.go file like this so its possible to get a JSON and grpc response at the same time

```
var (
		jsonAddr = flag.String("json", ":3000", "service is running on json transport")
		grpcAddr = flag.String("grpc", ":4000", "service is running on grpc transport")
		svc      = loggingService{priceService{}}
		ctx      = context.Background()
	)

	flag.Parse()

	grpcClient, err := client.NewGRPCClient(":4000")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(3 * time.Second)
		resp, err := grpcClient.FetchPrice(ctx, &proto.PriceRequest{Ticker: "BTC"})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", resp)

	}()

	go makeGRPCServerAndRun(*grpcAddr, svc)
```

now run `make run` and you recieve the price of the token
