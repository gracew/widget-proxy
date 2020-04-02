# widget-proxy

## Base API server

This repository contains the code and docker image for user-defined API servers.

Expected usage:

- A docker image with the source and dependencies already installed is defined via `Dockerfile.src`. This image can be
  built by running:

```
docker build . -f Dockerfile.src -t widget-proxy-src
```

- The resulting image is expected to be used as a base for building the runtime image (containing a go binary). When
  building the runtime image the file `generated/model.go` should be replaced by a file corresponding to the user-defined
  API.
- The runtime image is expected to be launched with the following inputs:
  - environment variable `API_NAME`
  - file containing the API definition at `/app/api.json`
  - file containing the auth definition at `/app/auth.json`
  - file containing the custom logic definition at `/app/customLogic.json`

Based on the custom logic definition, the API server will make requests to a custom logic server at
`http://custom-logic:8080`. All requests are POST requests. Paths are expected to be of the form
`/{when}{operation}`, for example `/beforecreate`, `/afterdelete`, or `/beforemarkComplete` for an update action
named `markComplete`.

## Custom logic

This repository also contains the docker images for running custom logic, found in the `docker/` directory. These images
are expected to be launched with a directory at `/app/customLogic` containing user-specified custom logic. Each file
in this directory corresponds to a POST HTTP endpoint, where the filename (with extension stripped) is the endpoint
path. For example, a file `beforecreate.js` or `beforecreate.py` would result in the endpoint `/beforecreate`.

## Tests

The tests rely on mocks which are generated via:
```
go generate ./...
```