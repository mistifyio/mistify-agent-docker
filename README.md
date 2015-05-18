# mdocker

[![mdocker](https://godoc.org/github.com/mistifyio/mistify-agent-docker?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent-docker)

Package mdocker is a mistify subagent that manages guest docker containers,
exposed via JSON-RPC over HTTP.

### HTTP API Endpoint

    /_mistify_RPC_
        * GET - Run a specified method

### Request Structure

    {
        "method": "RPC_METHOD",
        "params": [
            DATA_STRUCT
        ],
        "id": 0
    }

Where RPC_METHOD is the desired method and DATA_STRUCTURE is one of the request
structs defined in http://godoc.org/github.com/mistifyio/mistify-agent/rpc .

### Response Structure

    {
        "result": {
            KEY: RESPONSE_STRUCT
        },
        "error": null,
        "id": 0
    }

Where KEY is a string (e.g. "snapshot") and DATA is one of the response structs
defined in http://godoc.org/github.com/mistifyio/mistify-agent/rpc .

### RPC Methods

    ListContainers
    GetContainer
    DeleteContainer
    SaveContainer
    CreateContainer
    StartContainer
    StopContainer
    RestartContainer
    RebootContainer
    PauseContainer
    UnpauseContainer

    ListImages
    GetImages
    PullImage
    DeleteImage

See the godocs and function signatures for each method's purpose and expected
request/response structs.

## Usage

#### type MDocker

```go
type MDocker struct {
}
```

MDocker is the Mistify Docker subagent service

#### func  New

```go
func New(endpoint, imageService, tlsCertPath string) (*MDocker, error)
```
New creates a new MDocker with a docker client

#### func (*MDocker) CreateContainer

```go
func (md *MDocker) CreateContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
CreateContainer creates a new Docker container

#### func (*MDocker) DeleteContainer

```go
func (md *MDocker) DeleteContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
DeleteContainer deletes a Docker container

#### func (*MDocker) DeleteImage

```go
func (md *MDocker) DeleteImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error
```
DeleteImage deletes a Docker image

#### func (*MDocker) GetContainer

```go
func (md *MDocker) GetContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerResponse) error
```
GetContainer retrieves information about a specific Docker container

#### func (*MDocker) GetImage

```go
func (md *MDocker) GetImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error
```
GetImage retrieves information about a specific Docker image

#### func (*MDocker) GetInfo

```go
func (md *MDocker) GetInfo(h *http.Request, request *struct{}, response *docker.Env) error
```
GetInfo provides general information about the system from Docker

#### func (*MDocker) ImportImage

```go
func (md *MDocker) ImportImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error
```
ImportImage downloads a new container image from the image service and imports
it into Docker

#### func (*MDocker) ListContainers

```go
func (md *MDocker) ListContainers(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerResponse) error
```
ListContainers retrieves a list of Docker containers

#### func (*MDocker) ListImages

```go
func (md *MDocker) ListImages(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error
```
ListImages retrieves a list of Docker images

#### func (*MDocker) PauseContainer

```go
func (md *MDocker) PauseContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error
```
PauseContainer pauses a Docker container

#### func (*MDocker) RebootContainer

```go
func (md *MDocker) RebootContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error
```
RebootContainer restarts a Docker container

#### func (*MDocker) RequestOpts

```go
func (md *MDocker) RequestOpts(req RPCRequest, opts interface{}) error
```
RequestOpts extracts the request opts into an appropriate struct Nested structs
stored in interface{} don't convert directly, so use JSON as an intermediate

#### func (*MDocker) RestartContainer

```go
func (md *MDocker) RestartContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error
```
RestartContainer restarts a Docker container

#### func (*MDocker) RunHTTP

```go
func (md *MDocker) RunHTTP(port uint) error
```
RunHTTP creates and runs the RPC HTTP server

#### func (*MDocker) SaveContainer

```go
func (md *MDocker) SaveContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ImageResponse) error
```
SaveContainer saves a Docker container

#### func (*MDocker) StartContainer

```go
func (md *MDocker) StartContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
StartContainer starts a Docker container

#### func (*MDocker) StopContainer

```go
func (md *MDocker) StopContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error
```
StopContainer stops a Docker container or kills it after a timeout

#### func (*MDocker) UnpauseContainer

```go
func (md *MDocker) UnpauseContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error
```
UnpauseContainer restarts a Docker container

#### type RPCRequest

```go
type RPCRequest interface {
	GetOpts() interface{}
}
```

RPCRequest is an interface for incoming RPC requests

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
