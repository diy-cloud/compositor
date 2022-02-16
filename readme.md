# Compositor

## this using docker for container engine

### need to install docker

recommand to use ubuntu and snapcraft

```bash
sudo snap install docker
```

## install compositor

```bash
go install github.com/snowmerak/compositor/compositor@latest
```

installing is able on any OS has go compiler, but will work on linux only.

## TLS cert file

`<PWD>/cert.pem` is certificate file.  
`<PWD>/key.pem` is key file.

## register sub router

post request to `http://<SERVER>:8888/register/:id` or `https://<SERVER>:9999/register/:id` with image built by dockerfile.

and some client can request to `http://<SERVER>:80/:id` or `https://<SERVER>:443/:id` with any route or path variable, body, etc...
