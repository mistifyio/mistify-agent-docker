/*
Package mdocker is a mistify subagent that manages guest docker containers,
exposed via JSON-RPC over HTTP.

HTTP API Endpoint

    /_mistify_RPC_
        * GET - Run a specified method

Request Structure

    {
        "method": "MDocker.RPC_METHOD",
        "params": [
            DATA_STRUCT
        ],
        "id": 0
    }

Where RPC_METHOD is the desired method and DATA_STRUCTURE is one of the request
structs defined in http://godoc.org/github.com/mistifyio/mistify-agent/rpc .

Response Structure

    {
        "result": {
            KEY: RESPONSE_STRUCT
        },
        "error": null,
        "id": 0
    }

Where KEY is a string (e.g. "snapshot") and DATA is one of the response structs
defined in http://godoc.org/github.com/mistifyio/mistify-agent/rpc .

RPC Methods

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
    LoadImage
    DeleteImage

See the godocs and function signatures for each method's purpose and expected
request/response structs.
*/
package mdocker
